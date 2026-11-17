package service

import (
	"backend/model/model"
	"backend/repository"
	"backend/utils"
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"

	"google.golang.org/api/idtoken"
)

type AuthenByGoogleService interface {
	LoginWithGoogle(credential string) (*model.AuthResponse, error)
}

type authenByGoogleServiceImpl struct {
	userRepo      repository.UserRepository
	portfolioRepo repository.PortfolioRepository
	sheetRepo     repository.SheetConfigRepository
}

func NewAuthenByGoogleService(userRepo repository.UserRepository, portfolioRepo repository.PortfolioRepository, sheetRepo repository.SheetConfigRepository) AuthenByGoogleService {
	return &authenByGoogleServiceImpl{
		userRepo:      userRepo,
		portfolioRepo: portfolioRepo,
		sheetRepo:     sheetRepo,
	}
}

func (s *authenByGoogleServiceImpl) LoginWithGoogle(credential string) (*model.AuthResponse, error) {
	ctx := context.Background()
	payload, err := idtoken.Validate(ctx, credential, os.Getenv("GOOGLE_CLIENT_ID"))
	if err != nil {
		return nil, fmt.Errorf("ยืนยันบัญชี Google ไม่สำเร็จ: %v", err)
	}

	googleID, _ := payload.Claims["sub"].(string)
	name, _ := payload.Claims["name"].(string)
	googleEmail, _ := payload.Claims["email"].(string)

	if googleID == "" {
		return nil, fmt.Errorf("ไม่พบ Google ID")
	}

	member, err := s.userRepo.GetUserByGoogleID(googleID)
	if err != nil || member == nil {
		randomPass, _ := generateRandomHex(16)
		hashedPass, _ := utils.HashPassword(randomPass)

		finalEmail := googleEmail
		if finalEmail == "" {
			finalEmail = fmt.Sprintf("google-%s@google.com", googleID)
		} else if existing, _ := s.userRepo.GetUserByEmail(finalEmail); existing != nil {
			finalEmail = fmt.Sprintf("google-%s@google.com", googleID)
		}

		newMember := &model.Member{
			Fullname:      name,
			Email:         finalEmail,
			Password:      hashedPass,
			GoogleID:      googleID,
			IsPasswordSet: false,
		}
		if err := s.userRepo.CreateMember(newMember); err != nil {
			return nil, err
		}
		member = newMember
	}

	ensureMemberDefaultConfigs(ctx, member.MemberID, s.portfolioRepo, s.sheetRepo)

	userIDStr := strconv.Itoa(int(member.MemberID))
	accessToken, refreshToken, err := utils.GenerateTokenPair(userIDStr)
	if err != nil {
		return nil, err
	}

	return &model.AuthResponse{
		Success:      true,
		Token:        accessToken,
		RefreshToken: refreshToken,
		User: model.LoginResponse{
			ID:          int(member.MemberID),
			Fullname:    member.Fullname,
			Email:       member.Email,
			PhoneNumber: member.PhoneNumber,
			Role:        "Member",
			Type:        "Member",
			HasPassword: member.IsPasswordSet,
			GoogleID:    member.GoogleID,
		},
	}, nil
}

func generateRandomHex(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
