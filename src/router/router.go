package router

import (
    "FORUM-js/src/handlers"
    "FORUM-js/src/middleware"
    "github.com/gorilla/mux"
)

func SetupRoutes() *mux.Router {
    r := mux.NewRouter()


    r.Use(middleware.CORSMiddleware())


    r.HandleFunc("/api/register", handlers.Register).Methods("POST")
    r.HandleFunc("/api/verify-register", handlers.VerifyRegister).Methods("GET")
    r.HandleFunc("/api/login", handlers.Login).Methods("POST")
    r.HandleFunc("/api/refresh", handlers.RefreshToken).Methods("POST")
    

    r.HandleFunc("/api/forgot-password", handlers.ResetPassword).Methods("POST")
    r.HandleFunc("/api/confirm-reset", handlers.ConfirmResetPassword).Methods("POST")

    return r
} 