package handlers

import (
	"encoding/json"
	"fmt"
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
		http.Error(w, fmt.Sprintf("Error parsing data: %s", err), http.StatusBadRequest)
		return
	}

	// idAccommodation := rating.IDAccommodation

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

	rating.GuestID = id

	// reservations, err := rh.reservationClient.GetReservationsByUserIDExp(r.Context(), id)
	// if err != nil {
	// 	http.Error(w, fmt.Sprintf("Error fetching user reservations: %s", err), http.StatusBadRequest)
	// 	return
	// }

	// found := false
	// for _, reservation := range reservations {
	// 	if reservation.IDAccommodation == idAccommodation {
	// 		found = true
	// 		break
	// 	}
	// }

	// if !found {
	// 	http.Error(w, "Accommodation ID not found in user reservations", http.StatusBadRequest)
	// 	return
	// }

	ratings, err := rh.repo.GetAllAccommodationRatingsByUser(r.Context(), id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching user ratings accommodation: %s", err), http.StatusBadRequest)
		return
	}

	for _, r := range ratings {
		if r.IDAccommodation == rating.IDAccommodation {
			http.Error(w, "User already rated this accommodation", http.StatusBadRequest)
			return
		}
	}

	err = rh.repo.AddRating(&rating)
	if err != nil {
		http.Error(w, "Error adding rating", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Rating successfully added"))
}

func (r *NotificationsHandler) FindRatingById(rw http.ResponseWriter, h *http.Request) {
	vars := mux.Vars(h)
	ratingID := vars["id"]

	ctx := h.Context()

	objectID, err := primitive.ObjectIDFromHex(ratingID)
	if err != nil {
		http.Error(rw, "Invalid rating ID", http.StatusBadRequest)
		r.logger.Println("Invalid rating ID:", err)
		return
	}

	rating, err := r.repo.FindRatingById(ctx, objectID)
	if err != nil {
		r.logger.Println("Database exception: ", err)
		http.Error(rw, "Database exception", http.StatusInternalServerError)
		return
	}

	if rating == nil {
		r.logger.Println("No period with given ID in accommodation")
		http.Error(rw, "Rating not found", http.StatusNotFound)
		return
	}

	err = rating.ToJSON(rw)
	if err != nil {
		http.Error(rw, "Unable to convert to json", http.StatusInternalServerError)
		r.logger.Fatal("Unable to convert to json:", err)
		return
	}
}

func (r *NotificationsHandler) FindHostRatingById(rw http.ResponseWriter, h *http.Request) {
	vars := mux.Vars(h)
	ratingID := vars["id"]

	ctx := h.Context()

	objectID, err := primitive.ObjectIDFromHex(ratingID)
	if err != nil {
		http.Error(rw, "Invalid rating ID", http.StatusBadRequest)
		r.logger.Println("Invalid rating ID:", err)
		return
	}

	rating, err := r.repo.FindHostRatingById(ctx, objectID)
	if err != nil {
		r.logger.Println("Database exception: ", err)
		http.Error(rw, "Database exception", http.StatusInternalServerError)
		return
	}

	if rating == nil {
		r.logger.Println("No period with given ID in accommodation")
		http.Error(rw, "Rating not found", http.StatusNotFound)
		return
	}

	err = rating.ToJSON(rw)
	if err != nil {
		http.Error(rw, "Unable to convert to json", http.StatusInternalServerError)
		r.logger.Fatal("Unable to convert to json:", err)
		return
	}
}

func (rh *NotificationsHandler) GetAllAccommodationRatings(w http.ResponseWriter, r *http.Request) {
	ratings, err := rh.repo.GetAllAccommodationRatings(r.Context())
	if err != nil {
		rh.logger.Println("Error fetching all host ratings:", err)
		http.Error(w, "Error fetching host ratings", http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(ratings); err != nil {
		rh.logger.Println("Error encoding host ratings:", err)
		http.Error(w, "Error encoding host ratings", http.StatusInternalServerError)
		return
	}
}

func (rh *NotificationsHandler) GetAllAccommodationRatingsByUser(w http.ResponseWriter, r *http.Request) {
	tokenStr := rh.extractTokenFromHeader(r)
	username, err := rh.getUsername(tokenStr)
	if err != nil {
		rh.logger.Println("Failed to read username from token:", err)
		http.Error(w, "Failed to read username from token", http.StatusBadRequest)
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

	ratings, err := rh.repo.GetAllAccommodationRatingsByUser(r.Context(), id)
	if err != nil {
		rh.logger.Println("Error fetching all host ratings:", err)
		http.Error(w, "Error fetching host ratings", http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(ratings); err != nil {
		rh.logger.Println("Error encoding host ratings:", err)
		http.Error(w, "Error encoding host ratings", http.StatusInternalServerError)
		return
	}
}

func (rh *NotificationsHandler) GetAllHostRatings(w http.ResponseWriter, r *http.Request) {
	ratings, err := rh.repo.GetAllHostRatings(r.Context())
	if err != nil {
		rh.logger.Println("Error fetching all host ratings:", err)
		http.Error(w, "Error fetching host ratings", http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(ratings); err != nil {
		rh.logger.Println("Error encoding host ratings:", err)
		http.Error(w, "Error encoding host ratings", http.StatusInternalServerError)
		return
	}
}

func (rh *NotificationsHandler) GetAllHostRatingsByUser(w http.ResponseWriter, r *http.Request) {
	tokenStr := rh.extractTokenFromHeader(r)
	username, err := rh.getUsername(tokenStr)
	if err != nil {
		rh.logger.Println("Failed to read username from token:", err)
		http.Error(w, "Failed to read username from token", http.StatusBadRequest)
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
	ratings, err := rh.repo.GetAllHostRatingsByUser(r.Context(), id)
	if err != nil {
		rh.logger.Println("Error fetching all host ratings:", err)
		http.Error(w, "Error fetching host ratings", http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(ratings); err != nil {
		rh.logger.Println("Error encoding host ratings:", err)
		http.Error(w, "Error encoding host ratings", http.StatusInternalServerError)
		return
	}
}

func (rh *NotificationsHandler) GetHostRatings(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	hostUsername, ok := vars["hostUsername"]
	if !ok {
		http.Error(w, "Missing host username in the request path", http.StatusBadRequest)
		return
	}

	_, err := rh.profileClient.GetUserId(r.Context(), hostUsername)
	if err != nil {
		rh.logger.Println("Failed to get HostID from username:", err)
		http.Error(w, "Failed to get HostID from username", http.StatusBadRequest)
		return
	}

	// id, err := primitive.ObjectIDFromHex(hostID)
	// if err != nil {
	//     http.Error(w, "Invalid hostID", http.StatusBadRequest)
	//     return
	// }

	ratings, err := rh.repo.GetHostRatings(r.Context(), hostUsername)
	if err != nil {
		rh.logger.Println("Error fetching host ratings:", err)
		http.Error(w, "Error fetching host ratings", http.StatusInternalServerError)
		return
	}

	// Convert ratings to JSON and send the response
	err = json.NewEncoder(w).Encode(ratings)
	if err != nil {
		rh.logger.Println("Error encoding response:", err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}

func (rh *NotificationsHandler) AddHostRating(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	hostUsername, ok := vars["hostUsername"]
	if !ok {
		http.Error(w, "Missing host username in the request path", http.StatusBadRequest)
		return
	}
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
	rating.HostUsername = hostUsername

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

	rating.GuestID = id

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

	ratings, err := rh.repo.GetAllHostRatingsByUser(r.Context(), id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching user ratings accommodation: %s", err), http.StatusBadRequest)
		return
	}

	for _, r := range ratings {
		if r.HostUsername == rating.GuestUsername {
			http.Error(w, "User already rated this accommodation", http.StatusBadRequest)
			return
		}
	}

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

func (rh *NotificationsHandler) UpdateAccommodationRating(w http.ResponseWriter, r *http.Request) {
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

	var newRating data.RatingAccommodation
	if err := json.NewDecoder(r.Body).Decode(&newRating); err != nil {
		http.Error(w, "Error parsing data", http.StatusBadRequest)
		return
	}

	if err := rh.repo.UpdateRatingAccommodationByID(id, newRating.Rate); err != nil {
		http.Error(w, "Error updating host rating", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Host rating successfully updated"))
}

func (rh *NotificationsHandler) DeleteHostRating(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idParam, ok := vars["id"]

	if !ok {
		http.Error(w, "Missing ID parameter", http.StatusBadRequest)
		return
	}

	id, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		http.Error(w, "Invalid ID format", http.StatusBadRequest)
		return
	}

	tokenStr := rh.extractTokenFromHeader(r)
	username, err := rh.getUsername(tokenStr)
	if err != nil {
		rh.logger.Println("Failed to read username from token:", err)
		http.Error(w, "Failed to read username from token", http.StatusBadRequest)
		return
	}

	userID, err := rh.profileClient.GetUserId(r.Context(), username)
	if err != nil {
		rh.logger.Println("Failed to get HostID from username:", err)
		http.Error(w, "Failed to get HostID from username", http.StatusBadRequest)
		return
	}

	idUser, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		http.Error(w, "Invalid userID", http.StatusBadRequest)
		return
	}

	if err := rh.repo.DeleteHostRating(id, idUser); err != nil {
		http.Error(w, "Error deleting host rating", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Host rating successfully deleted"))

}

func (rh *NotificationsHandler) DeleteRatingAccommodationHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idParam, ok := vars["id"]

	if !ok {
		http.Error(w, "Missing ID parameter", http.StatusBadRequest)
		return
	}

	id, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		http.Error(w, "Invalid ID format", http.StatusBadRequest)
		return
	}

	tokenStr := rh.extractTokenFromHeader(r)
	username, err := rh.getUsername(tokenStr)
	if err != nil {
		rh.logger.Println("Failed to read username from token:", err)
		http.Error(w, "Failed to read username from token", http.StatusBadRequest)
		return
	}

	userID, err := rh.profileClient.GetUserId(r.Context(), username)
	if err != nil {
		rh.logger.Println("Failed to get HostID from username:", err)
		http.Error(w, "Failed to get HostID from username", http.StatusBadRequest)
		return
	}

	idUser, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		http.Error(w, "Invalid userID", http.StatusBadRequest)
		return
	}

	err = rh.repo.DeleteRatingAccommodationByID(id, idUser)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Document deleted successfully"))
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
