package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"gocom/main/internal/common/config"
	"gocom/main/internal/common/db"
	discovery "gocom/main/internal/marketplace/discovery/routes"
	"gocom/main/internal/models"
)

func main() {
	// 1) load configuration into AppConfig
	config.LoadConfig()

	// 2) connect DB (ConnectMySQL uses config.GetDatabaseDSN())
	db.ConnectMySQL()

	// 3) get *gorm.DB
	gormDB := db.GetDB()

	// 4) run migrations (optional but useful)
	if err := runAutoMigrate(gormDB); err != nil {
		log.Fatalf("auto-migrate failed: %v", err)
	}

	// 5) start router
	r := gin.New()
	r.Use(gin.Recovery())
	r.GET("/healthz", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })

	v1 := r.Group("/v1")
	discovery.SetupRoutes(v1, gormDB)

	// run server
	addr := ":" + getenv("MARKETPLACE_PORT", "8082")
	srvErr := make(chan error, 1)
	go func() {
		log.Println("marketplace-api listening on", addr)
		if err := r.Run(addr); err != nil {
			srvErr <- err
		}
	}()

	// graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	select {
	case sig := <-quit:
		log.Printf("shutting down server (%v) ...", sig)
	case err := <-srvErr:
		log.Fatalf("server error: %v", err)
	}

	cleanup()
}

func runAutoMigrate(gormDB *gorm.DB) error {
	return db.AutoMigrate(
		&models.Product{},
		&models.SKU{},
		&models.Category{},
		&models.Media{},
		&models.Inventory{},
		&models.Review{},
		&models.User{},
		&models.Order{},
		&models.OrderItem{},
		&models.Seller{},
		&models.Address{},
	)
}

func cleanup() {
	if db.GetDB() == nil {
		return
	}
	sqlDB, err := db.GetDB().DB()
	if err != nil {
		log.Printf("error getting sql.DB to close: %v", err)
		return
	}
	done := make(chan struct{})
	go func() {
		_ = sqlDB.Close()
		close(done)
	}()
	select {
	case <-done:
		log.Println("db closed")
	case <-time.After(3 * time.Second):
		log.Println("timeout closing db")
	}
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
