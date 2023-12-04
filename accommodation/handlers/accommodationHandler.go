package handlers

import (
	"accommodation/clients"
	"accommodation/data"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/dgrijalva/jwt-go"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AccommodationHandler struct {
	logger      *log.Logger
	repo        *data.AccommodationRepository
	reservation clients.ReservationClient
}

var secretKey = []byte("stayinn_secret")

func NewAccommodationsHandler(l *log.Logger, r *data.AccommodationRepository, rc clients.ReservationClient) *AccommodationHandler {
	return &AccommodationHandler{l, r, rc}
}

func (ah *AccommodationHandler) GetAllAccommodations(rw http.ResponseWriter, r *http.Request) {
	ah.logger.Printf("Usli smo u GetAllAccommodations funkciju")
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

	// Dodajemo smeštaj
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

	if err := ah.repo.DeleteAccommodation(r.Context(), id); err != nil {
		ah.logger.Println("Failed to delete accommodation:", err)
		http.Error(rw, "Failed to delete accommodation", http.StatusInternalServerError)
		return
	}

	rw.WriteHeader(http.StatusNoContent)
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

func (ah *AccommodationHandler) extractTokenFromHeader(rr *http.Request) string {
	token := rr.Header.Get("Authorization")
	if token != "" {
		return token[len("Bearer "):]
	}
	return ""
}

func (ah *AccommodationHandler) SearchAccommodations(rw http.ResponseWriter, r *http.Request) {
	ah.logger.Printf("Usli smo u SearchAccommodations funkciju")
	ctx := r.Context()

	// Parse query parameters
	location := r.URL.Query().Get("location")
	numberOfGuests := r.URL.Query().Get("numberOfGuests")
	startDate := r.URL.Query().Get("startDate")
	endDate := r.URL.Query().Get("endDate")

	numGuests, err := strconv.Atoi(numberOfGuests)
	if err != nil {
		http.Error(rw, "Failed to convert numberOfGuests", http.StatusInternalServerError)
		return
	}

	// Create a filter based on the parsed parameters
	filter := bson.M{}
	if location != "" {
		filter["location"] = location
	}

	if numGuests > 0 {
		filter["$and"] = bson.A{
			bson.M{"minGuests": bson.M{"$lte": numGuests}},
			bson.M{"maxGuests": bson.M{"$gte": numGuests}},
		}
	}

	if startDate != "" {
		filter["startDate"] = startDate
	}

	if endDate != "" {
		filter["endDate"] = endDate
	}

	accommodations, err := ah.repo.GetFilteredAccommodations(ctx, filter)
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
