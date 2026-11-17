package services

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"wms/internal/models"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	db        *sql.DB
	jwtSecret []byte
}

func NewAuthService(db *sql.DB, secret string) *AuthService {
	return &AuthService{db: db, jwtSecret: []byte(secret)}
}

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Role     string `json:"role" binding:"required,oneof=seller executor"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (s *AuthService) Register(ctx context.Context, req RegisterRequest) error {
	hash, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO users (email, password_hash, role, is_active)
		 VALUES ($1, $2, $3, true)`,
		req.Email, string(hash), req.Role)
	return err
}

func (s *AuthService) Login(ctx context.Context, req LoginRequest) (string, models.User, error) {
	var u models.User
	var hash string
	err := s.db.QueryRowContext(ctx,
		`SELECT id, email, password_hash, role, is_active FROM users WHERE email = $1`,
		req.Email).Scan(&u.ID, &u.Email, &hash, &u.Role, &u.IsActive)
	if err != nil {
		return "", u, errors.New("неверный email или пароль")
	}
	if !u.IsActive {
		return "", u, errors.New("пользователь заблокирован")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)); err != nil {
		return "", u, errors.New("неверный email или пароль")
	}

	// Токен живёт 30 дней — чтобы не перелогиниваться постоянно
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": u.ID,
		"role":    u.Role,
		"exp":     time.Now().Add(30 * 24 * time.Hour).Unix(),
	})
	str, err := token.SignedString(s.jwtSecret)
	return str, u, err
}

func (s *AuthService) Me(ctx context.Context, userID int) (models.User, error) {
	var u models.User
	err := s.db.QueryRowContext(ctx,
		`SELECT id, email, role, is_active FROM users WHERE id = $1`, userID).
		Scan(&u.ID, &u.Email, &u.Role, &u.IsActive)
	return u, err
}
