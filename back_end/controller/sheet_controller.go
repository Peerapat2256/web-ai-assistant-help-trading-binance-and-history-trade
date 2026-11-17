package controller

import (
	"backend/model/model"
	"backend/repository"
	"backend/service"
	"context"
	"fmt"
	"net/http"

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
	c.JSON(http.StatusOK, gin.H{"success": true, "data": cfg})
}

func updateSheetConfig(c *gin.Context) {
	memberID, ok := getMemberID(c)
	if !ok {
		return
	}
	var req model.SheetConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "กรุณาระบุ Spreadsheet ID หรือลิงก์ Google Sheet"})
		return
	}
	cfg, err := newSheetSyncService().SaveConfig(context.Background(), memberID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "บันทึกการตั้งค่า Google Sheets ไม่สำเร็จ"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "บันทึกการตั้งค่า Google Sheets สำเร็จ", "data": cfg})
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
