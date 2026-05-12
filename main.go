package main

import (
    "log"
    "net/http"
    "os"
    "FORUM-js/database"
    "FORUM-js/src/router"
    "github.com/joho/godotenv"
)

func main() {
    //Charger le fichier .env
    if err := godotenv.Load(); err != nil {
        log.Println("No .env file found")
    }

    //Initialiser la Base de données
    database.InitDB()

    //Configurer le Router
    r := router.SetupRoutes()

    //Lancer le serveur
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }

    log.Printf("Server started on http://localhost:%s", port)
    log.Fatal(http.ListenAndServe(":"+port, r))
}