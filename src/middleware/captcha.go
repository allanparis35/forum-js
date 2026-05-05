package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"FORUM-js/src/utils"
)


func CaptchaMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "erreur lecture body", http.StatusBadRequest)
			return
		}

		r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		var data map[string]interface{}
		if err := json.Unmarshal(bodyBytes, &data); err != nil {
			http.Error(w, "json invalide", http.StatusBadRequest)
			return
		}

		token, ok := data["captcha"].(string)
		if !ok || token == "" {
			http.Error(w, "captcha manquant", http.StatusBadRequest)
			return
		}

		valid, err := utils.VerifyCaptcha(token)
		if err != nil || !valid {
			http.Error(w, "captcha invalide", http.StatusForbidden)
			return
		}

		next(w, r)
	}
}