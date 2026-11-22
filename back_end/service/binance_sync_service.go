package service

import (
	"backend/model/model"
	"backend/repository"
	"backend/utils"
	"context"
	"fmt"
	"strings"
	"time"
)

type BinanceSyncService interface {
	GetConfig(ctx context.Context, memberID int32) (*model.BinanceConfigResponse, error)
	SaveConfig(ctx context.Context, memberID int32, req model.BinanceConfigRequest) (*model.BinanceConfigResponse, error)
	DeleteConfig(ctx context.Context, memberID int32) error
	TestConnection(ctx context.Context, memberID int32, req model.BinanceConfigRequest) (float64, error)
	SyncTradesNow(ctx context.Context, memberID int32, lookbackDays int) (*model.BinanceSyncResult, error)
}

type binanceSyncServiceImpl struct {
	binanceRepo repository.BinanceConfigRepository
	tradeRepo   repository.TradeRepository
	tradeSvc    TradeService
}

func NewBinanceSyncService(
	binanceRepo repository.BinanceConfigRepository,
	tradeRepo repository.TradeRepository,
	tradeSvc TradeService,
) BinanceSyncService {
	return &binanceSyncServiceImpl{
		binanceRepo: binanceRepo,
		tradeRepo:   tradeRepo,
		tradeSvc:    tradeSvc,
	}
}

func (s *binanceSyncServiceImpl) GetConfig(ctx context.Context, memberID int32) (*model.BinanceConfigResponse, error) {
	cfg, err := s.binanceRepo.GetByMemberID(ctx, memberID)
	if err != nil || cfg == nil {
		return &model.BinanceConfigResponse{
			Connected:  false,
			IsAutoSync: true,
		}, nil
	}
	return &model.BinanceConfigResponse{
		Connected:     cfg.APIKeyEnc != "" && cfg.APISecretEnc != "",
		APIKeyMasked:  cfg.APIKeyMasked,
		IsAutoSync:    cfg.IsAutoSync,
		LastSyncedAt:  cfg.LastSyncedAt,
		LastSyncError: cfg.LastSyncError,
	}, nil
}

func (s *binanceSyncServiceImpl) SaveConfig(ctx context.Context, memberID int32, req model.BinanceConfigRequest) (*model.BinanceConfigResponse, error) {
	existing, _ := s.binanceRepo.GetByMemberID(ctx, memberID)

	apiKey := strings.TrimSpace(req.APIKey)
	apiSecret := strings.TrimSpace(req.APISecret)

	cfg := &model.BinanceConfig{
		MemberID:   memberID,
		IsAutoSync: req.IsAutoSync,
	}

	if existing != nil {
		cfg.APIKeyEnc = existing.APIKeyEnc
		cfg.APISecretEnc = existing.APISecretEnc
		cfg.APIKeyMasked = existing.APIKeyMasked
		cfg.LastSyncedAt = existing.LastSyncedAt
	}

	if apiKey != "" {
		enc, err := utils.EncryptSecret(apiKey)
		if err != nil {
			return nil, fmt.Errorf("เข้ารหัส API Key ไม่สำเร็จ")
		}
		cfg.APIKeyEnc = enc
		cfg.APIKeyMasked = utils.MaskAPIKey(apiKey)
	}
	if apiSecret != "" {
		enc, err := utils.EncryptSecret(apiSecret)
		if err != nil {
			return nil, fmt.Errorf("เข้ารหัส API Secret ไม่สำเร็จ")
		}
		cfg.APISecretEnc = enc
	}

	if cfg.APIKeyEnc == "" || cfg.APISecretEnc == "" {
		return nil, fmt.Errorf("กรุณากรอกทั้ง API Key และ Secret Key")
	}

	cfg.LastSyncError = ""
	if err := s.binanceRepo.Upsert(ctx, cfg); err != nil {
		return nil, err
	}

	return s.GetConfig(ctx, memberID)
}

func (s *binanceSyncServiceImpl) DeleteConfig(ctx context.Context, memberID int32) error {
	return s.binanceRepo.DeleteByMemberID(ctx, memberID)
}

