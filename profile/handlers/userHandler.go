package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"profile/clients"
	"profile/data"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/dgrijalva/jwt-go"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const applicationJson = "application/json"
const contentType = "Content-Type"
const failedToEncodeUser = "Failed to encode user"
const failedToDecodeRequestBody = "Failed to decode request body"

type KeyProduct struct{}

type UserHandler struct {
	repo          *data.UserRepo
	accommodation clients.AccommodationClient
	auth          clients.AuthClient
	reservation   clients.ReservationClient
}

var secretKey = []byte("stayinn_secret")

// Injecting the logger makes this code much more testable
func NewUserHandler(r *data.UserRepo, ac clients.AccommodationClient,
	au clients.AuthClient, re clients.ReservationClient) *UserHandler {
	return &UserHandler{r, ac, au, re}
}

// Handler methods

func (uh *UserHandler) GetAllUsers(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	log.Info(fmt.Sprintf("[prof-handler]ph#37 Received request from '%s' for all users", r.RemoteAddr))

	users, err := uh.repo.GetAllUsers(ctx)
	if err != nil {
		http.Error(rw, "Failed to retrieve users", http.StatusInternalServerError)
		log.Error(fmt.Sprintf("[prof-handler]ph#1 Failed to retrieve users: %v", err))
		return
	}

	rw.Header().Set(contentType, applicationJson)
	rw.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(rw).Encode(users); err != nil {
		http.Error(rw, "Failed to encode users", http.StatusInternalServerError)
		log.Error(fmt.Sprintf("[prof-handler]ph#2 Failed to encode users: %v", err))
	}
	log.Info(fmt.Sprintf("[prof-handler]ph#38 Successfully fetched all accommodations"))
}

func (uh *UserHandler) GetUser(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username := vars["username"]

	log.Info(fmt.Sprintf("[prof-handler]ph#39 Received request from '%s' for user '%s'", r.RemoteAddr, username))

	ctx := r.Context()
	user, err := uh.repo.GetUser(ctx, username)
	if err != nil {
		http.Error(rw, "Failed to retrieve user", http.StatusInternalServerError)
		log.Error(fmt.Sprintf("[prof-handler]ph#3 Failed to retrieve user: %v", err))
		return
	}

	if user == nil {
		http.NotFound(rw, r)
		return
	}

	rw.Header().Set(contentType, applicationJson)
	rw.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(rw).Encode(user); err != nil {
		http.Error(rw, failedToEncodeUser, http.StatusInternalServerError)
		log.Error(fmt.Sprintf("[prof-handler]ph#4 Failed to encode user: %v", err))
	}
	log.Info(fmt.Sprintf("[prof-handler]ph#40 Successfully fetched user with username '%s'", username))
}

