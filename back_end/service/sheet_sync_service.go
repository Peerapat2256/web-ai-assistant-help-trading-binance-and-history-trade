package service

import (
	"backend/model/model"
	"backend/repository"
	"backend/utils"
	"context"
	"log"
	"os"
	"strings"
)

type SheetSyncService interface {
	GetConfig(ctx context.Context, memberID int32) (*model.SheetConfig, error)
	SaveConfig(ctx context.Context, memberID int32, req model.SheetConfigRequest) (*model.SheetConfig, error)
	TestConnection(ctx context.Context, memberID int32, req *model.SheetConfigRequest) (string, error)
	SyncSingleTradeCreate(ctx context.Context, trade *model.TradeHistory) string
	SyncSingleTradeUpdate(ctx context.Context, trade *model.TradeHistory) string
	SyncSingleTradeDelete(ctx context.Context, memberID int32, tradeCode string) error
	FullResyncAll(ctx context.Context, memberID int32) (int, error)
}

type sheetSyncServiceImpl struct {
	sheetRepo repository.SheetConfigRepository
	tradeRepo repository.TradeRepository
}

func NewSheetSyncService(sheetRepo repository.SheetConfigRepository, tradeRepo repository.TradeRepository) SheetSyncService {
	return &sheetSyncServiceImpl{
		sheetRepo: sheetRepo,
		tradeRepo: tradeRepo,
	}
}

func (s *sheetSyncServiceImpl) GetConfig(ctx context.Context, memberID int32) (*model.SheetConfig, error) {
	cfg, err := s.sheetRepo.GetByMemberID(ctx, memberID)
	if err == nil && cfg != nil {
		return cfg, nil
	}
	defaultID := strings.TrimSpace(os.Getenv("GOOGLE_SHEET_ID"))
	newCfg := &model.SheetConfig{
		MemberID:      memberID,
		SpreadsheetID: defaultID,
		SheetName:     "Trades",
		IsAutoSync:    defaultID != "",
	}
	if defaultID != "" {
		_ = s.sheetRepo.Upsert(ctx, newCfg)
	}
	return newCfg, nil
}

func (s *sheetSyncServiceImpl) SaveConfig(ctx context.Context, memberID int32, req model.SheetConfigRequest) (*model.SheetConfig, error) {
	sheetName := strings.TrimSpace(req.SheetName)
	if sheetName == "" {
		sheetName = "Trades"
	}
	cfg := &model.SheetConfig{
		MemberID:      memberID,
		SpreadsheetID: utils.ExtractSpreadsheetID(req.SpreadsheetID),
		SheetName:     sheetName,
		IsAutoSync:    req.IsAutoSync,
	}
	if err := s.sheetRepo.Upsert(ctx, cfg); err != nil {
		return nil, err
	}
	return s.sheetRepo.GetByMemberID(ctx, memberID)
}

func (s *sheetSyncServiceImpl) TestConnection(ctx context.Context, memberID int32, req *model.SheetConfigRequest) (string, error) {
	spreadsheetID := ""
	sheetName := "Trades"
	if req != nil && req.SpreadsheetID != "" {
		spreadsheetID = req.SpreadsheetID
		if req.SheetName != "" {
			sheetName = req.SheetName
		}
	} else {
		cfg, err := s.GetConfig(ctx, memberID)
		if err != nil {
			return "", err
		}
		spreadsheetID = cfg.SpreadsheetID
		sheetName = cfg.SheetName
	}

	title, err := utils.TestSheetConnection(spreadsheetID, sheetName)
	if err != nil {
		return "", err
	}
	_ = s.sheetRepo.UpdateLastSynced(ctx, memberID)
	return title, nil
}

func (s *sheetSyncServiceImpl) SyncSingleTradeCreate(ctx context.Context, trade *model.TradeHistory) string {
	cfg, err := s.GetConfig(ctx, trade.MemberID)
	if err != nil || !cfg.IsAutoSync || cfg.SpreadsheetID == "" {
		return "Pending"
	}
	if err := utils.AppendTradeToSheet(cfg.SpreadsheetID, cfg.SheetName, trade); err != nil {
		log.Printf("[Google Sheets] เพิ่มแถวไม่สำเร็จ: %v", err)
		return "Failed"
	}
	_ = s.sheetRepo.UpdateLastSynced(ctx, trade.MemberID)
	return "Synced"
}

func (s *sheetSyncServiceImpl) SyncSingleTradeUpdate(ctx context.Context, trade *model.TradeHistory) string {
	cfg, err := s.GetConfig(ctx, trade.MemberID)
	if err != nil || !cfg.IsAutoSync || cfg.SpreadsheetID == "" {
		return "Pending"
	}
	if err := utils.UpdateTradeInSheet(cfg.SpreadsheetID, cfg.SheetName, trade); err != nil {
		log.Printf("[Google Sheets] อัปเดตแถวไม่สำเร็จ: %v", err)
		return "Failed"
	}
	_ = s.sheetRepo.UpdateLastSynced(ctx, trade.MemberID)
	return "Synced"
}

func (s *sheetSyncServiceImpl) SyncSingleTradeDelete(ctx context.Context, memberID int32, tradeCode string) error {
	cfg, err := s.GetConfig(ctx, memberID)
	if err != nil || !cfg.IsAutoSync || cfg.SpreadsheetID == "" {
		return nil
	}
	if err := utils.DeleteTradeFromSheet(cfg.SpreadsheetID, cfg.SheetName, tradeCode); err != nil {
		log.Printf("[Google Sheets] ลบแถวไม่สำเร็จ: %v", err)
		return err
	}
	_ = s.sheetRepo.UpdateLastSynced(ctx, memberID)
	return nil
}

func (s *sheetSyncServiceImpl) FullResyncAll(ctx context.Context, memberID int32) (int, error) {
	cfg, err := s.GetConfig(ctx, memberID)
	if err != nil {
		return 0, err
	}
	trades, err := s.tradeRepo.ListByMember(ctx, memberID, "", "")
	if err != nil {
		return 0, err
	}
	if err := utils.FullSyncTradesToSheet(cfg.SpreadsheetID, cfg.SheetName, trades); err != nil {
		return 0, err
	}
	for _, t := range trades {
		_ = s.tradeRepo.UpdateSyncStatus(ctx, t.TradeID, "Synced")
	}
	_ = s.sheetRepo.UpdateLastSynced(ctx, memberID)
	return len(trades), nil
}
