package handlers

import (
    "encoding/json"
    "net/http"
    "os"
    "time"
    "FORUM-js/src/models"
    "FORUM-js/database"
    "github.com/golang-jwt/jwt/v5"
    "crypto/rand"
    "encoding/hex"
)

// Fonction utilitaire pour générer une chaîne aléatoire (Refresh Token)
func generateRandomString(n int) string {
    b := make([]byte, n)
    rand.Read(b)
    return hex.EncodeToString(b)
}

func RefreshToken(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Token string `json:"token"`
    }
    
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid json", http.StatusBadRequest)
        return
    }

    var user models.User
    //recherche de l'utilisateur possédant ce refresh token
    result := database.DB.Where("refresh_token = ?", req.Token).First(&user)
    if result.Error != nil {
        http.Error(w, "invalid or expired refresh token", http.StatusUnauthorized)
        return
    }

    //Générer le nouvel Access Token (2h) comme dans auth_handler.go
    secret := os.Getenv("JWT_SECRET")
    claims := jwt.MapClaims{
        "user_id": user.ID,
        "role":    user.IsAdmin,
        "exp":     time.Now().Add(2 * time.Hour).Unix(),
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    accessToken, err := token.SignedString([]byte(secret))
    if err != nil {
        http.Error(w, "error generating token", http.StatusInternalServerError)
        return
    }

    //Générer un nouveau Refresh Token pour plus de sécurité
    newRefreshToken := generateRandomString(32)
    user.RefreshToken = newRefreshToken
    database.DB.Save(&user) // Met à jour le champ RefreshToken en base

    //Réponse au front TypeScript
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{
        "access_token":  accessToken,
        "refresh_token": newRefreshToken,
    })
}