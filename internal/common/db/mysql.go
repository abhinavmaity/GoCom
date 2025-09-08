package db

import (
	"fmt"
	"log"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gocom/main/internal/common/config"
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
