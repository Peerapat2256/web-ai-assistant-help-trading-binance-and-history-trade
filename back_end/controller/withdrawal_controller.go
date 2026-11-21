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

func WithdrawalController(router *gin.Engine) {
	routes := router.Group("/withdrawals")
	{
		routes.POST("/parse-screenshot", parseWithdrawalScreenshot)
		routes.GET("", listWithdrawals)
		routes.POST("", createWithdrawal)
		routes.PUT("/:id", updateWithdrawal)
		routes.DELETE("/:id", deleteWithdrawal)
	}
}

func newWithdrawalService() service.WithdrawalService {
	return service.NewWithdrawalService(repository.NewWithdrawalRepository())
}

func parseWithdrawalScreenshot(c *gin.Context) {
	memberID, ok := getMemberID(c)
	if !ok {
		return
	}

	fh, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "กรุณาอัปโหลดไฟล์รูปภาพสลิปการถอน"})
		return
	}

	uploadDir := filepath.Join("uploads", "withdrawals")
	_ = os.MkdirAll(uploadDir, os.ModePerm)
	ext := filepath.Ext(fh.Filename)
	if ext == "" {
		ext = ".png"
	}
	filename := fmt.Sprintf("withdraw_%d%s", time.Now().UnixNano(), ext)
	savePath := filepath.Join(uploadDir, filename)
	if err := c.SaveUploadedFile(fh, savePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "บันทึกรูปภาพไม่สำเร็จ"})
		return
	}

	publicURL := "/uploads/withdrawals/" + filename
	parsed, err := utils.ParseWithdrawalScreenshot(fh, publicURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "วิเคราะห์รูปภาพไม่สำเร็จ: " + err.Error()})
		return
	}

	newWithdrawalService().AnnotateWithdrawalDuplicate(context.Background(), memberID, parsed)
	c.JSON(http.StatusOK, gin.H{"success": true, "data": parsed})
}

func listWithdrawals(c *gin.Context) {
	memberID, ok := getMemberID(c)
	if !ok {
		return
	}
	year, _ := strconv.Atoi(c.Query("year"))
	items, err := newWithdrawalService().ListWithdrawals(context.Background(), memberID, year)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "โหลดประวัติการถอนไม่สำเร็จ"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": items})
}

func createWithdrawal(c *gin.Context) {
	memberID, ok := getMemberID(c)
	if !ok {
		return
	}
	var req model.WithdrawalCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "กรุณากรอกจำนวนเงินและข้อมูลการถอนให้ครบถ้วน"})
		return
	}

	item, err := newWithdrawalService().CreateWithdrawal(context.Background(), memberID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "message": "บันทึกประวัติการถอนสำเร็จ", "data": item})
}

func updateWithdrawal(c *gin.Context) {
	memberID, ok := getMemberID(c)
	if !ok {
		return
	}
	idInt, err := strconv.Atoi(c.Param("id"))
	if err != nil || idInt <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "รหัสรายการถอนไม่ถูกต้อง"})
		return
	}

	var req model.WithdrawalUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "ข้อมูลแก้ไขไม่ถูกต้อง"})
		return
	}

	item, err := newWithdrawalService().UpdateWithdrawal(context.Background(), memberID, int32(idInt), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "อัปเดตประวัติการถอนสำเร็จ", "data": item})
}

func deleteWithdrawal(c *gin.Context) {
	memberID, ok := getMemberID(c)
	if !ok {
		return
	}
	idInt, err := strconv.Atoi(c.Param("id"))
	if err != nil || idInt <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "รหัสรายการถอนไม่ถูกต้อง"})
		return
	}

	if err := newWithdrawalService().DeleteWithdrawal(context.Background(), memberID, int32(idInt)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "ลบรายการถอนเรียบร้อยแล้ว"})
}
