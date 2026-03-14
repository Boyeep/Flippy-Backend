package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/Boyeep/flippy-backend/internal/config"
	"github.com/Boyeep/flippy-backend/internal/domain"
	"github.com/Boyeep/flippy-backend/internal/repository"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidInput       = errors.New("invalid input")
)

type AuthService struct {
	cfg   config.Config
	users repository.UserRepository
}

func NewAuthService(cfg config.Config, users repository.UserRepository) AuthService {
	return AuthService{
		cfg:   cfg,
		users: users,
	}
}

func (s AuthService) Register(ctx context.Context, input domain.RegisterInput) (domain.AuthResponse, error) {
	input.Username = strings.TrimSpace(input.Username)
	input.Email = strings.TrimSpace(strings.ToLower(input.Email))
	input.FullName = strings.TrimSpace(input.FullName)

	if input.Username == "" || input.Email == "" || len(input.Password) < 8 {
		return domain.AuthResponse{}, ErrInvalidInput
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return domain.AuthResponse{}, err
	}

	user, err := s.users.Create(ctx, input, string(passwordHash))
	if err != nil {
		return domain.AuthResponse{}, err
	}

	return s.buildAuthResponse(user)
}

func (s AuthService) Login(ctx context.Context, input domain.LoginInput) (domain.AuthResponse, error) {
	input.Email = strings.TrimSpace(strings.ToLower(input.Email))

	if input.Email == "" || input.Password == "" {
		return domain.AuthResponse{}, ErrInvalidInput
	}

	user, err := s.users.FindByEmail(ctx, input.Email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return domain.AuthResponse{}, ErrInvalidCredentials
		}
		return domain.AuthResponse{}, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		return domain.AuthResponse{}, ErrInvalidCredentials
	}

	if err := s.users.UpdateLastLogin(ctx, user.ID); err != nil {
		return domain.AuthResponse{}, err
	}

	refreshedUser, err := s.users.FindByID(ctx, user.ID)
	if err != nil {
		return domain.AuthResponse{}, err
	}

	return s.buildAuthResponse(refreshedUser)
}

func (s AuthService) Me(ctx context.Context, userID string) (domain.User, error) {
	return s.users.FindByID(ctx, userID)
}

func (s AuthService) ParseAccessToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidCredentials
		}
		return []byte(s.cfg.JWT.AccessSecret), nil
	})
	if err != nil {
		return "", ErrInvalidCredentials
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return "", ErrInvalidCredentials
	}

	subject, ok := claims["sub"].(string)
	if !ok || subject == "" {
		return "", ErrInvalidCredentials
	}

	return subject, nil
}

func (s AuthService) buildAuthResponse(user domain.User) (domain.AuthResponse, error) {
	expiresAt := time.Now().UTC().Add(time.Duration(s.cfg.JWT.AccessTTLMinutes) * time.Minute)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":   user.ID,
		"email": user.Email,
		"role":  user.Role,
		"exp":   expiresAt.Unix(),
		"iat":   time.Now().UTC().Unix(),
	})

	signedToken, err := token.SignedString([]byte(s.cfg.JWT.AccessSecret))
	if err != nil {
		return domain.AuthResponse{}, err
	}

	user.PasswordHash = ""

	return domain.AuthResponse{
		AccessToken: signedToken,
		TokenType:   "Bearer",
		ExpiresAt:   expiresAt,
		User:        user,
	}, nil
}
