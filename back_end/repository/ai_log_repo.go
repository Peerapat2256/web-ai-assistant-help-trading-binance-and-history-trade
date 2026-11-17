package repository

import (
	"backend/model/model"
	"backend/model/query"
	"context"
)

type AiLogRepository interface {
	Create(ctx context.Context, logEntry *model.AiAnalysisLog) error
	ListByMember(ctx context.Context, memberID int32, limit int) ([]model.AiAnalysisLog, error)
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
