package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"FORUM-js/database"
	"FORUM-js/src/models"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const maxPostTags = 5

type createPostRequest struct {
	Title    string   `json:"title"`
	Content  string   `json:"content"`
	ImageURL string   `json:"image_url"`
	Tags     []string `json:"tags"`
}

// Création d'un commentaire
type createCommentRequest struct {
	Content string `json:"content"`
}

// Voter pour un post
type voteRequest struct {
	IsLike bool `json:"is_like"`
}

// Liste des tags
type tagResponse struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

// Réponse pour un post
type postResponse struct {
	ID            uint          `json:"id"`
	Title         string        `json:"title"`
	Content       string        `json:"content"`
	ImageURL      string        `json:"image_url,omitempty"`
	Author        string        `json:"author"`
	UserID        uint          `json:"user_id"`
	Status        string        `json:"status"`
	CreatedAt     string        `json:"created_at"`
	Tags          []tagResponse `json:"tags"`
	LikesCount    int64         `json:"likes_count"`
	DislikesCount int64         `json:"dislikes_count"`
	CommentsCount int64         `json:"comments_count"`
	Score         int64         `json:"score"`
	IsFavorite    bool          `json:"is_favorite"`
}

// Réponse pour un commentaire
type commentResponse struct {
	ID        uint   `json:"id"`
	Content   string `json:"content"`
	Author    string `json:"author"`
	UserID    uint   `json:"user_id"`
	CreatedAt string `json:"created_at"`
}

// Réponse pour un post détaillé
type postDetailResponse struct {
	Post     postResponse      `json:"post"`
	Comments []commentResponse `json:"comments"`
}

// Liste des posts
func ListPosts(w http.ResponseWriter, r *http.Request) {
	var posts []models.Post
	query := database.DB.Preload("User").Preload("Tags")

	// Tri des posts
	switch r.URL.Query().Get("sort") {
	// Tri des posts par réponse
	case "unanswered":
		query = query.Where("NOT EXISTS (?)",
			database.DB.Model(&models.Comment{}).Select("1").Where("comments.post_id = posts.id"),
		).Order("created_at DESC")
	// Tri des posts par récent
	case "recent":
		query = query.Order("created_at DESC")
	default:
		// Tri des posts par récent
		query = query.Order("created_at DESC")
	}

	// Filtre par tag si le tag n'est pas vide et si le tag n'est pas "all"
	if tag := normalizeTagName(r.URL.Query().Get("tag")); tag != "" && tag != "all" {
		query = query.Joins("JOIN post_tags ON post_tags.post_id = posts.id").
			Joins("JOIN tags ON tags.id = post_tags.tag_id").
			Where("tags.name = ?", tag)
	}

	// Filtre par recherche
	if search := strings.TrimSpace(r.URL.Query().Get("q")); search != "" {
		like := "%" + strings.ToLower(search) + "%"
		query = query.Where("LOWER(title) LIKE ? OR LOWER(content_post) LIKE ?", like, like)
	}

	userID, _ := getUserIDFromContext(r)

	// Récupération des posts
	if err := query.Find(&posts).Error; err != nil {
		// Vérification si une erreur est survenue lors de la récupération des posts
		writeError(w, http.StatusInternalServerError, "erreur lors de la recuperation des posts")
		return
	}
	// Construction des réponses
	responses := make([]postResponse, 0, len(posts))
	for _, post := range posts {
		responses = append(responses, buildPostResponse(post, userID))
	}

	// Tri des posts par popularité
	if r.URL.Query().Get("sort") == "popular" {
		sort.SliceStable(responses, func(i, j int) bool {
			return responses[i].Score > responses[j].Score
		})
	}

	// Envoi des réponses
	writeJSON(w, http.StatusOK, map[string]any{"posts": responses})
}

