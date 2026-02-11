package services

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"time"

	"YeahMusic/internal/models"
)

type AuthService struct {
	store *Store
}

func NewAuthService(store *Store) *AuthService {
	return &AuthService{store: store}
}

func hashPassword(pw string) string {
	sum := sha256.Sum256([]byte(pw))
	return hex.EncodeToString(sum[:])
}

func genToken(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func normalizeEmail(s string) string {
	return strings.TrimSpace(strings.ToLower(s))
}

func (a *AuthService) Register(email, password, name, role string) (*models.User, error) {
	email = normalizeEmail(email)
	if email == "" || password == "" {
		return nil, ErrUnauthorized
	}
	u := &models.User{
		Email: email, PasswordHash: hashPassword(password), Name: name, Role: role,
	}
	return a.store.CreateUser(u)
}

func (a *AuthService) Login(email, password string) (*models.Session, *models.User, error) {
	email = normalizeEmail(email)
	u, err := a.store.GetUserByEmail(email)
	if err != nil {
		return nil, nil, ErrUnauthorized
	}
	if u.PasswordHash != hashPassword(password) {
		return nil, nil, ErrUnauthorized
	}
	sec := a.store.CreateSession(u.ID, genToken(16), 24*time.Hour)
	return sec, u, nil
}
