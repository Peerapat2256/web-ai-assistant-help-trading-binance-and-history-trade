package service

import (
	"backend/model/model"
	"backend/repository"
	"context"
	"fmt"
	"math"
	"strings"
)

type WithdrawalService interface {
	ListWithdrawals(ctx context.Context, memberID int32, year int) ([]model.WithdrawalHistory, error)
	CreateWithdrawal(ctx context.Context, memberID int32, req model.WithdrawalCreateRequest) (*model.WithdrawalHistory, error)
	UpdateWithdrawal(ctx context.Context, memberID int32, withdrawalID int32, req model.WithdrawalUpdateRequest) (*model.WithdrawalHistory, error)
	DeleteWithdrawal(ctx context.Context, memberID int32, withdrawalID int32) error
	AnnotateWithdrawalDuplicate(ctx context.Context, memberID int32, parsed *model.ParsedWithdrawalData)
}

type withdrawalServiceImpl struct {
	repo repository.WithdrawalRepository
}

func NewWithdrawalService(repo repository.WithdrawalRepository) WithdrawalService {
	return &withdrawalServiceImpl{repo: repo}
}

func (s *withdrawalServiceImpl) ListWithdrawals(ctx context.Context, memberID int32, year int) ([]model.WithdrawalHistory, error) {
	return s.repo.ListByMember(ctx, memberID, year)
}

func (s *withdrawalServiceImpl) AnnotateWithdrawalDuplicate(ctx context.Context, memberID int32, parsed *model.ParsedWithdrawalData) {
	if parsed == nil {
		return
	}
	txID := strings.TrimSpace(parsed.TransactionID)
	if txID != "" {
		if existing, err := s.repo.FindByTransactionID(ctx, memberID, txID); err == nil && existing != nil {
			parsed.IsDuplicate = true
			parsed.DuplicateReason = fmt.Sprintf("หมายเลขธุรกรรม %s ถูกบันทึกไว้ในระบบแล้ว", txID)
			return
		}
	}
	wTime := parseTradeTime(parsed.WithdrawalTime)
	if existing, err := s.repo.FindPotentialDuplicate(ctx, memberID, math.Abs(parsed.Amount), wTime); err == nil && existing != nil {
		parsed.IsDuplicate = true
		parsed.DuplicateReason = "พบรายการถอนยอดเงินและเวลาเดียวกันในระบบแล้ว"
	}
}

func (s *withdrawalServiceImpl) CreateWithdrawal(ctx context.Context, memberID int32, req model.WithdrawalCreateRequest) (*model.WithdrawalHistory, error) {
	txID := strings.TrimSpace(req.TransactionID)
	if txID != "" {
		if existing, err := s.repo.FindByTransactionID(ctx, memberID, txID); err == nil && existing != nil {
			return nil, fmt.Errorf("หมายเลขธุรกรรม %s มีอยู่ในประวัติการถอนแล้ว", txID)
		}
	}

	currency := strings.ToUpper(strings.TrimSpace(req.Currency))
	if currency == "" {
		currency = "THB"
	}
	status := strings.TrimSpace(req.Status)
	if status == "" {
		status = "สำเร็จ"
	}
	method := strings.TrimSpace(req.PaymentMethod)
	if method == "" {
		method = "BANK_TRANSFER"
	}

	w := &model.WithdrawalHistory{
		MemberID:       memberID,
		TransactionID:  txID,
		Amount:         math.Abs(req.Amount),
		Currency:       currency,
		Status:         status,
		PaymentMethod:  method,
		WithdrawalTime: parseTradeTime(req.WithdrawalTime),
		Notes:          strings.TrimSpace(req.Notes),
		ScreenshotURL:  strings.TrimSpace(req.ScreenshotURL),
	}

	if err := s.repo.Create(ctx, w); err != nil {
		return nil, err
	}
	return w, nil
}

func (s *withdrawalServiceImpl) UpdateWithdrawal(ctx context.Context, memberID int32, withdrawalID int32, req model.WithdrawalUpdateRequest) (*model.WithdrawalHistory, error) {
	w, err := s.repo.GetByID(ctx, withdrawalID, memberID)
	if err != nil {
		return nil, fmt.Errorf("ไม่พบรายการถอนที่ต้องการแก้ไข")
	}

	if req.TransactionID != "" {
		w.TransactionID = strings.TrimSpace(req.TransactionID)
	}
	if req.Amount > 0 {
		w.Amount = math.Abs(req.Amount)
	}
	if req.Currency != "" {
		w.Currency = strings.ToUpper(strings.TrimSpace(req.Currency))
	}
	if req.Status != "" {
		w.Status = strings.TrimSpace(req.Status)
	}
	if req.PaymentMethod != "" {
		w.PaymentMethod = strings.TrimSpace(req.PaymentMethod)
	}
	if req.WithdrawalTime != "" {
		w.WithdrawalTime = parseTradeTime(req.WithdrawalTime)
	}
	w.Notes = strings.TrimSpace(req.Notes)

	if err := s.repo.Update(ctx, w); err != nil {
		return nil, err
	}
	return w, nil
}

func (s *withdrawalServiceImpl) DeleteWithdrawal(ctx context.Context, memberID int32, withdrawalID int32) error {
	if _, err := s.repo.GetByID(ctx, withdrawalID, memberID); err != nil {
		return fmt.Errorf("ไม่พบรายการถอนที่ต้องการลบ")
	}
	return s.repo.Delete(ctx, withdrawalID, memberID)
}
