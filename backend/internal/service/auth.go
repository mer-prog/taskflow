package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/mer-prog/taskflow/internal/config"
	"github.com/mer-prog/taskflow/internal/model"
)

type RepositoryUser struct {
	ID           uuid.UUID
	Email        string
	PasswordHash string
	DisplayName  string
	AvatarURL    *string
	CreatedAt    time.Time
}

type RepositoryRefreshToken struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	TokenHash string
	ExpiresAt time.Time
}

type RepositoryTenantMember struct {
	TenantID uuid.UUID
	UserID   uuid.UUID
	Role     string
}

type AuthRepository interface {
	CreateUser(ctx context.Context, email, passwordHash, displayName string) (RepositoryUser, error)
	GetUserByEmail(ctx context.Context, email string) (RepositoryUser, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (RepositoryUser, error)
	UserExistsByEmail(ctx context.Context, email string) (bool, error)
	CreateRefreshToken(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) error
	GetRefreshTokenByHash(ctx context.Context, tokenHash string) (RepositoryRefreshToken, error)
	DeleteRefreshToken(ctx context.Context, id uuid.UUID) error
	DeleteRefreshTokensByUser(ctx context.Context, userID uuid.UUID) error
	CreateTenant(ctx context.Context, name, slug string) (uuid.UUID, error)
	CreateTenantMember(ctx context.Context, tenantID, userID uuid.UUID, role string) error
	GetTenantMemberByUserID(ctx context.Context, userID uuid.UUID) (RepositoryTenantMember, error)
}

type AuthService struct {
	repo AuthRepository
	cfg  *config.Config
}

func NewAuthService(repo AuthRepository, cfg *config.Config) *AuthService {
	return &AuthService{repo: repo, cfg: cfg}
}

func (s *AuthService) Register(ctx context.Context, req model.RegisterRequest) (*model.AuthResponse, string, error) {
	exists, err := s.repo.UserExistsByEmail(ctx, req.Email)
	if err != nil {
		return nil, "", fmt.Errorf("service.Register: %w", err)
	}
	if exists {
		return nil, "", ErrEmailAlreadyExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		return nil, "", fmt.Errorf("service.Register: %w", err)
	}

	user, err := s.repo.CreateUser(ctx, req.Email, string(hash), req.DisplayName)
	if err != nil {
		return nil, "", fmt.Errorf("service.Register: %w", err)
	}

	slug := strings.ToLower(strings.ReplaceAll(req.DisplayName, " ", "-")) + "-" + user.ID.String()[:8]
	tenantID, err := s.repo.CreateTenant(ctx, req.DisplayName+"'s Workspace", slug)
	if err != nil {
		return nil, "", fmt.Errorf("service.Register: %w", err)
	}

	if err := s.repo.CreateTenantMember(ctx, tenantID, user.ID, "owner"); err != nil {
		return nil, "", fmt.Errorf("service.Register: %w", err)
	}

	accessToken, err := s.generateAccessToken(user.ID, tenantID)
	if err != nil {
		return nil, "", fmt.Errorf("service.Register: %w", err)
	}

	refreshToken, err := s.createRefreshToken(ctx, user.ID)
	if err != nil {
		return nil, "", fmt.Errorf("service.Register: %w", err)
	}

	return &model.AuthResponse{
		AccessToken: accessToken,
		User:        toUserResponse(user),
	}, refreshToken, nil
}

func (s *AuthService) Login(ctx context.Context, req model.LoginRequest) (*model.AuthResponse, string, error) {
	user, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, "", ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, "", ErrInvalidCredentials
	}

	member, err := s.repo.GetTenantMemberByUserID(ctx, user.ID)
	if err != nil {
		return nil, "", fmt.Errorf("service.Login: %w", err)
	}

	accessToken, err := s.generateAccessToken(user.ID, member.TenantID)
	if err != nil {
		return nil, "", fmt.Errorf("service.Login: %w", err)
	}

	refreshToken, err := s.createRefreshToken(ctx, user.ID)
	if err != nil {
		return nil, "", fmt.Errorf("service.Login: %w", err)
	}

	return &model.AuthResponse{
		AccessToken: accessToken,
		User:        toUserResponse(user),
	}, refreshToken, nil
}

func (s *AuthService) RefreshAccessToken(ctx context.Context, rawToken string) (*model.AuthResponse, string, error) {
	hash := hashToken(rawToken)

	stored, err := s.repo.GetRefreshTokenByHash(ctx, hash)
	if err != nil {
		return nil, "", ErrInvalidRefreshToken
	}

	if time.Now().After(stored.ExpiresAt) {
		_ = s.repo.DeleteRefreshToken(ctx, stored.ID)
		return nil, "", ErrInvalidRefreshToken
	}

	// Rotation: delete old token
	if err := s.repo.DeleteRefreshToken(ctx, stored.ID); err != nil {
		return nil, "", fmt.Errorf("service.RefreshAccessToken: %w", err)
	}

	user, err := s.repo.GetUserByID(ctx, stored.UserID)
	if err != nil {
		return nil, "", fmt.Errorf("service.RefreshAccessToken: %w", err)
	}

	member, err := s.repo.GetTenantMemberByUserID(ctx, user.ID)
	if err != nil {
		return nil, "", fmt.Errorf("service.RefreshAccessToken: %w", err)
	}

	accessToken, err := s.generateAccessToken(user.ID, member.TenantID)
	if err != nil {
		return nil, "", fmt.Errorf("service.RefreshAccessToken: %w", err)
	}

	newRefreshToken, err := s.createRefreshToken(ctx, user.ID)
	if err != nil {
		return nil, "", fmt.Errorf("service.RefreshAccessToken: %w", err)
	}

	return &model.AuthResponse{
		AccessToken: accessToken,
		User:        toUserResponse(user),
	}, newRefreshToken, nil
}

func (s *AuthService) Logout(ctx context.Context, rawToken string) error {
	hash := hashToken(rawToken)

	stored, err := s.repo.GetRefreshTokenByHash(ctx, hash)
	if err != nil {
		return nil
	}

	if err := s.repo.DeleteRefreshToken(ctx, stored.ID); err != nil {
		return fmt.Errorf("service.Logout: %w", err)
	}

	return nil
}

func (s *AuthService) generateAccessToken(userID, tenantID uuid.UUID) (string, error) {
	claims := model.JWTClaims{
		UserID:   userID,
		TenantID: tenantID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.cfg.JWTAccessExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.JWTSecret))
}

func (s *AuthService) createRefreshToken(ctx context.Context, userID uuid.UUID) (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("service.createRefreshToken: %w", err)
	}

	rawToken := hex.EncodeToString(raw)
	hash := hashToken(rawToken)
	expiresAt := time.Now().Add(s.cfg.JWTRefreshExpiry)

	if err := s.repo.CreateRefreshToken(ctx, userID, hash, expiresAt); err != nil {
		return "", fmt.Errorf("service.createRefreshToken: %w", err)
	}

	return rawToken, nil
}

func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

func toUserResponse(u RepositoryUser) model.UserResponse {
	return model.UserResponse{
		ID:          u.ID,
		Email:       u.Email,
		DisplayName: u.DisplayName,
		AvatarURL:   u.AvatarURL,
		CreatedAt:   u.CreatedAt,
	}
}
