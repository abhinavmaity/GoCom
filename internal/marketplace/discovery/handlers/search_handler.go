package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"gocom/main/internal/marketplace/discovery/services"
)

func SearchProducts(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		page := atoiDefault(c.Query("page"), 1)
		limit := atoiDefault(c.Query("limit"), 20)

		params := services.SearchParams{
			Query:    c.Query("query"),
			Category: c.Query("category"),
			Brand:    c.Query("brand"),
			MinPrice: c.Query("min_price"),
			MaxPrice: c.Query("max_price"),
			Sort:     c.Query("sort"),
			Page:     page,
			PageSize: limit,
		}
		svc := services.NewSearchService(db)
		out, err := svc.Search(c.Request.Context(), params)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "search failed", "detail": err.Error()})
			return
		}
		c.JSON(http.StatusOK, out)
	}
}
