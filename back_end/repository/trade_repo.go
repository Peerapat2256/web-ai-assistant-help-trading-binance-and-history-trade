package repository

import (
	"backend/model/model"
	"backend/model/query"
	"context"
	"time"
)

type TradeRepository interface {
	Create(ctx context.Context, trade *model.TradeHistory) error
	GetByID(ctx context.Context, tradeID int32, memberID int32) (*model.TradeHistory, error)
	ListByMember(ctx context.Context, memberID int32, symbol string, side string) ([]model.TradeHistory, error)
	FindByExternalOrderID(ctx context.Context, memberID int32, extOrderID string) (*model.TradeHistory, error)
	FindPotentialDuplicate(ctx context.Context, memberID int32, symbol string, side string, entryPrice float64, exitPrice float64, pnl float64, tradeTime time.Time) (*model.TradeHistory, error)
	Update(ctx context.Context, trade *model.TradeHistory) error
	UpdateSyncStatus(ctx context.Context, tradeID int32, status string) error
	Delete(ctx context.Context, tradeID int32, memberID int32) error
	FindPendingSync(ctx context.Context, limit int) ([]model.TradeHistory, error)
}

type tradeRepository struct{}

func NewTradeRepository() TradeRepository {
	return &tradeRepository{}
}

func (r *tradeRepository) Create(ctx context.Context, trade *model.TradeHistory) error {
	return query.DB().WithContext(ctx).Create(trade).Error
}

func (r *tradeRepository) GetByID(ctx context.Context, tradeID int32, memberID int32) (*model.TradeHistory, error) {
	var trade model.TradeHistory
	err := query.DB().WithContext(ctx).Where("trade_id = ? AND member_id = ?", tradeID, memberID).First(&trade).Error
	if err != nil {
		return nil, err
	}
	return &trade, nil
}

func (r *tradeRepository) ListByMember(ctx context.Context, memberID int32, symbol string, side string) ([]model.TradeHistory, error) {
	var trades []model.TradeHistory
	q := query.DB().WithContext(ctx).Where("member_id = ?", memberID)
	if symbol != "" {
		q = q.Where("symbol LIKE ?", "%"+symbol+"%")
	}
	if side != "" && side != "ALL" {
		q = q.Where("position_side = ?", side)
	}
	err := q.Order("trade_time DESC, trade_id DESC").Find(&trades).Error
	return trades, err
}

func (r *tradeRepository) Update(ctx context.Context, trade *model.TradeHistory) error {
	return query.DB().WithContext(ctx).Save(trade).Error
}

func (r *tradeRepository) UpdateSyncStatus(ctx context.Context, tradeID int32, status string) error {
	return query.DB().WithContext(ctx).Model(&model.TradeHistory{}).
		Where("trade_id = ?", tradeID).
		Update("sync_sheet_status", status).Error
}

func (r *tradeRepository) Delete(ctx context.Context, tradeID int32, memberID int32) error {
	return query.DB().WithContext(ctx).Where("trade_id = ? AND member_id = ?", tradeID, memberID).Delete(&model.TradeHistory{}).Error
}

func (r *tradeRepository) FindPendingSync(ctx context.Context, limit int) ([]model.TradeHistory, error) {
	var trades []model.TradeHistory
	err := query.DB().WithContext(ctx).Where("sync_sheet_status IN ?", []string{"Pending", "Failed"}).Limit(limit).Find(&trades).Error
	return trades, err
}

func (r *tradeRepository) FindByExternalOrderID(ctx context.Context, memberID int32, extOrderID string) (*model.TradeHistory, error) {
	if extOrderID == "" {
		return nil, nil
	}
	var trade model.TradeHistory
	err := query.DB().WithContext(ctx).
		Where("member_id = ? AND external_order_id = ?", memberID, extOrderID).
		First(&trade).Error
	if err != nil {
		return nil, err
	}
	return &trade, nil
}

func (r *tradeRepository) FindPotentialDuplicate(ctx context.Context, memberID int32, symbol string, side string, entryPrice float64, exitPrice float64, pnl float64, tradeTime time.Time) (*model.TradeHistory, error) {
	var trade model.TradeHistory
	fromTime := tradeTime.Add(-5 * time.Minute)
	toTime := tradeTime.Add(5 * time.Minute)

	q := query.DB().WithContext(ctx).
		Where("member_id = ? AND symbol = ? AND position_side = ?", memberID, symbol, side).
		Where("ABS(pnl - ?) <= 0.05", pnl).
		Where("trade_time BETWEEN ? AND ?", fromTime, toTime)

	if entryPrice > 0 {
		q = q.Where("ABS(entry_price - ?) <= ?", entryPrice, entryPrice*0.001+0.05)
	}
	if exitPrice > 0 {
		q = q.Where("ABS(exit_price - ?) <= ?", exitPrice, exitPrice*0.001+0.05)
	}

	if err := q.First(&trade).Error; err != nil {
		return nil, err
	}
	return &trade, nil
}
