package repository

import (
	"backend/model/model"
	"backend/model/query"
	"context"
	"time"
)

type WithdrawalRepository interface {
	Create(ctx context.Context, w *model.WithdrawalHistory) error
	GetByID(ctx context.Context, withdrawalID int32, memberID int32) (*model.WithdrawalHistory, error)
	FindByTransactionID(ctx context.Context, memberID int32, txID string) (*model.WithdrawalHistory, error)
	FindPotentialDuplicate(ctx context.Context, memberID int32, amount float64, withdrawalTime time.Time) (*model.WithdrawalHistory, error)
	ListByMember(ctx context.Context, memberID int32, year int) ([]model.WithdrawalHistory, error)
	Update(ctx context.Context, w *model.WithdrawalHistory) error
	Delete(ctx context.Context, withdrawalID int32, memberID int32) error
}

type withdrawalRepository struct{}

func NewWithdrawalRepository() WithdrawalRepository {
	return &withdrawalRepository{}
}

func (r *withdrawalRepository) Create(ctx context.Context, w *model.WithdrawalHistory) error {
	return query.DB().WithContext(ctx).Create(w).Error
}

func (r *withdrawalRepository) GetByID(ctx context.Context, withdrawalID int32, memberID int32) (*model.WithdrawalHistory, error) {
	var w model.WithdrawalHistory
	err := query.DB().WithContext(ctx).Where("withdrawal_id = ? AND member_id = ?", withdrawalID, memberID).First(&w).Error
	if err != nil {
		return nil, err
	}
	return &w, nil
}

func (r *withdrawalRepository) FindByTransactionID(ctx context.Context, memberID int32, txID string) (*model.WithdrawalHistory, error) {
	if txID == "" {
		return nil, nil
	}
	var w model.WithdrawalHistory
	err := query.DB().WithContext(ctx).Where("member_id = ? AND transaction_id = ?", memberID, txID).First(&w).Error
	if err != nil {
		return nil, err
	}
	return &w, nil
}

func (r *withdrawalRepository) FindPotentialDuplicate(ctx context.Context, memberID int32, amount float64, withdrawalTime time.Time) (*model.WithdrawalHistory, error) {
	var w model.WithdrawalHistory
	fromTime := withdrawalTime.Add(-3 * time.Minute)
	toTime := withdrawalTime.Add(3 * time.Minute)
	err := query.DB().WithContext(ctx).
		Where("member_id = ? AND ABS(amount - ?) <= 0.01 AND withdrawal_time BETWEEN ? AND ?", memberID, amount, fromTime, toTime).
		First(&w).Error
	if err != nil {
		return nil, err
	}
	return &w, nil
}

func (r *withdrawalRepository) ListByMember(ctx context.Context, memberID int32, year int) ([]model.WithdrawalHistory, error) {
	var items []model.WithdrawalHistory
	q := query.DB().WithContext(ctx).Where("member_id = ?", memberID)
	if year > 0 {
		start := time.Date(year, 1, 1, 0, 0, 0, 0, time.FixedZone("ICT", 7*3600))
		end := time.Date(year+1, 1, 1, 0, 0, 0, 0, time.FixedZone("ICT", 7*3600))
		q = q.Where("withdrawal_time >= ? AND withdrawal_time < ?", start, end)
	}
	err := q.Order("withdrawal_time DESC, withdrawal_id DESC").Find(&items).Error
	return items, err
}

func (r *withdrawalRepository) Update(ctx context.Context, w *model.WithdrawalHistory) error {
	return query.DB().WithContext(ctx).Save(w).Error
}

func (r *withdrawalRepository) Delete(ctx context.Context, withdrawalID int32, memberID int32) error {
	return query.DB().WithContext(ctx).Where("withdrawal_id = ? AND member_id = ?", withdrawalID, memberID).Delete(&model.WithdrawalHistory{}).Error
}
