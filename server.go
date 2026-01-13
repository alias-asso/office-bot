package main

import (
	"bytes"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
)

var mux *http.ServeMux

func init() {
	mux = http.NewServeMux()

	mux.HandleFunc("GET /", handleStatus)

	mux.HandleFunc("GET /toggle", urlAuth(handleToggle))
}

func urlAuth(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()

		password := query.Get("p")

		if password != "" {
			passwordHash := sha256.Sum256([]byte(password))
			passwordHashHex := hex.EncodeToString(passwordHash[:])
			expectedPasswordHash := []byte(config.PasswordHash)

			passwordMatch := (subtle.ConstantTimeCompare([]byte(passwordHashHex), expectedPasswordHash[:]) == 1)

			if passwordMatch {
				if checkUserAgentBlacklist(r.UserAgent()) {
					next.ServeHTTP(w, r)
					return
				}
			}
		}

		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	})

}

func checkUserAgentBlacklist(userAgent string) bool {
	lowerUa := strings.ToLower(userAgent)
	for _, term := range config.BlacklistedUserAgents {
		lowerTerm := strings.ToLower(term)
		if strings.Contains(lowerUa, lowerTerm) {
			return false
		}
	}
	return true
}

func handleStatus(w http.ResponseWriter, req *http.Request) {
	status, err := GetStatusString()
	if err != nil {
		fmt.Fprint(w, "error: "+err.Error())
	}
	fmt.Fprint(w, status)
}

func handleToggle(w http.ResponseWriter, req *http.Request) {
	newStatus, err := ToggleStatus()
	if err != nil {
		fmt.Fprint(w, "error")
	}
	webhookJson := []byte("{\"content\": \"**[WEB]** Le statut du local est maintenant : **" + newStatus + "**\"}")
	http.Post("https://discord.com/api/webhooks/"+config.WebhookId+"/"+config.WebhookToken, "application/json", bytes.NewBuffer(webhookJson))
	w.Header().Set("Cache-Control", "no-store")
	http.Redirect(w, req, "/", http.StatusPermanentRedirect)
}
