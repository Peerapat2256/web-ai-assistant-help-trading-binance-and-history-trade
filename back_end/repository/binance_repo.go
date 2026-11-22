package repository

import (
	"backend/model/model"
	"backend/model/query"
	"context"
)

type BinanceConfigRepository interface {
	GetByMemberID(ctx context.Context, memberID int32) (*model.BinanceConfig, error)
	Upsert(ctx context.Context, cfg *model.BinanceConfig) error
	DeleteByMemberID(ctx context.Context, memberID int32) error
	ListAutoSyncEnabled(ctx context.Context) ([]model.BinanceConfig, error)
}

type binanceConfigRepository struct{}

func NewBinanceConfigRepository() BinanceConfigRepository {
	return &binanceConfigRepository{}
}

func (r *binanceConfigRepository) GetByMemberID(ctx context.Context, memberID int32) (*model.BinanceConfig, error) {
	var cfg model.BinanceConfig
	err := query.DB().WithContext(ctx).Where("member_id = ?", memberID).First(&cfg).Error
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (r *binanceConfigRepository) Upsert(ctx context.Context, cfg *model.BinanceConfig) error {
	existing, err := r.GetByMemberID(ctx, cfg.MemberID)
	if err == nil && existing != nil {
		cfg.ID = existing.ID
		cfg.CreatedAt = existing.CreatedAt
		return query.DB().WithContext(ctx).Save(cfg).Error
	}
	return query.DB().WithContext(ctx).Create(cfg).Error
}

func (r *binanceConfigRepository) DeleteByMemberID(ctx context.Context, memberID int32) error {
	return query.DB().WithContext(ctx).Where("member_id = ?", memberID).Delete(&model.BinanceConfig{}).Error
}

func (r *binanceConfigRepository) ListAutoSyncEnabled(ctx context.Context) ([]model.BinanceConfig, error) {
	var configs []model.BinanceConfig
	err := query.DB().WithContext(ctx).
		Where("is_auto_sync = ? AND api_key_enc <> '' AND api_secret_enc <> ''", true).
		Find(&configs).Error
	return configs, err
}
