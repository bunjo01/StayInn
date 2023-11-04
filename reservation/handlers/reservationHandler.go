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

func (r *ReservationHandler) GetAllReservations(rw http.ResponseWriter, h *http.Request) {
	reservations, err := r.repo.GetAll()
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

func (r *ReservationHandler) GetReservationById(rw http.ResponseWriter, h *http.Request) {
	vars := mux.Vars(h)
	id := vars["id"]

	reservation, err := r.repo.GetById(id)
	if err != nil {
		r.logger.Print("Database exception: ", err)
	}

	if reservation == nil {
		http.Error(rw, "Reservation with given id not found", http.StatusNotFound)
		r.logger.Printf("Reservation with id: '%s' not found", id)
		return
	}

	err = reservation.ToJSON(rw)
	if err != nil {
		http.Error(rw, "Unable to convert to json", http.StatusInternalServerError)
		r.logger.Fatal("Unable to convert to json :", err)
		return
	}
}

func (r *ReservationHandler) PostReservation(rw http.ResponseWriter, h *http.Request) {
	reservation := h.Context().Value(KeyProduct{}).(*data.Reservation)
	r.repo.Insert(reservation)
	rw.WriteHeader(http.StatusCreated)
}

func (r *ReservationHandler) DeleteReservation(rw http.ResponseWriter, h *http.Request) {
	vars := mux.Vars(h)
	id := vars["id"]

	r.repo.Delete(id)
	rw.WriteHeader(http.StatusOK)
}

func (r *ReservationHandler) MiddlewareReservationDeserialization(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, h *http.Request) {
		reservation := &data.Reservation{}
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
