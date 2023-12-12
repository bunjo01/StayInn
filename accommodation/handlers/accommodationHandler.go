package handlers

import (
	"accommodation/clients"
	"accommodation/data"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AccommodationHandler struct {
	logger      *log.Logger
	repo        *data.AccommodationRepository
	reservation clients.ReservationClient
	profile     clients.ProfileClient
}

var secretKey = []byte("stayinn_secret")

func NewAccommodationsHandler(l *log.Logger, r *data.AccommodationRepository,
	rc clients.ReservationClient, p clients.ProfileClient) *AccommodationHandler {
	return &AccommodationHandler{l, r, rc, p}
}

func (ah *AccommodationHandler) GetAllAccommodations(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	accommodations, err := ah.repo.GetAllAccommodations(ctx)
	if err != nil {
		http.Error(rw, "Failed to retrieve accommodations", http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(rw).Encode(accommodations); err != nil {
		http.Error(rw, "Failed to encode accommodations", http.StatusInternalServerError)
	}
}

func (ah *AccommodationHandler) GetAccommodation(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := primitive.ObjectIDFromHex(vars["id"])
	if err != nil {
		http.Error(rw, "Invalid ID", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	accommodation, err := ah.repo.GetAccommodation(ctx, id)
	if err != nil {
		http.Error(rw, "Failed to retrieve accommodation", http.StatusInternalServerError)
		return
	}

	if accommodation == nil {
		http.NotFound(rw, r)
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(rw).Encode(accommodation); err != nil {
		http.Error(rw, "Failed to encode accommodation", http.StatusInternalServerError)
	}
}

func (ah *AccommodationHandler) CreateAccommodation(rw http.ResponseWriter, r *http.Request) {
	var accommodation data.Accommodation
	if err := json.NewDecoder(r.Body).Decode(&accommodation); err != nil {
		http.Error(rw, "Failed to decode request body", http.StatusBadRequest)
		return
	}

	tokenStr := ah.extractTokenFromHeader(r)
	username, err := ah.getUsername(tokenStr)
	if err != nil {
		ah.logger.Println("Failed to read username from token:", err)
		http.Error(rw, "Failed to read username from token", http.StatusBadRequest)
		return
	}

	hostID, err := ah.profile.GetUserId(r.Context(), username)
	if err != nil {
		ah.logger.Println("Failed to get HostID from username:", err)
		http.Error(rw, "Failed to get HostID from username", http.StatusBadRequest)
		return
	}

	accommodation.HostID, err = primitive.ObjectIDFromHex(hostID)
	if err != nil {
		ah.logger.Println("Failed to set HostID for accommodation:", err)
		http.Error(rw, "Failed to set HostID for accommodation", http.StatusBadRequest)
		return
	}

	// Adding accommodation
	accommodation.ID = primitive.NewObjectID()
	if err := ah.repo.CreateAccommodation(r.Context(), &accommodation); err != nil {
		ah.logger.Println("Failed to create accommodation:", err)
		http.Error(rw, "Failed to create accommodation", http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(rw).Encode(accommodation); err != nil {
		ah.logger.Println("Failed to encode accommodation:", err)
		http.Error(rw, "Failed to encode accommodation", http.StatusInternalServerError)
	}
}

func (ah *AccommodationHandler) UpdateAccommodation(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := primitive.ObjectIDFromHex(vars["id"])
	if err != nil {
		http.Error(rw, "Invalid ID", http.StatusBadRequest)
		return
	}

	var updatedAccommodation data.Accommodation
	if err := json.NewDecoder(r.Body).Decode(&updatedAccommodation); err != nil {
		http.Error(rw, "Failed to decode request body", http.StatusBadRequest)
		return
	}

	updatedAccommodation.ID = id
	if err := ah.repo.UpdateAccommodation(r.Context(), &updatedAccommodation); err != nil {
		ah.logger.Println("Failed to update accommodation:", err)
		http.Error(rw, "Failed to update accommodation", http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(rw).Encode(updatedAccommodation); err != nil {
		ah.logger.Println("Failed to encode updated accommodation:", err)
		http.Error(rw, "Failed to encode updated accommodation", http.StatusInternalServerError)
	}
}

func (ah *AccommodationHandler) DeleteAccommodation(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := primitive.ObjectIDFromHex(vars["id"])
	if err != nil {
		http.Error(rw, "Invalid ID", http.StatusBadRequest)
		return
	}

	var accIDs []primitive.ObjectID
	accIDs = append(accIDs, id)

	ctx, cancel := context.WithTimeout(r.Context(), 4000*time.Millisecond)
	defer cancel()
	_, err = ah.reservation.CheckAndDeletePeriods(ctx, accIDs)
	if err != nil {
		ah.logger.Println("Error checking and deleting periods:", err)
		writeResp(err, http.StatusServiceUnavailable, rw)
		return
	}

	if err := ah.repo.DeleteAccommodation(r.Context(), id); err != nil {
		ah.logger.Println("Failed to delete accommodation:", err)
		http.Error(rw, "Failed to delete accommodation", http.StatusInternalServerError)
		return
	}

	rw.WriteHeader(http.StatusNoContent)
}

func (ah *AccommodationHandler) DeleteUserAccommodations(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := primitive.ObjectIDFromHex(vars["id"])
	if err != nil {
		http.Error(rw, "Invalid userID", http.StatusBadRequest)
		return
	}

	accommodations, err := ah.repo.GetAccommodationsForUser(r.Context(), userID)
	if err != nil {
		ah.logger.Println("Failed to get accommodations for userID:", err)
		http.Error(rw, "Failed to get accommodations for userID: "+userID.Hex(), http.StatusInternalServerError)
		return
	}

	var accIDs []primitive.ObjectID
	for _, accommodation := range accommodations {
		accIDs = append(accIDs, accommodation.ID)
	}

	// 4000 ms because it's second in chain of service calls
	ctx, cancel := context.WithTimeout(r.Context(), 4000*time.Millisecond)
	defer cancel()
	_, err = ah.reservation.CheckAndDeletePeriods(ctx, accIDs)
	if err != nil {
		ah.logger.Println(err)
		writeResp(err, http.StatusServiceUnavailable, rw)
		return
	}

	if err := ah.repo.DeleteAccommodationsForUser(r.Context(), userID); err != nil {
		ah.logger.Println("Failed to delete accommodations for userID:", err)
		http.Error(rw, "Failed to delete accommodations for userID: "+userID.Hex(), http.StatusInternalServerError)
		return
	}

	rw.WriteHeader(http.StatusNoContent)
}

func (ah *AccommodationHandler) getUsername(tokenString string) (string, error) {
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

func (ah *AccommodationHandler) AuthorizeRoles(allowedRoles ...string) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, rr *http.Request) {
			tokenString := ah.extractTokenFromHeader(rr)
			if tokenString == "" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			claims := jwt.MapClaims{}
			token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
				return secretKey, nil
			})
			ah.logger.Println("claims ok, token:", token.Valid)

			if err != nil || !token.Valid {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			ah.logger.Println("token valid")

			_, ok1 := claims["username"].(string)
			role, ok2 := claims["role"].(string)
			if !ok1 || !ok2 {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			ah.logger.Println("username and role ok")

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

func (ah *AccommodationHandler) extractTokenFromHeader(rr *http.Request) string {
	token := rr.Header.Get("Authorization")
	if token != "" {
		return token[len("Bearer "):]
	}
	return ""
}

func (ah *AccommodationHandler) SearchAccommodations(rw http.ResponseWriter, r *http.Request) {
	ah.logger.Printf("Entering SearchAccommodations function")
	ctx, cancel := context.WithTimeout(r.Context(), 5000*time.Millisecond)
	defer cancel()

	var accommodationIDs []primitive.ObjectID

	startDateStr := r.URL.Query().Get("startDate")
	endDateStr := r.URL.Query().Get("endDate")

	var startDate time.Time
	if startDateStr != "" {
		startDateTemp, err := time.Parse("2006-01-02T15:04:05Z", startDateStr)
		if err != nil {
			ah.logger.Println(err)
			http.Error(rw, "Invalid startDate format", http.StatusBadRequest)
			return
		}
		startDate = startDateTemp
	}

	var endDate time.Time
	if endDateStr != "" {
		endDateTemp, err := time.Parse("2006-01-02T15:04:05Z", endDateStr)
		if err != nil {
			ah.logger.Println(err)
			http.Error(rw, "Invalid endDate format", http.StatusBadRequest)
			return
		}
		endDate = endDateTemp
	}

	location := r.URL.Query().Get("location")
	numberOfGuests := r.URL.Query().Get("numberOfGuests")

	numGuests, err := strconv.Atoi(numberOfGuests)
	if err != nil {
		http.Error(rw, "Failed to convert numberOfGuests", http.StatusInternalServerError)
		return
	}

	filter := make(bson.M)

	if location != "" {
		filter["location"] = location
	}

	if numGuests > 0 {
		filter["$and"] = bson.A{
			bson.M{"minGuests": bson.M{"$lte": numGuests}},
			bson.M{"maxGuests": bson.M{"$gte": numGuests}},
		}
	}

	accommodations, err := ah.repo.GetFilteredAccommodations(ctx, filter)
	if err != nil {
		http.Error(rw, "Failed to retrieve accommodations", http.StatusInternalServerError)
		return
	}

	if startDateStr == "" && endDateStr != "" {
		http.Error(rw, "You forgot to select start date", http.StatusBadRequest)
		return
	}

	if endDateStr == "" && startDateStr != "" {
		http.Error(rw, "You forgot to select end date", http.StatusBadRequest)
		return
	}

	if endDateStr != "" && startDateStr != "" {
		if startDate.Before(time.Now()) {
			http.Error(rw, "Start date must be in future", http.StatusBadRequest)
			return
		}

		if endDate.Before(time.Now()) {
			http.Error(rw, "End date must be in future", http.StatusBadRequest)
			return
		}

		if startDate.After(endDate) {
			http.Error(rw, "Start date must be before end date", http.StatusBadRequest)
			return
		}

		for _, accommodation := range accommodations {
			accommodationIDs = append(accommodationIDs, accommodation.ID)
		}

		ids, err := ah.reservation.PassDatesToReservationService(ctx, accommodationIDs, startDate, endDate)
		if err != nil {
			ah.logger.Println(err)
			writeResp(err, http.StatusServiceUnavailable, rw)
			return
		}

		accommodationForReturn, err := ah.repo.FindAccommodationsByIDs(ctx, ids)
		if err != nil {
			ah.logger.Println(err)
			writeResp(err, http.StatusServiceUnavailable, rw)
			return
		}

		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(rw).Encode(accommodationForReturn); err != nil {
			http.Error(rw, "Failed to encode accommodations", http.StatusInternalServerError)
		}
	} else {
		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(rw).Encode(accommodations); err != nil {
			http.Error(rw, "Failed to encode accommodations", http.StatusInternalServerError)
		}
	}
}
