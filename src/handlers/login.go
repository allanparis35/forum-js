package handlers

import (
    "encoding/json"
    "net/http"
    "FORUM-js/src/models"
    "FORUM-js/database"
    "golang.org/x/crypto/bcrypt"
)

func Login(w http.ResponseWriter, r *http.Request) {
    var req LoginRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid JSON", http.StatusBadRequest)
        return
    }

    var user models.User 
    if err := database.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
        http.Error(w, "utilisateur non trouvé", http.StatusUnauthorized)
        return
    }

    if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
        http.Error(w, "mot de passe invalide", http.StatusUnauthorized)
        return
    }

    // Utilisation des fonctions utilitaires
    tokenString, err := generateToken(user.ID, user.IsAdmin)
    if err != nil {
        http.Error(w, "erreur token", http.StatusInternalServerError)
        return
    }

    refreshToken := generateRandomString(32)
    user.RefreshToken = refreshToken
    database.DB.Save(&user) // Sauvegarde en DB

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(LoginResponse{
        Token:        tokenString,
        RefreshToken: refreshToken,
    })
}