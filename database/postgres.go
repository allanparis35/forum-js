package database

import (
	"FORUM-js/src/models"
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	//Construction de la chaîne de connexion (DSN)
	host := os.Getenv("DB_HOST")
	if host == "" {
		host = "localhost"
	}
	port := os.Getenv("DB_PORT")
	if port == "" {
		port = "5432"
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		host,
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		port,
	)

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Erreur de connexion à la base de données : ", err)
	}

	//Migration automatique des tables
	err = DB.AutoMigrate(
		&models.User{},
		&models.UserToken{},
		&models.Post{},
		&models.Comment{},
		&models.Like{},
		&models.Favorite{},
		&models.Tag{},
	)

	if err != nil {
		log.Fatal("Erreur lors de la migration : ", err)
	}

	fmt.Println("Base de données PostgreSQL connectée et synchronisée.")
}