// Récupération d'un post
func GetPost(w http.ResponseWriter, r *http.Request) {
	postID, ok := parseIDParam(w, r, "id")
	// Vérification de l'ID du post
	if !ok {
		return
	}
	// Récupération du post
	var post models.Post
	// Vérification de l'existence du post
	if err := database.DB.Preload("User").Preload("Tags").First(&post, postID).Error; err != nil {
		// Vérification si le post n'existe pas
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "post introuvable")
			return
		}
		// Vérification si une erreur est survenue lors de la récupération du post
		writeError(w, http.StatusInternalServerError, "erreur lors de la recuperation du post")
		return
	}

	// Récupération des commentaires
	var comments []models.Comment
	// Vérification si une erreur est survenue lors de la récupération des commentaires
	if err := database.DB.Preload("User").Where("post_id = ?", post.ID).Order("created_at ASC").Find(&comments).Error; err != nil {
		// Vérification si une erreur est survenue lors de la récupération des commentaires
		writeError(w, http.StatusInternalServerError, "erreur lors de la recuperation des commentaires")
		return
	}

	// Construction des réponses des commentaires
	commentResponses := make([]commentResponse, 0, len(comments))
	for _, comment := range comments {
		commentResponses = append(commentResponses, commentResponse{
			ID:        comment.ID,
			Content:   comment.Content,
			Author:    comment.User.Username,
			UserID:    comment.UserID,
			CreatedAt: comment.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	userID, _ := getUserIDFromContext(r)

	// Envoi des réponses
	writeJSON(w, http.StatusOK, postDetailResponse{
		Post:     buildPostResponse(post, userID),
		Comments: commentResponses,
	})
}

// Création d'un post
func CreatePost(w http.ResponseWriter, r *http.Request) {
	// Vérification de l'utilisateur authentifié
	userID, ok := getAuthenticatedUserID(w, r)
	if !ok {
		return
	}
	// Récupération de la requête
	var req createPostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "json invalide")
		return
	}
	// Traitement du titre, du contenu et de l'URL de l'image
	req.Title = strings.TrimSpace(req.Title)
	req.Content = strings.TrimSpace(req.Content)
	req.ImageURL = strings.TrimSpace(req.ImageURL)

	// Vérification de la longueur du titre et du contenu
	if len(req.Title) < 5 {
		writeError(w, http.StatusBadRequest, "le titre doit faire au moins 5 caracteres")
		return
	}
	// Vérification de la longueur du contenu
	if len(req.Content) < 20 {
		writeError(w, http.StatusBadRequest, "le contenu doit faire au moins 20 caracteres")
		return
	}

	// Création des tags
	tags, err := findOrCreateTags(req.Tags)
	// Vérification si une erreur est survenue lors de la création des tags
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Création du post
	post := models.Post{
		Title:    req.Title,
		Content:  req.Content,
		ImageUrl: req.ImageURL,
		UserID:   uint(userID),
		Status:   "published",
		Tags:     tags,
	}

	// Vérification si une erreur est survenue lors de la création du post
	if err := database.DB.Create(&post).Error; err != nil {
		// Vérification si une erreur est survenue lors de la création du post
		writeError(w, http.StatusInternalServerError, "erreur lors de la creation du post")
		return
	}

	// Vérification si une erreur est survenue lors de la récupération du post
	if err := database.DB.Preload("User").Preload("Tags").First(&post, post.ID).Error; err != nil {
		// Vérification si une erreur est survenue lors de la récupération du post
		writeError(w, http.StatusInternalServerError, "post cree mais impossible a relire")
		return
	}

	// Envoi de la réponse
	writeJSON(w, http.StatusCreated, buildPostResponse(post, userID))
}

