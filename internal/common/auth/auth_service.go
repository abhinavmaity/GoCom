package auth

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gocom/main/internal/common/db"
	"gocom/main/internal/models"
)

type AuthService struct {
	DB *gorm.DB
}

func NewAuthService() *AuthService {
	return &AuthService{
		DB: db.GetDB(),
	}
}

func (as *AuthService) Register(req *RegisterRequest) (*models.User, error) {
	// Check if user exists
	var existingUser models.User
	if err := as.DB.Where("email = ? OR phone = ?", req.Email, req.Phone).First(&existingUser).Error; err == nil {
		return nil, errors.New("user already exists with this email or phone")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	// Create user
	user := &models.User{
		Name:         req.Name,
		Email:        req.Email,
		Phone:        req.Phone,
		PasswordHash: string(hashedPassword),
		Status:       models.UserStatusActive,
	}

	if err := as.DB.Create(user).Error; err != nil {
		return nil, errors.New("failed to create user")
	}

	return user, nil
}

func (as *AuthService) Login(req *LoginRequest) (*LoginResponse, error) {
	var user models.User
	if err := as.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		return nil, errors.New("invalid credentials")
	}

	// Check password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	// Check status
	if user.Status != models.UserStatusActive {
		return nil, errors.New("account is inactive")
	}

	// Generate tokens
	accessToken, err := GenerateAccessToken(user.ID, user.Email)
	if err != nil {
		return nil, errors.New("failed to generate access token")
	}

	refreshToken, err := GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, errors.New("failed to generate refresh token")
	}

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    3600,
		User: UserInfo{
			ID:    user.ID,
			Name:  user.Name,
			Email: user.Email,
			Phone: user.Phone,
		},
	}, nil
}

func (as *AuthService) RefreshToken(refreshToken string) (*RefreshResponse, error) {
	claims, err := ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}

	var user models.User
	if err := as.DB.First(&user, claims.UserID).Error; err != nil {
		return nil, errors.New("user not found")
	}

	newAccessToken, err := GenerateAccessToken(user.ID, user.Email)
	if err != nil {
		return nil, errors.New("failed to generate access token")
	}

	return &RefreshResponse{
		AccessToken: newAccessToken,
		TokenType:   "Bearer",
		ExpiresIn:   3600,
	}, nil
}

func (as *AuthService) GetUser(userID uint) (*models.User, error) {
	var user models.User
	if err := as.DB.First(&user, userID).Error; err != nil {
		return nil, errors.New("user not found")
	}
	return &user, nil
}

// DTOs
type RegisterRequest struct {
	Name     string `json:"name" binding:"required,min=2,max=50"`
	Email    string `json:"email" binding:"required,email"`
	Phone    string `json:"phone" binding:"required,min=10,max=15"`
	Password string `json:"password" binding:"required,min=8"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token"`
	TokenType    string   `json:"token_type"`
	ExpiresIn    int64    `json:"expires_in"`
	User         UserInfo `json:"user"`
}

type RefreshResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
}

type UserInfo struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Phone string `json:"phone"`
}
