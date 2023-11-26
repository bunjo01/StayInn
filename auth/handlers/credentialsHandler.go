package handlers

import (
	"auth/data"
	"encoding/json"
	"fmt"
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

	token, err := ch.repo.GenerateToken(credentials.Username, credentials.Role)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}

// Handler method for registration
func (ch *CredentialsHandler) Register(w http.ResponseWriter, r *http.Request) {
	var newUser data.NewUser
	if err := json.NewDecoder(r.Body).Decode(&newUser); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	err := ch.repo.RegisterUser(newUser.Username, newUser.Password, newUser.FirstName, newUser.LastName,
		newUser.Email, newUser.Address, newUser.Role)
	if err != nil && err.Error() == "username already exists" {
		http.Error(w, "Username is not unique!", http.StatusBadRequest)
		return
	} else if err != nil && err.Error() == "choose a more secure password" {
		http.Error(w, "Password did not pass the security check. Pick a stronger password", http.StatusBadRequest)
		return
	} else if err != nil {
		http.Error(w, "Failed to register new user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (ch *CredentialsHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	var reqBody data.ChangePasswordRequest

	// Parsiranje podataka iz zahtjeva
	err := reqBody.FromJSON(r.Body)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Provjera da li su svi potrebni podaci prisutni
	if reqBody.Username == "" || reqBody.OldPassword == "" || reqBody.NewPassword == "" {
		http.Error(w, "Missing username, old password, or new password", http.StatusBadRequest)
		return
	}

	// Pozivanje metode ChangePassword iz CredentialsRepo-a
	err = ch.repo.ChangePassword(reqBody.Username, reqBody.OldPassword, reqBody.NewPassword)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to change password: %v", err), http.StatusBadRequest)
		return
	}

	// Uspješno promijenjena lozinka
	w.WriteHeader(http.StatusOK)
}
