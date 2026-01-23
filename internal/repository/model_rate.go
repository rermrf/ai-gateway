package repository

import (
	"context"

	"ai-gateway/internal/domain"
	"ai-gateway/internal/repository/dao"
)

// ModelRateRepository 模型费率仓储接口
type ModelRateRepository interface {
	Create(ctx context.Context, rate *domain.ModelRate) error
	Update(ctx context.Context, rate *domain.ModelRate) error
	Delete(ctx context.Context, id int64) error
	GetByID(ctx context.Context, id int64) (*domain.ModelRate, error)
	List(ctx context.Context) ([]domain.ModelRate, error)
	GetAllEnabled(ctx context.Context) ([]domain.ModelRate, error)
}

type modelRateRepository struct {
	dao dao.ModelRateDAO
}

func NewModelRateRepository(dao dao.ModelRateDAO) ModelRateRepository {
	return &modelRateRepository{dao: dao}
}

func (r *modelRateRepository) toDomain(daoRate *dao.ModelRate) *domain.ModelRate {
	if daoRate == nil {
		return nil
	}
	return &domain.ModelRate{
		ID:              daoRate.ID,
		ModelPattern:    daoRate.ModelPattern,
		PromptPrice:     daoRate.PromptPrice,
		CompletionPrice: daoRate.CompletionPrice,
		Enabled:         daoRate.Enabled,
		CreatedAt:       daoRate.CreatedAt,
		UpdatedAt:       daoRate.UpdatedAt,
	}
}

func (r *modelRateRepository) toDAO(domainRate *domain.ModelRate) *dao.ModelRate {
	if domainRate == nil {
		return nil
	}
	return &dao.ModelRate{
		ID:              domainRate.ID,
		ModelPattern:    domainRate.ModelPattern,
		PromptPrice:     domainRate.PromptPrice,
		CompletionPrice: domainRate.CompletionPrice,
		Enabled:         domainRate.Enabled,
		CreatedAt:       domainRate.CreatedAt,
		UpdatedAt:       domainRate.UpdatedAt,
	}
}

func (r *modelRateRepository) Create(ctx context.Context, rate *domain.ModelRate) error {
	daoRate := r.toDAO(rate)
	if err := r.dao.Create(ctx, daoRate); err != nil {
		return err
	}
	rate.ID = daoRate.ID
	rate.CreatedAt = daoRate.CreatedAt
	rate.UpdatedAt = daoRate.UpdatedAt
	return nil
}

func (r *modelRateRepository) Update(ctx context.Context, rate *domain.ModelRate) error {
	return r.dao.Update(ctx, r.toDAO(rate))
}

func (r *modelRateRepository) Delete(ctx context.Context, id int64) error {
	return r.dao.Delete(ctx, id)
}

func (r *modelRateRepository) GetByID(ctx context.Context, id int64) (*domain.ModelRate, error) {
	daoRate, err := r.dao.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return r.toDomain(daoRate), nil
}

func (r *modelRateRepository) List(ctx context.Context) ([]domain.ModelRate, error) {
	daoRates, err := r.dao.List(ctx)
	if err != nil {
		return nil, err
	}
	rates := make([]domain.ModelRate, len(daoRates))
	for i, item := range daoRates {
		rates[i] = *r.toDomain(&item)
	}
	return rates, nil
}

func (r *modelRateRepository) GetAllEnabled(ctx context.Context) ([]domain.ModelRate, error) {
	daoRates, err := r.dao.GetAllEnabled(ctx)
	if err != nil {
		return nil, err
	}
	rates := make([]domain.ModelRate, len(daoRates))
	for i, item := range daoRates {
		rates[i] = *r.toDomain(&item)
	}
	return rates, nil
}
