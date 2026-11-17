package controller

import (
	"backend/repository"
	"backend/service"
	"backend/utils"
	"context"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"google.golang.org/api/idtoken"
)

func AuthenController(router *gin.Engine) {
	routes := router.Group("/auth")
	{
		routes.POST("/login", login)
		routes.POST("/google-login", googleLogin)
		routes.POST("/register", register)
		routes.POST("/refresh-token", refreshToken)
		routes.POST("/forgot-password", forgotPassword)
		routes.POST("/verify-reset-token", verifyResetToken)
		routes.POST("/reset-password", resetPassword)
		routes.GET("/profile/me", getProfileMe)
		routes.PUT("/profile/me", updateProfileMe)
		routes.POST("/link-google", linkGoogle)
		routes.POST("/unlink-google", unlinkGoogle)
	}
}

func newAuthService() service.AuthenService {
	return service.NewAuthenService(
		repository.NewUserRepository(),
		repository.NewPortfolioRepository(),
		repository.NewSheetConfigRepository(),
	)
}

func login(c *gin.Context) {
	var loginData struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&loginData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "กรุณากรอกอีเมลและรหัสผ่านให้ครบถ้วน"})
		return
	}

	resp, err := newAuthService().Login(loginData.Email, loginData.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func googleLogin(c *gin.Context) {
	var loginData struct {
		Token string `json:"token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&loginData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "ไม่พบข้อมูลโทเคน Google"})
		return
	}

	svc := service.NewAuthenByGoogleService(
		repository.NewUserRepository(),
		repository.NewPortfolioRepository(),
		repository.NewSheetConfigRepository(),
	)
	resp, err := svc.LoginWithGoogle(loginData.Token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func register(c *gin.Context) {
	var regData struct {
		Fullname    string `json:"fullname" binding:"required"`
		Email       string `json:"email" binding:"required,email"`
		Password    string `json:"password" binding:"required,min=6"`
		PhoneNumber string `json:"phone_number"`
	}
	if err := c.ShouldBindJSON(&regData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "ข้อมูลไม่ครบถ้วน หรือรหัสผ่านสั้นกว่า 6 ตัวอักษร"})
		return
	}

	resp, err := newAuthService().Register(regData.Fullname, regData.Email, regData.Password, regData.PhoneNumber)
	if err != nil {
		status := http.StatusInternalServerError
		if strings.Contains(err.Error(), "ถูกใช้งานแล้ว") {
			status = http.StatusConflict
		}
		c.JSON(status, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func refreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "ไม่พบ Refresh Token"})
		return
	}
	resp, err := newAuthService().RefreshToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func forgotPassword(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "รูปแบบอีเมลไม่ถูกต้อง"})
		return
	}
	if err := newAuthService().ForgotPassword(req.Email); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "ส่งรหัสยืนยันไปยังอีเมลแล้ว"})
}

func verifyResetToken(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
		Token string `json:"token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "ข้อมูลไม่ครบถ้วน"})
		return
	}
	if err := newAuthService().VerifyResetToken(req.Email, req.Token); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "ยืนยันรหัสสำเร็จ"})
}

func resetPassword(c *gin.Context) {
	var req struct {
		Email       string `json:"email" binding:"required,email"`
		Token       string `json:"token" binding:"required"`
		NewPassword string `json:"newPassword" binding:"required,min=6"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "รหัสผ่านใหม่ต้องมีอย่างน้อย 6 ตัวอักษร"})
		return
	}
	if err := newAuthService().ResetPassword(req.Email, req.Token, req.NewPassword); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "เปลี่ยนรหัสผ่านสำเร็จ"})
}

func getProfileMe(c *gin.Context) {
	memberID, ok := getMemberID(c)
	if !ok {
		return
	}
	user, err := repository.NewUserRepository().GetUserByID(memberID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "ไม่พบข้อมูลผู้ใช้"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"member_id":       user.MemberID,
			"fullname":        user.Fullname,
			"email":           user.Email,
			"phone_number":    user.PhoneNumber,
			"google_id":       user.GoogleID,
			"is_password_set": user.IsPasswordSet,
		},
	})
}

func updateProfileMe(c *gin.Context) {
	memberID, ok := getMemberID(c)
	if !ok {
		return
	}
	var req struct {
		Fullname        string `json:"fullname"`
		PhoneNumber     string `json:"phone_number"`
		Email           string `json:"email"`
		Password        string `json:"password"`
		CurrentPassword string `json:"current_password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "ข้อมูลไม่ถูกต้อง"})
		return
	}

	userRepo := repository.NewUserRepository()
	hashedPass := ""
	if req.Password != "" {
		if len(req.Password) < 6 {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "รหัสผ่านใหม่ต้องมีอย่างน้อย 6 ตัวอักษร"})
			return
		}
		currentUser, err := userRepo.GetUserByID(memberID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "ไม่พบข้อมูลผู้ใช้"})
			return
		}
		if currentUser.IsPasswordSet && currentUser.Password != "" {
			if !utils.CheckPasswordHash(req.CurrentPassword, currentUser.Password) {
				c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "รหัสผ่านปัจจุบันไม่ถูกต้อง"})
				return
			}
		}
		hashed, err := utils.HashPassword(req.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "เข้ารหัสรหัสผ่านไม่สำเร็จ"})
			return
		}
		hashedPass = hashed
	}

	if err := userRepo.UpdateMemberProfile(memberID, req.Fullname, req.PhoneNumber, req.Email, hashedPass); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "อัปเดตข้อมูลไม่สำเร็จ: " + err.Error()})
		return
	}

	updated, _ := userRepo.GetUserByID(memberID)
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "อัปเดตโปรไฟล์สำเร็จ", "data": updated})
}

func linkGoogle(c *gin.Context) {
	memberID, ok := getMemberID(c)
	if !ok {
		return
	}
	var req struct {
		Token string `json:"token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "ไม่พบข้อมูลโทเคน Google"})
		return
	}
	payload, err := idtoken.Validate(context.Background(), req.Token, os.Getenv("GOOGLE_CLIENT_ID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "ไม่สามารถยืนยันบัญชี Google ได้"})
		return
	}
	googleID, _ := payload.Claims["sub"].(string)
	if googleID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "ไม่พบ Google ID"})
		return
	}

	userRepo := repository.NewUserRepository()
	if other, err := userRepo.GetUserByGoogleID(googleID); err == nil && other != nil && other.MemberID != memberID {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "บัญชี Google นี้ถูกผูกกับบัญชีอื่นแล้ว"})
		return
	}
	if err := userRepo.UpdateGoogleID(memberID, googleID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "เชื่อมต่อ Google ไม่สำเร็จ"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "เชื่อมต่อบัญชี Google สำเร็จ", "google_id": googleID})
}

func unlinkGoogle(c *gin.Context) {
	memberID, ok := getMemberID(c)
	if !ok {
		return
	}
	userRepo := repository.NewUserRepository()
	user, err := userRepo.GetUserByID(memberID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "ไม่พบข้อมูลผู้ใช้"})
		return
	}
	if !user.IsPasswordSet || user.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "กรุณาตั้งรหัสผ่านก่อนยกเลิกการเชื่อมต่อ Google"})
		return
	}
	if err := userRepo.UpdateGoogleID(memberID, ""); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "ยกเลิกการเชื่อมต่อไม่สำเร็จ"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "ยกเลิกการเชื่อมต่อบัญชี Google สำเร็จ"})
}
