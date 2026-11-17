package worker

import (
	"backend/repository"
	"backend/service"
	"context"
	"log"
	"time"
)

func StartSheetRetryWorker() {
	ticker := time.NewTicker(2 * time.Minute)
	go func() {
		for range ticker.C {
			retryPendingSheetSync()
		}
	}()
}

func retryPendingSheetSync() {
	ctx := context.Background()
	tradeRepo := repository.NewTradeRepository()
	sheetRepo := repository.NewSheetConfigRepository()
	sheetSvc := service.NewSheetSyncService(sheetRepo, tradeRepo)

	pendingTrades, err := tradeRepo.FindPendingSync(ctx, 20)
	if err != nil || len(pendingTrades) == 0 {
		return
	}

	log.Printf("[Worker] กำลังซิงค์รายการเทรดที่ค้าง %d รายการ...", len(pendingTrades))
	for i := range pendingTrades {
		status := sheetSvc.SyncSingleTradeUpdate(ctx, &pendingTrades[i])
		_ = tradeRepo.UpdateSyncStatus(ctx, pendingTrades[i].TradeID, status)
	}
}
