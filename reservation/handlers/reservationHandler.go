package handlers

import (
	"context"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"main.go/data"
)

type KeyProduct struct{}

type ReservationHandler struct {
	logger *log.Logger
	repo   *data.ReservationRepo
}

func NewReservationHandler(l *log.Logger, r *data.ReservationRepo) *ReservationHandler {
	return &ReservationHandler{l, r}
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
		r.logger.Println("Ne postoji period sa datim ID-em ili u datom smestaju")
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
		rw.WriteHeader(http.StatusBadRequest)
		return
	}
	rw.WriteHeader(http.StatusCreated)
}

func (r *ReservationHandler) CreateReservation(rw http.ResponseWriter, h *http.Request) {
	reservation := h.Context().Value(KeyProduct{}).(*data.ReservationByAvailablePeriod)

	err := r.repo.InsertReservationByAvailablePeriod(reservation)
	if err != nil {
		r.logger.Print("Database exception: ", err)
		rw.WriteHeader(http.StatusConflict)
		return
	}
	rw.WriteHeader(http.StatusCreated)
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

func (r *ReservationHandler) DeleteReservation(rw http.ResponseWriter, h *http.Request) {
	vars := mux.Vars(h)
	periodID := vars["periodID"]
	reservationID := vars["reservationID"]

	err := r.repo.DeleteReservationByIdAndAvailablePeriodID(reservationID, periodID)
	if err != nil {
		r.logger.Println("Database exception: ", err)
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

func (r *ReservationHandler) MiddlewareContentTypeSet(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, h *http.Request) {
		rw.Header().Add("Content-Type", "application/json")

		next.ServeHTTP(rw, h)
	})
}
