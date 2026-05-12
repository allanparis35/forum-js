package database

import (
	"fmt"
	"log"
	"os"
	"FORUM-js/src/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	//Construction de la chaîne de connexion (DSN)
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		os.Getenv("DB_HOST"), 
		os.Getenv("DB_USER"), 
		os.Getenv("DB_PASSWORD"), 
		os.Getenv("DB_NAME"), 
		os.Getenv("DB_PORT"),
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
	)
	
	if err != nil {
		log.Fatal("Erreur lors de la migration : ", err)
	}

	fmt.Println("Base de données PostgreSQL connectée et synchronisée.")
}