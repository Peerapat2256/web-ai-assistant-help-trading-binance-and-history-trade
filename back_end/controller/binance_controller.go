package controller

import (
	"backend/model/model"
	"backend/repository"
	"backend/service"
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func BinanceController(router *gin.Engine) {
	routes := router.Group("/binance")
	{
		routes.GET("/config", getBinanceConfig)
		routes.PUT("/config", saveBinanceConfig)
		routes.DELETE("/config", deleteBinanceConfig)
		routes.POST("/test", testBinanceConnection)
		routes.POST("/sync", syncBinanceTrades)
	}
}

func newBinanceSyncService() service.BinanceSyncService {
	tradeRepo := repository.NewTradeRepository()
	sheetSvc := service.NewSheetSyncService(repository.NewSheetConfigRepository(), tradeRepo)
	tradeSvc := service.NewTradeService(tradeRepo, sheetSvc)
	return service.NewBinanceSyncService(repository.NewBinanceConfigRepository(), tradeRepo, tradeSvc)
}

func getBinanceConfig(c *gin.Context) {
	memberID, ok := getMemberID(c)
	if !ok {
		return
	}
	cfg, err := newBinanceSyncService().GetConfig(context.Background(), memberID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "โหลดการตั้งค่า Binance ไม่สำเร็จ"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": cfg})
}

func saveBinanceConfig(c *gin.Context) {
	memberID, ok := getMemberID(c)
	if !ok {
		return
	}
	var req model.BinanceConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "ข้อมูลการตั้งค่าไม่ถูกต้อง"})
		return
	}
	cfg, err := newBinanceSyncService().SaveConfig(context.Background(), memberID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "บันทึกการเชื่อมต่อ Binance API เรียบร้อยแล้ว", "data": cfg})
}

func deleteBinanceConfig(c *gin.Context) {
	memberID, ok := getMemberID(c)
	if !ok {
		return
	}
	if err := newBinanceSyncService().DeleteConfig(context.Background(), memberID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "ยกเลิกการเชื่อมต่อไม่สำเร็จ"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "ลบ API Key และยกเลิกการเชื่อมต่อ Binance แล้ว"})
}

func testBinanceConnection(c *gin.Context) {
	memberID, ok := getMemberID(c)
	if !ok {
		return
	}
	var req model.BinanceConfigRequest
	_ = c.ShouldBindJSON(&req)

	bal, err := newBinanceSyncService().TestConnection(context.Background(), memberID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": fmt.Sprintf("เชื่อมต่อ Binance Futures สำเร็จ! ยอดกระเป๋า Futures ปัจจุบัน: %.2f USDT", bal),
		"balance": bal,
	})
}

func syncBinanceTrades(c *gin.Context) {
	memberID, ok := getMemberID(c)
	if !ok {
		return
	}
	days, _ := strconv.Atoi(c.DefaultQuery("days", "7"))
	if days <= 0 {
		days = 7
	}

	res, err := newBinanceSyncService().SyncTradesNow(context.Background(), memberID, days)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}

	msg := fmt.Sprintf("ซิงค์สำเร็จ! เพิ่มรายการใหม่ %d ออเดอร์ (ข้ามรายการซ้ำ %d ออเดอร์)", res.CreatedCount, res.SkippedCount)
	if res.CreatedCount == 0 {
		msg = fmt.Sprintf("ข้อมูลเป็นปัจจุบันแล้ว (ตรวจพบ %d ออเดอร์ใน %d วันล่าสุด ทั้งหมดมีในระบบแล้ว)", res.FetchedCount, days)
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": msg, "data": res})
}
