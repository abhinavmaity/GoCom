package discovery

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"gocom/main/internal/common/config"
	"gocom/main/internal/marketplace/discovery/handlers"
	"gocom/main/internal/marketplace/middleware"
)

// SetupRoutes mounts discovery routes under the provided router group.
// Example: SetupRoutes(r.Group("/v1"), db)
func SetupRoutes(v1 *gin.RouterGroup, db *gorm.DB) {
	// load config for JWT secret (non-blocking; small cost)
	cfg := config.Load()

	// global middleware for this group: decode JWT if present and set user_id
	v1.Use(middleware.JWTAuthMiddleware(cfg.JWTSecret))

	// discovery endpoints (public reads; write requires auth)
	v1.GET("/products", handlers.SearchProducts(db))
	v1.GET("/products/:id", handlers.GetProduct(db))

	v1.GET("/categories", handlers.ListCategories(db))
	v1.GET("/categories/:id/products", handlers.ListCategoryProducts(db))

	v1.GET("/products/:id/reviews", handlers.ListReviews(db))

	// require auth for creating reviews
	v1.POST("/products/:id/reviews", requireAuth(handlers.CreateReview(db)))
}

// requireAuth wraps a handler and aborts with 401 if user_id is not present in context.
// This keeps the JWT middleware permissive (it allows anonymous requests) but enforces auth per route.
func requireAuth(h gin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		if _, ok := c.Get("user_id"); !ok {
			c.AbortWithStatusJSON(401, gin.H{"error": "authentication required"})
			return
		}
		h(c)
	}
}
