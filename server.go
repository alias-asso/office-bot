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

	mux.HandleFunc("POST /", basicAuth(handleToggle))
}

func basicAuth(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract the username and password from the request
		// Authorization header. If no Authentication header is present
		// or the header value is invalid, then the 'ok' return value
		// will be false.
		username, password, ok := r.BasicAuth()
		if ok {
			// Calculate SHA-256 hashes for the provided and expected
			// usernames and passwords.
			usernameHash := sha256.Sum256([]byte(username))
			usernameHashHex := hex.EncodeToString(usernameHash[:])
			passwordHash := sha256.Sum256([]byte(password))
			passwordHashHex := hex.EncodeToString(passwordHash[:])
			expectedUsernameHash := []byte(config.UsernameHash)
			expectedPasswordHash := []byte(config.PasswordHash)

			// Use the subtle.ConstantTimeCompare() function to check if
			// the provided username and password hashes equal the
			// expected username and password hashes. ConstantTimeCompare
			// will return 1 if the values are equal, or 0 otherwise.
			// Importantly, we should to do the work to evaluate both the
			// username and password before checking the return values to
			// avoid leaking information.
			usernameMatch := (subtle.ConstantTimeCompare([]byte(usernameHashHex), expectedUsernameHash[:]) == 1)
			passwordMatch := (subtle.ConstantTimeCompare([]byte(passwordHashHex), expectedPasswordHash[:]) == 1)

			// If the username and password are correct, then call
			// the next handler in the chain. Make sure to return
			// afterwards, so that none of the code below is run.
			if usernameMatch && passwordMatch {
				next.ServeHTTP(w, r)
				return
			}
		}

		// If the Authentication header is not present, is invalid, or the
		// username or password is wrong, then set a WWW-Authenticate
		// header to inform the client that we expect them to use basic
		// authentication and send a 401 Unauthorized response.
		w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
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
