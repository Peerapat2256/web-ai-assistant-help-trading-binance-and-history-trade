package controller

import (
	"backend/model/model"
	"backend/repository"
	"backend/service"
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

func PortfolioController(router *gin.Engine) {
	routes := router.Group("/portfolio")
	{
		routes.GET("/settings", getPortfolioSettings)
		routes.PUT("/settings", updatePortfolioSettings)
	}
}

func getPortfolioSettings(c *gin.Context) {
	memberID, ok := getMemberID(c)
	if !ok {
		return
	}
	svc := service.NewPortfolioService(repository.NewPortfolioRepository())
	data, err := svc.GetSettings(context.Background(), memberID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "โหลดการตั้งค่าพอร์ตไม่สำเร็จ"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": data})
}

func updatePortfolioSettings(c *gin.Context) {
	memberID, ok := getMemberID(c)
	if !ok {
		return
	}
	var req model.PortfolioUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "ข้อมูลการตั้งค่าพอร์ตไม่ถูกต้อง"})
		return
	}
	svc := service.NewPortfolioService(repository.NewPortfolioRepository())
	data, err := svc.UpdateSettings(context.Background(), memberID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "บันทึกการตั้งค่าพอร์ตไม่สำเร็จ"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "บันทึกการตั้งค่าพอร์ตสำเร็จ", "data": data})
}
