package repository

import (
	"backend/model/model"
	"backend/model/query"
	"context"
	"time"
)

type SheetConfigRepository interface {
	GetByMemberID(ctx context.Context, memberID int32) (*model.SheetConfig, error)
	Upsert(ctx context.Context, cfg *model.SheetConfig) error
	UpdateLastSynced(ctx context.Context, memberID int32) error
}

type sheetConfigRepository struct{}

func NewSheetConfigRepository() SheetConfigRepository {
	return &sheetConfigRepository{}
}

func (r *sheetConfigRepository) GetByMemberID(ctx context.Context, memberID int32) (*model.SheetConfig, error) {
	var cfg model.SheetConfig
	err := query.DB().WithContext(ctx).Where("member_id = ?", memberID).First(&cfg).Error
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (r *sheetConfigRepository) Upsert(ctx context.Context, cfg *model.SheetConfig) error {
	var existing model.SheetConfig
	err := query.DB().WithContext(ctx).Where("member_id = ?", cfg.MemberID).First(&existing).Error
	if err == nil {
		cfg.ID = existing.ID
		return query.DB().WithContext(ctx).Model(&existing).Updates(map[string]interface{}{
			"spreadsheet_id": cfg.SpreadsheetID,
			"sheet_name":     cfg.SheetName,
			"is_auto_sync":   cfg.IsAutoSync,
		}).Error
	}
	return query.DB().WithContext(ctx).Create(cfg).Error
}

func (r *sheetConfigRepository) UpdateLastSynced(ctx context.Context, memberID int32) error {
	now := time.Now()
	return query.DB().WithContext(ctx).Model(&model.SheetConfig{}).
		Where("member_id = ?", memberID).
		Update("last_synced_at", &now).Error
}
