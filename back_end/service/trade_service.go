package service

import (
	"backend/model/model"
	"backend/repository"
	"backend/utils"
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"
)

type TradeService interface {
	ListTrades(ctx context.Context, memberID int32, symbol string, side string) ([]model.TradeHistory, error)
	CreateTrade(ctx context.Context, memberID int32, req model.TradeCreateRequest) (*model.TradeHistory, error)
	CreateTradesBatch(ctx context.Context, memberID int32, req model.BatchTradeCreateRequest) (*model.BatchTradeCreateResponse, error)
	AnnotateDuplicates(ctx context.Context, memberID int32, parsed *model.ParsedScreenshotBatchResponse)
	UpdateTrade(ctx context.Context, memberID int32, tradeID int32, req model.TradeUpdateRequest) (*model.TradeHistory, error)
	DeleteTrade(ctx context.Context, memberID int32, tradeID int32) error
}

type tradeServiceImpl struct {
	tradeRepo repository.TradeRepository
	sheetSvc  SheetSyncService
}

func NewTradeService(tradeRepo repository.TradeRepository, sheetSvc SheetSyncService) TradeService {
	return &tradeServiceImpl{
		tradeRepo: tradeRepo,
		sheetSvc:  sheetSvc,
	}
}

func parseTradeTime(raw string) time.Time {
	s := strings.TrimSpace(raw)
	if s == "" {
		return time.Now()
	}
	layouts := []string{
		"2006-01-02T15:04",
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		time.RFC3339,
		"2006-01-02",
	}
	for _, layout := range layouts {
		if t, err := time.ParseInLocation(layout, s, time.FixedZone("ICT", 7*3600)); err == nil {
			return t
		}
	}
	return time.Now()
}

func generateTradeCode() string {
	return fmt.Sprintf("TRD-%s-%04d", time.Now().Format("0601021504"), rand.Intn(10000))
}

func (s *tradeServiceImpl) ListTrades(ctx context.Context, memberID int32, symbol string, side string) ([]model.TradeHistory, error) {
	return s.tradeRepo.ListByMember(ctx, memberID, strings.ToUpper(strings.TrimSpace(symbol)), strings.ToUpper(strings.TrimSpace(side)))
}

func (s *tradeServiceImpl) CreateTrade(ctx context.Context, memberID int32, req model.TradeCreateRequest) (*model.TradeHistory, error) {
	lev := req.Leverage
	if lev <= 0 {
		lev = 1
	}

	trade := &model.TradeHistory{
		TradeCode:       generateTradeCode(),
		MemberID:        memberID,
		Symbol:          utils.NormalizeBinanceSymbol(req.Symbol),
		PositionSide:    strings.ToUpper(strings.TrimSpace(req.PositionSide)),
		Leverage:        lev,
		EntryPrice:      req.EntryPrice,
		ExitPrice:       req.ExitPrice,
		Quantity:        req.Quantity,
		MarginUsed:      req.MarginUsed,
		Pnl:             req.Pnl,
		PnlPercentage:   req.PnlPercentage,
		Fee:             req.Fee,
		TradeTime:       parseTradeTime(req.TradeTime),
		SetupTag:        strings.TrimSpace(req.SetupTag),
		EmotionTag:      strings.TrimSpace(req.EmotionTag),
		Notes:           strings.TrimSpace(req.Notes),
		ScreenshotURL:   strings.TrimSpace(req.ScreenshotURL),
		SyncSheetStatus: "Pending",
	}

	if err := s.tradeRepo.Create(ctx, trade); err != nil {
		return nil, err
	}

	status := s.sheetSvc.SyncSingleTradeCreate(ctx, trade)
	trade.SyncSheetStatus = status
	_ = s.tradeRepo.UpdateSyncStatus(ctx, trade.TradeID, status)

	return trade, nil
}

