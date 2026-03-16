package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"

	"frp-proxy/internal/model"

	"gorm.io/gorm"
)

type DomainService struct {
	db *gorm.DB
}

func NewDomainService(db *gorm.DB) *DomainService {
	return &DomainService{db: db}
}

func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (s *DomainService) Create(userID uint, subdomain string) (*model.Domain, error) {
	var user model.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return nil, errors.New("user not found")
	}

	var count int64
	s.db.Model(&model.Domain{}).Where("user_id = ?", userID).Count(&count)
	if int(count) >= user.MaxDomains {
		return nil, errors.New("domain quota exceeded")
	}

	var existing model.Domain
	if err := s.db.Where("subdomain = ?", subdomain).First(&existing).Error; err == nil {
		return nil, errors.New("subdomain already taken")
	}

	token, err := generateToken()
	if err != nil {
		return nil, err
	}

	domain := model.Domain{
		UserID:    userID,
		Subdomain: subdomain,
		Token:     token,
		Status:    "active",
	}
	if err := s.db.Create(&domain).Error; err != nil {
		return nil, err
	}
	return &domain, nil
}

func (s *DomainService) ListByUser(userID uint) ([]model.Domain, error) {
	var domains []model.Domain
	err := s.db.Where("user_id = ?", userID).Find(&domains).Error
	return domains, err
}

func (s *DomainService) Delete(domainID, userID uint) error {
	var domain model.Domain
	if err := s.db.First(&domain, domainID).Error; err != nil {
		return errors.New("domain not found")
	}
	if domain.UserID != userID {
		return errors.New("permission denied")
	}
	return s.db.Delete(&domain).Error
}

func (s *DomainService) ListAll() ([]model.Domain, error) {
	var domains []model.Domain
	err := s.db.Preload("User").Find(&domains).Error
	return domains, err
}

func (s *DomainService) AdminCreate(userID uint, subdomain string) (*model.Domain, error) {
	return s.Create(userID, subdomain)
}

func (s *DomainService) AdminUpdate(domainID uint, status string) error {
	return s.db.Model(&model.Domain{}).Where("id = ?", domainID).Update("status", status).Error
}

func (s *DomainService) AdminDelete(domainID uint) error {
	return s.db.Delete(&model.Domain{}, domainID).Error
}

func (s *DomainService) VerifyToken(token string) bool {
	var domain model.Domain
	err := s.db.Where("token = ? AND status = ?", token, "active").First(&domain).Error
	return err == nil
}

func (s *DomainService) VerifyTokenSubdomain(token, subdomain string) bool {
	var domain model.Domain
	err := s.db.Where("token = ? AND subdomain = ? AND status = ?", token, subdomain, "active").First(&domain).Error
	return err == nil
}
