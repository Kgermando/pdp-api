package database

import (
	"fmt"
	"log"

	"github.com/kgermando/pdp-api/config"
	"github.com/kgermando/pdp-api/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Connect() {
	cfg := config.AppConfig
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=Africa/Kinshasa",
		cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort, cfg.DBSSLMode,
	)

	logLevel := logger.Silent
	if cfg.AppEnv == "development" {
		logLevel = logger.Info
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get sql.DB: %v", err)
	}
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)

	DB = db
	log.Println("Database connected successfully")
}

func Migrate() {
	err := DB.AutoMigrate(
		&models.User{},
		&models.Category{},
		&models.Proposal{},
		&models.ProposalVote{},
		&models.ProposalComment{},
		&models.Institution{},
		&models.Report{},
		&models.ReportEvidence{},
		&models.Evaluation{},
		&models.Notification{},
		&models.SMSLog{},
		&models.Forum{},
		&models.ForumMessage{},
		&models.ForumLike{},
	)
	if err != nil {
		log.Fatalf("Migration failed: %v", err)
	}
	log.Println("Database migration completed")
}
