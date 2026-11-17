package controller

import (
	"backend/repository"
	"backend/service"
	"context"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func AiController(router *gin.Engine) {
	ai := router.Group("/ai")
	{
		ai.POST("/analyze-chart", analyzeChart)
		ai.POST("/review-journal", reviewJournal)
		ai.GET("/logs", listAiLogs)
		ai.DELETE("/logs/:id", deleteAiLog)
	}
	router.GET("/market/klines", getMarketKlines)
}

func newAiCoachService() service.AiCoachService {
	dashSvc := service.NewDashboardService(repository.NewDashboardRepository(), repository.NewPortfolioRepository())
	return service.NewAiCoachService(repository.NewAiLogRepository(), dashSvc, repository.NewTradeRepository())
}

func getMarketKlines(c *gin.Context) {
	if _, ok := getMemberID(c); !ok {
		return
	}
	symbol := c.DefaultQuery("symbol", "BTCUSDT")
	timeframe := c.DefaultQuery("timeframe", "1h")

	data, err := newAiCoachService().GetMarketData(symbol, timeframe)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": data})
}

func analyzeChart(c *gin.Context) {
	memberID, ok := getMemberID(c)
	if !ok {
		return
	}

	var symbol, timeframe, question string
	teachMode := true
	var fh *multipart.FileHeader

	ct := c.GetHeader("Content-Type")
	if strings.Contains(ct, "multipart/form-data") {
		symbol = c.PostForm("symbol")
		timeframe = c.PostForm("timeframe")
		question = c.PostForm("question")
		teachMode = c.PostForm("teach_mode") != "false"
		fh, _ = c.FormFile("file")
	} else {
		var req struct {
			Symbol    string `json:"symbol"`
			Timeframe string `json:"timeframe"`
			Question  string `json:"question"`
			TeachMode bool   `json:"teach_mode"`
		}
		if err := c.ShouldBindJSON(&req); err == nil {
			symbol = req.Symbol
			timeframe = req.Timeframe
			question = req.Question
			teachMode = req.TeachMode
		}
	}

	if symbol == "" {
		symbol = "BTCUSDT"
	}
	if timeframe == "" {
		timeframe = "1h"
	}

	res, err := newAiCoachService().AnalyzeChart(context.Background(), memberID, symbol, timeframe, question, teachMode, fh)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "วิเคราะห์กราฟไม่สำเร็จ: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": res})
}

func reviewJournal(c *gin.Context) {
	memberID, ok := getMemberID(c)
	if !ok {
		return
	}
	res, err := newAiCoachService().ReviewJournal(context.Background(), memberID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "วิเคราะห์พอร์ตไม่สำเร็จ: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": res})
}

func listAiLogs(c *gin.Context) {
	memberID, ok := getMemberID(c)
	if !ok {
		return
	}
	logs, err := newAiCoachService().ListLogs(context.Background(), memberID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "โหลดประวัติการวิเคราะห์ไม่สำเร็จ"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": logs})
}

func deleteAiLog(c *gin.Context) {
	memberID, ok := getMemberID(c)
	if !ok {
		return
	}
	idInt, err := strconv.Atoi(c.Param("id"))
	if err != nil || idInt <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "รหัสรายการไม่ถูกต้อง"})
		return
	}
	if err := newAiCoachService().DeleteLog(context.Background(), memberID, int32(idInt)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "ลบประวัติการวิเคราะห์สำเร็จ"})
}
