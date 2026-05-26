package handlers

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	"FORUM-js/database"
	"FORUM-js/src/models"
	"FORUM-js/src/utils"

	"github.com/jackc/pgconn"
	"golang.org/x/crypto/bcrypt"
)

// Fonction utilitaire pour générer 6 chiffres aléatoires
func generate2FACode() string {
	b := make([]byte, 3)
	_, _ = rand.Read(b)
	val := int(b[0])<<16 | int(b[1])<<8 | int(b[2])
	return fmt.Sprintf("%06d", val%1000000)
}

func Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	
	twoFactorCode := generate2FACode()
	expiration := time.Now().Add(15 * time.Minute)

	
	user := models.User{
		Email:              req.Email,
		Username:           req.Username,
		Password:           string(hash),
		TwoFactorCode:      twoFactorCode,     
		TwoFactorExpiresAt: expiration,  
		IsActive:           false,           
	}

	result := database.DB.Create(&user)
	if result.Error != nil {
		if pgErr, ok := result.Error.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			http.Error(w, "email or username already exists", http.StatusConflict)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	err := utils.Send2FACodeEmail(user.Email, twoFactorCode)
	if err != nil {
		fmt.Println("Erreur d'envoi du mail 2FA:", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Inscription réussie. Veuillez vérifier votre boîte mail pour activer votre compte.",
		"email":   user.Email,
	})
}