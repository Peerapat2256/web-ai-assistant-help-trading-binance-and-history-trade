package worker

import (
	"backend/repository"
	"backend/service"
	"context"
	"log"
	"time"
)

// StartBinanceAutoSyncWorker ตรวจสอบและดึงออเดอร์ที่เพิ่งปิดจาก Binance Futures อัตโนมัติทุก 1 นาที
func StartBinanceAutoSyncWorker() {
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			runBinanceAutoSyncCycle()
		}
	}()
}

func runBinanceAutoSyncCycle() {
	ctx := context.Background()
	binanceRepo := repository.NewBinanceConfigRepository()
	configs, err := binanceRepo.ListAutoSyncEnabled(ctx)
	if err != nil || len(configs) == 0 {
		return
	}

	tradeRepo := repository.NewTradeRepository()
	sheetSvc := service.NewSheetSyncService(repository.NewSheetConfigRepository(), tradeRepo)
	tradeSvc := service.NewTradeService(tradeRepo, sheetSvc)
	syncSvc := service.NewBinanceSyncService(binanceRepo, tradeRepo, tradeSvc)

	for _, cfg := range configs {
		res, err := syncSvc.SyncTradesNow(ctx, cfg.MemberID, 1)
		if err != nil {
			log.Printf("[Binance AutoSync] member=%d error: %v", cfg.MemberID, err)
			continue
		}
		if res.CreatedCount > 0 {
			log.Printf("[Binance AutoSync] member=%d synced %d new closed trades!", cfg.MemberID, res.CreatedCount)
		}
	}
}
