package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"
	"FORUM-js/src/models"
	"FORUM-js/database"
	"github.com/jackc/pgconn" // pour gérer les erreurs de duplication d'email
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterResponse struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Captcha  string `json:"captcha"`
}


type LoginResponse struct {
	Token string `json:"token"`
}

func Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Password == "" {
		http.Error(w, "email and password required", http.StatusBadRequest)
		return
	}

	hash, err := bcrypt.GenerateFromPassword(
		[]byte(req.Password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		http.Error(w, "error hashing password", http.StatusInternalServerError)
		return
	}

	var userID int

	log.Println("Tentative d'ajout user :", req.Email)

	user := models.User{Email: req.Email, Password: string(hash)}
	result := database.DB.Create(&user)
	// gérer les erreurs de duplication d'email
	if result.Error != nil {
		if pgErr, ok := result.Error.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			http.Error(w, "email already exists", http.StatusConflict)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError) // pour les autres erreurs
		return
	}

	resp := RegisterResponse{
		ID:    userID,
		Email: req.Email,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Password == "" {
		http.Error(w, "email and password requi", http.StatusBadRequest)
		return
	}

    var user models.User 


    //Recherche de l'utilisateur par email avec GORM
    result := database.DB.Where("email = ?", req.Email).First(&user)

    if result.Error != nil {
        http.Error(w, "utilisateur non trouvé", http.StatusUnauthorized)
        return
    }

    //Vérification du mot de passe avec bcrypt
    if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
        http.Error(w, "mot de passe invalide", http.StatusUnauthorized)
        return
    }

    //Préparation des Claims JWT
    secret := os.Getenv("JWT_SECRET")
    if secret == "" {
        http.Error(w, "JWT secret non configuré", http.StatusInternalServerError)
        return
    }

    claims := jwt.MapClaims{
        "user_id": user.ID,       
        "role":    user.IsAdmin,  
        "exp":     time.Now().Add(2 * time.Hour).Unix(),
    }


    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    tokenString, err := token.SignedString([]byte(secret))
    if err != nil {
        http.Error(w, "erreur de génération du token", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(LoginResponse{Token: tokenString})
}
