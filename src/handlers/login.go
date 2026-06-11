package handlers

import (
	"FORUM-js/database"
	"FORUM-js/src/models"
	"encoding/json"
	"net/http"

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

	if !user.IsActive {
		http.Error(w, "compte non activé", http.StatusForbidden)
		return
	}

	tokenString, err := generateToken(user.ID, user.IsAdmin)
	if err != nil {
		http.Error(w, "erreur token", http.StatusInternalServerError)
		return
	}

	refreshToken := generateRandomString(32)
	user.RefreshToken = refreshToken
	if err := database.DB.Save(&user).Error; err != nil {
		http.Error(w, "erreur lors de la mise à jour du refresh token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(LoginResponse{
		Token:        tokenString,
		RefreshToken: refreshToken,
	})
}
