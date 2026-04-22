package database

import (
	"fmt"
	"log"

	"github.com/danielfillol/waste/internal/config"
	"github.com/danielfillol/waste/internal/domain/entity"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Connect(cfg *config.Config) *gorm.DB {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable TimeZone=America/Sao_Paulo",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPass, cfg.DBName,
	)

	logLevel := logger.Silent
	if cfg.AppEnv == "development" {
		logLevel = logger.Info
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	if err := migrate(db); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	return db
}

func migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		// Domain (global)
		&entity.UF{},
		&entity.City{},
		&entity.Material{},
		&entity.Packaging{},
		&entity.Treatment{},
		// Core
		&entity.Tenant{},
		&entity.User{},
		&entity.Generator{},
		&entity.Receiver{},
		&entity.Driver{},
		&entity.Truck{},
		&entity.Route{},
		&entity.Collect{},
		&entity.Alert{},
		// Financial
		&entity.PricingRule{},
		&entity.Invoice{},
		&entity.InvoiceItem{},
		&entity.TruckCost{},
		&entity.PersonnelCost{},
		// Audit
		&entity.AuditLog{},
	)
}
