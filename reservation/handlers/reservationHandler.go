package handlers

import (
	"context"
	"fmt"
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

func (r *ReservationHandler) ReservePeriod(rw http.ResponseWriter, h *http.Request) {
	vars := mux.Vars(h)
	reservationId := vars["reservationId"]
	periodId := vars["periodId"]

	reservation, err := r.repo.GetById(reservationId)
	if reservation == nil {
		http.Error(rw, "Reservation with given id not found", http.StatusNotFound)
		r.logger.Printf("Reservation with id: '%s' not found", reservationId)
		return
	}

	err = reservation.ToJSON(rw)
	if err != nil {
		http.Error(rw, "Unable to convert to json", http.StatusInternalServerError)
		r.logger.Fatal("Unable to convert to json :", err)
		return
	}

	r.repo.ReservePeriod(reservationId, periodId)
	rw.WriteHeader(http.StatusOK)
}

func (r *ReservationHandler) PostReservation(rw http.ResponseWriter, h *http.Request) {
	reservation := h.Context().Value(KeyProduct{}).(*data.Reservation)
	r.repo.PostReservation(reservation)
	rw.WriteHeader(http.StatusCreated)
}

func (r *ReservationHandler) UpdatePeriod(rw http.ResponseWriter, h *http.Request) {
	vars := mux.Vars(h)
	reservaationId := vars["reservationId"]

	period := h.Context().Value(KeyProduct{}).(*data.AvailabilityPeriod)
	r.repo.UpdateAvailablePeriod(reservaationId, period)
	rw.WriteHeader(http.StatusCreated)
}

func (r *ReservationHandler) AddAvaiablePeriod(rw http.ResponseWriter, h *http.Request) {
	vars := mux.Vars(h)
	id := vars["id"]

	// Debugging: Print id to check its value
	fmt.Printf("ID: %s\n", id)

	period, ok := h.Context().Value(KeyProduct{}).(*data.AvailabilityPeriod)

	// Debugging: Print period and ok to inspect them
	fmt.Printf("Value from context: %#v, Conversion success: %v\n", period, ok)

	if !ok {
		log.Printf("Error: Conversion failed, value is not of type *data.AvailabilityPeriod")
		// Handle the error appropriately
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	r.repo.AddAvaiablePeriod(id, period)
	rw.WriteHeader(http.StatusOK)
}

func (r *ReservationHandler) DeleteReservation(rw http.ResponseWriter, h *http.Request) {
	vars := mux.Vars(h)
	id := vars["id"]

	r.repo.DeleteById(id)
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

func (r *ReservationHandler) MiddlewareAvaiablePeriodsDeserialization(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, h *http.Request) {
		avaiablePeriod := &data.AvailabilityPeriod{}
		err := avaiablePeriod.FromJSON(h.Body)
		if err != nil {
			http.Error(rw, "Unable to decode json", http.StatusBadRequest)
			r.logger.Fatal(err)
			return
		}

		ctx := context.WithValue(h.Context(), KeyProduct{}, avaiablePeriod)
		h = h.WithContext(ctx)

		println(avaiablePeriod)

		next.ServeHTTP(rw, h)
	})
}

func (r *ReservationHandler) MiddlewareContentTypeSet(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, h *http.Request) {
		rw.Header().Add("Content-Type", "application/json")

		next.ServeHTTP(rw, h)
	})
}
