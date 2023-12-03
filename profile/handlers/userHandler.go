package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"profile/clients"
	"profile/data"
	"time"

	"github.com/dgrijalva/jwt-go"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type KeyProduct struct{}

type UserHandler struct {
	logger        *log.Logger
	repo          *data.UserRepo
	accommodation clients.AccommodationClient
	auth          clients.AuthClient
	reservation   clients.ReservationClient
}

var secretKey = []byte("stayinn_secret")

// Injecting the logger makes this code much more testable
func NewUserHandler(l *log.Logger, r *data.UserRepo, ac clients.AccommodationClient,
	au clients.AuthClient, re clients.ReservationClient) *UserHandler {
	return &UserHandler{l, r, ac, au, re}
}

// Handler methods

func (uh *UserHandler) GetAllUsers(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	users, err := uh.repo.GetAllUsers(ctx)
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

func (uh *UserHandler) GetUser(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username := vars["username"]

	ctx := r.Context()
	user, err := uh.repo.GetUser(ctx, username)
	if err != nil {
		http.Error(rw, "Failed to retrieve user", http.StatusInternalServerError)
		return
	}

	if user == nil {
		http.NotFound(rw, r)
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(rw).Encode(user); err != nil {
		http.Error(rw, "Failed to encode user", http.StatusInternalServerError)
	}
}

func (uh *UserHandler) CreateUser(rw http.ResponseWriter, r *http.Request) {
	var user data.NewUser
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		uh.logger.Println("Failed to decode body:", err)
		http.Error(rw, "Failed to decode request body", http.StatusBadRequest)
		return
	}

	user.ID = primitive.NewObjectID()

	err := uh.repo.CreateProfileDetails(r.Context(), &user)
	if err != nil {
		uh.logger.Println("Failed to create user:", err)
		http.Error(rw, "Failed to create user", http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(rw).Encode(user); err != nil {
		uh.logger.Println("Failed to encode user:", err)
		http.Error(rw, "Failed to encode user", http.StatusInternalServerError)
	}
}

func (uh *UserHandler) CheckUsernameAvailability(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username := vars["username"]

	available, err := uh.repo.CheckUsernameAvailability(r.Context(), username)
	if err != nil {
		uh.logger.Println("Error checking username availability:", err)
		http.Error(w, "Failed to check username availability", http.StatusInternalServerError)
		return
	}

	response := struct {
		Available bool `json:"available"`
	}{
		Available: available,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		uh.logger.Println("Failed to encode JSON response:", err)
		http.Error(w, "Failed to encode JSON response", http.StatusInternalServerError)
	}
}

func (uh *UserHandler) UpdateUser(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username := vars["username"]

	var updatedUser data.NewUser
	if err := json.NewDecoder(r.Body).Decode(&updatedUser); err != nil {
		http.Error(rw, "Failed to decode request body", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5000*time.Millisecond)
	defer cancel()
	_, err := uh.auth.PassUsernameToAuthService(ctx, username, updatedUser.Username)
	if err != nil {
		uh.logger.Println(err)
		writeResp(err, http.StatusServiceUnavailable, rw)
		return
	}

	if err := uh.repo.UpdateUser(r.Context(), username, &updatedUser); err != nil {
		uh.logger.Println("Failed to update user:", err)
		http.Error(rw, "Failed to update user", http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(rw).Encode(updatedUser); err != nil {
		uh.logger.Println("Failed to encode updated user:", err)
		http.Error(rw, "Failed to encode updated user", http.StatusInternalServerError)
	}
}

func (uh *UserHandler) DeleteUser(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username := vars["username"]

	if err := uh.repo.DeleteUser(r.Context(), username); err != nil {
		uh.logger.Println("Failed to delete user:", err)
		http.Error(rw, "Failed to delete user", http.StatusInternalServerError)
		return
	}

	rw.WriteHeader(http.StatusNoContent)
}

func (uh *UserHandler) AuthorizeRoles(allowedRoles ...string) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, rr *http.Request) {
			if rr.URL.Path == "/users" && rr.Method == http.MethodPost {
				next.ServeHTTP(w, rr)
				return
			}

			tokenString := uh.extractTokenFromHeader(rr)
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

func (uh *UserHandler) extractTokenFromHeader(rr *http.Request) string {
	token := rr.Header.Get("Authorization")
	if token != "" {
		return token[len("Bearer "):]
	}
	return ""
}