func (uh *UserHandler) GetUserById(rw http.ResponseWriter, r *http.Request) {
	var id data.UserId

	log.Info(fmt.Sprintf("[prof-handler]ph#41 Received request from '%s' for user with ID '%s'", r.RemoteAddr, id.ID.Hex()))

	if err := json.NewDecoder(r.Body).Decode(&id); err != nil {
		log.Error(fmt.Sprintf("[prof-handler]ph#5 Failed to decode request body: %v", err))
		http.Error(rw, failedToDecodeRequestBody, http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	user, err := uh.repo.GetUserById(ctx, id.ID)
	if err != nil {
		http.Error(rw, "Failed to retrieve user", http.StatusInternalServerError)
		log.Error(fmt.Sprintf("[prof-handler]ph#6 Failed to retrieve user: %v", err))
		return
	}

	if user == nil {
		http.NotFound(rw, r)
		return
	}

	rw.Header().Set(contentType, applicationJson)
	rw.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(rw).Encode(user); err != nil {
		http.Error(rw, failedToEncodeUser, http.StatusInternalServerError)
		log.Error(fmt.Sprintf("[prof-handler]ph#7 Failed to encode user: %v", err))
	}
	log.Info(fmt.Sprintf("[prof-handler]ph#42 Successfully fetched user with id '%s'", id.ID.Hex()))
}

func (uh *UserHandler) CreateUser(rw http.ResponseWriter, r *http.Request) {
	var user data.NewUser

	log.Info(fmt.Sprintf("[prof-handler]ph#8 Creating new user from '%s'", r.RemoteAddr))

	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		log.Error(fmt.Sprintf("[prof-handler]ph#9 Failed to decode request body: %v", err))
		http.Error(rw, failedToDecodeRequestBody, http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	user.ID = primitive.NewObjectID()

	avaible, err := uh.repo.CheckUsernameAvailability(ctx, user.Username)
	if !avaible || err != nil {
		log.Error(fmt.Sprintf("[prof-handler]ph#10 Username is not unique: %v", err))
		http.Error(rw, "Username is not unique!", http.StatusBadRequest)
		return
	}

	err = uh.repo.CreateProfileDetails(ctx, &user)
	if err != nil {
		log.Error(fmt.Sprintf("[prof-handler]ph#11 Failed to create user: %v", err))
		http.Error(rw, "Failed to create user", http.StatusInternalServerError)
		return
	}

	rw.Header().Set(contentType, applicationJson)
	rw.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(rw).Encode(user); err != nil {
		log.Error(fmt.Sprintf("[prof-handler]ph#12 Failed to encode user: %v", err))
		http.Error(rw, failedToEncodeUser, http.StatusInternalServerError)
	}

	log.Info(fmt.Sprintf("[prof-handler]ph#13 User successfully created with id '%s'", user.ID.Hex()))
}

func (uh *UserHandler) CheckUsernameAvailability(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username := vars["username"]

	log.Info(fmt.Sprintf("[prof-handler]ph#43 Checking username availability from '%s' for user '%s'", r.RemoteAddr, username))

	available, err := uh.repo.CheckUsernameAvailability(r.Context(), username)
	if err != nil {
		log.Error(fmt.Sprintf("[prof-handler]ph#14 Failed to check username availability: %v", err))
		http.Error(w, "Failed to check username availability", http.StatusInternalServerError)
		return
	}

	response := struct {
		Available bool `json:"available"`
	}{
		Available: available,
	}

	w.Header().Set(contentType, applicationJson)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Error(fmt.Sprintf("[prof-handler]ph#15 Failed to encode JSON response: %v", err))
		http.Error(w, "Failed to encode JSON response", http.StatusInternalServerError)
	}
	log.Info(fmt.Sprintf("[prof-handler]ph#44 Successfully checked username availability for user '%s'", username))
}

func (uh *UserHandler) CheckEmailAvailability(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	email := vars["email"]

	log.Info(fmt.Sprintf("[prof-handler]ph#45 Checking email availability from '%s' for user with email '%s'", r.RemoteAddr, email))

	available, err := uh.repo.CheckEmailAvailability(r.Context(), email)
	if err != nil {
		log.Error(fmt.Sprintf("[prof-handler]ph#16 Failed to check email availability: %v", err))
		http.Error(w, "Failed to check email availability", http.StatusInternalServerError)
		return
	}

	response := struct {
		Available bool `json:"available"`
	}{
		Available: available,
	}

	w.Header().Set(contentType, applicationJson)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Error(fmt.Sprintf("[prof-handler]ph#17 Failed to encode JSON response: %v", err))
		http.Error(w, "Failed to encode JSON response", http.StatusInternalServerError)
	}
	log.Info(fmt.Sprintf("[prof-handler]ph#46 Successfully checked email availability for user with email '%v'", email))
}

func (uh *UserHandler) UpdateUser(rw http.ResponseWriter, r *http.Request) {
	tokenStr := uh.extractTokenFromHeader(r)
	vars := mux.Vars(r)
	username := vars["username"]

	log.Info(fmt.Sprintf("[prof-handler]ph#18 Received request to update user '%s' from '%s'", username, r.RemoteAddr))

	// Get current user for email check
	currentUser, err := uh.repo.GetUser(r.Context(), username)
	if err != nil {
		log.Error(fmt.Sprintf("[prof-handler]ph#19 Failed to get user '%s': %v", username, err))
		http.Error(rw, "Failed to get user", http.StatusInternalServerError)
		return
	}

	email := currentUser.Email

	var updatedUser data.NewUser
	if err := json.NewDecoder(r.Body).Decode(&updatedUser); err != nil {
		http.Error(rw, failedToDecodeRequestBody, http.StatusBadRequest)
		log.Error(fmt.Sprintf("[prof-handler]ph#20 Failed to decode request body: %v", err))
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5000*time.Millisecond)
	defer cancel()

	if username != updatedUser.Username {
		_, err = uh.auth.PassUsernameToAuthService(ctx, username, updatedUser.Username, tokenStr)
		if err != nil {
			log.Error(fmt.Sprintf("[prof-handler]ph#21 Error while passing username to auth service: %v", err))
			writeResp(err, http.StatusServiceUnavailable, rw)
			return
		}
	}

	if email != updatedUser.Email {
		_, err = uh.auth.PassEmailToAuthService(ctx, email, updatedUser.Email, tokenStr)
		if err != nil {
			log.Error(fmt.Sprintf("[prof-handler]ph#22 Error while passing email to auth service: %v", err))
			writeResp(err, http.StatusServiceUnavailable, rw)
			return
		}
	}

	if err := uh.repo.UpdateUser(r.Context(), username, &updatedUser, email); err != nil {
		log.Error(fmt.Sprintf("[prof-handler]ph#23 Failed to update user: %v", err))
		http.Error(rw, "Failed to update user", http.StatusInternalServerError)
		return
	}

	rw.Header().Set(contentType, applicationJson)
	rw.WriteHeader(http.StatusOK)

	log.Info(fmt.Sprintf("[prof-handler]ph#24 Successfully updated user '%s'", updatedUser.Username))
}

func (uh *UserHandler) DeleteUser(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username := vars["username"]

	log.Info(fmt.Sprintf("[prof-handler]ph#25 Recieved request from '%s' to delete user '%s'", r.RemoteAddr, username))

	// Extracting role from token
	tokenString := uh.extractTokenFromHeader(r)
	role, err := uh.getRole(tokenString)
	if err != nil {
		log.Error(fmt.Sprintf("[prof-handler]ph#26 Failed to read role from token: %v", err))
		http.Error(rw, "Failed to read role from token", http.StatusBadRequest)
		return
	}

	// Extracting userID for Cassandra
	user, err := uh.repo.GetUser(context.Background(), username)
	if err != nil {
		log.Error(fmt.Sprintf("[prof-handler]ph#27 Failed to retrieve user for username: %v", err))
		http.Error(rw, "Failed to retrieve user for username: "+username, http.StatusBadRequest)
		return
	}

	// Check reservation service for reservations if user is 'GUEST'
	// Check accommodation service if user is 'HOST' and delete all his accommodations
	if role == "GUEST" {
		ctx, cancel := context.WithTimeout(r.Context(), 5000*time.Millisecond)
		defer cancel()
		_, err = uh.reservation.CheckUserReservations(ctx, user.ID, tokenString)
		if err != nil {
			log.Error(fmt.Sprintf("[prof-handler]ph#28 Error while checking user reservations: %v", err))
			writeResp(err, http.StatusServiceUnavailable, rw)
			return
		}
	} else if role == "HOST" {
		ctx, cancel := context.WithTimeout(r.Context(), 5000*time.Millisecond)
		defer cancel()
		_, err = uh.accommodation.CheckAndDeleteUserAccommodations(ctx, user.ID, tokenString)
		if err != nil {
			log.Error(fmt.Sprintf("[prof-handler]ph#29 Error while checking user accommodations: %v", err))
			writeResp(err, http.StatusServiceUnavailable, rw)
			return
		}
	} else {
		http.Error(rw, "Invalid role", http.StatusForbidden)
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5000*time.Millisecond)
	defer cancel()
	_, err = uh.auth.DeleteUserInAuthService(ctx, username, tokenString)
	if err != nil {
		log.Error(fmt.Sprintf("[prof-handler]ph#30 Error while deleting user in auth service: %v", err))
		writeResp(err, http.StatusServiceUnavailable, rw)
		return
	}

	if err := uh.repo.DeleteUser(r.Context(), username); err != nil {
		log.Error(fmt.Sprintf("[prof-handler]ph#31 Failed to delete user: %v", err))
		http.Error(rw, "Failed to delete user", http.StatusInternalServerError)
		return
	}

	rw.WriteHeader(http.StatusNoContent)
	log.Info(fmt.Sprintf("[prof-handler]ph#32 User with username '%s' deleted successfully", username))
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
				log.Warning(fmt.Sprintf("[prof-handler]ph#33 No token found in request from '%s'", rr.RemoteAddr))
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			claims := jwt.MapClaims{}
			token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
				return secretKey, nil
			})

			if err != nil || !token.Valid {
				log.Warning(fmt.Sprintf("[prof-handler]ph#34 Invalid signature token found in request from '%s'", rr.RemoteAddr))
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			username, ok1 := claims["username"].(string)
			role, ok2 := claims["role"].(string)
			if !ok1 || !ok2 {
				log.Warning(fmt.Sprintf("[prof-handler]ph#35 Username or role not found in token in request from '%s'", rr.RemoteAddr))
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			for _, allowedRole := range allowedRoles {
				if allowedRole == role {
					next.ServeHTTP(w, rr)
					return
				}
			}

			log.Warning(fmt.Sprintf("[prof-handler]ph#36 User '%s' from '%s' tried to do unauthorized action", username, rr.RemoteAddr))
			http.Error(w, "Forbidden", http.StatusForbidden)
		})
	}
}

func (uh *UserHandler) getRole(tokenString string) (string, error) {
	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})

	if err != nil || !token.Valid {
		return "", err
	}

	_, ok1 := claims["username"].(string)
	role, ok2 := claims["role"].(string)
	if !ok1 || !ok2 {
		return "", err
	}

	return role, nil
}

func (uh *UserHandler) extractTokenFromHeader(rr *http.Request) string {
	token := rr.Header.Get("Authorization")
	if token != "" {
		return token[len("Bearer "):]
	}
	return ""
}
