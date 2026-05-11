package handlers

import (
	"FORUM-js/src/utils"
    "FORUM-js/database"
    "FORUM-js/src/models"
    "encoding/json"
    "net/http"
    "time"
	"fmt"
)

func ResetPassword(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Email string `json:"email"`
    }

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid json", http.StatusBadRequest)
        return
    }

    var user models.User
    result := database.DB.Where("email = ?", req.Email).First(&user)

    if result.Error == nil {
        tokenStr := generateRandomString(32) 
        
        //Création de l'entrée dans UserToken
        userToken := models.UserToken{
            UserID:    user.ID,
            Token:     tokenStr,
            Type:      "reset_password",
            ExpiresAt: time.Now().Add(15 * time.Minute),
        }

        database.DB.Create(&userToken)

		err := utils.SendResetEmail(user.Email, tokenStr)
    if err != nil {
        fmt.Println("Erreur d'envoi du mail :", err)
    }
	}

    //Réponse toujours positive (Sécurité : on ne confirme pas si l'email existe ou non)
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{
        "message": "Si cet email existe, un lien de réinitialisation a été envoyé.",
    })
}	
