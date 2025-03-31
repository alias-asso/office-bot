package main

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"net/http"
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
				next.ServeHTTP(w, r)
				return
			}
		}

		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	})

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
	fmt.Fprint(w, "new status: "+newStatus)
}
