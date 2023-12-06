package handlers

import (
	"auth/clients"
	"auth/data"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
)

var secretKey = []byte("stayinn_secret")

type KeyProduct struct{}

type CredentialsHandler struct {
	logger  *log.Logger
	repo    *data.CredentialsRepo
	profile clients.ProfileClient
}

const (
	INTENTION_ACTIVATION = "activation"
	errorDeletingUser    = "Error deleting user with email "
)

// Injecting the logger makes this code much more testable
func NewCredentialsHandler(l *log.Logger, r *data.CredentialsRepo, p clients.ProfileClient) *CredentialsHandler {
	return &CredentialsHandler{l, r, p}
}

// TODO Handler methods

func (ch *CredentialsHandler) Login(w http.ResponseWriter, r *http.Request) {
	var credentials data.Credentials
	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	dbUser, err := ch.repo.FindUserByUsername(credentials.Username)
	if err != nil {
		http.Error(w, "User not found with username: "+credentials.Username, http.StatusBadRequest)
		return
	}

	if err := ch.repo.ValidateCredentials(credentials.Username, credentials.Password); err != nil {
		if err.Error() == "account not activated" {
			http.Error(w, "Account not activated", http.StatusForbidden)
			return
		}
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	token, err := ch.repo.GenerateToken(credentials.Username, dbUser.Role)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}

func (ch *CredentialsHandler) GetAllUsers(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	users, err := ch.repo.GetAllCredentials(ctx)
	if err != nil {
		http.Error(rw, "Failed to retrieve users", http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(rw).Encode(users); err != nil {
		http.Error(rw, "Failed to encode users", http.StatusInternalServerError)
	}
}

func (ch *CredentialsHandler) UpdateUsername(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	oldUsername := vars["oldUsername"]
	username := vars["username"]

	if err := ch.repo.ChangeUsername(r.Context(), oldUsername, username); err != nil {
		http.Error(w, fmt.Sprintf("Failed to change username: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// Handler method for registration
func (ch *CredentialsHandler) Register(w http.ResponseWriter, r *http.Request) {
	var newUser data.NewUser
	if err := json.NewDecoder(r.Body).Decode(&newUser); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5000*time.Millisecond)
	defer cancel()
	_, err := ch.profile.PassInfoToProfileService(ctx, newUser)
	if err != nil {
		ch.logger.Println(err)
		writeResp(err, http.StatusServiceUnavailable, w)
		return
	}

	err = ch.repo.RegisterUser(newUser.Username, newUser.Password, newUser.FirstName, newUser.LastName,
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

	w.WriteHeader(http.StatusCreated)
}

func (ch *CredentialsHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	var reqBody data.ChangePasswordRequest

	err := reqBody.FromJSON(r.Body)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if reqBody.Username == "" || reqBody.CurrentPassword == "" || reqBody.NewPassword == "" {
		http.Error(w, "Missing username, old password, or new password", http.StatusBadRequest)
		return
	}

	err = ch.repo.ChangePassword(reqBody.Username, reqBody.CurrentPassword, reqBody.NewPassword)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to change password: %v", err), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (ch *CredentialsHandler) ActivateAccount(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	activationUUID := params["activationUUID"]

	// Activating user account
	err := ch.repo.ActivateUserAccount(activationUUID)
	if err != nil {
		if err.Error() == "link for activation has expired" {
			ch.logger.Printf("Error during activation: %v", err)
			http.Error(w, "Link for activation has expired", http.StatusGone)
			return
		}
		http.Error(w, "Failed to activate user account", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("User account successfully activated"))
}

func (ch *CredentialsHandler) SendRecoveryEmail(w http.ResponseWriter, r *http.Request) {
	var requestBody struct {
		Email string `json:"email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	recoveryUUID, err := ch.repo.SendRecoveryEmail(requestBody.Email)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to send recovery mail: %v", err), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("Recovery email sent successfully with UUID: %s", recoveryUUID)))
}

func (ch *CredentialsHandler) UpdatePasswordWithRecoveryUUID(w http.ResponseWriter, r *http.Request) {
	var reqBody struct {
		RecoveryUUID string `json:"recoveryUUID"`
		NewPassword  string `json:"newPassword"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := ch.repo.UpdatePasswordWithRecoveryUUID(reqBody.RecoveryUUID, reqBody.NewPassword)
	if err != nil {
		http.Error(w, "Failed to update password with recoveryUUID", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (ch *CredentialsHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username := vars["username"]

	if err := ch.repo.DeleteUser(r.Context(), username); err != nil {
		ch.logger.Println("Failed to delete user: ", err)
		http.Error(w, "Failed to delete user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (ch *CredentialsHandler) AuthorizeRoles(allowedRoles ...string) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, rr *http.Request) {
			tokenString := ch.extractTokenFromHeader(rr)
			if tokenString == "" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			claims := jwt.MapClaims{}
			token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
				return secretKey, nil
			})

			if err != nil || !token.Valid {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			_, ok1 := claims["username"].(string)
			role, ok2 := claims["role"].(string)
			if !ok1 || !ok2 {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			for _, allowedRole := range allowedRoles {
				if allowedRole == role {
					next.ServeHTTP(w, rr)
					return
				}
			}

			http.Error(w, "Forbidden", http.StatusForbidden)
		})
	}
}

func (ch *CredentialsHandler) extractTokenFromHeader(rr *http.Request) string {
	token := rr.Header.Get("Authorization")
	if token != "" {
		return token[len("Bearer "):]
	}
	return ""
}
