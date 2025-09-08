package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"gocom/main/internal/marketplace/discovery/services"
)

func GetProduct(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id64, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil || id64 == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
			return
		}

		svc := services.NewProductService(db)
		out, err := svc.GetProductDetail(c.Request.Context(), uint(id64))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, out)
	}
}
