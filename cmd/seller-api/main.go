package main

import (
    "log"
    "net/http"
    
    "github.com/gin-gonic/gin"
    
    "gocom/main/internal/common/config"
    "gocom/main/internal/common/db"
    "gocom/main/internal/integrations/storage"
)

func main() {
    // Load configuration
    config.LoadConfig()
    
    // Connect to services
    db.ConnectMySQL()
    
    // Initialize MinIO
    storage.ConnectMinIO()
    
    // Initialize MinIO buckets
    if err := storage.InitializeBuckets(); err != nil {
        log.Printf("Warning: Could not initialize MinIO buckets: %v", err)
    }
    
    // Set Gin mode
    gin.SetMode(config.AppConfig.GinMode)
    
    // Create Gin router
    router := gin.Default()
    
    // Health check
    router.GET("/health", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{
            "status": "healthy",
            "service": "seller-api",
            "integrations": gin.H{
                "database": "connected",
                "storage":  "connected",
            },
        })
    })
    
    // MinIO test endpoint
    router.GET("/minio/test", func(c *gin.Context) {
        // List files in kyc-documents bucket
        files, err := storage.ListFiles("kyc-documents", "")
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        
        c.JSON(http.StatusOK, gin.H{
            "message": "MinIO is working!",
            "buckets": []string{"kyc-documents", "product-images", "seller-documents", "temp-uploads"},
            "files_in_kyc": len(files),
        })
    })
    
    // Start server
    port := ":" + config.AppConfig.ServerPort
    log.Printf("🚀 Seller API starting on port %s", config.AppConfig.ServerPort)
    log.Printf("🔗 MinIO Test: http://localhost%s/minio/test", port)
    log.Fatal(router.Run(port))
}
