package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"reservation/clients"
	"reservation/data"

	"github.com/gorilla/mux"
)

type KeyProduct struct{}

type ReservationHandler struct {
	logger        *log.Logger
	repo          *data.ReservationRepo
	notification  clients.NotificationClient
	profile       clients.ProfileClient
	accommodation clients.AccommodationClient
}

var secretKey = []byte("stayinn_secret")

func NewReservationHandler(l *log.Logger, r *data.ReservationRepo, n clients.NotificationClient,
	p clients.ProfileClient, a clients.AccommodationClient) *ReservationHandler {
	return &ReservationHandler{l, r, n, p, a}
}

func (r *ReservationHandler) GetAllAvailablePeriodsByAccommodation(rw http.ResponseWriter, h *http.Request) {
	vars := mux.Vars(h)
	id := vars["id"]

	availablePeriods, err := r.repo.GetAvailablePeriodsByAccommodation(id)
	if err != nil {
		r.logger.Println("Database exception: ", err)
	}

	if availablePeriods == nil {
		return
	}

	err = availablePeriods.ToJSON(rw)
	if err != nil {
		http.Error(rw, "Unable to convert to json:", http.StatusInternalServerError)
		r.logger.Fatal("Unable to convert to json :", err)
		return
	}
}

func (r *ReservationHandler) FindAvailablePeriodByIdAndByAccommodationId(rw http.ResponseWriter, h *http.Request) {
	vars := mux.Vars(h)
	periodID := vars["periodID"]
	accomodationID := vars["accomodationID"]

	availablePeriod, err := r.repo.FindAvailablePeriodById(periodID, accomodationID)
	if err != nil {
		r.logger.Println("Database exception: ", err)
	}

	if availablePeriod == nil {
		r.logger.Println("No period with given ID in accommodation")
		return
	}

	err = availablePeriod.ToJSON(rw)
	if err != nil {
		http.Error(rw, "Unable to convert to json:", http.StatusInternalServerError)
		r.logger.Fatal("Unable to convert to json :", err)
		return
	}
}

func (r *ReservationHandler) GetAllReservationByAvailablePeriod(rw http.ResponseWriter, h *http.Request) {
	vars := mux.Vars(h)
	id := vars["id"]

	reservations, err := r.repo.GetReservationsByAvailablePeriod(id)
	if err != nil {
		r.logger.Println("Database exception: ", err)
	}

	if reservations == nil {
		return
	}

	err = reservations.ToJSON(rw)
	if err != nil {
		http.Error(rw, "Unable to convert to json:", http.StatusInternalServerError)
		r.logger.Fatal("Unable to convert to json :", err)
		return
	}
}

func (r *ReservationHandler) CreateAvailablePeriod(rw http.ResponseWriter, h *http.Request) {
	availablePeriod := h.Context().Value(KeyProduct{}).(*data.AvailablePeriodByAccommodation)
	err := r.repo.InsertAvailablePeriodByAccommodation(availablePeriod)
	if err != nil {
		r.logger.Print("Database exception: ", err)
		http.Error(rw, fmt.Sprintf("Failed to create available period: %v", err), http.StatusBadRequest)
		return
	}
	rw.WriteHeader(http.StatusCreated)
}

func (r *ReservationHandler) CreateReservation(rw http.ResponseWriter, h *http.Request) {
	reservation := h.Context().Value(KeyProduct{}).(*data.ReservationByAvailablePeriod)

	err := r.repo.InsertReservationByAvailablePeriod(reservation)
	if err != nil {
		r.logger.Print("Database exception: ", err)
		http.Error(rw, fmt.Sprintf("Failed to create reservation: %v", err), http.StatusBadRequest)
		return
	}
	rw.WriteHeader(http.StatusCreated)
}

func (r *ReservationHandler) FindAccommodationIdsByDates(rw http.ResponseWriter, h *http.Request) {
	dates := h.Context().Value(KeyProduct{}).(data.Dates)
	ids, err := r.repo.FindAccommodationIdsByDates(&dates)
	if err != nil {
		r.logger.Print("Database exception: ", err)
		http.Error(rw, fmt.Sprintf("Failed to find accommodation ids: %v", err), http.StatusBadRequest)
		return
	}
	err = ids.ToJSON(rw)
	if err != nil {
		http.Error(rw, "Unable to convert to json:", http.StatusInternalServerError)
		r.logger.Fatal("Unable to convert to json :", err)
		return
	}
}

