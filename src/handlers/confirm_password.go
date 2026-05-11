package handlers

import (
	"FORUM-js/database"
	"FORUM-js/src/models"
	"encoding/json"
	"net/http"
	"time"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func ConfirmResetPassword(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token           string `json:"token"`
		Password        string `json:"password"`
		ConfirmPassword string `json:"confirm_password"`
	}

	//Décodage du JSON
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	//VERIFICATION : Est-ce que les deux mots de passe sont identiques ?
	if req.Password != req.ConfirmPassword {
		http.Error(w, "les mots de passe ne correspondent pas", http.StatusBadRequest)
		return
	}

	//VERIFICATION : Est-ce que le mot de passe n'est pas vide
	if len(req.Password) < 8 {
		http.Error(w, "le mot de passe doit faire au moins 8 caractères", http.StatusBadRequest)
		return
	}

	//VERIFICATION du Token en base
	var userToken models.UserToken
	result := database.DB.Where("token = ? AND type = ?", req.Token, "reset_password").First(&userToken)
	
	if result.Error != nil {
		http.Error(w, "token invalide", http.StatusUnauthorized)
		return
	}

	//VERIFICATION de l'expiration
	if time.Now().After(userToken.ExpiresAt) {
		database.DB.Delete(&userToken)
		http.Error(w, "le lien a expiré", http.StatusUnauthorized)
		return
	}

	newHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "erreur serveur", http.StatusInternalServerError)
		return
	}

	//MISE À JOUR de l'utilisateur et SUPPRESSION du token
	err = database.DB.Transaction(func(tx *gorm.DB) error {

		if err := tx.Model(&models.User{}).Where("id = ?", userToken.UserID).Update("password", string(newHash)).Error; err != nil {
			return err
		}
		if err := tx.Delete(&userToken).Error; err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		http.Error(w, "erreur lors de la mise à jour", http.StatusInternalServerError)
		return
	}

	//RÉPONSE
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Mot de passe réinitialisé avec succès.",
	})
}