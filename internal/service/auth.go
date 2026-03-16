package service

import (
	"errors"
	"time"

	"frp-proxy/internal/model"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService struct {
	db         *gorm.DB
	jwtSecret  string
	expireHour int
}

func NewAuthService(db *gorm.DB, jwtSecret string, expireHour int) *AuthService {
	return &AuthService{db: db, jwtSecret: jwtSecret, expireHour: expireHour}
}

func (s *AuthService) Register(username, password, inviteCode string) (*model.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := model.User{
		Username:     username,
		PasswordHash: string(hash),
		Role:         "user",
		Status:       "pending",
		MaxDomains:   1,
	}

	if inviteCode != "" {
		var code model.InviteCode
		err := s.db.Where("code = ?", inviteCode).First(&code).Error
		if err != nil {
			return nil, errors.New("invalid invite code")
		}
		if code.UsedCount >= code.MaxUses {
			return nil, errors.New("invite code exhausted")
		}
		if code.ExpiresAt != nil && code.ExpiresAt.Before(time.Now()) {
			return nil, errors.New("invite code expired")
		}
		user.Status = "active"

		if err := s.db.Create(&user).Error; err != nil {
			return nil, errors.New("username already exists")
		}

		s.db.Model(&code).Update("used_count", code.UsedCount+1)
		return &user, nil
	}

	if err := s.db.Create(&user).Error; err != nil {
		return nil, errors.New("username already exists")
	}
	return &user, nil
}

func (s *AuthService) Login(username, password string) (string, error) {
	var user model.User
	if err := s.db.Where("username = ?", username).First(&user).Error; err != nil {
		return "", errors.New("invalid credentials")
	}

	if user.Status != "active" {
		return "", errors.New("account not activated")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", errors.New("invalid credentials")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"role":     user.Role,
		"exp":      time.Now().Add(time.Duration(s.expireHour) * time.Hour).Unix(),
	})

	return token.SignedString([]byte(s.jwtSecret))
}

func (s *AuthService) ParseToken(tokenStr string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		return []byte(s.jwtSecret), nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid claims")
	}
	return claims, nil
}
