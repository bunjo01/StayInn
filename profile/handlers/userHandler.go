package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"profile/data"

	"github.com/dgrijalva/jwt-go"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type KeyProduct struct{}

type UserHandler struct {
	logger *log.Logger
	repo   *data.UserRepo
}

var secretKey = []byte("stayinn_secret")

// Injecting the logger makes this code much more testable
func NewUserHandler(l *log.Logger, r *data.UserRepo) *UserHandler {
	return &UserHandler{l, r}
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
	id, err := primitive.ObjectIDFromHex(vars["id"])
	if err != nil {
		http.Error(rw, "Invalid ID", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	user, err := uh.repo.GetUser(ctx, id)
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

func (uh *UserHandler) UpdateUser(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := primitive.ObjectIDFromHex(vars["id"])
	if err != nil {
		http.Error(rw, "Invalid ID", http.StatusBadRequest)
		return
	}

	var updatedUser data.NewUser
	if err := json.NewDecoder(r.Body).Decode(&updatedUser); err != nil {
		http.Error(rw, "Failed to decode request body", http.StatusBadRequest)
		return
	}

	updatedUser.ID = id
	if err := uh.repo.UpdateUser(r.Context(), &updatedUser); err != nil {
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
	id, err := primitive.ObjectIDFromHex(vars["id"])
	if err != nil {
		http.Error(rw, "Invalid ID", http.StatusBadRequest)
		return
	}

	if err := uh.repo.DeleteUser(r.Context(), id); err != nil {
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
