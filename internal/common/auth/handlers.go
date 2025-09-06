// package auth

// import (

// 	"gocom/main/internal/common/db"
// 	"gocom/main/internal/models"
// 	"net/http"

// 	"github.com/gin-gonic/gin"
// )

// type AuthHandler struct {
// 	authService *AuthService
// }

// func NewAuthHandler(authService *AuthService) *AuthHandler {
// 	return &AuthHandler{
// 		authService: authService,
// 	}
// }

// type RegisterRequest struct {
// 	Name     string `json:"name" binding:"required"`
// 	Email    string `json:"email" binding:"required,email"`
// 	Phone    string `json:"phone" binding:"required"`
// 	Password string `json:"password" binding:"required,min=6"`
// 	Role     string `json:"role"` // optional, defaults to "user"
// }

// type LoginRequest struct {
// 	Email    string `json:"email" binding:"required,email"`
// 	Password string `json:"password" binding:"required"`
// }

// type AuthResponse struct {
// 	Token   string      `json:"token"`
// 	User    models.User `json:"user"`
// 	Message string      `json:"message"`
// }

// // Register new user
// func (h *AuthHandler) Register(c *gin.Context) {
// 	var req RegisterRequest
// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}

// 	// Use your existing database connection
// 	database := db.GetDB()

// 	// Check if user already exists
// 	var existingUser models.User
// 	result := database.Where("email = ? OR phone = ?", req.Email, req.Phone).First(&existingUser)
// 	if result.Error == nil {
// 		c.JSON(http.StatusConflict, gin.H{"error": "User with this email or phone already exists"})
// 		return
// 	}

// 	// Hash password
// 	hashedPassword, err := h.authService.HashPassword(req.Password)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
// 		return
// 	}

// 	// Determine role
// 	role := RoleUser // default
// 	if req.Role == "admin" {
// 		role = RoleAdmin
// 	} else if req.Role == "seller" {
// 		role = RoleSeller
// 	}

// 	// Create user
// 	user := models.User{
// 		Name:         req.Name,
// 		Email:        req.Email,
// 		Phone:        req.Phone,
// 		PasswordHash: hashedPassword,
// 		Status:       1, // Active
// 	}

// 	if err := database.Create(&user).Error; err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
// 		return
// 	}

// 	// Generate JWT token
// 	token, err := h.authService.GenerateJWT(user.ID, user.Email, role, nil)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
// 		return
// 	}

// 	c.JSON(http.StatusCreated, AuthResponse{
// 		Token:   token,
// 		User:    user,
// 		Message: "User registered successfully",
// 	})
// }

// // Login user
// func (h *AuthHandler) Login(c *gin.Context) {
// 	var req LoginRequest
// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}

// 	// Use your existing database connection
// 	database := db.GetDB()

// 	// Find user by email
// 	var user models.User
// 	if err := database.Where("email = ? AND status = 1", req.Email).First(&user).Error; err != nil {
// 		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
// 		return
// 	}

// 	// Check password
// 	if !h.authService.CheckPasswordHash(req.Password, user.PasswordHash) {
// 		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
// 		return
// 	}

// 	// Determine role and seller ID
// 	role := RoleUser
// 	var sellerID *uint

// 	// Check if user is a seller
// 	var sellerUser models.SellerUser
// 	if err := database.Where("user_id = ? AND status = 1", user.ID).First(&sellerUser).Error; err == nil {
// 		role = RoleSeller
// 		sellerID = &sellerUser.SellerID
// 	}

// 	// Generate JWT token
// 	token, err := h.authService.GenerateJWT(user.ID, user.Email, role, sellerID)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
// 		return
// 	}

// 	c.JSON(http.StatusOK, AuthResponse{
// 		Token:   token,
// 		User:    user,
// 		Message: "Login successful",
// 	})
// }

// // Test protected endpoint
// func (h *AuthHandler) GetProfile(c *gin.Context) {
// 	userID, exists := GetUserIDFromContext(c)
// 	if !exists {
// 		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
// 		return
// 	}

// 	database := db.GetDB()
// 	var user models.User
// 	if err := database.First(&user, userID).Error; err != nil {
// 		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
// 		return
// 	}

// 	c.JSON(http.StatusOK, gin.H{
// 		"user":        user,
// 		"permissions": c.MustGet("permissions"),
// 	})
// }

package auth

import (
	"crypto/rand"
	"encoding/hex"
	"gocom/main/internal/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pquerna/otp/totp"
	"gorm.io/gorm"
)

type AuthHandler struct {
	authService *AuthService
	DB          *gorm.DB
}

func NewAuthHandler(authService *AuthService, database *gorm.DB) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		DB:          database,
	}
}

type RegisterRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Phone    string `json:"phone" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
	Role     string `json:"role"` // optional, defaults to "user"
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	Token   string      `json:"token"`
	User    models.User `json:"user"`
	Message string      `json:"message"`
}

// Generate refresh token
func generateRefreshToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// Create refresh token
func (h *AuthHandler) createRefreshToken(userID uint) (string, error) {
	token, err := generateRefreshToken()
	if err != nil {
		return "", err
	}

	refreshToken := models.RefreshToken{
		UserID:    userID,
		Token:     token,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour), // 7 days
		IsRevoked: false,
	}

	if err := h.DB.Create(&refreshToken).Error; err != nil {
		return "", err
	}

	return token, nil
}

// Register new user
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if user already exists
	var existingUser models.User
	result := h.DB.Where("email = ? OR phone = ?", req.Email, req.Phone).First(&existingUser)
	if result.Error == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "User with this email or phone already exists"})
		return
	}

	// Hash password
	hashedPassword, err := h.authService.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Determine role
	role := RoleUser // default
	if req.Role == "admin" {
		role = RoleAdmin
	} else if req.Role == "seller" {
		role = RoleSeller
	}

	// Create user
	user := models.User{
		Name:         req.Name,
		Email:        req.Email,
		Phone:        req.Phone,
		PasswordHash: hashedPassword,
		Status:       1, // Active
	}

	if err := h.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// Generate JWT token
	token, err := h.authService.GenerateJWT(user.ID, user.Email, role, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusCreated, AuthResponse{
		Token:   token,
		User:    user,
		Message: "User registered successfully",
	})
}

// Login user with refresh token
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Find user by email
	var user models.User
	if err := h.DB.Where("email = ? AND status = 1", req.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Check password
	if !h.authService.CheckPasswordHash(req.Password, user.PasswordHash) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Determine role and seller ID
	role := RoleUser
	var sellerID *uint

	// Check if user is a seller
	var sellerUser models.SellerUser
	if err := h.DB.Where("user_id = ? AND status = 1", user.ID).First(&sellerUser).Error; err == nil {
		role = RoleSeller
		sellerID = &sellerUser.SellerID
	}

	// Generate JWT token
	token, err := h.authService.GenerateJWT(user.ID, user.Email, role, sellerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Generate refresh token
	refreshToken, err := h.createRefreshToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate refresh token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "Login successful",
		"token":         token,
		"refresh_token": refreshToken,
		"user": gin.H{
			"id":    user.ID,
			"name":  user.Name,
			"email": user.Email,
		},
	})
}

// Refresh token endpoint
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var refreshToken models.RefreshToken
	if err := h.DB.Where("token = ? AND is_revoked = false AND expires_at > ?",
		req.RefreshToken, time.Now()).Preload("User").First(&refreshToken).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	// Determine role
	role := RoleUser
	var sellerID *uint
	var sellerUser models.SellerUser
	if err := h.DB.Where("user_id = ? AND status = 1", refreshToken.User.ID).First(&sellerUser).Error; err == nil {
		role = RoleSeller
		sellerID = &sellerUser.SellerID
	}

	// Generate new access token
	token, err := h.authService.GenerateJWT(refreshToken.User.ID, refreshToken.User.Email, role, sellerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":   token,
		"message": "Token refreshed successfully",
	})
}

// Generate OTP for user
func (h *AuthHandler) GenerateOTP(c *gin.Context) {
	userID, exists := GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	var user models.User
	if err := h.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Generate TOTP secret if not exists
	if user.OTPSecret == "" {
		key, err := totp.Generate(totp.GenerateOpts{
			Issuer:      "Commerce Platform",
			AccountName: user.Email,
			SecretSize:  32,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate OTP"})
			return
		}

		user.OTPSecret = key.Secret()
		h.DB.Save(&user)
	}

	// Generate current OTP token
	token, err := totp.GenerateCode(user.OTPSecret, time.Now())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate OTP code"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "OTP generated successfully",
		"otp_code":   token, // Remove this in production
		"expires_in": "30 seconds",
	})
}

// Verify OTP
func (h *AuthHandler) VerifyOTP(c *gin.Context) {
	var req struct {
		Email   string `json:"email" binding:"required,email"`
		OTPCode string `json:"otp_code" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := h.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Verify TOTP
	valid := totp.Validate(req.OTPCode, user.OTPSecret)
	if !valid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid OTP code"})
		return
	}

	// Enable OTP for user
	user.OTPEnabled = true
	user.OTPVerified = true
	h.DB.Save(&user)

	c.JSON(http.StatusOK, gin.H{
		"message":     "OTP verified successfully",
		"otp_enabled": true,
	})
}

// Test protected endpoint
func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID, exists := GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	var user models.User
	if err := h.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user":        user,
		"permissions": c.MustGet("permissions"),
	})
}
