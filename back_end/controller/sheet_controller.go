package controller

import (
	"backend/model/model"
	"backend/repository"
	"backend/service"
	"backend/utils"
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func SheetController(router *gin.Engine) {
	routes := router.Group("/sheets")
	{
		routes.GET("/config", getSheetConfig)
		routes.PUT("/config", updateSheetConfig)
		routes.POST("/test", testSheetConnection)
		routes.POST("/resync", resyncAllTradesToSheet)
	}
}

func newSheetSyncService() service.SheetSyncService {
	return service.NewSheetSyncService(repository.NewSheetConfigRepository(), repository.NewTradeRepository())
}

func toSheetConfigResponse(cfg *model.SheetConfig) model.SheetConfigResponse {
	sheetURL := ""
	if cfg.SpreadsheetID != "" {
		sheetURL = fmt.Sprintf("https://docs.google.com/spreadsheets/d/%s/edit", cfg.SpreadsheetID)
	}
	return model.SheetConfigResponse{
		ID:            cfg.ID,
		MemberID:      cfg.MemberID,
		SpreadsheetID: cfg.SpreadsheetID,
		SheetName:     cfg.SheetName,
		IsAutoSync:    cfg.IsAutoSync,
		LastSyncedAt:  cfg.LastSyncedAt,
		BotEmail:      utils.GetServiceAccountEmail(),
		SheetURL:      sheetURL,
	}
}

func getSheetConfig(c *gin.Context) {
	memberID, ok := getMemberID(c)
	if !ok {
		return
	}
	cfg, err := newSheetSyncService().GetConfig(context.Background(), memberID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "โหลดการตั้งค่า Google Sheets ไม่สำเร็จ"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": toSheetConfigResponse(cfg)})
}

func updateSheetConfig(c *gin.Context) {
	memberID, ok := getMemberID(c)
	if !ok {
		return
	}
	var req model.SheetConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "ข้อมูลการตั้งค่า Google Sheets ไม่ถูกต้อง"})
		return
	}
	if req.IsAutoSync && strings.TrimSpace(req.SpreadsheetID) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "กรุณาระบุ Spreadsheet ID หรือลิงก์ Google Sheet ก่อนเปิดซิงค์อัตโนมัติ"})
		return
	}
	cfg, err := newSheetSyncService().SaveConfig(context.Background(), memberID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "บันทึกการตั้งค่า Google Sheets ไม่สำเร็จ"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "บันทึกการตั้งค่า Google Sheets สำเร็จ", "data": toSheetConfigResponse(cfg)})
}

func testSheetConnection(c *gin.Context) {
	memberID, ok := getMemberID(c)
	if !ok {
		return
	}
	var req model.SheetConfigRequest
	_ = c.ShouldBindJSON(&req)

	title, err := newSheetSyncService().TestConnection(context.Background(), memberID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": fmt.Sprintf("เชื่อมต่อไฟล์ '%s' และเตรียมหัวตารางสำเร็จ!", title),
		"title":   title,
	})
}

func resyncAllTradesToSheet(c *gin.Context) {
	memberID, ok := getMemberID(c)
	if !ok {
		return
	}
	count, err := newSheetSyncService().FullResyncAll(context.Background(), memberID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "ซิงค์ข้อมูลไม่สำเร็จ: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": fmt.Sprintf("ซิงค์รายการเทรดทั้งหมด %d รายการลง Google Sheets สำเร็จ", count),
		"count":   count,
	})
}
