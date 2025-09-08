package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"

	"gocom/main/internal/common/config"
	"gocom/main/internal/common/db"
	discovery "gocom/main/internal/marketplace/discovery"
)

func main() {
	cfg := config.Load()
	database := db.MustConnect(cfg)

	r := gin.New()
	r.Use(gin.Recovery())
	r.GET("/healthz", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })

	// mount v1 group
	v1 := r.Group("/v1")
	discovery.SetupRoutes(v1, database)

	addr := ":" + getenv("MARKETPLACE_PORT", "8082")
	log.Println("marketplace-api listening on", addr)
	if err := r.Run(addr); err != nil {
		log.Fatal(err)
	}
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
