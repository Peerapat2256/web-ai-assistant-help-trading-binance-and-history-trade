package repository

import (
	"backend/model/model"
	"backend/model/query"
	"context"
)

type PortfolioRepository interface {
	GetByMemberID(ctx context.Context, memberID int32) (*model.PortfolioSetting, error)
	Upsert(ctx context.Context, setting *model.PortfolioSetting) error
}

type portfolioRepository struct{}

func NewPortfolioRepository() PortfolioRepository {
	return &portfolioRepository{}
}

func (r *portfolioRepository) GetByMemberID(ctx context.Context, memberID int32) (*model.PortfolioSetting, error) {
	var setting model.PortfolioSetting
	err := query.DB().WithContext(ctx).Where("member_id = ?", memberID).First(&setting).Error
	if err != nil {
		return nil, err
	}
	return &setting, nil
}

func (r *portfolioRepository) Upsert(ctx context.Context, setting *model.PortfolioSetting) error {
	var existing model.PortfolioSetting
	err := query.DB().WithContext(ctx).Where("member_id = ?", setting.MemberID).First(&existing).Error
	if err == nil {
		setting.ID = existing.ID
		return query.DB().WithContext(ctx).Model(&existing).Updates(map[string]interface{}{
			"initial_balance":    setting.InitialBalance,
			"currency":           setting.Currency,
			"monthly_target_pct": setting.MonthlyTargetPct,
			"max_daily_loss_usd": setting.MaxDailyLossUSD,
		}).Error
	}
	return query.DB().WithContext(ctx).Create(setting).Error
}
