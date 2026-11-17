package repository

import (
	"backend/model/model"
	"backend/model/query"
	"context"
)

type DashboardRepository interface {
	GetAllTradesAsc(ctx context.Context, memberID int32) ([]model.TradeHistory, error)
	GetTradesByMonth(ctx context.Context, memberID int32, year int, month int) ([]model.TradeHistory, error)
}

type dashboardRepository struct{}

func NewDashboardRepository() DashboardRepository {
	return &dashboardRepository{}
}

func (r *dashboardRepository) GetAllTradesAsc(ctx context.Context, memberID int32) ([]model.TradeHistory, error) {
	var trades []model.TradeHistory
	err := query.DB().WithContext(ctx).
		Where("member_id = ?", memberID).
		Order("trade_time ASC, trade_id ASC").
		Find(&trades).Error
	return trades, err
}

func (r *dashboardRepository) GetTradesByMonth(ctx context.Context, memberID int32, year int, month int) ([]model.TradeHistory, error) {
	var trades []model.TradeHistory
	err := query.DB().WithContext(ctx).
		Where("member_id = ? AND YEAR(trade_time) = ? AND MONTH(trade_time) = ?", memberID, year, month).
		Order("trade_time ASC").
		Find(&trades).Error
	return trades, err
}
