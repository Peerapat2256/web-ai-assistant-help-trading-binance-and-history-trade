package repository

import (
	"backend/model/model"
	"backend/model/query"
	"context"
)

type TradeRepository interface {
	Create(ctx context.Context, trade *model.TradeHistory) error
	GetByID(ctx context.Context, tradeID int32, memberID int32) (*model.TradeHistory, error)
	ListByMember(ctx context.Context, memberID int32, symbol string, side string) ([]model.TradeHistory, error)
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
