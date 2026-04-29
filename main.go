package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
	"FORUM-js/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	//Récupération des variables d'environnement
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=5432 sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
	)

	//Connexion à PostgreSQL via l'ORM GORM
	var db *gorm.DB
	var err error

	// Boucle de tentative pour attendre que le conteneur Postgres soit prêt
	for i := 0; i < 5; i++ {
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err == nil {
			break
		}
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		log.Fatal("Impossible de se connecter à la DB :", err)
	}

	//Migrations automatiques 
	err = db.AutoMigrate(&models.User{}, &models.Post{}, &models.Comment{}, &models.Like{}, &models.Tag{}, &models.PostTag{})
	if err != nil {
		log.Fatal("Erreur de migration :", err)
	}
	fmt.Println("Base de données prête et migrations terminées")

	//Serveur de fichiers statiques
	fileServer := http.FileServer(http.Dir("./public"))
	http.Handle("/", fileServer)

	//Lancement du serveur
	port := "8080"
	fmt.Printf("Serveur Forum Warframe lancé sur http://localhost:%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}