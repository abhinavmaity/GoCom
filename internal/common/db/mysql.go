package db

import (
	"fmt"
	"gocom/main/internal/common/config"
	"gocom/main/internal/models"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectMySQL() {
	dsn := config.GetDatabaseDSN()

	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to MySQL:", err)
	}

	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatal("Failed to get underlying sql.DB:", err)
	}

	if err := sqlDB.Ping(); err != nil {
		log.Fatal("Failed to ping MySQL:", err)
	}

	log.Printf("Connected to MySQL: %s@%s:%s/%s",
		config.AppConfig.DBUser,
		config.AppConfig.DBHost,
		config.AppConfig.DBPort,
		config.AppConfig.DBName,
	)
}

func GetDB() *gorm.DB {
	if DB == nil {
		log.Fatal("DB is nil — call ConnectMySQL() first")
	}
	return DB
}

func AutoMigrate(models ...interface{}) error {
	if DB == nil {
		return fmt.Errorf("DB is nil")
	}

	log.Println("Running AutoMigrate...")
	if err := DB.AutoMigrate(models...); err != nil {
		return err
	}

	log.Printf("Migration completed (%d models)", len(models))
	return nil
}

func InitCommerceDB() (*gorm.DB, error) {

	ConnectMySQL()

	// Auto-migrate all Commerce Platform models
	err := AutoMigrate(
		&models.User{},
		&models.Seller{},
		&models.SellerUser{},
		&models.KYC{},
		&models.Category{},
		&models.Product{},
		&models.SKU{},
		&models.Inventory{},
		&models.Cart{},
		&models.CartItem{},
		&models.Order{},
		&models.OrderItem{},
		&models.Payment{},
		&models.Shipment{},
		&models.Return{},
		&models.Refund{},
		&models.Coupon{},
		&models.Review{},
		&models.Address{},
		&models.AuditLog{},
		&models.Media{},
	)

	if err != nil {
		return nil, fmt.Errorf("failed to migrate models: %w", err)
	}

	return GetDB(), nil
}
