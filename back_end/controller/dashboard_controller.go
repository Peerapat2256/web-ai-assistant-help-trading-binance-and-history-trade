package controller

import (
	"backend/repository"
	"backend/service"
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func DashboardController(router *gin.Engine) {
	routes := router.Group("/dashboard")
	{
		routes.GET("/summary", getDashboardSummary)
		routes.GET("/calendar", getDashboardCalendar)
	}
}

func getDashboardSummary(c *gin.Context) {
	memberID, ok := getMemberID(c)
	if !ok {
		return
	}
	svc := service.NewDashboardService(repository.NewDashboardRepository(), repository.NewPortfolioRepository())
	summary, err := svc.GetSummary(context.Background(), memberID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "โหลดข้อมูลแดชบอร์ดไม่สำเร็จ"})
		return
	}
	c.JSON(http.StatusOK, summary)
}

func getDashboardCalendar(c *gin.Context) {
	memberID, ok := getMemberID(c)
	if !ok {
		return
	}
	now := time.Now()
	year, err := strconv.Atoi(c.Query("year"))
	if err != nil || year < 2000 {
		year = now.Year()
	}
	month, err := strconv.Atoi(c.Query("month"))
	if err != nil || month < 1 || month > 12 {
		month = int(now.Month())
	}

	svc := service.NewDashboardService(repository.NewDashboardRepository(), repository.NewPortfolioRepository())
	days, err := svc.GetCalendar(context.Background(), memberID, year, month)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "โหลดปฏิทินกำไรขาดทุนไม่สำเร็จ"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "year": year, "month": month, "data": days})
}