func (s *tradeServiceImpl) AnnotateDuplicates(ctx context.Context, memberID int32, parsed *model.ParsedScreenshotBatchResponse) {
	if parsed == nil || len(parsed.Trades) == 0 {
		return
	}
	dupCount := 0
	seenKeys := make(map[string]bool)

	for i := range parsed.Trades {
		item := &parsed.Trades[i]
		sym := utils.NormalizeBinanceSymbol(item.Symbol)
		side := strings.ToUpper(strings.TrimSpace(item.PositionSide))
		tTime := parseTradeTime(item.TradeTime)

		key := fmt.Sprintf("%s|%s|%.2f|%.2f|%s", sym, side, item.EntryPrice, item.Pnl, tTime.Format("2006-01-02 15:04"))
		if seenKeys[key] {
			item.IsDuplicate = true
			item.DuplicateReason = "รายการซ้ำภายในรูปภาพเดียวกัน"
			dupCount++
			continue
		}
		seenKeys[key] = true

		if existing, err := s.tradeRepo.FindPotentialDuplicate(ctx, memberID, sym, side, item.EntryPrice, item.ExitPrice, item.Pnl, tTime); err == nil && existing != nil {
			item.IsDuplicate = true
			item.DuplicateReason = fmt.Sprintf("มีรายการนี้ในระบบแล้ว (%s)", existing.TradeCode)
			dupCount++
		}
	}
	parsed.DuplicateCount = dupCount
}

func (s *tradeServiceImpl) CreateTradesBatch(ctx context.Context, memberID int32, req model.BatchTradeCreateRequest) (*model.BatchTradeCreateResponse, error) {
	res := &model.BatchTradeCreateResponse{
		Trades: make([]model.TradeHistory, 0, len(req.Trades)),
	}

	for _, item := range req.Trades {
		sym := utils.NormalizeBinanceSymbol(item.Symbol)
		side := strings.ToUpper(strings.TrimSpace(item.PositionSide))
		tTime := parseTradeTime(item.TradeTime)

		if req.SkipDuplicates {
			if existing, err := s.tradeRepo.FindPotentialDuplicate(ctx, memberID, sym, side, item.EntryPrice, item.ExitPrice, item.Pnl, tTime); err == nil && existing != nil {
				res.SkippedCount++
				continue
			}
		}

		created, err := s.CreateTrade(ctx, memberID, item)
		if err != nil {
			return nil, err
		}
		res.Trades = append(res.Trades, *created)
		res.CreatedCount++
	}

	return res, nil
}

func (s *tradeServiceImpl) UpdateTrade(ctx context.Context, memberID int32, tradeID int32, req model.TradeUpdateRequest) (*model.TradeHistory, error) {
	trade, err := s.tradeRepo.GetByID(ctx, tradeID, memberID)
	if err != nil {
		return nil, fmt.Errorf("ไม่พบรายการเทรดที่ต้องการแก้ไข")
	}

	if req.Symbol != "" {
		trade.Symbol = utils.NormalizeBinanceSymbol(req.Symbol)
	}
	if req.PositionSide != "" {
		trade.PositionSide = strings.ToUpper(strings.TrimSpace(req.PositionSide))
	}
	if req.Leverage > 0 {
		trade.Leverage = req.Leverage
	}
	trade.EntryPrice = req.EntryPrice
	trade.ExitPrice = req.ExitPrice
	trade.Quantity = req.Quantity
	trade.MarginUsed = req.MarginUsed
	trade.Pnl = req.Pnl
	trade.PnlPercentage = req.PnlPercentage
	trade.Fee = req.Fee
	if req.TradeTime != "" {
		trade.TradeTime = parseTradeTime(req.TradeTime)
	}
	trade.SetupTag = strings.TrimSpace(req.SetupTag)
	trade.EmotionTag = strings.TrimSpace(req.EmotionTag)
	trade.Notes = strings.TrimSpace(req.Notes)

	if err := s.tradeRepo.Update(ctx, trade); err != nil {
		return nil, err
	}

	status := s.sheetSvc.SyncSingleTradeUpdate(ctx, trade)
	trade.SyncSheetStatus = status
	_ = s.tradeRepo.UpdateSyncStatus(ctx, trade.TradeID, status)

	return trade, nil
}

func (s *tradeServiceImpl) DeleteTrade(ctx context.Context, memberID int32, tradeID int32) error {
	trade, err := s.tradeRepo.GetByID(ctx, tradeID, memberID)
	if err != nil {
		return fmt.Errorf("ไม่พบรายการเทรดที่ต้องการลบ")
	}

	_ = s.sheetSvc.SyncSingleTradeDelete(ctx, memberID, trade.TradeCode)
	return s.tradeRepo.Delete(ctx, tradeID, memberID)
}
