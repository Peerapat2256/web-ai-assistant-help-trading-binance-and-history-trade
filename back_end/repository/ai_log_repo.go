package repository

import (
	"backend/model/model"
	"backend/model/query"
	"context"
	"fmt"
)

type AiLogRepository interface {
	Create(ctx context.Context, logEntry *model.AiAnalysisLog) error
	ListByMember(ctx context.Context, memberID int32, limit int) ([]model.AiAnalysisLog, error)
	Delete(ctx context.Context, id int32, memberID int32) error
}

type aiLogRepository struct{}

func NewAiLogRepository() AiLogRepository {
	return &aiLogRepository{}
}

func (r *aiLogRepository) Create(ctx context.Context, logEntry *model.AiAnalysisLog) error {
	return query.DB().WithContext(ctx).Create(logEntry).Error
}

func (r *aiLogRepository) ListByMember(ctx context.Context, memberID int32, limit int) ([]model.AiAnalysisLog, error) {
	var logs []model.AiAnalysisLog
	err := query.DB().WithContext(ctx).
		Where("member_id = ?", memberID).
		Order("created_at DESC").
		Limit(limit).
		Find(&logs).Error
	return logs, err
}

func (r *aiLogRepository) Delete(ctx context.Context, id int32, memberID int32) error {
	res := query.DB().WithContext(ctx).
		Where("id = ? AND member_id = ?", id, memberID).
		Delete(&model.AiAnalysisLog{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("ไม่พบรายการประวัติที่ต้องการลบ")
	}
	return nil
}
