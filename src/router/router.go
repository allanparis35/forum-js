package router

import (
	"net/http"

	"FORUM-js/src/handlers"
	"FORUM-js/src/middleware"

	"github.com/gorilla/mux"
)

func SetupRoutes() *mux.Router {
	r := mux.NewRouter()

	r.Use(middleware.CORSMiddleware())

	r.HandleFunc("/api/register", handlers.Register).Methods("POST")
	r.HandleFunc("/api/verify-register", handlers.VerifyRegister).Methods("GET")
	r.HandleFunc("/api/login", middleware.CaptchaMiddleware(handlers.Login)).Methods("POST")
	r.HandleFunc("/api/refresh", handlers.RefreshToken).Methods("POST")

	r.HandleFunc("/api/forgot-password", handlers.ResetPassword).Methods("POST")
	r.HandleFunc("/api/confirm-reset", handlers.ConfirmResetPassword).Methods("POST")

	r.HandleFunc("/api/posts", handlers.ListPosts).Methods("GET")
	r.Handle("/api/posts", middleware.JWT(http.HandlerFunc(handlers.CreatePost))).Methods("POST")
	r.HandleFunc("/api/posts/{id}", handlers.GetPost).Methods("GET")
	r.Handle("/api/posts/{id}/comments", middleware.JWT(http.HandlerFunc(handlers.CreateComment))).Methods("POST")
	r.Handle("/api/posts/{id}/vote", middleware.JWT(http.HandlerFunc(handlers.VotePost))).Methods("POST")
	r.HandleFunc("/api/tags", handlers.ListTags).Methods("GET")

	return r
}
