package service

import (
	"testing"

	"frp-proxy/internal/model"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)
	err = db.AutoMigrate(&model.User{}, &model.Domain{}, &model.InviteCode{})
	assert.NoError(t, err)
	return db
}

func TestRegisterUser(t *testing.T) {
	db := setupTestDB(t)
	svc := NewAuthService(db, "test-secret", 72)

	user, err := svc.Register("testuser", "password123", "")
	assert.NoError(t, err)
	assert.Equal(t, "testuser", user.Username)
	assert.Equal(t, "pending", user.Status)
}

func TestRegisterWithInviteCode(t *testing.T) {
	db := setupTestDB(t)
	svc := NewAuthService(db, "test-secret", 72)

	db.Create(&model.InviteCode{Code: "INVITE1", MaxUses: 1, UsedCount: 0, CreatedBy: 1})

	user, err := svc.Register("testuser", "password123", "INVITE1")
	assert.NoError(t, err)
	assert.Equal(t, "active", user.Status)

	var code model.InviteCode
	db.Where("code = ?", "INVITE1").First(&code)
	assert.Equal(t, 1, code.UsedCount)
}

func TestRegisterDuplicateUsername(t *testing.T) {
	db := setupTestDB(t)
	svc := NewAuthService(db, "test-secret", 72)

	_, err := svc.Register("testuser", "password123", "")
	assert.NoError(t, err)

	_, err = svc.Register("testuser", "password456", "")
	assert.Error(t, err)
}

func TestLogin(t *testing.T) {
	db := setupTestDB(t)
	svc := NewAuthService(db, "test-secret", 72)

	_, err := svc.Register("testuser", "password123", "")
	assert.NoError(t, err)

	db.Model(&model.User{}).Where("username = ?", "testuser").Update("status", "active")

	token, err := svc.Login("testuser", "password123")
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestLoginPendingUser(t *testing.T) {
	db := setupTestDB(t)
	svc := NewAuthService(db, "test-secret", 72)

	_, err := svc.Register("testuser", "password123", "")
	assert.NoError(t, err)

	_, err = svc.Login("testuser", "password123")
	assert.Error(t, err)
}

func TestLoginWrongPassword(t *testing.T) {
	db := setupTestDB(t)
	svc := NewAuthService(db, "test-secret", 72)

	_, err := svc.Register("testuser", "password123", "")
	assert.NoError(t, err)
	db.Model(&model.User{}).Where("username = ?", "testuser").Update("status", "active")

	_, err = svc.Login("testuser", "wrongpass")
	assert.Error(t, err)
}
