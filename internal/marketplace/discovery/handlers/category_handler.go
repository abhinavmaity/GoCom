package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ListCategories returns id,parent,name,seo_slug
func ListCategories(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var cats []struct {
			ID       uint   `json:"id"`
			ParentID *uint  `json:"parent_id"`
			Name     string `json:"name"`
			SEOSlug  string `json:"seo_slug"`
		}
		if err := db.Table("categories").Select("id, parent_id, name, seo_slug").Order("id ASC").Scan(&cats).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch categories", "detail": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"categories": cats})
	}
}
