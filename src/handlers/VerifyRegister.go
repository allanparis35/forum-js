package handlers

import (
	"fmt"
	"net/http"
	"FORUM-js/database"
	"FORUM-js/src/models"
	"time"
)

func VerifyRegister(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	code := r.URL.Query().Get("code")

	var user models.User
	result := database.DB.Where("email = ?", email).First(&user)
	if result.Error != nil {
		http.Error(w, "utilisateur non trouvé", http.StatusNotFound)
		return
	}

	if user.TwoFactorCode != code || time.Now().After(user.TwoFactorExpiresAt) {
		http.Error(w, "code invalide ou expiré", http.StatusForbidden)
		return
	}
	user.IsActive = true
	user.TwoFactorCode = "" 
	database.DB.Save(&user)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"message": "compte vérifié avec succès"}`)
}