// Création d'un commentaire
func CreateComment(w http.ResponseWriter, r *http.Request) {
	// Vérification de l'utilisateur authentifié
	userID, ok := getAuthenticatedUserID(w, r)
	// Vérification si l'utilisateur est authentifié
	if !ok {
		return
	}
	// Vérification de l'ID du post
	postID, ok := parseIDParam(w, r, "id")
	// Vérification si l'ID du post est valide
	if !ok {
		return
	}
	// Récupération de la requête
	var req createCommentRequest
	// Vérification si la requête est valide
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "json invalide")
		return
	}
	// Traitement du contenu du commentaire
	req.Content = strings.TrimSpace(req.Content)
	// Vérification de la longueur du contenu du commentaire
	if len(req.Content) < 2 {
		writeError(w, http.StatusBadRequest, "le commentaire est trop court")
		return
	}

	// Récupération du post
	var post models.Post
	// Vérification si une erreur est survenue lors de la récupération du post
	if err := database.DB.First(&post, postID).Error; err != nil {
		// Vérification si le post n'existe pas
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Vérification si le post n'existe pas
			writeError(w, http.StatusNotFound, "post introuvable")
			return
		}
		// Vérification si une erreur est survenue lors de la vérification du post
		writeError(w, http.StatusInternalServerError, "erreur lors de la verification du post")
		return
	}

	// Création du commentaire
	comment := models.Comment{
		Content: req.Content,
		UserID:  uint(userID),
		PostID:  post.ID,
		Status:  "published",
	}

	// Vérification si une erreur est survenue lors de la création du commentaire
	if err := database.DB.Create(&comment).Error; err != nil {
		// Vérification si une erreur est survenue lors de la création du commentaire
		writeError(w, http.StatusInternalServerError, "erreur lors de la creation du commentaire")
		return
	}

	// Vérification si une erreur est survenue lors de la récupération du commentaire
	if err := database.DB.Preload("User").First(&comment, comment.ID).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "commentaire cree mais impossible a relire")
		return
	}

	// Envoi de la réponse
	writeJSON(w, http.StatusCreated, commentResponse{
		ID:        comment.ID,
		Content:   comment.Content,
		Author:    comment.User.Username,
		UserID:    comment.UserID,
		CreatedAt: comment.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

// Voter pour un post
func VotePost(w http.ResponseWriter, r *http.Request) {
	// Vérification de l'utilisateur authentifié
	userID, ok := getAuthenticatedUserID(w, r)
	// Vérification si l'utilisateur est authentifié
	if !ok {
		return
	}
	// Vérification de l'ID du post
	postID, ok := parseIDParam(w, r, "id")
	// Vérification si l'ID du post est valide
	if !ok {
		return
	}
	// Récupération de la requête
	var req voteRequest
	// Vérification si la requête est valide
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "json invalide")
		return
	}
	// Récupération du post
	var post models.Post
	// Vérification si une erreur est survenue lors de la récupération du post
	if err := database.DB.First(&post, postID).Error; err != nil {
		// Vérification si le post n'existe pas
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "post introuvable")
			return
		}
		// Vérification si une erreur est survenue lors de la vérification du post
		writeError(w, http.StatusInternalServerError, "erreur lors de la verification du post")
		return
	}
	// Création du vote
	vote := models.Like{
		UserID: uint(userID),
		PostID: post.ID,
		IsLike: req.IsLike,
	}
	// Vérification si une erreur est survenue lors de la création du vote
	err := database.DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "post_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"is_like"}),
	}).Create(&vote).Error
	// Vérification si une erreur est survenue lors de l'enregistrement du vote
	if err != nil {
		// Vérification si une erreur est survenue lors de l'enregistrement du vote
		writeError(w, http.StatusInternalServerError, "erreur lors de l'enregistrement du vote")
		return
	}
	// Récupération des likes et des dislikes
	likes, dislikes := countVotes(post.ID)
	// Envoi de la réponse
	writeJSON(w, http.StatusOK, map[string]int64{
		"likes_count":    likes,
		"dislikes_count": dislikes,
		"score":          likes - dislikes,
	})
}

// Toggle favorite d'un post
func ToggleFavorite(w http.ResponseWriter, r *http.Request) {
	userID, ok := getAuthenticatedUserID(w, r)
	if !ok {
		return
	}

	postID, ok := parseIDParam(w, r, "id")
	if !ok {
		return
	}

	var post models.Post
	if err := database.DB.First(&post, postID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "post introuvable")
			return
		}
		writeError(w, http.StatusInternalServerError, "erreur lors de la verification du post")
		return
	}

	var favorite models.Favorite
	err := database.DB.Where("user_id = ? AND post_id = ?", userID, post.ID).First(&favorite).Error
	if err == nil {
		if err := database.DB.Delete(&favorite).Error; err != nil {
			writeError(w, http.StatusInternalServerError, "erreur lors de la suppression du favori")
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"is_favorite": false})
		return
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		writeError(w, http.StatusInternalServerError, "erreur lors de la lecture du favori")
		return
	}

	favorite = models.Favorite{UserID: uint(userID), PostID: post.ID}
	if err := database.DB.Create(&favorite).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "erreur lors de l'ajout du favori")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"is_favorite": true})
}

// Liste des tags
func ListTags(w http.ResponseWriter, r *http.Request) {
	var tags []models.Tag
	// Vérification si une erreur est survenue lors de la récupération des tags
	if err := database.DB.Order("name ASC").Find(&tags).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "erreur lors de la recuperation des tags")
		return
	}
	// Construction des réponses des tags
	responses := make([]tagResponse, 0, len(tags))
	for _, tag := range tags {
		responses = append(responses, tagResponse{ID: tag.ID, Name: tag.Name})
	}
	// Envoi de la réponse
	writeJSON(w, http.StatusOK, map[string]any{"tags": responses})
}