func (r *ReservationHandler) UpdateAvailablePeriodByAccommodation(rw http.ResponseWriter, h *http.Request) {
	availablePeriod := h.Context().Value(KeyProduct{}).(*data.AvailablePeriodByAccommodation)
	err := r.repo.UpdateAvailablePeriodByAccommodation(availablePeriod)
	if err != nil {
		r.logger.Print("Database exception: ", err)
		rw.WriteHeader(http.StatusBadRequest)
		return
	}
	rw.WriteHeader(http.StatusCreated)
}

func (r *ReservationHandler) DeletePeriodsForAccommodations(rw http.ResponseWriter, h *http.Request) {
	accIDs := h.Context().Value(KeyProduct{}).([]primitive.ObjectID)
	err := r.repo.DeletePeriodsForAccommodations(accIDs)
	if err != nil {
		r.logger.Print("Database exception: ", err)
		rw.WriteHeader(http.StatusBadRequest)
		return
	}
	rw.WriteHeader(http.StatusNoContent)
}

func (r *ReservationHandler) CheckAndDeleteReservationsForUser(rw http.ResponseWriter, h *http.Request) {
	vars := mux.Vars(h)
	userID, err := primitive.ObjectIDFromHex(vars["id"])
	if err != nil {
		http.Error(rw, "Invalid ID", http.StatusBadRequest)
		return
	}

	err = r.repo.CheckAndDeleteReservationsByUserID(userID)
	if err != nil {
		r.logger.Println("Database exception: ", err)
		rw.WriteHeader(http.StatusBadRequest)
	}

	rw.WriteHeader(http.StatusNoContent)
}

func (r *ReservationHandler) DeleteReservation(rw http.ResponseWriter, h *http.Request) {
	vars := mux.Vars(h)
	periodID := vars["periodID"]
	reservationID := vars["reservationID"]

	err := r.repo.DeleteReservationByIdAndAvailablePeriodID(reservationID, periodID)
	if err != nil {
		r.logger.Println("Database exception: ", err)
		rw.WriteHeader(http.StatusNotFound)
	}

	rw.WriteHeader(http.StatusAccepted)
}

func (r *ReservationHandler) MiddlewareAvailablePeriodDeserialization(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, h *http.Request) {
		availablePeriod := &data.AvailablePeriodByAccommodation{}
		err := availablePeriod.FromJSON(h.Body)
		if err != nil {
			http.Error(rw, "Unable to decode json", http.StatusBadRequest)
			r.logger.Fatal(err)
			return
		}
		ctx := context.WithValue(h.Context(), KeyProduct{}, availablePeriod)
		h = h.WithContext(ctx)
		next.ServeHTTP(rw, h)
	})
}

func (r *ReservationHandler) MiddlewareReservationDeserialization(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, h *http.Request) {
		reservation := &data.ReservationByAvailablePeriod{}
		err := reservation.FromJSON(h.Body)
		if err != nil {
			http.Error(rw, "Unable to decode json", http.StatusBadRequest)
			r.logger.Fatal(err)
			return
		}
		ctx := context.WithValue(h.Context(), KeyProduct{}, reservation)
		h = h.WithContext(ctx)
		next.ServeHTTP(rw, h)
	})
}

func (r *ReservationHandler) MiddlewareDatesDeserialization(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, h *http.Request) {
		dates := data.Dates{}
		err := dates.FromJSON(h.Body)
		if err != nil {
			http.Error(rw, "Unable to decode json", http.StatusBadRequest)
			r.logger.Fatal(err)
			return
		}
		ctx := context.WithValue(h.Context(), KeyProduct{}, dates)
		h = h.WithContext(ctx)
		next.ServeHTTP(rw, h)
	})
}

func (r *ReservationHandler) MiddlewareContentTypeSet(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, h *http.Request) {
		rw.Header().Add("Content-Type", "application/json")

		next.ServeHTTP(rw, h)
	})
}

func (r *ReservationHandler) AuthorizeRoles(allowedRoles ...string) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, rr *http.Request) {
			tokenString := r.extractTokenFromHeader(rr)
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
				fmt.Println("allowed role : ", allowedRole)
				fmt.Println("JWT role : ", role)
				if allowedRole == role {
					next.ServeHTTP(w, rr)
					return
				}
			}

			http.Error(w, "Forbidden", http.StatusForbidden)
		})
	}
}

func (r *ReservationHandler) extractTokenFromHeader(rr *http.Request) string {
	token := rr.Header.Get("Authorization")
	if token != "" {
		return token[len("Bearer "):]
	}
	return ""
}
