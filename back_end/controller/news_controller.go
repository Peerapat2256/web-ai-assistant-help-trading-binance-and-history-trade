package controller

import (
	"backend/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

func NewsController(router *gin.Engine) {
	router.GET("/news", getCryptoNews)
}

func getCryptoNews(c *gin.Context) {
	if _, ok := getMemberID(c); !ok {
		return
	}
	symbol := c.DefaultQuery("symbol", "BTC")
	items, err := service.NewNewsService().GetCryptoNews(symbol)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "ดึงข่าวไม่สำเร็จ"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": items})
}
