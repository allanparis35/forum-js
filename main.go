package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

//Modèles pour la migration automatique
type User struct {
	id        uint   `gorm:"primaryKey"`
	Username  string `gorm:"unique;not null"`
	Email     string `gorm:"unique;not null"`
	Password  string `gorm:"not null"`
}

type Post struct {
    ID          uint   `gorm:"primaryKey"`
    Title       string `gorm:"not null"`
    ContentPost string `gorm:"not null"`
    ImageURL    string
    UserID      uint
    CreatedAt   time.Time
}

type Comment struct {
    ID        uint   `gorm:"primaryKey"`
    Content   string `gorm:"not null"`
    UserID    uint
    PostID    uint
    CreatedAt time.Time
}

type Like struct {
    ID     uint `gorm:"primaryKey"`
    UserID uint `gorm:"not null"`
    PostID uint `gorm:"not null"`
    IsLike bool `gorm:"not null"`
}

type Tag struct {
    ID   uint   `gorm:"primaryKey"`
    Name string `gorm:"unique;not null"`
}

type PostTag struct {
    PostID uint `gorm:"primaryKey"`
    TagID  uint `gorm:"primaryKey"`
}

func main() {
	//Récupération des variables d'environnement 
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=5432 sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
	)

	//Connexion à PostgreSQL via l'ORM
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Erreur connexion DB:", err)
	}

	//Migration automatique
	//crée les tables dans Postgres si elles n'existent pas
	db.AutoMigrate(&User{}) 
	fmt.Println("Base de données prête et migrations terminées")

	//Serveur de fichiers statiques (Public) [cite: 72, 120]
	fileServer := http.FileServer(http.Dir("./public"))
	http.Handle("/", fileServer)

	http.HandleFunc("/api/status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status": "API Forum Warframe opérationnelle"}`)
	})

	//Lancement du serveur
	fmt.Println("Serveur lancé sur http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}