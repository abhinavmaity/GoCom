package main

import (
	"log"

	"gocom/main/internal/common/config"
	"gocom/main/internal/common/db"
	"gocom/main/internal/models"
)

func main() {
	config.LoadConfig()
	db.ConnectMySQL()
	err := db.AutoMigrate(
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
		log.Fatalf("Migration failed: %v", err)
	}

	log.Println("Migration completed successfully!")
}
