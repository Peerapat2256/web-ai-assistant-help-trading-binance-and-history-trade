package service

import (
	"backend/model/model"
	"backend/repository"
	"context"
)

type PortfolioService interface {
	GetSettings(ctx context.Context, memberID int32) (*model.PortfolioSetting, error)
	UpdateSettings(ctx context.Context, memberID int32, req model.PortfolioUpdateRequest) (*model.PortfolioSetting, error)
}

type portfolioServiceImpl struct {
	repo repository.PortfolioRepository
}

func NewPortfolioService(repo repository.PortfolioRepository) PortfolioService {
	return &portfolioServiceImpl{repo: repo}
}

func (s *portfolioServiceImpl) GetSettings(ctx context.Context, memberID int32) (*model.PortfolioSetting, error) {
	setting, err := s.repo.GetByMemberID(ctx, memberID)
	if err != nil {
		def := &model.PortfolioSetting{
			MemberID:         memberID,
			InitialBalance:   1000,
			Currency:         "USDT",
			MonthlyTargetPct: 10,
			MaxDailyLossUSD:  50,
		}
		_ = s.repo.Upsert(ctx, def)
		return def, nil
	}
	return setting, nil
}

func (s *portfolioServiceImpl) UpdateSettings(ctx context.Context, memberID int32, req model.PortfolioUpdateRequest) (*model.PortfolioSetting, error) {
	if req.Currency == "" {
		req.Currency = "USDT"
	}
	if req.InitialBalance <= 0 {
		req.InitialBalance = 1000
	}
	setting := &model.PortfolioSetting{
		MemberID:         memberID,
		InitialBalance:   req.InitialBalance,
		Currency:         req.Currency,
		MonthlyTargetPct: req.MonthlyTargetPct,
		MaxDailyLossUSD:  req.MaxDailyLossUSD,
	}
	if err := s.repo.Upsert(ctx, setting); err != nil {
		return nil, err
	}
	return s.repo.GetByMemberID(ctx, memberID)
}
