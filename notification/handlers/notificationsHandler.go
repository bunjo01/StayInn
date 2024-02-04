package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"notification/clients"
	"notification/data"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const FailedToFetchRatings = "Failed to fetch ratings"
const ErrorEncodingHostRatings = "Error encoding host ratings"
const ParseErrorDataFormat = "Error parsing data: %s"
const FailedToReadUsernameFromToken = "Failed to read username from token"
const FailedToGetHostIDFromUsername = "Failed to get HostID from username"
const InvalidUserId = "Invalid userID"
const FailedToCreateNotification = "Failed to create notification"
const FailedToRetriveUserIdFromProfileService = "Failed to retrive user id from profile service"
const ErrorFethingHostRatings = "Error fetching host ratings"
const ContentType = "Content-Type"
const ApplicationJson = "application/json"

type NotificationsHandler struct {
	repo              *data.NotificationsRepo
	reservationClient clients.ReservationClient
	profileClient     clients.ProfileClient
}

var secretKey = []byte("stayinn_secret")

// Injecting the logger makes this code much more testable
func NewNotificationsHandler(r *data.NotificationsRepo, rc clients.ReservationClient, p clients.ProfileClient) *NotificationsHandler {
	return &NotificationsHandler{r, rc, p}
}

// TODO Handler methods

func (nh *NotificationsHandler) GetAccommodationRatings(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	accommodationID := vars["idAccommodation"]

	objectID, err := primitive.ObjectIDFromHex(accommodationID)
	if err != nil {
		http.Error(w, "Invalid accommodation ID", http.StatusBadRequest)
		log.Error(fmt.Sprintf("[noti-handler]nh#1 Invalid accommodation ID: %v", err))
		return
	}

	log.Info(fmt.Sprintf("[noti-handler]nh#116 Received request from '%s' for ratings '%s'", r.RemoteAddr, objectID.Hex()))

	ratings, err := nh.repo.GetRatingsByAccommodationID(objectID)
	if err != nil {
		http.Error(w, FailedToFetchRatings, http.StatusBadRequest)
		return
	}

	if err := json.NewEncoder(w).Encode(ratings); err != nil {
		log.Error(fmt.Sprintf("[noti-handler]nh#2 Error encoding accommodation ratings: %v", err))
		http.Error(w, "Error encoding accommodation ratings", http.StatusInternalServerError)
		return
	}

	log.Info(("[noti-handler]nh#102 Successfully found ratings with accommodationID"))
}

