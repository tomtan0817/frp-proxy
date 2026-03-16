package service

import (
	"testing"

	"frp-proxy/internal/model"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func createTestUser(db *gorm.DB, username, status string, maxDomains int) *model.User {
	user := model.User{
		Username:     username,
		PasswordHash: "hash",
		Role:         "user",
		Status:       status,
		MaxDomains:   maxDomains,
	}
	db.Create(&user)
	return &user
}

func TestCreateDomain(t *testing.T) {
	db := setupTestDB(t)
	svc := NewDomainService(db)
	user := createTestUser(db, "testuser", "active", 2)

	domain, err := svc.Create(user.ID, "myapp")
	assert.NoError(t, err)
	assert.Equal(t, "myapp", domain.Subdomain)
	assert.NotEmpty(t, domain.Token)
	assert.Equal(t, "active", domain.Status)
}

func TestCreateDomainExceedQuota(t *testing.T) {
	db := setupTestDB(t)
	svc := NewDomainService(db)
	user := createTestUser(db, "testuser", "active", 1)

	_, err := svc.Create(user.ID, "first")
	assert.NoError(t, err)

	_, err = svc.Create(user.ID, "second")
	assert.Error(t, err)
}

func TestCreateDomainDuplicate(t *testing.T) {
	db := setupTestDB(t)
	svc := NewDomainService(db)
	user1 := createTestUser(db, "user1", "active", 2)
	user2 := createTestUser(db, "user2", "active", 2)

	_, err := svc.Create(user1.ID, "myapp")
	assert.NoError(t, err)

	_, err = svc.Create(user2.ID, "myapp")
	assert.Error(t, err)
}

func TestListDomainsByUser(t *testing.T) {
	db := setupTestDB(t)
	svc := NewDomainService(db)
	user := createTestUser(db, "testuser", "active", 3)

	svc.Create(user.ID, "app1")
	svc.Create(user.ID, "app2")

	domains, err := svc.ListByUser(user.ID)
	assert.NoError(t, err)
	assert.Len(t, domains, 2)
}

func TestDeleteDomain(t *testing.T) {
	db := setupTestDB(t)
	svc := NewDomainService(db)
	user := createTestUser(db, "testuser", "active", 2)

	domain, _ := svc.Create(user.ID, "myapp")
	err := svc.Delete(domain.ID, user.ID)
	assert.NoError(t, err)

	domains, _ := svc.ListByUser(user.ID)
	assert.Len(t, domains, 0)
}

func TestDeleteDomainWrongUser(t *testing.T) {
	db := setupTestDB(t)
	svc := NewDomainService(db)
	user1 := createTestUser(db, "user1", "active", 2)
	user2 := createTestUser(db, "user2", "active", 2)

	domain, _ := svc.Create(user1.ID, "myapp")
	err := svc.Delete(domain.ID, user2.ID)
	assert.Error(t, err)
}

func TestVerifyToken(t *testing.T) {
	db := setupTestDB(t)
	svc := NewDomainService(db)
	user := createTestUser(db, "testuser", "active", 2)

	domain, _ := svc.Create(user.ID, "myapp")
	assert.True(t, svc.VerifyToken(domain.Token))
	assert.False(t, svc.VerifyToken("invalid-token"))
}

func TestVerifyTokenSubdomain(t *testing.T) {
	db := setupTestDB(t)
	svc := NewDomainService(db)
	user := createTestUser(db, "testuser", "active", 2)

	domain, _ := svc.Create(user.ID, "myapp")
	assert.True(t, svc.VerifyTokenSubdomain(domain.Token, "myapp"))
	assert.False(t, svc.VerifyTokenSubdomain(domain.Token, "wrong"))
	assert.False(t, svc.VerifyTokenSubdomain("invalid", "myapp"))
}
