package service

import (
	"backend/model/model"
	"backend/repository"
	"backend/utils"
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"math/rand"
	"net/smtp"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type AuthenService interface {
	Login(email string, password string) (*model.AuthResponse, error)
	RefreshToken(refreshToken string) (*model.AuthResponse, error)
	Register(fullname, email, password, phoneNumber string) (*model.AuthResponse, error)
	ForgotPassword(email string) error
	VerifyResetToken(email, token string) error
	ResetPassword(email, token, newPassword string) error
}

type authenServiceImpl struct {
	userRepo      repository.UserRepository
	portfolioRepo repository.PortfolioRepository
	sheetRepo     repository.SheetConfigRepository
}

func NewAuthenService(userRepo repository.UserRepository, portfolioRepo repository.PortfolioRepository, sheetRepo repository.SheetConfigRepository) AuthenService {
	return &authenServiceImpl{
		userRepo:      userRepo,
		portfolioRepo: portfolioRepo,
		sheetRepo:     sheetRepo,
	}
}

func (s *authenServiceImpl) ensureDefaultConfigs(memberID int32) {
	ctx := context.Background()
	if _, err := s.portfolioRepo.GetByMemberID(ctx, memberID); err != nil {
		_ = s.portfolioRepo.Upsert(ctx, &model.PortfolioSetting{
			MemberID:         memberID,
			InitialBalance:   1000,
			Currency:         "USDT",
			MonthlyTargetPct: 10,
			MaxDailyLossUSD:  50,
		})
	}
	defaultSheetID := strings.TrimSpace(os.Getenv("GOOGLE_SHEET_ID"))
	if defaultSheetID != "" {
		if _, err := s.sheetRepo.GetByMemberID(ctx, memberID); err != nil {
			_ = s.sheetRepo.Upsert(ctx, &model.SheetConfig{
				MemberID:      memberID,
				SpreadsheetID: defaultSheetID,
				SheetName:     "Trades",
				IsAutoSync:    true,
			})
		}
	}
}

func (s *authenServiceImpl) Login(email string, password string) (*model.AuthResponse, error) {
	mem, err := s.userRepo.GetUserByEmail(email)
	if err != nil || mem == nil {
		return nil, fmt.Errorf("ไม่พบบัญชีผู้ใช้นี้")
	}

	if !utils.CheckPasswordHash(password, mem.Password) {
		return nil, fmt.Errorf("อีเมลหรือรหัสผ่านไม่ถูกต้อง")
	}

	s.ensureDefaultConfigs(mem.MemberID)

	userIDStr := strconv.Itoa(int(mem.MemberID))
	accessToken, refreshToken, err := utils.GenerateTokenPair(userIDStr)
	if err != nil {
		return nil, err
	}

	return &model.AuthResponse{
		Success:      true,
		Token:        accessToken,
		RefreshToken: refreshToken,
		User: model.LoginResponse{
			ID:          int(mem.MemberID),
			Fullname:    mem.Fullname,
			Email:       mem.Email,
			PhoneNumber: mem.PhoneNumber,
			Role:        "Member",
			Type:        "Member",
			HasPassword: mem.IsPasswordSet,
			GoogleID:    mem.GoogleID,
		},
	}, nil
}

func (s *authenServiceImpl) RefreshToken(refreshToken string) (*model.AuthResponse, error) {
	refreshSecret := []byte(os.Getenv("JWT_REFRESH_SECRET"))
	if len(refreshSecret) == 0 {
		refreshSecret = []byte("binance_journal_super_secret_refresh_key_2026")
	}
	token, err := utils.ValidateToken(refreshToken, refreshSecret)
	if err != nil || !token.Valid {
		return nil, fmt.Errorf("โทเคนสำหรับต่ออายุไม่ถูกต้อง")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("ข้อมูลโทเคนไม่ถูกต้อง")
	}

	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		return nil, fmt.Errorf("รหัสผู้ใช้ในโทเคนไม่ถูกต้อง")
	}

	idInt, err := strconv.Atoi(userIDStr)
	if err != nil {
		return nil, fmt.Errorf("รหัสผู้ใช้ไม่ถูกต้อง")
	}

	mem, err := s.userRepo.GetUserByID(int32(idInt))
	if err != nil || mem == nil {
		return nil, fmt.Errorf("ไม่พบบัญชีผู้ใช้งาน")
	}

	newAccessToken, newRefreshToken, err := utils.GenerateTokenPair(userIDStr)
	if err != nil {
		return nil, fmt.Errorf("สร้างโทเคนใหม่ไม่สำเร็จ")
	}

	return &model.AuthResponse{
		Success:      true,
		Token:        newAccessToken,
		RefreshToken: newRefreshToken,
		User: model.LoginResponse{
			ID:          int(mem.MemberID),
			Fullname:    mem.Fullname,
			Email:       mem.Email,
			PhoneNumber: mem.PhoneNumber,
			Role:        "Member",
			Type:        "Member",
			HasPassword: mem.IsPasswordSet,
			GoogleID:    mem.GoogleID,
		},
	}, nil
}

