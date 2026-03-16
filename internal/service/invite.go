package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"frp-proxy/internal/model"

	"gorm.io/gorm"
)

type InviteService struct {
	db *gorm.DB
}

func NewInviteService(db *gorm.DB) *InviteService {
	return &InviteService{db: db}
}

func (s *InviteService) Create(createdBy uint, maxUses int, expiresAt *time.Time) (*model.InviteCode, error) {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}
	code := model.InviteCode{
		Code:      hex.EncodeToString(b),
		MaxUses:   maxUses,
		CreatedBy: createdBy,
		ExpiresAt: expiresAt,
	}
	if err := s.db.Create(&code).Error; err != nil {
		return nil, err
	}
	return &code, nil
}

func (s *InviteService) List() ([]model.InviteCode, error) {
	var codes []model.InviteCode
	err := s.db.Preload("Creator").Find(&codes).Error
	return codes, err
}

func (s *InviteService) Delete(id uint) error {
	result := s.db.Delete(&model.InviteCode{}, id)
	if result.RowsAffected == 0 {
		return errors.New("invite code not found")
	}
	return result.Error
}
