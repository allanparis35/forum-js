package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

const maxUploadImageSize = 5 << 20 // 5 MiB
const maxUploadRequestSize = maxUploadImageSize + (1 << 20)

var allowedImageTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
	"image/gif":  ".gif",
}

// Upload d'une image pour un post
func UploadPostImage(w http.ResponseWriter, r *http.Request) {
	// Vérification si l'utilisateur est authentifié
	if _, ok := getAuthenticatedUserID(w, r); !ok {
		return
	}

	// Vérification si la taille de la requête est trop grande
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadRequestSize)
	// Vérification si le formulaire est invalide
	if err := r.ParseMultipartForm(maxUploadImageSize); err != nil {
		writeError(w, http.StatusBadRequest, "image trop volumineuse ou formulaire invalide")
		return
	}

	// Récupération du fichier
	file, header, err := r.FormFile("image")
	// Vérification si une erreur est survenue lors de la récupération du fichier
	if err != nil {
		writeError(w, http.StatusBadRequest, "image manquante")
		return
	}
	defer file.Close()

	// Vérification si la taille de l'image est trop grande
	if header.Size > maxUploadImageSize {
		writeError(w, http.StatusBadRequest, "image trop volumineuse")
		return
	}

	// Vérification si le type de l'image est valide
	contentType, extension, err := detectImageType(file)
	// Vérification si une erreur est survenue lors de la détection du type de l'image
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Génération d'un nom de fichier aléatoire
	fileName, err := randomFileName(extension)
	// Vérification si une erreur est survenue lors de la génération du nom de fichier
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erreur lors de la preparation du fichier")
		return
	}

	// Création du dossier d'upload
	uploadDir := filepath.Join("public", "uploads", "posts")
	// Vérification si une erreur est survenue lors de la création du dossier d'upload
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		writeError(w, http.StatusInternalServerError, "erreur lors de la creation du dossier upload")
		return
	}

	// Création du chemin de destination
	destinationPath := filepath.Join(uploadDir, fileName)
	// Vérification si une erreur est survenue lors de la création du chemin de destination
	destination, err := os.OpenFile(destinationPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	// Vérification si une erreur est survenue lors de la création du fichier
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erreur lors de la creation du fichier")
		return
	}
	defer destination.Close()

	// Copie de l'image dans le dossier d'upload
	if _, err := io.Copy(destination, file); err != nil {
		writeError(w, http.StatusInternalServerError, "erreur lors de l'enregistrement de l'image")
		return
	}
	// Envoi de la réponse
	writeJSON(w, http.StatusCreated, map[string]string{
		"url":          "/uploads/posts/" + fileName,
		"content_type": contentType,
	})
}

// Détection du type d'image
func detectImageType(file interface {
	io.Reader
	io.Seeker
}) (string, string, error) {
	// Création d'un buffer
	buffer := make([]byte, 512)
	// Lecture du buffer
	bytesRead, err := file.Read(buffer)
	if err != nil && !errors.Is(err, io.EOF) {
		return "", "", errors.New("image illisible")
	}

	// Détection du type de contenu
	contentType := http.DetectContentType(buffer[:bytesRead])
	// Vérification si le type de contenu est valide
	extension, ok := allowedImageTypes[contentType]
	// Vérification si le type de contenu n'est pas valide
	if !ok {
		return "", "", errors.New("format image non supporte")
	}

	// Réinitialisation du pointeur du fichier
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		// Vérification si une erreur est survenue lors de la réinitialisation du pointeur du fichier
		return "", "", errors.New("image illisible")
	}
	// Retour du type de contenu et de l'extension
	return contentType, extension, nil
}

// Génération d'un nom de fichier aléatoire
func randomFileName(extension string) (string, error) {
	buffer := make([]byte, 16)
	// Génération du buffer
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}
	// Encodage du buffer en hexadécimal
	// Ajout de l'extension au nom de fichier
	return hex.EncodeToString(buffer) + extension, nil
}
