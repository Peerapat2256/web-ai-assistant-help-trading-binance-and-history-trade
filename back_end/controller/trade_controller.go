package controller

import (
	"backend/model/model"
	"backend/repository"
	"backend/service"
	"backend/utils"
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func TradeController(router *gin.Engine) {
	routes := router.Group("/trades")
	{
		routes.POST("/parse-screenshot", parseScreenshot)
		routes.GET("", listTrades)
		routes.POST("", createTrade)
		routes.PUT("/:id", updateTrade)
		routes.DELETE("/:id", deleteTrade)
	}
}

func newTradeService() service.TradeService {
	tradeRepo := repository.NewTradeRepository()
	sheetSvc := service.NewSheetSyncService(repository.NewSheetConfigRepository(), tradeRepo)
	return service.NewTradeService(tradeRepo, sheetSvc)
}

func parseScreenshot(c *gin.Context) {
	if _, ok := getMemberID(c); !ok {
		return
	}

	fh, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "กรุณาอัปโหลดไฟล์รูปภาพ"})
		return
	}

	uploadDir := filepath.Join("uploads", "screenshots")
	_ = os.MkdirAll(uploadDir, os.ModePerm)
	ext := filepath.Ext(fh.Filename)
	if ext == "" {
		ext = ".png"
	}
	filename := fmt.Sprintf("trade_%d%s", time.Now().UnixNano(), ext)
	savePath := filepath.Join(uploadDir, filename)
	if err := c.SaveUploadedFile(fh, savePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "บันทึกรูปภาพไม่สำเร็จ"})
		return
	}

	publicURL := "/uploads/screenshots/" + filename
	parsed, err := utils.ParseBinanceScreenshot(fh, publicURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "วิเคราะห์รูปภาพไม่สำเร็จ: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": parsed})
}

func listTrades(c *gin.Context) {
	memberID, ok := getMemberID(c)
	if !ok {
		return
	}
	symbol := c.Query("symbol")
	side := c.Query("side")

	trades, err := newTradeService().ListTrades(context.Background(), memberID, symbol, side)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "โหลดประวัติการเทรดไม่สำเร็จ"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": trades})
}

func createTrade(c *gin.Context) {
	memberID, ok := getMemberID(c)
	if !ok {
		return
	}
	var req model.TradeCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "กรุณากรอกข้อมูลการเทรดให้ครบถ้วน"})
		return
	}

	trade, err := newTradeService().CreateTrade(context.Background(), memberID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "บันทึกรายการเทรดไม่สำเร็จ: " + err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "message": "บันทึกรายการเทรดสำเร็จ", "data": trade})
}

func updateTrade(c *gin.Context) {
	memberID, ok := getMemberID(c)
	if !ok {
		return
	}
	idInt, err := strconv.Atoi(c.Param("id"))
	if err != nil || idInt <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "รหัสรายการเทรดไม่ถูกต้อง"})
		return
	}

	var req model.TradeUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "ข้อมูลแก้ไขไม่ถูกต้อง"})
		return
	}

	trade, err := newTradeService().UpdateTrade(context.Background(), memberID, int32(idInt), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "อัปเดตรายการเทรดสำเร็จ", "data": trade})
}

func deleteTrade(c *gin.Context) {
	memberID, ok := getMemberID(c)
	if !ok {
		return
	}
	idInt, err := strconv.Atoi(c.Param("id"))
	if err != nil || idInt <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "รหัสรายการเทรดไม่ถูกต้อง"})
		return
	}

	if err := newTradeService().DeleteTrade(context.Background(), memberID, int32(idInt)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "ลบรายการเทรดและซิงค์ข้อมูลเรียบร้อยแล้ว"})
}
