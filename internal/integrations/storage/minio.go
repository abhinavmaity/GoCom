package storage

import (
    "context"
    "fmt"
    "io"
    "log"
    "time"
    "github.com/minio/minio-go/v7"
    "github.com/minio/minio-go/v7/pkg/credentials"
    "gocom/main/internal/common/config"
)

var MinIOClient *minio.Client

func ConnectMinIO() {
    cfg := config.AppConfig
    
    var err error
    MinIOClient, err = minio.New(cfg.MinIOEndpoint, &minio.Options{
        Creds:  credentials.NewStaticV4(cfg.MinIOAccessKey, cfg.MinIOSecretKey, ""),
        Secure: cfg.MinIOUseSSL,
    })
    
    if err != nil {
        log.Fatal("Failed to initialize MinIO client:", err)
    }
    
    log.Println("MinIO client initialized successfully")
}


func CreateBucketIfNotExists(bucketName string) error {
    ctx := context.Background()
    
    exists, err := MinIOClient.BucketExists(ctx, bucketName)
    if err != nil {
        return err
    }
    
    if !exists {
        err = MinIOClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
        if err != nil {
            return err
        }
        log.Printf("Created bucket: %s", bucketName)
    } else {
        log.Printf("Bucket already exists: %s", bucketName)
    }
    
    return nil
}

func UploadFile(bucketName, objectName string, reader io.Reader, objectSize int64, contentType string) (string, error) {
    ctx := context.Background()
    
    info, err := MinIOClient.PutObject(ctx, bucketName, objectName, reader, objectSize, minio.PutObjectOptions{
        ContentType: contentType,
    })
    
    if err != nil {
        return "", err
    }
    
    log.Printf("Uploaded file: %s/%s (size: %d bytes)", bucketName, objectName, info.Size)
    return fmt.Sprintf("/%s/%s", bucketName, objectName), nil
}

func GetPresignedURL(bucketName, objectName string, expiry time.Duration) (string, error) {
    ctx := context.Background()
    
    presignedURL, err := MinIOClient.PresignedGetObject(ctx, bucketName, objectName, expiry, nil)
    if err != nil {
        return "", err
    }
    
    return presignedURL.String(), nil
}


func DownloadFile(bucketName, objectName string) (*minio.Object, error) {
    ctx := context.Background()
    
    object, err := MinIOClient.GetObject(ctx, bucketName, objectName, minio.GetObjectOptions{})
    if err != nil {
        return nil, err
    }
    
    return object, nil
}

func DeleteFile(bucketName, objectName string) error {
    ctx := context.Background()
    
    err := MinIOClient.RemoveObject(ctx, bucketName, objectName, minio.RemoveObjectOptions{})
    if err != nil {
        return err
    }
    
    log.Printf("Deleted file: %s/%s", bucketName, objectName)
    return nil
}

// ListFiles lists all files in a bucket with prefix
func ListFiles(bucketName, prefix string) ([]minio.ObjectInfo, error) {
    ctx := context.Background()
    
    var objects []minio.ObjectInfo
    
    for object := range MinIOClient.ListObjects(ctx, bucketName, minio.ListObjectsOptions{
        Prefix:    prefix,
        Recursive: true,
    }) {
        if object.Err != nil {
            return nil, object.Err
        }
        objects = append(objects, object)
    }
    
    return objects, nil
}

func InitializeBuckets() error {
    buckets := []string{
        "kyc-documents",    // KYC verification files
        "product-images",   // Product photos
        "seller-documents", // Business documents
        "temp-uploads",     // Temporary file storage
    }
    
    for _, bucket := range buckets {
        if err := CreateBucketIfNotExists(bucket); err != nil {
            return fmt.Errorf("failed to create bucket %s: %v", bucket, err)
        }
    }
    
    return nil
}

// GetMinIOClient returns the MinIO 
func GetMinIOClient() *minio.Client {
    return MinIOClient
}
