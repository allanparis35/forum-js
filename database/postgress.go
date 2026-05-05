package database

import (
	"os"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)
var DB *gorm.DB

func Connect() error {
	// Récupère l'URL de connexion depuis le .env
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return gorm.ErrInvalidDB 
	}

	// Connexion via GORM
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	}

	DB = db
	return nil
}