func (nh *NotificationsHandler) GetRatingsHost(w http.ResponseWriter, r *http.Request) {
	var requestData struct {
		HostID string `json:"idHost"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		log.Error(fmt.Sprintf("[noti-handler]nh#3 Invalid request body: %v", err))
		return
	}

	log.Info(fmt.Sprintf("[noti-handler]nh#116 Received request from '%s' for ratings", r.RemoteAddr))

	hostID, err := primitive.ObjectIDFromHex(requestData.HostID)
	if err != nil {
		http.Error(w, "Invalid host ID", http.StatusBadRequest)
		log.Error(fmt.Sprintf("[noti-handler]nh#4 Invalid host ID: %v", err))
		return
	}

	ratings, err := nh.repo.GetRatingsByHostID(hostID)
	if err != nil {
		http.Error(w, FailedToFetchRatings, http.StatusBadRequest)
		log.Error(("[noti-handler]nh#5 Failed to fetch ratings"))
		return
	}

	if err := json.NewEncoder(w).Encode(ratings); err != nil {
		log.Error(fmt.Sprintf("[noti-handler]nh#6 Error encoding host ratings: %v", err))
		http.Error(w, ErrorEncodingHostRatings, http.StatusInternalServerError)
		return
	}

	log.Info(("[noti-handler]nh#103 Successfully found ratings with hostID"))
}

func (nh *NotificationsHandler) AddRating(w http.ResponseWriter, r *http.Request) {
	var rating data.RatingAccommodation

	log.Info(fmt.Sprintf("[noti-handler]nh#117 User from '%s' creating a new rating", r.RemoteAddr))

	err := json.NewDecoder(r.Body).Decode(&rating)
	if err != nil {
		log.Error(fmt.Sprintf("[noti-handler]nh#7 Failed to parse data: %v", err))
		http.Error(w, fmt.Sprintf(ParseErrorDataFormat, err), http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5000*time.Millisecond)
	defer cancel()

	idAccommodation := rating.IDAccommodation

	tokenStr := nh.extractTokenFromHeader(r)
	guestUsername, err := nh.getUsername(tokenStr)
	if err != nil {
		log.Error(fmt.Sprintf("[noti-handler]nh#8 Failed to read username from token: %v", err))
		http.Error(w, FailedToReadUsernameFromToken, http.StatusBadRequest)
		return
	}
	rating.GuestUsername = guestUsername
	rating.Time = time.Now()

	host, err := nh.profileClient.GetUsernameById(ctx, rating.HostID, tokenStr)
	if err != nil {
		log.Error(fmt.Sprintf("[noti-handler]nh#9 Failed to find username with id: %v", err))
		http.Error(w, "Failed to get host", http.StatusBadRequest)
		return
	}

	rating.HostUsername = host.Username

	if rating.Rate < 1 || rating.Rate > 5 {
		log.Error(("[noti-handler]nh#10 Rating must be between 1 and 5"))
		http.Error(w, "Rating must be between 1 and 5", http.StatusBadRequest)
		return
	}

	userID, err := nh.profileClient.GetUserId(r.Context(), guestUsername, tokenStr)
	if err != nil {
		log.Error(fmt.Sprintf("[noti-handler]nh#11 Failed to get HostID from username: %v", err))
		http.Error(w, FailedToGetHostIDFromUsername, http.StatusBadRequest)
		return
	}

	id, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		log.Error(fmt.Sprintf("[noti-handler]nh#12 Invalid userID: %v", err))
		http.Error(w, InvalidUserId, http.StatusBadRequest)
		return
	}

	rating.GuestID = id

	reservations, err := nh.reservationClient.GetReservationsByUserIDExp(r.Context(), id, tokenStr)
	if err != nil {
		log.Error(fmt.Sprintf("[noti-handler]nh#13 Error fetching user reservations: %v", err))
		http.Error(w, fmt.Sprintf("Error fetching user reservations: %s", err), http.StatusBadRequest)
		return
	}

	if len(reservations) == 0 {
		log.Error(("[noti-handler]nh#14 You don't have any reservations"))
		http.Error(w, "You don't have any reservations", http.StatusBadRequest)
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
		log.Error(("[noti-handler]nh#15 AccommodationID not found in user reservations"))
		http.Error(w, "Accommodation ID not found in user reservations", http.StatusBadRequest)
		return
	}

	ratings, err := nh.repo.GetAllAccommodationRatingsByUser(r.Context(), id)
	if err != nil {
		log.Error(fmt.Sprintf("[noti-handler]nh#16 Error fetching user ratings accommodation: %v", err))
		http.Error(w, fmt.Sprintf("Error fetching user ratings accommodation: %s", err), http.StatusBadRequest)
		return
	}

	for _, r := range ratings {
		if r.IDAccommodation == rating.IDAccommodation {
			nh.repo.UpdateRatingAccommodationByID(r.ID, id, rating.Rate)
			log.Info(("[noti-handler]nh#17 Rating successfully added"))
			http.Error(w, "Rating successfully added", http.StatusCreated)
			return
		}
	}

	err = nh.repo.AddRating(&rating)
	if err != nil {
		log.Error(fmt.Sprintf("[noti-handler]nh#18 Error adding rating: %v", err))
		http.Error(w, "Error adding rating", http.StatusBadRequest)
		return
	}

	notification := data.Notification{
		HostID:       host.ID,
		HostUsername: host.Username,
		HostEmail:    host.Email,
		Text:         fmt.Sprintf("User %s rated one of your accommodations %d stars", rating.GuestUsername, rating.Rate),
		Time:         time.Now(),
	}

	err = nh.repo.CreateNotification(r.Context(), &notification)
	if err != nil {
		log.Error(fmt.Sprintf("[noti-handler]nh#19 Failed to create notification: %v", err))
		http.Error(w, FailedToCreateNotification, http.StatusInternalServerError)
		return
	}

	success, err := data.SendNotificationEmail(notification.HostEmail, "rating-accommodation")
	if !success {
		log.Error(("[noti-handler]nh#120 Failed to send notification mail"))
	}

	w.WriteHeader(http.StatusCreated)
	log.Info(("[noti-handler]nh#103 Successfully add rating"))
	w.Write([]byte("Rating successfully added"))
}

func (nh *NotificationsHandler) FindAccommodationRatingByGuest(rw http.ResponseWriter, h *http.Request) {
	vars := mux.Vars(h)
	ratingID := vars["idAccommodation"]

	ctx := h.Context()

	objectID, err := primitive.ObjectIDFromHex(ratingID)
	if err != nil {
		http.Error(rw, "Invalid rating ID", http.StatusBadRequest)
		log.Error(fmt.Sprintf("[noti-handler]nh#21 Invalid ratingID: %v", err))
		return
	}

	log.Info(fmt.Sprintf("[noti-handler]nh#118 Received request from '%s' for accommodation '%s' ratings ", h.RemoteAddr, objectID.Hex()))

	tokenStr := nh.extractTokenFromHeader(h)
	guestUsername, err := nh.getUsername(tokenStr)
	if err != nil {
		log.Error(fmt.Sprintf("[noti-handler]nh#22 Failed to read username from token: %v", err))
		http.Error(rw, FailedToReadUsernameFromToken, http.StatusBadRequest)
		return
	}

	guestId, err := nh.profileClient.GetUserId(ctx, guestUsername, tokenStr)
	if err != nil {
		log.Error(fmt.Sprintf("[noti-handler]nh#23 Failed to retrive userID from profile service: %v", err))
		http.Error(rw, FailedToRetriveUserIdFromProfileService, http.StatusBadRequest)
		return
	}

	guestIdObject, err := primitive.ObjectIDFromHex(guestId)
	if err != nil {
		log.Error(fmt.Sprintf("[noti-handler]nh#24 Failed to parse id to primitive object id: %v", err))
		http.Error(rw, "Failed to parse id to primitive object id", http.StatusBadRequest)
		return
	}

	rating, err := nh.repo.FindAccommodationRatingByGuest(ctx, objectID, guestIdObject)
	if err != nil {
		log.Error(fmt.Sprintf("[noti-handler]nh#25 Database exception: %v", err))
		http.Error(rw, "Database exception", http.StatusInternalServerError)
		return
	}

	if rating == nil {
		log.Error(("[noti-handler]nh#26 Rating not found"))
		http.Error(rw, "Rating not found", http.StatusNotFound)
		return
	}

	err = rating.ToJSON(rw)
	if err != nil {
		http.Error(rw, "Unable to convert to json", http.StatusInternalServerError)
		log.Error(fmt.Sprintf("[noti-handler]nh#27 Unable to convert to json: %v", err))
		return
	}

	log.Info(fmt.Sprintf("[noti-handler]nh#104 Successfully found rating for accommodation by guest"))
}

func (nh *NotificationsHandler) FindHostRatingByGuest(rw http.ResponseWriter, h *http.Request) {
	var userId data.UserId
	err := json.NewDecoder(h.Body).Decode(&userId)
	if err != nil {
		http.Error(rw, fmt.Sprintf(ParseErrorDataFormat, err), http.StatusBadRequest)
		log.Error(fmt.Sprintf("[noti-handler]nh#28 Error parsing data: %v", err))
		return
	}

	log.Info(fmt.Sprintf("[noti-handler]nh#119 Received request from '%s' for host '%s' ratings", h.RemoteAddr, userId.ID.Hex()))

	ctx := h.Context()

	tokenStr := nh.extractTokenFromHeader(h)
	guestUsername, err := nh.getUsername(tokenStr)
	if err != nil {
		log.Error(fmt.Sprintf("[noti-handler]nh#29 Failed to read username from token: %v", err))
		http.Error(rw, FailedToReadUsernameFromToken, http.StatusBadRequest)
		return
	}

	guestId, err := nh.profileClient.GetUserId(ctx, guestUsername, tokenStr)
	if err != nil {
		log.Error(fmt.Sprintf("[noti-handler]nh#30 Failed to retrive userID from profile service: %v", err))
		http.Error(rw, FailedToRetriveUserIdFromProfileService, http.StatusBadRequest)
		return
	}

	guestIdObject, err := primitive.ObjectIDFromHex(guestId)
	if err != nil {
		log.Error(fmt.Sprintf("[noti-handler]nh#31 Failed to parse id to primitive object id: %v", err))
		http.Error(rw, "Failed to parse id to primitive object id", http.StatusBadRequest)
		return
	}

	rating, err := nh.repo.FindHostRatingByGuest(ctx, userId.ID, guestIdObject)
	if err != nil {
		log.Error(fmt.Sprintf("[noti-handler]nh#32 Database exception: %v", err))
		http.Error(rw, "Database exception", http.StatusInternalServerError)
		return
	}

	if rating == nil {
		log.Error(("[noti-handler]nh#33 Rating not found"))
		http.Error(rw, "Rating not found", http.StatusNotFound)
		return
	}

	err = rating.ToJSON(rw)
	if err != nil {
		http.Error(rw, "Unable to convert to json", http.StatusInternalServerError)
		log.Error(fmt.Sprintf("[noti-handler]nh#34 Unable to convert to json: %v", err))
		return
	}

	log.Info(fmt.Sprintf("[noti-handler]nh#105 Successfully found rating for host by guest"))
}

func (nh *NotificationsHandler) GetAllAccommodationRatings(w http.ResponseWriter, r *http.Request) {
	ratings, err := nh.repo.GetAllAccommodationRatings(r.Context())
	if err != nil {
		log.Error(fmt.Sprintf("[noti-handler]nh#35 Error fetching host ratings: %v", err))
		http.Error(w, ErrorFethingHostRatings, http.StatusInternalServerError)
		return
	}

	log.Info(fmt.Sprintf("[noti-handler]nh#120 Received request from '%s' for accommodation ratings", r.RemoteAddr))

	if err := json.NewEncoder(w).Encode(ratings); err != nil {
		log.Error(fmt.Sprintf("[noti-handler]nh#36 Error encoding host ratings: %v", err))
		http.Error(w, ErrorEncodingHostRatings, http.StatusInternalServerError)
		return
	}

	log.Info(("[noti-handler]nh#106 Successfully found ratings for accommodation"))
}

func (nh *NotificationsHandler) GetAllAccommodationRatingsForLoggedHost(w http.ResponseWriter, r *http.Request) {
	tokenStr := nh.extractTokenFromHeader(r)
	hostUsername, err := nh.getUsername(tokenStr)
	if err != nil {
		log.Error(fmt.Sprintf("[noti-handler]nh#37 Failed to read username from token: %v", err))
		http.Error(w, FailedToReadUsernameFromToken, http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	hostId, err := nh.profileClient.GetUserId(ctx, hostUsername, tokenStr)
	if err != nil {
		log.Error(fmt.Sprintf("[noti-handler]nh#38 Failed to retrive userID from profile service: %v", err))
		http.Error(w, FailedToRetriveUserIdFromProfileService, http.StatusBadRequest)
		return
	}

	hostIdObject, err := primitive.ObjectIDFromHex(hostId)
	if err != nil {
		log.Error(fmt.Sprintf("[noti-handler]nh#39 Failed to parse id to primitive object: %v", err))
		http.Error(w, "Failed to parse id to primitive object", http.StatusBadRequest)
		return
	}

	log.Info(fmt.Sprintf("[noti-handler]nh#121 Received request from '%s' for accommodation ratings for host '%s'", r.RemoteAddr, hostIdObject.Hex()))

	ratings, err := nh.repo.GetAllAccommodationRatingsForLoggedHost(r.Context(), hostIdObject)
	if err != nil {
		log.Error(fmt.Sprintf("[noti-handler]nh#40 Error fetching host ratings: %v", err))
		http.Error(w, ErrorFethingHostRatings, http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(ratings); err != nil {
		log.Error(fmt.Sprintf("[noti-handler]nh#41 Error encoding host ratings: %v", err))
		http.Error(w, ErrorEncodingHostRatings, http.StatusInternalServerError)
		return
	}

	log.Info(("[noti-handler]nh#107 Successfully found ratings for accommodation by logged host"))
}

func (nh *NotificationsHandler) GetAllAccommodationRatingsByUser(w http.ResponseWriter, r *http.Request) {
	tokenStr := nh.extractTokenFromHeader(r)
	username, err := nh.getUsername(tokenStr)
	if err != nil {
		log.Error(fmt.Sprintf("[noti-handler]nh#42 Failed to read username from token: %v", err))
		http.Error(w, FailedToReadUsernameFromToken, http.StatusBadRequest)
		return
	}

	userID, err := nh.profileClient.GetUserId(r.Context(), username, tokenStr)
	if err != nil {
		log.Error(fmt.Sprintf("[noti-handler]nh#43 Failed to get HostID from username: %v", err))
		http.Error(w, FailedToGetHostIDFromUsername, http.StatusBadRequest)
		return
	}

	id, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		log.Error(fmt.Sprintf("[noti-handler]nh#44 Invalid userID: %v", err))
		http.Error(w, InvalidUserId, http.StatusBadRequest)
		return
	}

	log.Info(fmt.Sprintf("[noti-handler]nh#122 Received request from '%s' for accommodation ratings by user: '%v'", r.RemoteAddr, id))

	ratings, err := nh.repo.GetAllAccommodationRatingsByUser(r.Context(), id)
	if err != nil {
		log.Error(fmt.Sprintf("[noti-handler]nh#45 Error fetching host ratings: %v", err))
		http.Error(w, ErrorFethingHostRatings, http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(ratings); err != nil {
		log.Error(fmt.Sprintf("[noti-handler]nh#46 Error encoding host ratings: %v", err))
		http.Error(w, ErrorEncodingHostRatings, http.StatusInternalServerError)
		return
	}

	log.Info(("[noti-handler]nh#108 Successfully found ratings for accommodation by user"))
}

func (nh *NotificationsHandler) GetAllHostRatings(w http.ResponseWriter, r *http.Request) {
	ratings, err := nh.repo.GetAllHostRatings(r.Context())
	log.Info(fmt.Sprintf("[noti-handler]nh#123 Received request from '%s' for host ratings", r.RemoteAddr))
	if err != nil {
		log.Error(fmt.Sprintf("[noti-handler]nh#47 Error fetching host ratings: %v", err))
		http.Error(w, ErrorFethingHostRatings, http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(ratings); err != nil {
		log.Error(fmt.Sprintf("[noti-handler]nh#48 Error encoding host ratings: %v", err))
		http.Error(w, ErrorEncodingHostRatings, http.StatusInternalServerError)
		return
	}

	log.Info(("[noti-handler]nh#109 Successfully found ratings for host"))
}

func (nh *NotificationsHandler) GetAllHostRatingsByUser(w http.ResponseWriter, r *http.Request) {
	tokenStr := nh.extractTokenFromHeader(r)
	username, err := nh.getUsername(tokenStr)
	if err != nil {
		log.Error(fmt.Sprintf("[noti-handler]nh#49 Failed to read username from token: %v", err))
		http.Error(w, FailedToReadUsernameFromToken, http.StatusBadRequest)
		return
	}

	userID, err := nh.profileClient.GetUserId(r.Context(), username, tokenStr)
	if err != nil {
		log.Error(fmt.Sprintf("[noti-handler]nh#50 Failed to get HostID from username: %v", err))
		http.Error(w, FailedToGetHostIDFromUsername, http.StatusBadRequest)
		return
	}

	id, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		log.Error(fmt.Sprintf("[noti-handler]nh#51 Invalid userID: %v", err))
		http.Error(w, InvalidUserId, http.StatusBadRequest)
		return
	}

	log.Info(fmt.Sprintf("[noti-handler]nh#124 Received request from '%s' for host ratings by user '%s'", r.RemoteAddr, id.Hex()))

	ratings, err := nh.repo.GetAllHostRatingsByUser(r.Context(), id)
	if err != nil {
		log.Error(fmt.Sprintf("[noti-handler]nh#52 Error fetching host ratings: %v", err))
		http.Error(w, ErrorFethingHostRatings, http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(ratings); err != nil {
		log.Error(fmt.Sprintf("[noti-handler]nh#53 Error encoding host ratings: %v", err))
		http.Error(w, ErrorEncodingHostRatings, http.StatusInternalServerError)
		return
	}

	log.Info(("[noti-handler]nh#110 Successfully found ratings for host by user"))
}

func (nh *NotificationsHandler) GetHostRatings(w http.ResponseWriter, r *http.Request) {
	tokenStr := nh.extractTokenFromHeader(r)
	vars := mux.Vars(r)
	hostUsername, ok := vars["hostUsername"]
	if !ok {
		log.Error(("[noti-handler]nh#54 Missing host username in the request path"))
		http.Error(w, "Missing host username in the request path", http.StatusBadRequest)
		return
	}

	log.Info(fmt.Sprintf("[noti-handler]nh#125 Received request from '%s' for host ratings by username: '%s'", r.RemoteAddr, hostUsername))

	_, err := nh.profileClient.GetUserId(r.Context(), hostUsername, tokenStr)
	if err != nil {
		log.Error(fmt.Sprintf("[noti-handler]nh#55 Failed HostID from username: %v", err))
		http.Error(w, FailedToGetHostIDFromUsername, http.StatusBadRequest)
		return
	}

	ratings, err := nh.repo.GetHostRatings(r.Context(), hostUsername)
	if err != nil {
		log.Error(fmt.Sprintf("[noti-handler]nh#56 Error fetching host ratings: %v", err))
		http.Error(w, ErrorFethingHostRatings, http.StatusInternalServerError)
		return
	}

	// Convert ratings to JSON and send the response
	err = json.NewEncoder(w).Encode(ratings)
	if err != nil {
		log.Error(fmt.Sprintf("[noti-handler]nh#57 Error encoding response: %v", err))
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}

	log.Info(("[noti-handler]nh#111 Successfully found ratings for host"))
}

func (nh *NotificationsHandler) AddHostRating(w http.ResponseWriter, r *http.Request) {
	var rating data.RatingHost
	ctx, cancel := context.WithTimeout(r.Context(), 5000*time.Millisecond)
	defer cancel()

	log.Info(fmt.Sprintf("[noti-handler]nh#126 User from '%s' creating a new rating", r.RemoteAddr))

	err := json.NewDecoder(r.Body).Decode(&rating)
	if err != nil {
		log.Error(fmt.Sprintf("[noti-handler]nh#58 Error parsing data: %v", err))
		http.Error(w, "Error parsing data", http.StatusBadRequest)
		return
	}

	tokenStr := nh.extractTokenFromHeader(r)
	guestUsername, err := nh.getUsername(tokenStr)
	if err != nil {
		log.Error(fmt.Sprintf("[noti-handler]nh#59 Failed to red username from token: %v", err))
		http.Error(w, FailedToReadUsernameFromToken, http.StatusBadRequest)
		return
	}

	host, err := nh.profileClient.GetUsernameById(ctx, rating.HostID, tokenStr)
	if err != nil {
		log.Error(fmt.Sprintf("[noti-handler]nh#60 Failed to get host: %v", err))
		http.Error(w, "Failed to get host", http.StatusBadRequest)
		return
	}

	guestId, err := nh.profileClient.GetUserId(ctx, guestUsername, tokenStr)
	if err != nil {
		log.Error(fmt.Sprintf("[noti-handler]nh#61 Failed to get guest: %v", err))
		http.Error(w, "Failed to get guest", http.StatusBadRequest)
		return
	}

	rating.GuestUsername = guestUsername
	rating.HostUsername = host.Username

	rating.GuestID, err = primitive.ObjectIDFromHex(guestId)
	if err != nil {
		log.Error(fmt.Sprintf("[noti-handler]nh#62 Failed to parse primitive object id: %v", err))
		http.Error(w, "Failed to parse primitive object id", http.StatusBadRequest)
		return
	}

	if rating.Rate < 1 || rating.Rate > 5 {
		log.Error(("[noti-handler]nh#63 Rating must be between 1 and 5"))
		http.Error(w, "Rating must be between 1 and 5", http.StatusBadRequest)
		return
	}

	hostID, err := nh.profileClient.GetUserId(r.Context(), rating.HostUsername, tokenStr)
	if err != nil {
		log.Error(fmt.Sprintf("[noti-handler]nh#64 Failed to get HostID from username: %v", err))
		http.Error(w, FailedToGetHostIDFromUsername, http.StatusBadRequest)
		return
	}

	id, err := primitive.ObjectIDFromHex(hostID)
	if err != nil {
		log.Error(fmt.Sprintf("[noti-handler]nh#65 Invalid hostID: %v", err))
		http.Error(w, "Invalid hostID", http.StatusBadRequest)
		return
	}

	rating.GuestID = id

	hasExpiredReservations, err := nh.reservationClient.GetReservationsByUserIDExp(r.Context(), rating.GuestID, tokenStr)
	if err != nil {
		log.Error(fmt.Sprintf("[noti-handler]nh#66 Error checking expired reservations: %v", err))
		http.Error(w, "Error checking expired reservations", http.StatusBadRequest)
		return
	}

	if len(hasExpiredReservations) == 0 {
		log.Error(("[noti-handler]nh#67 Guest does not have any expired reservations with the specified host"))
		http.Error(w, "Guest does not have any expired reservations with the specified host", http.StatusBadRequest)
		return
	}

	rating.Time = time.Now()

	ratings, err := nh.repo.GetAllHostRatingsByUser(r.Context(), rating.HostID)
	if err != nil {
		log.Error(fmt.Sprintf("[noti-handler]nh#68 Error fetching user ratings host: %v", err))
		http.Error(w, fmt.Sprintf("Error fetching user ratings host: %s", err), http.StatusBadRequest)
		return
	}

	for _, r := range ratings {
		if r.HostUsername == rating.HostUsername && r.GuestUsername == rating.GuestUsername {
			nh.repo.UpdateHostRating(r.ID, rating.GuestID, &rating)
			log.Info(("[noti-handler]nh#69 Host rating successfully added"))
			http.Error(w, "Host rating successfully added", http.StatusCreated)
			return
		}
	}

	err = nh.repo.AddHostRating(&rating)
	if err != nil {
		log.Error(fmt.Sprintf("[noti-handler]nh#70 Error adding host rating: %v", err))
		http.Error(w, "Error adding host rating", http.StatusBadRequest)
		return
	}

	notification := data.Notification{
		HostID:       host.ID,
		HostUsername: host.Username,
		HostEmail:    host.Email,
		Text:         fmt.Sprintf("User %s rated you %d stars", rating.GuestUsername, rating.Rate),
		Time:         time.Now(),
	}

	err = nh.repo.CreateNotification(r.Context(), &notification)
	if err != nil {
		log.Error(fmt.Sprintf("[noti-handler]nh#71 Failed to create notification: %v", err))
		http.Error(w, FailedToCreateNotification, http.StatusInternalServerError)
		return
	}

	success, err := data.SendNotificationEmail(notification.HostEmail, "rating-host")
	if !success {
		log.Error(("[noti-handler]nh#72 Failed to send notification mail"))
	}

	log.Info(("[noti-handler]nh#112 Successfully add rating for host"))
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Host rating successfully added"))
}

func (nh *NotificationsHandler) GetAverageAccommodationRating(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	accommodationID := params["accommodationID"]

	objectID, err := primitive.ObjectIDFromHex(accommodationID)
	if err != nil {
		http.Error(w, "Invalid rating ID", http.StatusBadRequest)
		log.Error(fmt.Sprintf("[noti-handler]nh#73 Invalid rating ID: %v", err))
		return
	}

	log.Info(fmt.Sprintf("[noti-handler]nh#127 Received request from '%s' for average accommodation '%s' rating: ", r.RemoteAddr, objectID.Hex()))

	ratings, err := nh.repo.GetRatingsByAccommodationID(objectID)
	if err != nil {
		http.Error(w, FailedToFetchRatings, http.StatusBadRequest)
		log.Error(fmt.Sprintf("[noti-handler]nh#74 Failed to fetch ratings: %v", err))
		return
	}

	totalRatings := len(ratings)
	if totalRatings == 0 {
		http.Error(w, "No ratings found for this accommodation", http.StatusNotFound)
		log.Error(("[noti-handler]nh#74 No ratings found for this accommodaiton"))
		return
	}

	sum := 0
	for _, rating := range ratings {
		sum += rating.Rate
	}

	averageRating := float64(sum) / float64(totalRatings)

	avgRatingAccommodation := data.AverageRatingAccommodation{
		AccommodationID: objectID,
		AverageRating:   averageRating,
	}

	jsonResponse, err := json.Marshal(avgRatingAccommodation)
	if err != nil {
		log.Error(fmt.Sprintf("[noti-handler]nh#75 Error encoding average rating: %v", err))
		http.Error(w, "Error encoding average rating", http.StatusBadRequest)
		return
	}

	log.Info(("[noti-handler]nh#113 Successfully get average accommodation rating"))
	w.Header().Set(ContentType, ApplicationJson)
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)

}

func (nh *NotificationsHandler) GetAverageHostRating(w http.ResponseWriter, r *http.Request) {
	tokenStr := nh.extractTokenFromHeader(r)
	var userId data.UserId
	err := json.NewDecoder(r.Body).Decode(&userId)
	if err != nil {
		log.Error(fmt.Sprintf("[noti-handler]nh#76 Error parsing data: %v", err))
		http.Error(w, fmt.Sprintf(ParseErrorDataFormat, err), http.StatusBadRequest)
		return
	}

	log.Info(fmt.Sprintf("[noti-handler]nh#128 Received request from '%s' for average rating on host '%s'", r.RemoteAddr, userId.ID.Hex()))

	ctx := r.Context()

	host, err := nh.profileClient.GetUsernameById(ctx, userId.ID, tokenStr)
	if err != nil {
		log.Error(fmt.Sprintf("[noti-handler]nh#77 Error trying to find user by id: %v", err))
		http.Error(w, fmt.Sprintf("Error trying to find user by id: %s", err), http.StatusBadRequest)
		return
	}

	ratings, err := nh.repo.GetRatingsByHostUsername(host.Username)
	if err != nil {
		log.Error(fmt.Sprintf("[noti-handler]nh#78 Failed to fetch ratings: %v", err))
		http.Error(w, FailedToFetchRatings, http.StatusBadRequest)
		return
	}

	totalRatings := len(ratings)
	if totalRatings == 0 {
		log.Error(("[noti-handler]nh#79 No ratings found for this accommodation"))
		http.Error(w, "No ratings found for this accommodation", http.StatusNotFound)
		return
	}

	sum := 0
	for _, rating := range ratings {
		sum += rating.Rate
	}

	averageRating := float64(sum) / float64(totalRatings)

	avgRatingAccommodation := data.AverageRatingHost{
		Username:      host.Username,
		AverageRating: averageRating,
	}

	jsonResponse, err := json.Marshal(avgRatingAccommodation)
	if err != nil {
		log.Error(fmt.Sprintf("[noti-handler]nh#80 Error encoding average rating: %v", err))
		http.Error(w, "Error encoding average rating", http.StatusBadRequest)
		return
	}

	log.Info(("[noti-handler]nh#114 Successfully get average host rating"))
	w.Header().Set(ContentType, ApplicationJson)
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)

}

func (nh *NotificationsHandler) DeleteHostRating(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idParam, ok := vars["id"]

	if !ok {
		log.Error(("[noti-handler]nh#81 Missing ID parameter"))
		http.Error(w, "Missing ID parameter", http.StatusBadRequest)
		return
	}

	id, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		log.Error(fmt.Sprintf("[noti-handler]nh#82 Invalid ID format: %v", err))
		http.Error(w, "Invalid ID format", http.StatusBadRequest)
		return
	}

	log.Info(fmt.Sprintf("[noti-handler]nh#129 Recieved request from '%s' to delete rating '%s'", r.RemoteAddr, id.Hex()))

	tokenStr := nh.extractTokenFromHeader(r)
	username, err := nh.getUsername(tokenStr)
	if err != nil {
		log.Error(fmt.Sprintf("[noti-handler]nh#83 Failed to read username from token: %v", err))
		http.Error(w, FailedToReadUsernameFromToken, http.StatusBadRequest)
		return
	}

	userID, err := nh.profileClient.GetUserId(r.Context(), username, tokenStr)
	if err != nil {
		log.Error(fmt.Sprintf("[noti-handler]nh#84 Failed to get HostID from username: %v", err))
		http.Error(w, FailedToGetHostIDFromUsername, http.StatusBadRequest)
		return
	}

	idUser, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		log.Error(fmt.Sprintf("[noti-handler]nh#85 Invalid userID: %v", err))
		http.Error(w, InvalidUserId, http.StatusBadRequest)
		return
	}

	if err := nh.repo.DeleteHostRating(id, idUser); err != nil {
		log.Error(fmt.Sprintf("[noti-handler]nh#86 Error deleting host rating: %v", err))
		http.Error(w, "Error deleting host rating", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	log.Info(("[noti-handler]nh#87 Host rating successfully deleted"))
	w.Write([]byte("Host rating successfully deleted"))

}

func (nh *NotificationsHandler) DeleteRatingAccommodationHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idParam, ok := vars["id"]

	if !ok {
		log.Error(("[noti-handler]nh#88 Missing ID parameter"))
		http.Error(w, "Missing ID parameter", http.StatusBadRequest)
		return
	}

	id, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		log.Error(fmt.Sprintf("[noti-handler]nh#89 Invalid ID format: %v", err))
		http.Error(w, "Invalid ID format", http.StatusBadRequest)
		return
	}

	log.Info(fmt.Sprintf("[noti-handler]nh#130 Recieved request from '%s' to delete rating '%s'", r.RemoteAddr, id.Hex()))

	tokenStr := nh.extractTokenFromHeader(r)
	username, err := nh.getUsername(tokenStr)
	if err != nil {
		log.Error(fmt.Sprintf("[noti-handler]nh#90 Failed to read username from token: %v", err))
		http.Error(w, FailedToReadUsernameFromToken, http.StatusBadRequest)
		return
	}

	userID, err := nh.profileClient.GetUserId(r.Context(), username, tokenStr)
	if err != nil {
		log.Error(fmt.Sprintf("[noti-handler]nh#91 Failed to get HostID from username: %v", err))
		http.Error(w, FailedToGetHostIDFromUsername, http.StatusBadRequest)
		return
	}

	idUser, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		log.Error(fmt.Sprintf("[noti-handler]nh#92 Invalid userID: %v", err))
		http.Error(w, InvalidUserId, http.StatusBadRequest)
		return
	}

	err = nh.repo.DeleteRatingAccommodationByID(id, idUser)
	if err != nil {
		log.Error(fmt.Sprintf("[noti-handler]nh#93 Error deleting rating: %v", err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	log.Info(fmt.Sprintf("[noti-handler]nh#94 Rating successfully deleted: %v", err))
	w.Write([]byte("Document deleted successfully"))
}

func (nh *NotificationsHandler) NotifyForReservation(w http.ResponseWriter, r *http.Request) {
	var notification data.Notification
	if err := json.NewDecoder(r.Body).Decode(&notification); err != nil {
		log.Error(fmt.Sprintf("[noti-handler]nh#95 Failed to decode request body: %v", err))
		http.Error(w, "Failed to decode request body", http.StatusBadRequest)
		return
	}

	log.Info(fmt.Sprintf("[noti-handler]nh#131 Recieved request from '%s' to notify user for reservation accommodation", r.RemoteAddr))

	err := nh.repo.CreateNotification(r.Context(), &notification)
	if err != nil {
		log.Error(fmt.Sprintf("[noti-handler]nh#96 Failed to create notification: %v", err))
		http.Error(w, FailedToCreateNotification, http.StatusInternalServerError)
		return
	}

	var intent string
	if strings.Contains(notification.Text, "created") {
		intent = "reservation-new"
	} else {
		intent = "reservation-deleted"
	}

	success, err := data.SendNotificationEmail(notification.HostEmail, intent)
	if !success {
		log.Error(("[noti-handler]nh#97 Failed to send notification mail"))
	}

	w.Header().Set(ContentType, ApplicationJson)
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(notification); err != nil {
		log.Error(fmt.Sprintf("[noti-handler]nh#98 Failed to encode notification: %v", err))
		http.Error(w, "Failed to encode notification", http.StatusInternalServerError)
	}
}

func (nh *NotificationsHandler) GetAllNotifications(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username, ok := vars["username"]

	if !ok {
		log.Error(("[noti-handler]nh#99 Missing username parameter"))
		http.Error(w, "Missing username parameter", http.StatusBadRequest)
		return
	}

	log.Info(fmt.Sprintf("[noti-handler]nh#132 Recieved request from '%s' to get all notification", r.RemoteAddr))

	ratings, err := nh.repo.GetAllNotifications(r.Context(), username)
	if err != nil {
		log.Error(fmt.Sprintf("[noti-handler]nh#100 Error fetching host ratings: %v", err))
		http.Error(w, ErrorFethingHostRatings, http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(ratings); err != nil {
		log.Error(fmt.Sprintf("[noti-handler]nh#101 Error encoding host rating: %v", err))
		http.Error(w, ErrorEncodingHostRatings, http.StatusInternalServerError)
		return
	}
}

func (nh *NotificationsHandler) extractTokenFromHeader(rr *http.Request) string {
	token := rr.Header.Get("Authorization")
	if token != "" {
		return token[len("Bearer "):]
	}
	return ""
}

func (nh *NotificationsHandler) getUsername(tokenString string) (string, error) {
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

func (nh *NotificationsHandler) MiddlewareContentTypeSet(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, h *http.Request) {
		rw.Header().Add(ContentType, ApplicationJson)

		next.ServeHTTP(rw, h)
	})
}