func (s *authenServiceImpl) Register(fullname, email, password, phoneNumber string) (*model.AuthResponse, error) {
	existingUser, _ := s.userRepo.GetUserByEmail(email)
	if existingUser != nil {
		return nil, fmt.Errorf("อีเมลนี้ถูกใช้งานแล้ว")
	}

	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("เข้ารหัสรหัสผ่านไม่สำเร็จ: %v", err)
	}

	newMember := &model.Member{
		Fullname:      fullname,
		Email:         email,
		Password:      hashedPassword,
		PhoneNumber:   phoneNumber,
		IsPasswordSet: true,
	}

	if err := s.userRepo.CreateMember(newMember); err != nil {
		return nil, fmt.Errorf("สร้างบัญชีไม่สำเร็จ: %v", err)
	}

	s.ensureDefaultConfigs(newMember.MemberID)

	userIDStr := strconv.Itoa(int(newMember.MemberID))
	accessToken, refreshToken, err := utils.GenerateTokenPair(userIDStr)
	if err != nil {
		return nil, fmt.Errorf("สร้างโทเคนไม่สำเร็จ: %v", err)
	}

	return &model.AuthResponse{
		Success:      true,
		Token:        accessToken,
		RefreshToken: refreshToken,
		User: model.LoginResponse{
			ID:          int(newMember.MemberID),
			Fullname:    newMember.Fullname,
			Email:       newMember.Email,
			PhoneNumber: newMember.PhoneNumber,
			Role:        "Member",
			Type:        "Member",
			HasPassword: true,
			GoogleID:    "",
		},
	}, nil
}

var resetTokenStore sync.Map

type resetTokenData struct {
	Token    string
	ExpireAt time.Time
}

func (s *authenServiceImpl) ForgotPassword(email string) error {
	mem, err := s.userRepo.GetUserByEmail(email)
	if err != nil || mem == nil {
		return fmt.Errorf("ไม่พบอีเมลนี้ในระบบ")
	}

	otp := fmt.Sprintf("%06d", rand.Intn(1000000))
	expireTime := time.Now().Add(15 * time.Minute)
	resetTokenStore.Store(email, resetTokenData{
		Token:    otp,
		ExpireAt: expireTime,
	})

	go sendOTPEmail(email, otp)
	return nil
}

func sendOTPEmail(toEmail, otp string) {
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	senderEmail := os.Getenv("SMTP_EMAIL")
	senderPassword := os.Getenv("SMTP_PASSWORD")

	if senderEmail == "" || senderPassword == "" {
		log.Printf("[OTP] ถึง %s รหัสยืนยัน: %s", toEmail, otp)
		return
	}

	auth := smtp.PlainAuth("", senderEmail, senderPassword, smtpHost)
	msg := []byte("To: " + toEmail + "\r\n" +
		"Subject: รหัสยืนยันการเปลี่ยนรหัสผ่าน\r\n" +
		"MIME-version: 1.0;\r\nContent-Type: text/plain; charset=\"UTF-8\";\r\n\r\n" +
		"รหัส OTP ของคุณคือ: " + otp + "\r\nรหัสนี้จะหมดอายุใน 15 นาที")

	if smtpPort == "465" {
		tlsConfig := &tls.Config{ServerName: smtpHost}
		conn, err := tls.Dial("tcp", smtpHost+":"+smtpPort, tlsConfig)
		if err != nil {
			return
		}
		defer conn.Close()
		client, err := smtp.NewClient(conn, smtpHost)
		if err != nil {
			return
		}
		defer client.Close()
		if err = client.Auth(auth); err != nil {
			return
		}
		_ = client.Mail(senderEmail)
		_ = client.Rcpt(toEmail)
		w, err := client.Data()
		if err != nil {
			return
		}
		_, _ = w.Write(msg)
		_ = w.Close()
		_ = client.Quit()
	} else {
		_ = smtp.SendMail(smtpHost+":"+smtpPort, auth, senderEmail, []string{toEmail}, msg)
	}
}

func (s *authenServiceImpl) VerifyResetToken(email, token string) error {
	data, ok := resetTokenStore.Load(email)
	if !ok {
		return fmt.Errorf("รหัสยืนยันไม่ถูกต้องหรือยังไม่ได้ขอรหัส")
	}
	tokenData := data.(resetTokenData)
	if time.Now().After(tokenData.ExpireAt) {
		resetTokenStore.Delete(email)
		return fmt.Errorf("รหัสยืนยันหมดอายุแล้ว กรุณาขอรหัสใหม่")
	}
	if tokenData.Token != token {
		return fmt.Errorf("รหัสยืนยันไม่ถูกต้อง")
	}
	return nil
}

func (s *authenServiceImpl) ResetPassword(email, token, newPassword string) error {
	if err := s.VerifyResetToken(email, token); err != nil {
		return err
	}
	hashed, err := utils.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("ตั้งรหัสผ่านใหม่ไม่สำเร็จ")
	}
	if err := s.userRepo.UpdatePassword(email, hashed); err != nil {
		return err
	}
	resetTokenStore.Delete(email)
	return nil
}
