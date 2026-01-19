// Package gateway provides the core gateway service logic.
package gateway

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"go.uber.org/zap"

	"ai-gateway/config"
	"ai-gateway/internal/domain"
	"ai-gateway/internal/errs"
	"ai-gateway/internal/pkg/loadbalancer"
	"ai-gateway/internal/providers"
	"ai-gateway/internal/providers/anthropic"
	"ai-gateway/internal/providers/openai"
	"ai-gateway/internal/repository"
)

// GatewayService defines the interface for gateway operations.
type GatewayService interface {
	Chat(ctx context.Context, req *domain.ChatRequest) (*domain.ChatResponse, error)
	ChatStream(ctx context.Context, req *domain.ChatRequest) (<-chan domain.StreamDelta, error)
	ListModels(ctx context.Context) ([]string, error)
	GetProvider(model string) (providers.Provider, string, error)
	// Reload reloads configuration from database.
	Reload(ctx context.Context) error
}

// providerNode wraps a Provider to implement loadbalancer.Node interface.
type providerNode struct {
	provider providers.Provider
	name     string
}

func (n *providerNode) ID() string {
	return n.name
}

type gatewayService struct {
	providerRepo    repository.ProviderRepository
	routingRuleRepo repository.RoutingRuleRepository
	loadBalanceRepo repository.LoadBalanceRepository

	providers     map[string]providers.Provider                       // name -> provider
	typeDefaults  map[string]string                                   // type -> default provider name
	routes        map[string]config.ModelRoute                        // exact model routes
	prefixRoutes  []prefixRouteEntry                                  // sorted by priority
	loadBalancers map[string]loadbalancer.LoadBalancer[*providerNode] // model pattern -> load balancer
	httpClient    *http.Client
	logger        *zap.Logger
}

type prefixRouteEntry struct {
	prefix   string
	provider string
	priority int
}

var _ GatewayService = (*gatewayService)(nil)

// NewGatewayService creates a new gateway service that loads config from database.
func NewGatewayService(
	providerRepo repository.ProviderRepository,
	routingRuleRepo repository.RoutingRuleRepository,
	loadBalanceRepo repository.LoadBalanceRepository,
	logger *zap.Logger,
) GatewayService {
	g := &gatewayService{
		providerRepo:    providerRepo,
		routingRuleRepo: routingRuleRepo,
		loadBalanceRepo: loadBalanceRepo,
		providers:       make(map[string]providers.Provider),
		typeDefaults:    make(map[string]string),
		routes:          make(map[string]config.ModelRoute),
		loadBalancers:   make(map[string]loadbalancer.LoadBalancer[*providerNode]),
		httpClient:      &http.Client{Timeout: 120 * time.Second},
		logger:          logger.Named("gateway"),
	}

	// Initial load from database
	if err := g.Reload(context.Background()); err != nil {
		logger.Error("failed to load initial configuration from database", zap.Error(err))
	}

	return g
}

// Reload reloads configuration from database.
func (g *gatewayService) Reload(ctx context.Context) error {
	// Load providers from database
	dbProviders, err := g.providerRepo.List(ctx)
	if err != nil {
		return fmt.Errorf("failed to load providers from database: %w", err)
	}

	// Initialize providers
	newProviders := make(map[string]providers.Provider)
	newTypeDefaults := make(map[string]string)

	for _, p := range dbProviders {
		if p.APIKey == "" {
			continue
		}

		var provider providers.Provider
		switch p.Type {
		case "openai":
			provider = openai.NewProvider(
				p.APIKey,
				p.BaseURL,
				g.httpClient,
				g.logger,
			)
		case "anthropic":
			provider = anthropic.NewProvider(
				p.APIKey,
				p.BaseURL,
				g.httpClient,
				g.logger,
			)
		default:
			g.logger.Warn("unknown provider type", zap.String("type", p.Type))
			continue
		}

		newProviders[p.Name] = provider
		g.logger.Info("registered provider from database",
			zap.String("name", p.Name),
			zap.String("type", p.Type),
			zap.String("baseURL", p.BaseURL),
		)

		// Track default provider for each type
		if p.IsDefault {
			newTypeDefaults[p.Type] = p.Name
		} else if newTypeDefaults[p.Type] == "" {
			newTypeDefaults[p.Type] = p.Name
		}
	}

	// Load routing rules from database
	routingRules, err := g.routingRuleRepo.List(ctx)
	if err != nil {
		return fmt.Errorf("failed to load routing rules from database: %w", err)
	}

	newRoutes := make(map[string]config.ModelRoute)
	var newPrefixRoutes []prefixRouteEntry

	for _, rule := range routingRules {
		if rule.RuleType == "exact" {
			newRoutes[rule.Pattern] = config.ModelRoute{
				Provider:    rule.ProviderName,
				ActualModel: rule.ActualModel,
			}
		} else if rule.RuleType == "prefix" {
			newPrefixRoutes = append(newPrefixRoutes, prefixRouteEntry{
				prefix:   rule.Pattern,
				provider: rule.ProviderName,
				priority: rule.Priority,
			})
		}
	}

	// Sort prefix routes by priority descending, then by length descending
	sort.Slice(newPrefixRoutes, func(i, j int) bool {
		if newPrefixRoutes[i].priority != newPrefixRoutes[j].priority {
			return newPrefixRoutes[i].priority > newPrefixRoutes[j].priority
		}
		return len(newPrefixRoutes[i].prefix) > len(newPrefixRoutes[j].prefix)
	})

	// Load load balancing groups from database
	lbGroups, err := g.loadBalanceRepo.ListGroups(ctx)
	if err != nil {
		return fmt.Errorf("failed to load load balance groups from database: %w", err)
	}

	newLoadBalancers := make(map[string]loadbalancer.LoadBalancer[*providerNode])
	for _, group := range lbGroups {
		members, err := g.loadBalanceRepo.GetMembers(ctx, group.ID)
		if err != nil {
			g.logger.Warn("failed to load members for load balance group",
				zap.String("group", group.Name),
				zap.Error(err),
			)
			continue
		}

		var nodes []*providerNode
		var weights []int

		for _, member := range members {
			if p, ok := newProviders[member.ProviderName]; ok {
				nodes = append(nodes, &providerNode{provider: p, name: member.ProviderName})
				weights = append(weights, member.Weight)
			}
		}

		if len(nodes) == 0 {
			continue
		}

		var lb loadbalancer.LoadBalancer[*providerNode]
		switch group.Strategy {
		case "round-robin":
			lb = loadbalancer.NewRoundRobin(nodes)
		case "random":
			lb = loadbalancer.NewRandom(nodes)
		case "failover":
			lb = loadbalancer.NewFailover(nodes)
		case "weighted":
			lb = loadbalancer.NewWeighted(nodes, weights)
		default:
			lb = loadbalancer.NewRoundRobin(nodes)
		}

		newLoadBalancers[group.ModelPattern] = lb
		g.logger.Info("created load balancer from database",
			zap.String("model", group.ModelPattern),
			zap.String("strategy", group.Strategy),
			zap.Int("providers", len(nodes)),
		)
	}

	// Atomic update
	g.providers = newProviders
	g.typeDefaults = newTypeDefaults
	g.routes = newRoutes
	g.prefixRoutes = newPrefixRoutes
	g.loadBalancers = newLoadBalancers

	g.logger.Info("configuration reloaded from database",
		zap.Int("providers", len(g.providers)),
		zap.Int("routes", len(g.routes)),
		zap.Int("prefixRoutes", len(g.prefixRoutes)),
		zap.Int("loadBalancers", len(g.loadBalancers)),
	)

	return nil
}

