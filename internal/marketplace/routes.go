package discovery

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"gocom/main/internal/marketplace/discovery/handlers"
)

// SetupRoutes mounts discovery routes under the provided router group.
// Example: SetupRoutes(r.Group("/v1"), db)
func SetupRoutes(v1 *gin.RouterGroup, db *gorm.DB) {
	// product discovery
	v1.GET("/products", handlers.SearchProducts(db))
	v1.GET("/products/:id", handlers.GetProduct(db))

	// categories
	v1.GET("/categories", handlers.ListCategories(db))
	v1.GET("/categories/:id/products", handlers.ListCategories(db))

	// reviews
	v1.GET("/products/:id/reviews", handlers.ListReviews(db))
	v1.POST("/products/:id/reviews", handlers.CreateReview(db))
}
