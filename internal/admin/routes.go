package admin

import (
	"gocom/main/internal/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SetupAdminRoutes sets up all admin-related routes
// func SetupAdminRoutes(router *gin.Engine, database *gorm.DB, authService *auth.AuthService) {
// 	adminGroup := router.Group("/v1/admin")

// 	// Apply JWT middleware and admin role requirement
// 	adminGroup.Use(authService.JWTMiddleware())
// 	adminGroup.Use(authService.RequireRole(auth.RoleAdmin))

// 	// KYC Management Routes
// 	adminGroup.GET("/kyc/pending", GetKYCPending(database))
// 	adminGroup.POST("/kyc/:id/approve", ApproveKYC(database))
// 	adminGroup.POST("/kyc/:id/reject", RejectKYC(database))

// 	// Catalog Management Routes
// 	adminGroup.POST("/catalog/categories", CreateCategory(database))
// 	adminGroup.PATCH("/catalog/attributes", UpdateCategoryAttributes(database))
// }

// GetKYCPending retrieves all pending KYC submissions
func GetKYCPending(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var kycList []models.KYC

		// Get all pending KYC records (status = 0) with seller info
		result := db.Preload("Seller").Where("status = ?", 0).Find(&kycList)
		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to fetch pending KYC records",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data":    kycList,
			"count":   len(kycList),
			"message": "Pending KYC records retrieved successfully",
		})
	}
}

// ApproveKYC approves a KYC submission
func ApproveKYC(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		kycID, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid KYC ID",
			})
			return
		}

		var kyc models.KYC
		if err := db.First(&kyc, uint(kycID)).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "KYC record not found",
			})
			return
		}

		// Update KYC status to approved (1)
		kyc.Status = 1
		kyc.Remarks = "Approved by admin"

		if err := db.Save(&kyc).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to approve KYC",
			})
			return
		}

		// Also update the seller status to approved
		var seller models.Seller
		if err := db.First(&seller, kyc.SellerID).Error; err == nil {
			seller.Status = 1 // Approved
			db.Save(&seller)
		}

		// Log audit trail
		userEmail, exists := c.Get("email")
		if !exists {
			userEmail = "admin"
		}

		auditLog := models.AuditLog{
			Actor:    userEmail.(string),
			Action:   "APPROVE_KYC",
			Entity:   "KYC",
			EntityID: kyc.ID,
			Meta:     []byte(`{"approved_at":"` + kyc.UpdatedAt.String() + `"}`),
		}
		db.Create(&auditLog)

		c.JSON(http.StatusOK, gin.H{
			"message": "KYC approved successfully",
			"data":    kyc,
		})
	}
}

// RejectKYC rejects a KYC submission
func RejectKYC(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		kycID, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid KYC ID",
			})
			return
		}

		var req struct {
			Remarks string `json:"remarks" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Remarks are required for rejection",
			})
			return
		}

		var kyc models.KYC
		if err := db.First(&kyc, uint(kycID)).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "KYC record not found",
			})
			return
		}

		// Update KYC status to rejected (2)
		kyc.Status = 2
		kyc.Remarks = req.Remarks

		if err := db.Save(&kyc).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to reject KYC",
			})
			return
		}

		// Also update the seller status to rejected
		var seller models.Seller
		if err := db.First(&seller, kyc.SellerID).Error; err == nil {
			seller.Status = 2 // Rejected
			db.Save(&seller)
		}

		// Log audit trail
		userEmail, exists := c.Get("email")
		if !exists {
			userEmail = "admin"
		}

		auditLog := models.AuditLog{
			Actor:    userEmail.(string),
			Action:   "REJECT_KYC",
			Entity:   "KYC",
			EntityID: kyc.ID,
			Meta:     []byte(`{"rejected_at":"` + kyc.UpdatedAt.String() + `","remarks":"` + req.Remarks + `"}`),
		}
		db.Create(&auditLog)

		c.JSON(http.StatusOK, gin.H{
			"message": "KYC rejected successfully",
			"data":    kyc,
		})
	}
}

// CreateCategory creates a new product category
func CreateCategory(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var category models.Category

		if err := c.ShouldBindJSON(&category); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid category data",
				"details": err.Error(),
			})
			return
		}

		// Validate required fields
		if category.Name == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Category name is required",
			})
			return
		}

		// Generate SEO slug if not provided
		if category.SEOSlug == "" {
			category.SEOSlug = generateSlug(category.Name)
		}

		if err := db.Create(&category).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to create category",
			})
			return
		}

		// Log audit trail
		userEmail, exists := c.Get("email")
		if !exists {
			userEmail = "admin"
		}

		auditLog := models.AuditLog{
			Actor:    userEmail.(string),
			Action:   "CREATE_CATEGORY",
			Entity:   "Category",
			EntityID: category.ID,
			Meta:     []byte(`{"category_name":"` + category.Name + `"}`),
		}
		db.Create(&auditLog)

		c.JSON(http.StatusCreated, gin.H{
			"message": "Category created successfully",
			"data":    category,
		})
	}
}

// UpdateCategoryAttributes updates the attributes schema for a category
func UpdateCategoryAttributes(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			CategoryID       uint   `json:"category_id" binding:"required"`
			AttributesSchema string `json:"attributes_schema" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid request data",
				"details": err.Error(),
			})
			return
		}

		var category models.Category
		if err := db.First(&category, req.CategoryID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Category not found",
			})
			return
		}

		// Update attributes schema
		category.AttributesSchema = []byte(req.AttributesSchema)

		if err := db.Save(&category).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to update category attributes",
			})
			return
		}

		// Log audit trail
		userEmail, exists := c.Get("email")
		if !exists {
			userEmail = "admin"
		}

		auditLog := models.AuditLog{
			Actor:    userEmail.(string),
			Action:   "UPDATE_CATEGORY_ATTRIBUTES",
			Entity:   "Category",
			EntityID: category.ID,
			Meta:     []byte(`{"category_id":` + strconv.FormatUint(uint64(req.CategoryID), 10) + `}`),
		}
		db.Create(&auditLog)

		c.JSON(http.StatusOK, gin.H{
			"message": "Category attributes updated successfully",
			"data":    category,
		})
	}
}

// Helper function to generate SEO slug
func generateSlug(name string) string {
	// Simple slug generation - you can use a more robust library
	slug := ""
	for _, char := range name {
		if char == ' ' {
			slug += "-"
		} else if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') {
			if char >= 'A' && char <= 'Z' {
				slug += string(char + 32) // Convert to lowercase
			} else {
				slug += string(char)
			}
		}
	}
	return slug
}
