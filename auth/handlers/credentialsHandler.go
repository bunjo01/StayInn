package handlers

import (
	"auth/data"
	"encoding/json"
	"log"
	"net/http"
)

type KeyProduct struct{}

type CredentialsHandler struct {
	logger *log.Logger
	repo   *data.CredentialsRepo
}

// Injecting the logger makes this code much more testable
func NewCredentialsHandler(l *log.Logger, r *data.CredentialsRepo) *CredentialsHandler {
	return &CredentialsHandler{l, r}
}

// TODO Handler methods

func (ch *CredentialsHandler) Login(w http.ResponseWriter, r *http.Request) {
    var credentials data.Credentials
    if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    if err := ch.repo.ValidateCredentials(credentials.Username, credentials.Password); err != nil {
        http.Error(w, "Invalid username or password", http.StatusUnauthorized)
        return
    }

    w.WriteHeader(http.StatusOK)
    w.Write([]byte("Successfully logged in"))
}
