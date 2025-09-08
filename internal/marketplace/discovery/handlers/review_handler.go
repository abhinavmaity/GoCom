package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"gocom/main/internal/marketplace/discovery/services"
)

func ListReviews(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id64, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil || id64 == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
			return
		}
		page := atoiDefault(c.Query("page"), 1)
		limit := atoiDefault(c.Query("limit"), 20)

		svc := services.NewReviewService(db)
		rows, total, err := svc.ListReviews(c.Request.Context(), uint(id64), page, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list reviews", "detail": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"reviews": rows, "page": page, "limit": limit, "total": total})
	}
}

func CreateReview(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id64, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil || id64 == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
			return
		}

		var req struct {
			Rating int         `json:"rating" binding:"required,min=1,max=5"`
			Text   string      `json:"text"`
			Media  interface{} `json:"media"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload", "detail": err.Error()})
			return
		}

		// Temporary: hardcode userID=1 (remove once auth is ready)
		userID := uint(1)

		svc := services.NewReviewService(db)
		if err := svc.CreateReview(c.Request.Context(), userID, uint(id64), req.Rating, req.Text, req.Media, true); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create review", "detail": err.Error()})
			return
		}

		// recompute aggregates (non-fatal)
		_, _, _ = svc.RecomputeAndPersistAggregate(c.Request.Context(), uint(id64))

		c.JSON(http.StatusCreated, gin.H{"ok": true})
	}
}

// small helper for safe int conversion
func atoiDefault(s string, def int) int {
	if s == "" {
		return def
	}
	if v, err := strconv.Atoi(s); err == nil && v > 0 {
		return v
	}
	return def
}