func (s *binanceSyncServiceImpl) resolveCredentials(ctx context.Context, memberID int32, reqKey, reqSecret string) (string, string, error) {
	k := strings.TrimSpace(reqKey)
	sec := strings.TrimSpace(reqSecret)
	if k != "" && sec != "" {
		return k, sec, nil
	}

	cfg, err := s.binanceRepo.GetByMemberID(ctx, memberID)
	if err != nil || cfg == nil || cfg.APIKeyEnc == "" || cfg.APISecretEnc == "" {
		return "", "", fmt.Errorf("ยังไม่ได้ตั้งค่า Binance API Key")
	}
	if k == "" {
		k, err = utils.DecryptSecret(cfg.APIKeyEnc)
		if err != nil {
			return "", "", fmt.Errorf("ถอดรหัส API Key ไม่สำเร็จ")
		}
	}
	if sec == "" {
		sec, err = utils.DecryptSecret(cfg.APISecretEnc)
		if err != nil {
			return "", "", fmt.Errorf("ถอดรหัส API Secret ไม่สำเร็จ")
		}
	}
	return k, sec, nil
}

func (s *binanceSyncServiceImpl) TestConnection(ctx context.Context, memberID int32, req model.BinanceConfigRequest) (float64, error) {
	k, sec, err := s.resolveCredentials(ctx, memberID, req.APIKey, req.APISecret)
	if err != nil {
		return 0, err
	}
	return utils.TestBinanceFuturesConnection(k, sec)
}

func (s *binanceSyncServiceImpl) SyncTradesNow(ctx context.Context, memberID int32, lookbackDays int) (*model.BinanceSyncResult, error) {
	cfg, err := s.binanceRepo.GetByMemberID(ctx, memberID)
	if err != nil || cfg == nil || cfg.APIKeyEnc == "" || cfg.APISecretEnc == "" {
		return nil, fmt.Errorf("ยังไม่ได้เชื่อมต่อ Binance API Key")
	}

	apiKey, err := utils.DecryptSecret(cfg.APIKeyEnc)
	if err != nil {
		return nil, err
	}
	apiSecret, err := utils.DecryptSecret(cfg.APISecretEnc)
	if err != nil {
		return nil, err
	}

	fetched, err := utils.FetchClosedFuturesTrades(apiKey, apiSecret, lookbackDays)
	if err != nil {
		cfg.LastSyncError = err.Error()
		_ = s.binanceRepo.Upsert(ctx, cfg)
		return nil, err
	}

	res := &model.BinanceSyncResult{
		FetchedCount: len(fetched),
	}

	for _, item := range fetched {
		// 1) เช็คซ้ำด้วย ExternalOrderID ก่อน (แม่นยำ 100%)
		if existing, err := s.tradeRepo.FindByExternalOrderID(ctx, memberID, item.ExternalOrderID); err == nil && existing != nil {
			res.SkippedCount++
			continue
		}
		// 2) เช็คซ้ำกับรายการที่เคยสแกนจากรูปภาพไว้ก่อนหน้า
		if existing, err := s.tradeRepo.FindPotentialDuplicate(ctx, memberID, item.Symbol, item.PositionSide, item.EntryPrice, item.ExitPrice, item.Pnl, item.TradeTime); err == nil && existing != nil {
			res.SkippedCount++
			continue
		}

		_, err := s.tradeSvc.CreateTrade(ctx, memberID, model.TradeCreateRequest{
			ExternalOrderID: item.ExternalOrderID,
			Symbol:          item.Symbol,
			PositionSide:    item.PositionSide,
			Leverage:        item.Leverage,
			EntryPrice:      item.EntryPrice,
			ExitPrice:       item.ExitPrice,
			Quantity:        item.Quantity,
			MarginUsed:      item.MarginUsed,
			Pnl:             item.Pnl,
			PnlPercentage:   item.PnlPercentage,
			Fee:             item.Fee,
			TradeTime:       item.TradeTime.Format("2006-01-02 15:04:05"),
			Notes:           "ดึงอัตโนมัติจาก Binance Futures API",
		})
		if err != nil {
			continue
		}
		res.CreatedCount++
	}

	now := time.Now()
	cfg.LastSyncedAt = &now
	cfg.LastSyncError = ""
	_ = s.binanceRepo.Upsert(ctx, cfg)

	return res, nil
}
