package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
	"encoding/json"
	"net/url"

	"FORUM-js/database"
	"FORUM-js/src/handlers"
	"FORUM-js/src/middleware"
	"FORUM-js/src/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Récupération des variables d'environnement
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=5432 sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
	)

	// Connexion à PostgreSQL via l'ORM GORM
	var db *gorm.DB
	var err error

	for i := 0; i < 5; i++ {
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err == nil {
			break
		}
		fmt.Println("Attente de la base de données...")
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		log.Fatal("Impossible de se connecter à la DB :", err)
	}

	database.DB = db

	// Migrations automatiques
	modelsToMigrate := []interface{}{
		&models.User{},
		&models.Post{},
		&models.Comment{},
		&models.Like{},
		&models.Tag{},
		&models.PostTag{},
	}
	err = db.AutoMigrate(modelsToMigrate...)
	if err != nil {
		log.Fatal("Erreur de migration :", err)
	}
	fmt.Println("Base de données prête et migrations terminées")

	// On applique le middleware CORS sur nos routes API
	corsMiddleware := middleware.CORS()

	http.Handle("/api/register", corsMiddleware(http.HandlerFunc(handlers.Register)))
	http.Handle("/api/login", corsMiddleware(http.HandlerFunc(handlers.Login)))
	http.HandleFunc("/login", middleware.CaptchaMiddleware(handlers.Login))
	//Serveur de fichiers statiques
	fileServer := http.FileServer(http.Dir("./public"))
	http.Handle("/", fileServer)

	// Lancement du serveur
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Serveur Forum Warframe lancé sur http://localhost:%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

	type CaptchaResponse struct {
	Success bool `json:"success"`
}
// Fonction de vérification du captcha
func verifyCaptcha(token string) (bool, error) {
	resp, err := http.PostForm(
		"https://www.google.com/recaptcha/api/siteverify",
		url.Values{
			"secret":   {"TOKEN_SECRET_RECAPTCHA"},
			"response": {token},
		},
	)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	var result CaptchaResponse
	json.NewDecoder(resp.Body).Decode(&result)

	return result.Success, nil
}
