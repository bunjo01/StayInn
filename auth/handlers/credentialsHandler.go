package handlers

import (
	"auth/clients"
	"auth/data"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

var secretKey = []byte("stayinn_secret")

type KeyProduct struct{}

type CredentialsHandler struct {
	repo    *data.CredentialsRepo
	profile clients.ProfileClient
}

const InvalidRequestBody = "Invalid request body"

const (
	INTENTION_ACTIVATION = "activation"
	errorDeletingUser    = "Error deleting user with email "
)

// Injecting the logger makes this code much more testable
func NewCredentialsHandler(r *data.CredentialsRepo, p clients.ProfileClient) *CredentialsHandler {
	return &CredentialsHandler{r, p}
}

// TODO Handler methods

func (ch *CredentialsHandler) Login(w http.ResponseWriter, r *http.Request) {
	var credentials data.Credentials
	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		http.Error(w, InvalidRequestBody, http.StatusBadRequest)
		return
	}

	log.Info(fmt.Sprintf("[auth-handler]ah#10 Received login request from '%s' for user '%s'", r.RemoteAddr, credentials.Username))

	dbUser, err := ch.repo.FindUserByUsername(credentials.Username)
	if err != nil {
		http.Error(w, "User not found with username: "+credentials.Username, http.StatusBadRequest)
		log.Error(fmt.Sprintf("[auth-handler]ah#11 Failed to log in '%s'", credentials.Username))
		return
	}

	if err := ch.repo.ValidateCredentials(credentials.Username, credentials.Password); err != nil {
		if err.Error() == "account not activated" {
			http.Error(w, "Account not activated", http.StatusForbidden)
			log.Error(fmt.Sprintf("[auth-handler]ah#12 Failed to log in '%s'", credentials.Username))
			return
		}
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		log.Error(fmt.Sprintf("[auth-handler]ah#13 Failed to log in '%s'", credentials.Username))
		return
	}

	token, err := ch.repo.GenerateToken(credentials.Username, dbUser.Role)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		log.Error(fmt.Sprintf("[auth-handler]ah#14 Failed to generate token for '%s'", credentials.Username))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": token})

	log.Info(fmt.Sprintf("[auth-handler]ah#15 User '%s' successfully logged in from '%s'", credentials.Username, r.RemoteAddr))
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

	log.Info(fmt.Sprintf("[auth-handler]ah#16 User '%s' changing username to '%s'", oldUsername, username))

	if err := ch.repo.ChangeUsername(r.Context(), oldUsername, username); err != nil {
		http.Error(w, fmt.Sprintf("Failed to change username: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	log.Info(fmt.Sprintf("[auth-handler]ah#17 User '%s' successfully changed username to '%s'", oldUsername, username))
}

func (ch *CredentialsHandler) UpdateEmail(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	oldEmail := vars["oldEmail"]
	email := vars["email"]

	log.Info(fmt.Sprintf("[auth-handler]ah#1 User '%s' changing email to '%s'", oldEmail, email))

	if err := ch.repo.ChangeEmail(r.Context(), oldEmail, email); err != nil {
		http.Error(w, fmt.Sprintf("Failed to change email: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	log.Info(fmt.Sprintf("[auth-handler]ah#18 User '%s' successfully changed email to '%s'", oldEmail, email))
}

// Handler method for registration
func (ch *CredentialsHandler) Register(w http.ResponseWriter, r *http.Request) {
	log.Info(fmt.Sprintf("[auth-handler]ah#19 Registering new user"))
	tokenStr := ch.extractTokenFromHeader(r)
	var newUser data.NewUser
	if err := json.NewDecoder(r.Body).Decode(&newUser); err != nil {
		http.Error(w, InvalidRequestBody, http.StatusBadRequest)
		log.Error(fmt.Sprintf("[auth-handler]ah#20 Failed to register new user: invalid request body"))
		return
	}

	err := ch.repo.RegisterUser(newUser.Username, newUser.Password, newUser.FirstName, newUser.LastName,
		newUser.Email, newUser.Address, newUser.Role)
	if err != nil && err.Error() == "username already exists" {
		http.Error(w, "Username is not unique!", http.StatusBadRequest)
		log.Error(fmt.Sprintf("[auth-handler]ah#21 Failed to register new user: username '%s' is not unique", newUser.Username))
		return
	} else if err != nil && err.Error() == "choose a more secure password" {
		http.Error(w, "Password did not pass the security check. Pick a stronger password", http.StatusBadRequest)
		log.Error(fmt.Sprintf("[auth-handler]ah#22 Failed to register new user: weak password"))
		return
	} else if err != nil {
		http.Error(w, "Failed to register new user", http.StatusInternalServerError)
		log.Error(fmt.Sprintf("[auth-handler]ah#23 Failed to register new user: internal server error"))
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5000*time.Millisecond)
	defer cancel()
	_, err = ch.profile.PassInfoToProfileService(ctx, newUser, tokenStr)
	if err != nil {
		log.Error(fmt.Sprintf("[auth-handler]ah#2 Error while passing info to profile service: %v", err))
		ch.repo.DeleteUser(r.Context(), newUser.Username)
		writeResp(err, http.StatusServiceUnavailable, w)
		return
	}

	w.WriteHeader(http.StatusCreated)
	log.Info(fmt.Sprintf("[auth-handler]ah#24 Successfully registered new user"))
}

func (ch *CredentialsHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	var reqBody data.ChangePasswordRequest

	log.Info(fmt.Sprintf("[auth-handler]ah#25 Recieved request from '%s' to change password", reqBody.Username))

	err := reqBody.FromJSON(r.Body)
	if err != nil {
		http.Error(w, InvalidRequestBody, http.StatusBadRequest)
		log.Error(fmt.Sprintf("[auth-handler]ah#26 Failed to change password: invalid request body"))
		return
	}
	if reqBody.Username == "" || reqBody.CurrentPassword == "" || reqBody.NewPassword == "" {
		http.Error(w, "Missing username, old password, or new password", http.StatusBadRequest)
		log.Error(fmt.Sprintf("[auth-handler]ah#27 Failed to change password: request body contains empty values"))
		return
	}

	err = ch.repo.ChangePassword(reqBody.Username, reqBody.CurrentPassword, reqBody.NewPassword)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to change password: %v", err), http.StatusBadRequest)
		log.Error(fmt.Sprintf("[auth-handler]ah#28 Failed to change password: bad request"))
		return
	}

	w.WriteHeader(http.StatusOK)
	log.Info(fmt.Sprintf("[auth-handler]ah#29 Successfully changed password for user '%s'", reqBody.Username))
}

func (ch *CredentialsHandler) ActivateAccount(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	activationUUID := params["activationUUID"]

	log.Info(fmt.Sprintf("[auth-handler]ah#30 Activating user account for '%s'", activationUUID))

	// Activating user account
	err := ch.repo.ActivateUserAccount(activationUUID)
	if err != nil {
		if err.Error() == "link for activation has expired" {
			log.Error(fmt.Sprintf("[auth-handler]ah#3 Error during account activation: %v", err))
			http.Error(w, "Link for activation has expired", http.StatusGone)
			return
		}
		http.Error(w, "Failed to activate user account", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("User account successfully activated"))
	log.Info(fmt.Sprintf("[auth-handler]ah#31 Successfully activated user account for '%s'", activationUUID))
}

func (ch *CredentialsHandler) SendRecoveryEmail(w http.ResponseWriter, r *http.Request) {
	var requestBody struct {
		Email string `json:"email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, InvalidRequestBody, http.StatusBadRequest)
		return
	}

	log.Info(fmt.Sprintf("[auth-handler]ah#32 Sending recovery email to '%s'", requestBody.Email))

	recoveryUUID, err := ch.repo.SendRecoveryEmail(requestBody.Email)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to send recovery mail: %v", err), http.StatusBadRequest)
		log.Error(fmt.Sprintf("[auth-handler]ah#33 Failed to send recovery email to '%s': %v", requestBody.Email, err))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("Recovery email sent successfully with UUID: %s", recoveryUUID)))
	log.Info(fmt.Sprintf("[auth-handler]ah#34 Recovery email sent successfully with UUID: %s", recoveryUUID))
}

func (ch *CredentialsHandler) UpdatePasswordWithRecoveryUUID(w http.ResponseWriter, r *http.Request) {
	var reqBody struct {
		RecoveryUUID string `json:"recoveryUUID"`
		NewPassword  string `json:"newPassword"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, InvalidRequestBody, http.StatusBadRequest)
		return
	}

	log.Info(fmt.Sprintf("[auth-handler]ah#35 Updating password with recovery uuid '%s'", reqBody.RecoveryUUID))

	err := ch.repo.UpdatePasswordWithRecoveryUUID(reqBody.RecoveryUUID, reqBody.NewPassword)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %s", err), http.StatusBadRequest)
		log.Error(fmt.Sprintf("[auth-handler]ah#36 Failed to update password with recovery: %v", err))
		return
	}

	w.WriteHeader(http.StatusOK)
	log.Info(fmt.Sprintf("[auth-handler]ah#37 Successfully updated password with recovery uuid '%s'", reqBody.RecoveryUUID))
}

func (ch *CredentialsHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username := vars["username"]

	log.Info(fmt.Sprintf("[auth-handler]ah#38 Recieved request to delete user '%s'", username))

	if err := ch.repo.DeleteUser(r.Context(), username); err != nil {
		log.Error(fmt.Sprintf("[auth-handler]ah#4 Failed to delete user '%s': %v", username, err))
		http.Error(w, "Failed to delete user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
	log.Info(fmt.Sprintf("[auth-handler]ah#39 Successfully deleted user '%s'", username))
}

func (ch *CredentialsHandler) AuthorizeRoles(allowedRoles ...string) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, rr *http.Request) {
			tokenString := ch.extractTokenFromHeader(rr)
			if tokenString == "" {
				log.Warning(fmt.Sprintf("[auth-handler]ah#5 No token found in request from '%s'", rr.RemoteAddr))
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			claims := jwt.MapClaims{}
			token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
				return secretKey, nil
			})

			if err != nil || !token.Valid {
				log.Warning(fmt.Sprintf("[auth-handler]ah#6 Invalid signature token found in request from '%s'", rr.RemoteAddr))
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			username, ok1 := claims["username"].(string)
			role, ok2 := claims["role"].(string)
			if !ok1 {
				log.Warning(fmt.Sprintf("[auth-handler]ah#7 Username not found in token in request from '%s'", rr.RemoteAddr))
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			if !ok2 {
				log.Warning(fmt.Sprintf("[auth-handler]ah#8 Role not found in token in request from '%s'", rr.RemoteAddr))
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			for _, allowedRole := range allowedRoles {
				if allowedRole == role {
					next.ServeHTTP(w, rr)
					return
				}
			}

			log.Warning(fmt.Sprintf("[auth-handler]ah#9 User '%s' from '%s' tried to do unauthorized action", username, rr.RemoteAddr))
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
