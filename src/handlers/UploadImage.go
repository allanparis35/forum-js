package handlers

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	file := r.URL.Query().Get("file")

	// Si le fichier est bien renseigné dans l'URL
	if file != "" {
		safeFile := filepath.Base(file)

		//Suppression du fichier dans le dossier tmp/
		err := os.Remove("tmp/" + safeFile)
		if err != nil {
			log.Println("Erreur lors de la suppression :", err)
			http.Error(w, "Impossible de supprimer le fichier", http.StatusInternalServerError)
			return
		}

		//Redirection vers la page d'accueil
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	// Chargement du template HTML
	tmpl, err := template.ParseFiles("views/form_upload.html")
	if err != nil {
		http.Error(w, "Erreur template : "+err.Error(), http.StatusInternalServerError)
		return
	}
	entries, err := os.ReadDir("tmp/")
	if err != nil {
		http.Error(w, "Impossible de lire le dossier : "+err.Error(), http.StatusInternalServerError)
		return
	}

	type FileInfo struct {
		Name    string
		Size    int64
		Updated string
	}
	var files []FileInfo

	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		f := FileInfo{
			Name:    entry.Name(), //Nom du fichier
			Size:    info.Size(),	//Taille du fichier en octet
			Updated: info.ModTime().Format("02/01/2006 15:04"), // Formatage de la date en string
		}
		files = append(files, f)
	}

	data := map[string]interface{}{
		"Files": files,
	}

	tmpl.Execute(w, data)
}