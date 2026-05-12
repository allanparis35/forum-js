package middleware

import (
	"net/http"

	"github.com/rs/cors"
)

// CORSMiddleware configure un middleware CORS pour le backend
func CORSMiddleware() func(next http.Handler) http.Handler {
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Authorization"},
		AllowCredentials: true,
		MaxAge:           300, // 5 minutes
	})
	return c.Handler
}
