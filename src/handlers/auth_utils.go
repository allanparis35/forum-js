
package handlers

import (
    "os"
    "time"
    "github.com/golang-jwt/jwt/v5"
)

func generateToken(userID uint, isAdmin bool) (string, error) {
    secret := os.Getenv("JWT_SECRET")
    claims := jwt.MapClaims{
        "user_id": userID,
        "role":    isAdmin,
        "exp":     time.Now().Add(2 * time.Hour).Unix(),
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(secret))
}

