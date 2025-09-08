package main

import (
    "bytes"
    "fmt"
    "io"
    "log"
    "time"
    
    "gocom/main/internal/common/config"
    "gocom/main/internal/integrations/storage"
)

func main() {
    log.Println("🧪 Testing MinIO Integration...")
    
    // Load configuration
    config.LoadConfig()
    
    // Connect to MinIO
    storage.ConnectMinIO()
    
    // Test 1: Initialize buckets
    log.Println("📦 Creating buckets...")
    if err := storage.InitializeBuckets(); err != nil {
        log.Fatalf("Failed to initialize buckets: %v", err)
    }
    
    // Test 2: Upload a test file
    log.Println("📤 Testing file upload...")
    testContent := "Hello from Team Seller! This is a test KYC document."
    reader := bytes.NewReader([]byte(testContent))
    
    fileName := fmt.Sprintf("test-document-%d.txt", time.Now().Unix())
    filePath, err := storage.UploadFile("kyc-documents", fileName, reader, int64(len(testContent)), "text/plain")
    if err != nil {
        log.Fatalf("Failed to upload test file: %v", err)
    }
    
    log.Printf("✅ File uploaded successfully: %s", filePath)
    
    // Test 3: Generate presigned URL
    log.Println("🔗 Generating presigned URL...")
    presignedURL, err := storage.GetPresignedURL("kyc-documents", fileName, 24*time.Hour)
    if err != nil {
        log.Fatalf("Failed to generate presigned URL: %v", err)
    }
    
    log.Printf("✅ Presigned URL generated: %s", presignedURL)
    
    // Test 4: Download file (FIXED - using io.ReadAll instead of buf.ReadFrom)
    log.Println("📥 Testing file download...")
    object, err := storage.DownloadFile("kyc-documents", fileName)
    if err != nil {
        log.Fatalf("Failed to download file: %v", err)
    }
    defer object.Close()
    
    // ✅ FIXED: Use io.ReadAll to read the object content
    downloadedData, err := io.ReadAll(object)
    if err != nil {
        log.Fatalf("Failed to read downloaded content: %v", err)
    }
    
    log.Printf("✅ Downloaded content: %s", string(downloadedData))
    
    // Test 5: List files
    log.Println("📋 Listing files in bucket...")
    files, err := storage.ListFiles("kyc-documents", "")
    if err != nil {
        log.Fatalf("Failed to list files: %v", err)
    }
    
    log.Printf("✅ Found %d files in kyc-documents bucket:", len(files))
    for _, file := range files {
        log.Printf("  - %s (size: %d bytes, modified: %s)", file.Key, file.Size, file.LastModified.Format("2006-01-02 15:04:05"))
    }
    
    // Test 6: Delete test file
    log.Println("🗑️ Cleaning up test file...")
    if err := storage.DeleteFile("kyc-documents", fileName); err != nil {
        log.Fatalf("Failed to delete test file: %v", err)
    }
    
    log.Println("🎉 All MinIO tests passed successfully!")
    log.Println("✅ MinIO is ready for seller operations:")
    log.Println("   - KYC document uploads ✓")
    log.Println("   - Product image storage ✓") 
    log.Println("   - File management ✓")
}
