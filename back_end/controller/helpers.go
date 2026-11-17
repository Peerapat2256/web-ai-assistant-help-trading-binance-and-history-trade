package controller

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func getMemberID(c *gin.Context) (int32, bool) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "กรุณาเข้าสู่ระบบ"})
		return 0, false
	}
	idInt, err := strconv.Atoi(fmt.Sprintf("%v", userIDStr))
	if err != nil || idInt <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "รหัสผู้ใช้ไม่ถูกต้อง"})
		return 0, false
	}
	return int32(idInt), true
}
