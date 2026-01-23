// Package gateway 提供核心网关服务逻辑。
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

// GatewayService 定义了网关操作的接口。
type GatewayService interface {
	Chat(ctx context.Context, req *domain.ChatRequest) (*domain.ChatResponse, error)
	ChatStream(ctx context.Context, req *domain.ChatRequest) (<-chan domain.StreamDelta, string, error)
	ListModels(ctx context.Context) ([]string, error)
	GetProvider(model string) (providers.Provider, string, error)
	// Reload 从数据库重新加载配置。
	Reload(ctx context.Context) error
}

// providerNode 包装了一个 Provider 以实现 loadbalancer.Node 接口。
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

	providers        map[string]providers.Provider                       // 名称 -> 供应商
	configuredModels map[string][]string                                 // 供应商名称 -> 模型列表
	typeDefaults     map[string]string                                   // 类型 -> 默认供应商名称
	routes           map[string]config.ModelRoute                        // 精确的模型路由
	prefixRoutes     []prefixRouteEntry                                  // 按优先级排序
	loadBalancers    map[string]loadbalancer.LoadBalancer[*providerNode] // 模型模式 -> 负载均衡器
	httpClient       *http.Client
	logger           *zap.Logger
}

type prefixRouteEntry struct {
	prefix   string
	provider string
	priority int
}

var _ GatewayService = (*gatewayService)(nil)

// NewGatewayService 创建一个新的网关服务，从数据库加载配置。
func NewGatewayService(
	providerRepo repository.ProviderRepository,
	routingRuleRepo repository.RoutingRuleRepository,
	loadBalanceRepo repository.LoadBalanceRepository,
	logger *zap.Logger,
) GatewayService {
	g := &gatewayService{
		providerRepo:     providerRepo,
		routingRuleRepo:  routingRuleRepo,
		loadBalanceRepo:  loadBalanceRepo,
		providers:        make(map[string]providers.Provider),
		configuredModels: make(map[string][]string),
		typeDefaults:     make(map[string]string),
		routes:           make(map[string]config.ModelRoute),
		loadBalancers:    make(map[string]loadbalancer.LoadBalancer[*providerNode]),
		httpClient:       &http.Client{Timeout: 120 * time.Second},
		logger:           logger.Named("gateway"),
	}

	// 初始从数据库加载
	if err := g.Reload(context.Background()); err != nil {
		logger.Error("failed to load initial configuration from database", zap.Error(err))
	}

	return g
}

// Reload 从数据库重新加载配置。
func (g *gatewayService) Reload(ctx context.Context) error {
	// 从数据库加载供应商
	dbProviders, err := g.providerRepo.List(ctx)
	if err != nil {
		return fmt.Errorf("从数据库加载供应商失败: %w", err)
	}

	// 初始化供应商
	newProviders := make(map[string]providers.Provider)
	newConfiguredModels := make(map[string][]string)
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
		// 存储配置的模型列表
		if len(p.Models) > 0 {
			newConfiguredModels[p.Name] = p.Models
		}

		g.logger.Info("registered provider from database",
			zap.String("name", p.Name),
			zap.String("type", p.Type),
			zap.String("baseURL", p.BaseURL),
			zap.Int("models", len(p.Models)),
		)

		// 跟踪每种类型的默认供应商
		if p.IsDefault {
			newTypeDefaults[p.Type] = p.Name
		} else if newTypeDefaults[p.Type] == "" {
			newTypeDefaults[p.Type] = p.Name
		}
	}

	// 从数据库加载路由规则
	routingRules, err := g.routingRuleRepo.List(ctx)
	if err != nil {
		return fmt.Errorf("从数据库加载路由规则失败: %w", err)
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

	// 按优先级降序排序前缀路由，然后按长度降序排序
	sort.Slice(newPrefixRoutes, func(i, j int) bool {
		if newPrefixRoutes[i].priority != newPrefixRoutes[j].priority {
			return newPrefixRoutes[i].priority > newPrefixRoutes[j].priority
		}
		return len(newPrefixRoutes[i].prefix) > len(newPrefixRoutes[j].prefix)
	})

	// 从数据库加载负载均衡组
	lbGroups, err := g.loadBalanceRepo.ListGroups(ctx)
	if err != nil {
		return fmt.Errorf("从数据库加载负载均衡组失败: %w", err)
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

	// 原子更新
	g.providers = newProviders
	g.configuredModels = newConfiguredModels
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

// GetProvider 返回给定模型的供应商。
// 优先级：精确匹配 -> 负载均衡 -> 前缀匹配 -> 类型默认
func (g *gatewayService) GetProvider(model string) (providers.Provider, string, error) {
	// 1. 检查精确路由
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

	// 2. 检查负载均衡
	if lb, ok := g.loadBalancers[model]; ok {
		node, err := lb.Select()
		if err == nil && node != nil {
			g.logger.Debug("using load balancer", zap.String("model", model), zap.String("provider", node.ID()))
			return node.provider, model, nil
		}
	}

	// 3. 检查前缀路由
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

	// 4. 回退到类型默认值
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

// Chat 处理非流式聊天请求。
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

	resp, err := provider.Chat(ctx, req)
	if err != nil {
		return nil, err
	}
	resp.Provider = provider.Name()
	return resp, nil
}

// ChatStream 处理流式聊天请求。
func (g *gatewayService) ChatStream(ctx context.Context, req *domain.ChatRequest) (<-chan domain.StreamDelta, string, error) {
	provider, actualModel, err := g.GetProvider(req.Model)
	if err != nil {
		return nil, "", err
	}

	req.Model = actualModel

	g.logger.Info("routing streaming chat request",
		zap.String("model", req.Model),
		zap.String("provider", provider.Name()),
	)

	ch, err := provider.ChatStream(ctx, req)
	if err != nil {
		return nil, "", err
	}
	return ch, provider.Name(), nil
}

// ListModels 返回所有供应商提供的所有可用模型。
func (g *gatewayService) ListModels(ctx context.Context) ([]string, error) {
	var allModels []string
	seen := make(map[string]bool)

	// 1. 从配置的静态模型列表中获取
	for _, models := range g.configuredModels {
		for _, model := range models {
			if !seen[model] {
				allModels = append(allModels, model)
				seen[model] = true
			}
		}
	}

	// 2. 如果需要，可以从 Provider 动态获取（目前作为回退或补充）
	// 如果配置了静态模型，通常以此为准。
	// 但如果某个 Provider 没有配置静态模型，我们可能希望尝试调用 API 获取。
	for name, provider := range g.providers {
		// 如果该 Provider 已经有配置的模型，跳过动态获取以节省 API 调用
		if len(g.configuredModels[name]) > 0 {
			continue
		}

		models, err := provider.ListModels(ctx)
		if err != nil {
			g.logger.Warn("failed to list models from provider",
				zap.String("provider", name),
				zap.Error(err),
			)
			continue
		}
		for _, model := range models {
			if !seen[model] {
				allModels = append(allModels, model)
				seen[model] = true
			}
		}
	}

	sort.Strings(allModels)
	return allModels, nil
}
