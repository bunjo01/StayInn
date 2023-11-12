package handlers

import (
	"auth/data"
	"encoding/json"
	"errors"
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

// Handler method for registration
func (ch *CredentialsHandler) Register(w http.ResponseWriter, r *http.Request) {
	var newUser data.NewUser
	if err := json.NewDecoder(r.Body).Decode(&newUser); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	err := ch.repo.RegisterUser(newUser.Username, newUser.Password, newUser.FirstName, newUser.LastName,
		newUser.Email, newUser.Address)
	if err != nil && errors.Is(err, errors.New("username already exists")) {
		http.Error(w, "Username not unique", http.StatusBadRequest)
		return
	} else if err != nil {
		http.Error(w, "Failed to register new user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Successfully registered"))
}