// Construction d'une réponse pour un post
func buildPostResponse(post models.Post, userID int) postResponse {
	// Récupération des likes et des dislikes
	likes, dislikes := countVotes(post.ID)
	// Récupération du nombre de commentaires
	var commentsCount int64
	database.DB.Model(&models.Comment{}).Where("post_id = ?", post.ID).Count(&commentsCount)
	isFavorite := false
	if userID > 0 {
		isFavorite = isPostFavorite(uint(userID), post.ID)
	}
	// Construction des réponses des tags
	tags := make([]tagResponse, 0, len(post.Tags))
	for _, tag := range post.Tags {
		tags = append(tags, tagResponse{ID: tag.ID, Name: tag.Name})
	}
	// Construction de la réponse pour le post
	return postResponse{
		ID:            post.ID,
		Title:         post.Title,
		Content:       post.Content,
		ImageURL:      post.ImageUrl,
		Author:        post.User.Username,
		UserID:        post.UserID,
		Status:        post.Status,
		CreatedAt:     post.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		Tags:          tags,
		LikesCount:    likes,
		DislikesCount: dislikes,
		CommentsCount: commentsCount,
		Score:         likes - dislikes,
		IsFavorite:    isFavorite,
	}
}

// Création de tags
func findOrCreateTags(input []string) ([]models.Tag, error) {
	names := make([]string, 0, len(input))
	seen := map[string]bool{}

	// Traitement des noms de tags
	for _, rawName := range input {
		name := normalizeTagName(rawName)
		// Vérification si le nom de tag est vide ou déjà vu
		if name == "" || seen[name] {
			// Vérification si le nom de tag est vide ou déjà vu
			continue
		}
		// Vérification si le nom de tag est trop long
		if len(name) > 27 {
			return nil, errors.New("un tag ne doit pas depasser 27 caracteres")
		}
		// Ajout du nom de tag à la liste des noms de tags vus
		seen[name] = true
		// Ajout du nom de tag à la liste des noms de tags
		names = append(names, name)
	}
	// Vérification si le post a au moins deux tags
	if len(names) < 2 {
		return nil, errors.New("un post doit avoir au moins deux tags")
	}
	// Vérification si le post a plus de 5 tags
	if len(names) > maxPostTags {
		return nil, errors.New("un post ne peut pas avoir plus de 5 tags")
	}
	// Construction des tags
	tags := make([]models.Tag, 0, len(names))
	for _, name := range names {
		// Création du tag
		tag := models.Tag{Name: name}
		// Vérification si une erreur est survenue lors de la création du tag
		err := database.DB.Where(models.Tag{Name: name}).FirstOrCreate(&tag).Error
		// Vérification si une erreur est survenue lors de la création du tag
		if err != nil {
			return nil, errors.New("erreur lors de la preparation des tags")
		}
		tags = append(tags, tag)
	}
	// Retour des tags
	return tags, nil
}

// Normalisation du nom d'un tag
func normalizeTagName(name string) string {
	name = strings.TrimSpace(strings.ToLower(name))
	name = strings.TrimPrefix(name, "#")
	return name
}

// Compte les likes et les dislikes pour un post
func countVotes(postID uint) (int64, int64) {
	var likes int64
	var dislikes int64

	database.DB.Model(&models.Like{}).Where("post_id = ? AND is_like = ?", postID, true).Count(&likes)
	database.DB.Model(&models.Like{}).Where("post_id = ? AND is_like = ?", postID, false).Count(&dislikes)

	return likes, dislikes
}

func isPostFavorite(userID uint, postID uint) bool {
	var favorite models.Favorite
	err := database.DB.Where("user_id = ? AND post_id = ?", userID, postID).First(&favorite).Error
	return err == nil
}

func getUserIDFromContext(r *http.Request) (int, bool) {
	userID, ok := r.Context().Value("user_id").(int)
	if !ok || userID <= 0 {
		return 0, false
	}
	return userID, true
}

// Parsing de l'ID d'un paramètre
func parseIDParam(w http.ResponseWriter, r *http.Request, name string) (uint, bool) {
	rawID := mux.Vars(r)[name]
	id, err := strconv.ParseUint(rawID, 10, 64)
	// Vérification si l'ID est invalide
	if err != nil || id == 0 {
		// Vérification si l'ID est invalide
		writeError(w, http.StatusBadRequest, "identifiant invalide")
		return 0, false
	}
	// Retour de l'ID
	return uint(id), true
}

// Récupération de l'ID de l'utilisateur authentifié
func getAuthenticatedUserID(w http.ResponseWriter, r *http.Request) (int, bool) {
	userID, ok := r.Context().Value("user_id").(int)
	if !ok || userID <= 0 {
		writeError(w, http.StatusUnauthorized, "utilisateur non authentifié")
		return 0, false
	}
	// Retour de l'ID de l'utilisateur authentifié
	return userID, true
}

// Envoi d'une réponse JSON
func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}

// Envoi d'une erreur JSON
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"message": message})
}
