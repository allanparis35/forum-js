package handlers

import (
    "encoding/json"
    "net/http"
    "FORUM-js/src/models"
    "FORUM-js/database"
    "github.com/jackc/pgconn"
    "golang.org/x/crypto/bcrypt"
)

func Register(w http.ResponseWriter, r *http.Request) {
    var req RegisterRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid json", http.StatusBadRequest)
        return
    }

    hash, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
    user := models.User{Email: req.Email, Password: string(hash)}
    
    result := database.DB.Create(&user)
    if result.Error != nil {
        if pgErr, ok := result.Error.(*pgconn.PgError); ok && pgErr.Code == "23505" {
            http.Error(w, "email already exists", http.StatusConflict)
            return
        }
        http.Error(w, "internal server error", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(RegisterResponse{ID: user.ID, Email: user.Email})
}