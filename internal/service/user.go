package service

import (
	"errors"

	"frp-proxy/internal/model"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserService struct {
	db *gorm.DB
}

func NewUserService(db *gorm.DB) *UserService {
	return &UserService{db: db}
}

func (s *UserService) List() ([]model.User, error) {
	var users []model.User
	err := s.db.Find(&users).Error
	return users, err
}

func (s *UserService) ListByStatus(status string) ([]model.User, error) {
	var users []model.User
	err := s.db.Where("status = ?", status).Find(&users).Error
	return users, err
}

func (s *UserService) GetByID(id uint) (*model.User, error) {
	var user model.User
	err := s.db.Preload("Domains").First(&user, id).Error
	return &user, err
}

func (s *UserService) Create(username, password, role string, maxDomains int) (*model.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	user := model.User{
		Username:     username,
		PasswordHash: string(hash),
		Role:         role,
		Status:       "active",
		MaxDomains:   maxDomains,
	}
	if err := s.db.Create(&user).Error; err != nil {
		return nil, errors.New("username already exists")
	}
	return &user, nil
}

func (s *UserService) Update(id uint, updates map[string]interface{}) error {
	return s.db.Model(&model.User{}).Where("id = ?", id).Updates(updates).Error
}

func (s *UserService) Activate(id uint) error {
	return s.db.Model(&model.User{}).Where("id = ?", id).Update("status", "active").Error
}

func (s *UserService) Delete(id uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", id).Delete(&model.Domain{}).Error; err != nil {
			return err
		}
		return tx.Delete(&model.User{}, id).Error
	})
}
