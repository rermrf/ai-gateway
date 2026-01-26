// Package chat 封装网关调用并下沉计费/用量逻辑。
package chat

import (
	"context"
	"time"

	"ai-gateway/internal/domain"
	"ai-gateway/internal/errs"
	"ai-gateway/internal/pkg/logger"
	"ai-gateway/internal/service/apikey"
	"ai-gateway/internal/service/gateway"
	"ai-gateway/internal/service/modelrate"
	"ai-gateway/internal/service/usage"
	"ai-gateway/internal/service/wallet"
)

// RequestMeta 表示本次请求的上下文信息（由 API 层采集）。
type RequestMeta struct {
	UserID    int64
	APIKeyID  *int64
	RequestID string
	ClientIP  string
	UserAgent string
}

// Service 统一封装 Chat + 计费/用量记录。
//
//go:generate mockgen -source=./chat.go -destination=./mocks/chat.mock.go -package=chatmocks Service
type Service interface {
	Chat(ctx context.Context, req *domain.ChatRequest, meta RequestMeta) (*domain.ChatResponse, error)
	ChatStream(ctx context.Context, req *domain.ChatRequest, meta RequestMeta) (<-chan domain.StreamDelta, string, error)
}

type service struct {
	gw           gateway.GatewayService
	walletSvc    wallet.Service
	usageSvc     usage.Service
	apiKeySvc    apikey.Service
	modelRateSvc modelrate.Service
	logger       logger.Logger
}

func NewService(
	gw gateway.GatewayService,
	walletSvc wallet.Service,
	usageSvc usage.Service,
	apiKeySvc apikey.Service,
	modelRateSvc modelrate.Service,
	l logger.Logger,
) Service {
	return &service{
		gw:           gw,
		walletSvc:    walletSvc,
		usageSvc:     usageSvc,
		apiKeySvc:    apiKeySvc,
		modelRateSvc: modelRateSvc,
		logger:       l.With(logger.String("service", "chat")),
	}
}

func (s *service) Chat(ctx context.Context, req *domain.ChatRequest, meta RequestMeta) (*domain.ChatResponse, error) {
	if err := s.preflight(ctx, meta.UserID); err != nil {
		return nil, err
	}

	start := time.Now()
	resp, err := s.gw.Chat(ctx, req)
	if err != nil {
		return nil, err
	}

	model := req.Model // 注意：gateway 会把 model 重写成实际模型
	provider := resp.Provider
	usageData := resp.Usage
	latency := int(time.Since(start).Milliseconds())

	if usageData != nil {
		s.recordAsync(meta, model, provider, usageData.PromptTokens, usageData.CompletionTokens, httpStatusOK, latency)
	} else {
		s.recordAsync(meta, model, provider, 0, 0, httpStatusOK, latency)
	}

	return resp, nil
}

func (s *service) ChatStream(ctx context.Context, req *domain.ChatRequest, meta RequestMeta) (<-chan domain.StreamDelta, string, error) {
	if err := s.preflight(ctx, meta.UserID); err != nil {
		return nil, "", err
	}

	start := time.Now()
	in, provider, err := s.gw.ChatStream(ctx, req)
	if err != nil {
		return nil, "", err
	}

	model := req.Model // 注意：gateway 会把 model 重写成实际模型
	out := make(chan domain.StreamDelta, 16)

	go func() {
		defer close(out)

		var inputTokens, outputTokens int
		statusCode := httpStatusOK

		for {
			select {
			case <-ctx.Done():
				statusCode = httpStatusClientClosed
				latency := int(time.Since(start).Milliseconds())
				s.recordAsync(meta, model, provider, inputTokens, outputTokens, statusCode, latency)
				return
			case delta, ok := <-in:
				if !ok {
					latency := int(time.Since(start).Milliseconds())
					s.recordAsync(meta, model, provider, inputTokens, outputTokens, statusCode, latency)
					return
				}

				if delta.Usage != nil {
					inputTokens = delta.Usage.PromptTokens
					outputTokens = delta.Usage.CompletionTokens
				} else if delta.Content != nil {
					if delta.Content.Text != "" || delta.Content.Thinking != "" {
						outputTokens++
					}
				}

				select {
				case <-ctx.Done():
					statusCode = httpStatusClientClosed
					latency := int(time.Since(start).Milliseconds())
					s.recordAsync(meta, model, provider, inputTokens, outputTokens, statusCode, latency)
					return
				case out <- delta:
				}

				if delta.Type == "done" {
					latency := int(time.Since(start).Milliseconds())
					s.recordAsync(meta, model, provider, inputTokens, outputTokens, statusCode, latency)
					return
				}
			}
		}
	}()

	return out, provider, nil
}

func (s *service) preflight(ctx context.Context, userID int64) error {
	if userID <= 0 {
		return nil
	}
	if s.walletSvc == nil {
		return nil
	}
	has, err := s.walletSvc.HasBalance(ctx, userID)
	if err != nil {
		return errs.Wrap(errs.CodeInternalError, "Failed to check balance", err)
	}
	if !has {
		return errs.New(errs.CodeInsufficientBalance, "Insufficient balance. Please top up your wallet.")
	}
	return nil
}

const (
	httpStatusOK           = 200
	httpStatusClientClosed = 499
)

func (s *service) recordAsync(meta RequestMeta, model, provider string, inputTokens, outputTokens, statusCode, latency int) {
	// 未认证/未关联用户时不记录
	if meta.UserID <= 0 {
		return
	}

	// 与请求生命周期解耦，避免 ctx cancel 导致记录丢失
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		log := &domain.UsageLog{
			UserID:       meta.UserID,
			APIKeyID:     meta.APIKeyID,
			Model:        model,
			Provider:     provider,
			InputTokens:  inputTokens,
			OutputTokens: outputTokens,
			LatencyMs:    latency,
			StatusCode:   statusCode,
			ClientIP:     meta.ClientIP,
			UserAgent:    meta.UserAgent,
			RequestID:    meta.RequestID,
		}

		if err := s.usageSvc.LogRequest(ctx, log); err != nil {
			s.logger.Error("failed to log usage", logger.Error(err))
		}

		if meta.APIKeyID == nil {
			return
		}

		promptPrice, completionPrice, err := s.modelRateSvc.GetRateForModel(ctx, model)
		if err != nil {
			// modelrate service 已经做过降级，这里只记录日志
			s.logger.Warn("failed to get model rate", logger.String("model", model), logger.Error(err))
			return
		}

		cost := (float64(inputTokens)/1_000_000.0)*promptPrice + (float64(outputTokens)/1_000_000.0)*completionPrice
		if cost <= 0 {
			return
		}

		if err := s.apiKeySvc.IncrementUsage(ctx, *meta.APIKeyID, cost); err != nil {
			s.logger.Error("failed to increment api key usage",
				logger.Error(err),
				logger.Int64("key_id", *meta.APIKeyID),
				logger.Float64("cost", cost),
			)
		}
	}()
}