// GetProvider returns the provider for the given model.
// Priority: exact match -> load balancing -> prefix match -> type default
func (g *gatewayService) GetProvider(model string) (providers.Provider, string, error) {
	// 1. Check exact route
	if route, ok := g.routes[model]; ok {
		provider, ok := g.providers[route.Provider]
		if !ok {
			return nil, "", fmt.Errorf("%w: %s", errs.ErrProviderNotFound, route.Provider)
		}
		actualModel := route.ActualModel
		if actualModel == "" {
			actualModel = model
		}
		g.logger.Debug("using exact route", zap.String("model", model), zap.String("provider", route.Provider))
		return provider, actualModel, nil
	}

	// 2. Check load balancing
	if lb, ok := g.loadBalancers[model]; ok {
		node, err := lb.Select()
		if err == nil && node != nil {
			g.logger.Debug("using load balancer", zap.String("model", model), zap.String("provider", node.ID()))
			return node.provider, model, nil
		}
	}

	// 3. Check prefix routing
	for _, entry := range g.prefixRoutes {
		if strings.HasPrefix(strings.ToLower(model), strings.ToLower(entry.prefix)) {
			provider, ok := g.providers[entry.provider]
			if ok {
				g.logger.Debug("using prefix route",
					zap.String("model", model),
					zap.String("prefix", entry.prefix),
					zap.String("provider", entry.provider),
				)
				return provider, model, nil
			}
		}
	}

	// 4. Fall back to type default
	providerType := g.detectProviderType(model)
	providerName := g.typeDefaults[providerType]
	if providerName == "" {
		return nil, "", fmt.Errorf("%w: no provider for type %s", errs.ErrProviderNotFound, providerType)
	}

	provider, ok := g.providers[providerName]
	if !ok {
		return nil, "", fmt.Errorf("%w: %s", errs.ErrProviderNotFound, providerName)
	}

	g.logger.Debug("using type default",
		zap.String("model", model),
		zap.String("type", providerType),
		zap.String("provider", providerName),
	)
	return provider, model, nil
}

func (g *gatewayService) detectProviderType(model string) string {
	lower := strings.ToLower(model)
	switch {
	case strings.HasPrefix(lower, "gpt") || strings.HasPrefix(lower, "o1"):
		return "openai"
	case strings.HasPrefix(lower, "claude"):
		return "anthropic"
	default:
		return "openai"
	}
}

// Chat handles a non-streaming chat request.
func (g *gatewayService) Chat(ctx context.Context, req *domain.ChatRequest) (*domain.ChatResponse, error) {
	provider, actualModel, err := g.GetProvider(req.Model)
	if err != nil {
		return nil, err
	}

	req.Model = actualModel

	g.logger.Info("routing chat request",
		zap.String("model", req.Model),
		zap.String("provider", provider.Name()),
		zap.Bool("stream", req.Stream),
	)

	return provider.Chat(ctx, req)
}

// ChatStream handles a streaming chat request.
func (g *gatewayService) ChatStream(ctx context.Context, req *domain.ChatRequest) (<-chan domain.StreamDelta, error) {
	provider, actualModel, err := g.GetProvider(req.Model)
	if err != nil {
		return nil, err
	}

	req.Model = actualModel

	g.logger.Info("routing streaming chat request",
		zap.String("model", req.Model),
		zap.String("provider", provider.Name()),
	)

	return provider.ChatStream(ctx, req)
}

// ListModels returns all available models from all providers.
func (g *gatewayService) ListModels(ctx context.Context) ([]string, error) {
	var allModels []string
	for name, provider := range g.providers {
		models, err := provider.ListModels(ctx)
		if err != nil {
			g.logger.Warn("failed to list models",
				zap.String("provider", name),
				zap.Error(err),
			)
			continue
		}
		allModels = append(allModels, models...)
	}
	return allModels, nil
}
