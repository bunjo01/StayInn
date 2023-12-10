package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"notification/clients"
	"notification/data"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type NotificationsHandler struct {
	logger            *log.Logger
	repo              *data.NotificationsRepo
	reservationClient clients.ReservationClient
	profileClient     clients.ProfileClient
}

var secretKey = []byte("stayinn_secret")

// Injecting the logger makes this code much more testable
func NewNotificationsHandler(l *log.Logger, r *data.NotificationsRepo, rc clients.ReservationClient, p clients.ProfileClient) *NotificationsHandler {
	return &NotificationsHandler{l, r, rc, p}
}

// TODO Handler methods

func (rh *NotificationsHandler) AddRating(w http.ResponseWriter, r *http.Request) {
	var rating data.RatingAccommodation
	err := json.NewDecoder(r.Body).Decode(&rating)
	if err != nil {
		http.Error(w, "Error parsing data", http.StatusBadRequest)
		return
	}

	idAccommodation := rating.IDAccommodation

	tokenStr := rh.extractTokenFromHeader(r)
	username, err := rh.getUsername(tokenStr)
	if err != nil {
		rh.logger.Println("Failed to read username from token:", err)
		http.Error(w, "Failed to read username from token", http.StatusBadRequest)
		return
	}

	rating.GuestUsername = username
	rating.Time = time.Now()

	if rating.Rate < 1 || rating.Rate > 5 {
		http.Error(w, "Rating must be between 1 and 5", http.StatusBadRequest)
		return
	}

	userID, err := rh.profileClient.GetUserId(r.Context(), username)
	if err != nil {
		rh.logger.Println("Failed to get HostID from username:", err)
		http.Error(w, "Failed to get HostID from username", http.StatusBadRequest)
		return
	}

	id, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		http.Error(w, "Invalid userID", http.StatusBadRequest)
		return
	}

	reservations, err := rh.reservationClient.GetReservationsByUserIDExp(r.Context(), id)
	if err != nil {
		http.Error(w, "Error fetching user reservations", http.StatusBadRequest)
		return
	}

	found := false
	for _, reservation := range reservations {
		if reservation.IDAccommodation == idAccommodation {
			found = true
			break
		}
	}

	if !found {
		http.Error(w, "Accommodation ID not found in user reservations", http.StatusBadRequest)
		return
	}

	err = rh.repo.AddRating(&rating)
	if err != nil {
		http.Error(w, "Error adding rating", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Rating successfully added"))
}

func (rh *NotificationsHandler) AddHostRating(w http.ResponseWriter, r *http.Request) {
	var rating data.RatingHost
	err := json.NewDecoder(r.Body).Decode(&rating)
	if err != nil {
		http.Error(w, "Error parsing data", http.StatusBadRequest)
		return
	}

	tokenStr := rh.extractTokenFromHeader(r)
	username, err := rh.getUsername(tokenStr)
	if err != nil {
		rh.logger.Println("Failed to read username from token:", err)
		http.Error(w, "Failed to read username from token", http.StatusBadRequest)
		return
	}

	rating.GuestUsername = username

	if rating.Rate < 1 || rating.Rate > 5 {
		http.Error(w, "Rating must be between 1 and 5", http.StatusBadRequest)
		return
	}

	hostID, err := rh.profileClient.GetUserId(r.Context(), rating.HostUsername)
	if err != nil {
		rh.logger.Println("Failed to get HostID from username:", err)
		http.Error(w, "Failed to get HostID from username", http.StatusBadRequest)
		return
	}

	id, err := primitive.ObjectIDFromHex(hostID)
	if err != nil {
		http.Error(w, "Invalid hostID", http.StatusBadRequest)
		return
	}

	hasExpiredReservations, err := rh.reservationClient.GetReservationsByUserIDExp(r.Context(), id)
	if err != nil {
		http.Error(w, "Error checking expired reservations", http.StatusBadRequest)
		return
	}

	if len(hasExpiredReservations) == 0 {
		http.Error(w, "Guest does not have any expired reservations with the specified host", http.StatusBadRequest)
		return
	}

	rating.Time = time.Now()

	err = rh.repo.AddHostRating(&rating)
	if err != nil {
		http.Error(w, "Error adding host rating", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Host rating successfully added"))
}

func (rh *NotificationsHandler) UpdateHostRating(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    ratingID, ok := vars["id"]
    if !ok {
        http.Error(w, "Missing rating ID in the request path", http.StatusBadRequest)
        return
    }

    id, err := primitive.ObjectIDFromHex(ratingID)
    if err != nil {
        http.Error(w, "Invalid rating ID", http.StatusBadRequest)
        return
    }

    var newRating data.RatingHost
    if err := json.NewDecoder(r.Body).Decode(&newRating); err != nil {
        http.Error(w, "Error parsing data", http.StatusBadRequest)
        return
    }

    newRating.Time = time.Now()

    if err := rh.repo.UpdateHostRating(id, &newRating); err != nil {
        http.Error(w, "Error updating host rating", http.StatusBadRequest)
        return
    }

    w.WriteHeader(http.StatusOK)
    w.Write([]byte("Host rating successfully updated"))
}

func (rh *NotificationsHandler) DeleteHostRating(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    ratingID, ok := vars["id"]
    if !ok {
        http.Error(w, "Missing rating ID in the request path", http.StatusBadRequest)
        return
    }

    id, err := primitive.ObjectIDFromHex(ratingID)
    if err != nil {
        http.Error(w, "Invalid rating ID", http.StatusBadRequest)
        return
    }

    if err := rh.repo.DeleteHostRating(id); err != nil {
        http.Error(w, "Error deleting host rating", http.StatusBadRequest)
        return
    }

    w.WriteHeader(http.StatusOK)
    w.Write([]byte("Host rating successfully deleted"))
}

func (ah *NotificationsHandler) extractTokenFromHeader(rr *http.Request) string {
	token := rr.Header.Get("Authorization")
	if token != "" {
		return token[len("Bearer "):]
	}
	return ""
}

func (ah *NotificationsHandler) getUsername(tokenString string) (string, error) {
	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})

	if err != nil || !token.Valid {
		return "", err
	}

	username, ok1 := claims["username"].(string)
	_, ok2 := claims["role"].(string)
	if !ok1 || !ok2 {
		return "", err
	}

	return username, nil
}

func (r *NotificationsHandler) MiddlewareContentTypeSet(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, h *http.Request) {
		rw.Header().Add("Content-Type", "application/json")

		next.ServeHTTP(rw, h)
	})
}
