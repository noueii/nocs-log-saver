package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/noueii/nocs-log-saver/internal/domain/entities"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotActive      = errors.New("user account is not active")
	ErrTokenExpired       = errors.New("token has expired")
	ErrInvalidToken       = errors.New("invalid token")
)

// AuthService handles authentication and authorization
type AuthService struct {
	userRepo     UserRepository
	sessionRepo  SessionRepository
	jwtSecret    []byte
	accessTTL    time.Duration
	refreshTTL   time.Duration
}

// UserRepository interface for user operations
type UserRepository interface {
	Create(ctx context.Context, user *entities.User) error
	FindByID(ctx context.Context, id uuid.UUID) (*entities.User, error)
	FindByEmail(ctx context.Context, email string) (*entities.User, error)
	FindByUsername(ctx context.Context, username string) (*entities.User, error)
	Update(ctx context.Context, user *entities.User) error
	UpdateLastLogin(ctx context.Context, userID uuid.UUID) error
	ListUsers(ctx context.Context, limit, offset int) ([]*entities.User, error)
}

// SessionRepository interface for session operations
type SessionRepository interface {
	Create(ctx context.Context, session *entities.UserSession) error
	FindByToken(ctx context.Context, token string) (*entities.UserSession, error)
	DeleteByUserID(ctx context.Context, userID uuid.UUID) error
	DeleteExpired(ctx context.Context) error
}

// JWTClaims represents the JWT token claims
type JWTClaims struct {
	UserID   uuid.UUID           `json:"user_id"`
	Username string              `json:"username"`
	Email    string              `json:"email"`
	Role     entities.UserRole   `json:"role"`
	jwt.RegisteredClaims
}

// NewAuthService creates a new authentication service
func NewAuthService(userRepo UserRepository, sessionRepo SessionRepository, jwtSecret string) *AuthService {
	return &AuthService{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		jwtSecret:   []byte(jwtSecret),
		accessTTL:   15 * time.Minute,
		refreshTTL:  7 * 24 * time.Hour,
	}
}

// Login authenticates a user and returns tokens
func (s *AuthService) Login(ctx context.Context, emailOrUsername, password, ipAddress, userAgent string) (accessToken, refreshToken string, user *entities.User, err error) {
	// Find user by email or username
	user, err = s.userRepo.FindByEmail(ctx, emailOrUsername)
	if err != nil {
		user, err = s.userRepo.FindByUsername(ctx, emailOrUsername)
		if err != nil {
			return "", "", nil, ErrInvalidCredentials
		}
	}

	// Check if user is active
	if !user.IsActive {
		return "", "", nil, ErrUserNotActive
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", "", nil, ErrInvalidCredentials
	}

	// Generate tokens
	accessToken, err = s.generateAccessToken(user)
	if err != nil {
		return "", "", nil, fmt.Errorf("generate access token: %w", err)
	}

	refreshToken, err = s.generateRefreshToken(user, ipAddress, userAgent)
	if err != nil {
		return "", "", nil, fmt.Errorf("generate refresh token: %w", err)
	}

	// Update last login
	_ = s.userRepo.UpdateLastLogin(ctx, user.ID)

	return accessToken, refreshToken, user, nil
}

// Register creates a new user account
func (s *AuthService) Register(ctx context.Context, email, username, password, fullName string) (*entities.User, error) {
	// Check if user exists
	if _, err := s.userRepo.FindByEmail(ctx, email); err == nil {
		return nil, errors.New("email already registered")
	}
	if _, err := s.userRepo.FindByUsername(ctx, username); err == nil {
		return nil, errors.New("username already taken")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	// Create user
	user := &entities.User{
		ID:           uuid.New(),
		Email:        email,
		Username:     username,
		PasswordHash: string(hashedPassword),
		FullName:     fullName,
		Role:         entities.RoleViewer, // Default role
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	return user, nil
}

// ValidateAccessToken validates and parses an access token
func (s *AuthService) ValidateAccessToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.jwtSecret, nil
	})

	if err != nil {
		return nil, ErrInvalidToken
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrInvalidToken
}

// RefreshAccessToken generates a new access token using a refresh token
func (s *AuthService) RefreshAccessToken(ctx context.Context, refreshToken string) (string, error) {
	// Find session
	session, err := s.sessionRepo.FindByToken(ctx, refreshToken)
	if err != nil {
		return "", ErrInvalidToken
	}

	// Check if expired
	if time.Now().After(session.ExpiresAt) {
		return "", ErrTokenExpired
	}

	// Get user
	user, err := s.userRepo.FindByID(ctx, session.UserID)
	if err != nil {
		return "", ErrInvalidToken
	}

	// Check if user is active
	if !user.IsActive {
		return "", ErrUserNotActive
	}

	// Generate new access token
	return s.generateAccessToken(user)
}

// Logout invalidates a user's refresh token
func (s *AuthService) Logout(ctx context.Context, userID uuid.UUID) error {
	return s.sessionRepo.DeleteByUserID(ctx, userID)
}

// generateAccessToken creates a new JWT access token
func (s *AuthService) generateAccessToken(user *entities.User) (string, error) {
	claims := JWTClaims{
		UserID:   user.ID,
		Username: user.Username,
		Email:    user.Email,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.accessTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

// generateRefreshToken creates a new refresh token and stores it in the session
func (s *AuthService) generateRefreshToken(user *entities.User, ipAddress, userAgent string) (string, error) {
	// Generate random token
	refreshToken := uuid.New().String()

	// Create session
	session := &entities.UserSession{
		ID:           uuid.New(),
		UserID:       user.ID,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(s.refreshTTL),
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		CreatedAt:    time.Now(),
	}

	// Store session
	ctx := context.Background()
	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return "", fmt.Errorf("create session: %w", err)
	}

	return refreshToken, nil
}

// ChangePassword changes a user's password
func (s *AuthService) ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) error {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return errors.New("user not found")
	}

	// Verify old password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword)); err != nil {
		return errors.New("incorrect current password")
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	// Update password
	user.PasswordHash = string(hashedPassword)
	user.UpdatedAt = time.Now()

	return s.userRepo.Update(ctx, user)
}

// UpdateUserRole updates a user's role (admin only)
func (s *AuthService) UpdateUserRole(ctx context.Context, adminID, targetUserID uuid.UUID, newRole entities.UserRole) error {
	// Get admin user
	admin, err := s.userRepo.FindByID(ctx, adminID)
	if err != nil {
		return errors.New("admin user not found")
	}

	// Check admin permissions
	if !admin.CanManageUsers() {
		return errors.New("insufficient permissions")
	}

	// Get target user
	user, err := s.userRepo.FindByID(ctx, targetUserID)
	if err != nil {
		return errors.New("target user not found")
	}

	// Update role
	user.Role = newRole
	user.UpdatedAt = time.Now()

	return s.userRepo.Update(ctx, user)
}