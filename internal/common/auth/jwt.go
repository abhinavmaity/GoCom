package auth

import (
	"errors"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// User roles and status constants
type UserRole string
type UserStatus int

const (
	RoleAdmin  UserRole = "admin"
	RoleSeller UserRole = "seller"
	RoleUser   UserRole = "user"

	StatusActive   UserStatus = 1
	StatusInactive UserStatus = 0

	SellerStatusPending  = 0
	SellerStatusApproved = 1
	SellerStatusRejected = 2
)

// JWT Claims structure
type JWTClaims struct {
	UserID      uint     `json:"user_id"`
	Email       string   `json:"email"`
	Role        UserRole `json:"role"`
	SellerID    *uint    `json:"seller_id,omitempty"`
	Permissions []string `json:"permissions,omitempty"`
	jwt.RegisteredClaims
}

// Auth service structure
type AuthService struct {
	jwtSecret []byte
	jwtExpiry time.Duration
	db        *gorm.DB
}

// Create new auth service
func NewAuthService(secret string, db *gorm.DB) *AuthService {
	// Get JWT expiration from environment or use default (72 hours)
	jwtExpiry := 72 * time.Hour
	if envExpiry := os.Getenv("JWT_EXPIRATION_HOURS"); envExpiry != "" {
		if hours, err := strconv.Atoi(envExpiry); err == nil {
			jwtExpiry = time.Duration(hours) * time.Hour
		}
	}

	return &AuthService{
		jwtSecret: []byte(secret),
		jwtExpiry: jwtExpiry,
		db:        db,
	}
}

// Password utilities
func (a *AuthService) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func (a *AuthService) CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// Generate JWT token
func (a *AuthService) GenerateJWT(userID uint, email string, role UserRole, sellerID *uint) (string, error) {
	claims := JWTClaims{
		UserID:   userID,
		Email:    email,
		Role:     role,
		SellerID: sellerID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(a.jwtExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "commerce-platform",
		},
	}

	// Add role-specific permissions
	claims.Permissions = a.getRolePermissions(role)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(a.jwtSecret)
}

// Get permissions based on role
func (a *AuthService) getRolePermissions(role UserRole) []string {
	switch role {
	case RoleAdmin:
		return []string{
			"admin.kyc.approve",
			"admin.kyc.reject",
			"admin.category.create",
			"admin.category.update",
			"admin.users.view",
			"admin.sellers.view",
			"admin.audit.view",
		}
	case RoleSeller:
		return []string{
			"seller.profile.update",
			"seller.kyc.upload",
			"seller.products.create",
			"seller.products.update",
			"seller.inventory.update",
			"seller.orders.view",
		}
	case RoleUser:
		return []string{
			"user.cart.manage",
			"user.orders.create",
			"user.reviews.create",
			"user.profile.update",
		}
	default:
		return []string{}
	}
}

// Validate JWT token
func (a *AuthService) ValidateJWT(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return a.jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// JWT Middleware - validates tokens and sets user context
func (a *AuthService) JWTMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header required",
				"code":  "AUTH_HEADER_MISSING",
			})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid authorization format. Use 'Bearer <token>'",
				"code":  "INVALID_AUTH_FORMAT",
			})
			c.Abort()
			return
		}

		claims, err := a.ValidateJWT(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token: " + err.Error(),
				"code":  "INVALID_TOKEN",
			})
			c.Abort()
			return
		}

		// Store claims in context for use in handlers
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_role", string(claims.Role))
		c.Set("seller_id", claims.SellerID)
		c.Set("permissions", claims.Permissions)
		c.Next()
	}
}

// Role-based middleware
func (a *AuthService) RequireRole(role UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("user_role")
		if !exists || userRole != string(role) {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Access denied. Required role: " + string(role),
				"code":  "INSUFFICIENT_ROLE",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// Permission-based middleware
func (a *AuthService) RequirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		permissions, exists := c.Get("permissions")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "No permissions found",
				"code":  "NO_PERMISSIONS",
			})
			c.Abort()
			return
		}

		permissionList, ok := permissions.([]string)
		if !ok {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Invalid permissions format",
				"code":  "INVALID_PERMISSIONS",
			})
			c.Abort()
			return
		}

		for _, p := range permissionList {
			if p == permission {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{
			"error":      "Permission required: " + permission,
			"code":       "PERMISSION_REQUIRED",
			"permission": permission,
		})
		c.Abort()
	}
}

// Seller-specific middleware with KYC validation
func (a *AuthService) RequireApprovedSeller() gin.HandlerFunc {
	return func(c *gin.Context) {
		sellerIDInterface, exists := c.Get("seller_id")
		if !exists || sellerIDInterface == nil {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Seller account required",
				"code":  "SELLER_ACCOUNT_REQUIRED",
			})
			c.Abort()
			return
		}

		sellerID := sellerIDInterface.(*uint)
		if sellerID == nil {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Invalid seller account",
				"code":  "INVALID_SELLER_ACCOUNT",
			})
			c.Abort()
			return
		}

		// Check seller status in database
		var seller struct {
			Status int `gorm:"column:status"`
		}

		err := a.db.Table("sellers").Select("status").Where("id = ?", *sellerID).First(&seller).Error
		if err != nil {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Seller not found",
				"code":  "SELLER_NOT_FOUND",
			})
			c.Abort()
			return
		}

		if seller.Status != SellerStatusApproved {
			statusMessage := map[int]string{
				SellerStatusPending:  "KYC verification pending",
				SellerStatusRejected: "KYC verification rejected",
			}

			c.JSON(http.StatusForbidden, gin.H{
				"error":         "Seller account not approved",
				"code":          "SELLER_NOT_APPROVED",
				"seller_status": seller.Status,
				"message":       statusMessage[seller.Status],
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// Optional authentication middleware for marketplace (allows browsing without login)
func (a *AuthService) OptionalJWTMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// No token provided, continue without authentication
			c.Next()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			// Invalid format, but don't block - just continue
			c.Next()
			return
		}

		claims, err := a.ValidateJWT(tokenString)
		if err == nil {
			// Valid token, set user context
			c.Set("user_id", claims.UserID)
			c.Set("user_email", claims.Email)
			c.Set("user_role", string(claims.Role))
			c.Set("seller_id", claims.SellerID)
			c.Set("permissions", claims.Permissions)
		}

		c.Next()
	}
}

// Helper function to get user ID from context
func GetUserIDFromContext(c *gin.Context) (uint, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}

	id, ok := userID.(uint)
	return id, ok
}

// Helper function to check if user has permission
func HasPermission(c *gin.Context, permission string) bool {
	permissions, exists := c.Get("permissions")
	if !exists {
		return false
	}

	permissionList, ok := permissions.([]string)
	if !ok {
		return false
	}

	for _, p := range permissionList {
		if p == permission {
			return true
		}
	}
	return false
}
