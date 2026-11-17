package repository

import (
	"backend/model/model"
	"backend/model/query"
	"context"
)

type UserRepository interface {
	GetUserByEmail(email string) (*model.Member, error)
	GetUserByID(memberID int32) (*model.Member, error)
	GetUserByGoogleID(googleID string) (*model.Member, error)
	CreateMember(member *model.Member) error
	UpdateGoogleID(memberID int32, googleID string) error
	UpdatePassword(email string, hashedPassword string) error
	UpdateMemberProfile(memberID int32, fullname, phone, email, password string) error
}

type userRepository struct{}

func NewUserRepository() UserRepository {
	return &userRepository{}
}

func (r *userRepository) GetUserByEmail(email string) (*model.Member, error) {
	var user model.Member
	err := query.DB().WithContext(context.Background()).Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetUserByID(memberID int32) (*model.Member, error) {
	var user model.Member
	err := query.DB().WithContext(context.Background()).Where("member_id = ?", memberID).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetUserByGoogleID(googleID string) (*model.Member, error) {
	var user model.Member
	err := query.DB().WithContext(context.Background()).Where("google_id = ?", googleID).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) CreateMember(member *model.Member) error {
	return query.DB().WithContext(context.Background()).Create(member).Error
}

func (r *userRepository) UpdateGoogleID(memberID int32, googleID string) error {
	return query.DB().WithContext(context.Background()).Model(&model.Member{}).
		Where("member_id = ?", memberID).
		Update("google_id", googleID).Error
}

func (r *userRepository) UpdatePassword(email string, hashedPassword string) error {
	return query.DB().WithContext(context.Background()).Model(&model.Member{}).
		Where("email = ?", email).
		Updates(map[string]interface{}{
			"password":        hashedPassword,
			"is_password_set": true,
		}).Error
}

func (r *userRepository) UpdateMemberProfile(memberID int32, fullname, phone, email, password string) error {
	updates := map[string]interface{}{}
	if fullname != "" {
		updates["fullname"] = fullname
	}
	if phone != "" {
		updates["phone_number"] = phone
	}
	if email != "" {
		updates["email"] = email
	}
	if password != "" {
		updates["password"] = password
		updates["is_password_set"] = true
	}
	if len(updates) == 0 {
		return nil
	}
	return query.DB().WithContext(context.Background()).Model(&model.Member{}).
		Where("member_id = ?", memberID).
		Updates(updates).Error
}
