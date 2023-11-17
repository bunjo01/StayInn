package handlers

import (
	"context"
	"github.com/gorilla/mux"
	"log"
	"main.go/data"
	"net/http"
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
		rw.WriteHeader(http.StatusBadRequest)
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

//
//func (r *ReservationHandler) GetReservationById(rw http.ResponseWriter, h *http.Request) {
//	vars := mux.Vars(h)
//	id := vars["id"]
//
//	reservation, err := r.repo.GetReservationById(id)
//	if err != nil {
//		r.logger.Print("Database exception: ", err)
//	}
//
//	if reservation == nil {
//		http.Error(rw, "Reservation with given id not found", http.StatusNotFound)
//		r.logger.Printf("Reservation with id: '%s' not found", id)
//		return
//	}
//
//	err = reservation.ToJSON(rw)
//	if err != nil {
//		http.Error(rw, "Unable to convert to json", http.StatusInternalServerError)
//		r.logger.Fatal("Unable to convert to json :", err)
//		return
//	}
//}
//
//func (r *ReservationHandler) PostReservation(rw http.ResponseWriter, h *http.Request) {
//	reservation := h.Context().Value(KeyProduct{}).(*data.Reservation)
//	r.repo.PostReservation(reservation)
//	rw.WriteHeader(http.StatusCreated)
//}
//
//// TODO : NOT WORKING
//func (r *ReservationHandler) UpdateReservedPeriod(rw http.ResponseWriter, h *http.Request) {
//	vars := mux.Vars(h)
//	reservationId := vars["reservationId"]
//
//	period := h.Context().Value(KeyProduct{}).(*data.ReservedPeriod)
//	err := r.repo.UpdateReservedPeriod(reservationId, period)
//	if err != nil {
//		rw.WriteHeader(http.StatusNotFound)
//		return
//	}
//	rw.WriteHeader(http.StatusCreated)
//}
//
//func (r *ReservationHandler) ReservePeriod(rw http.ResponseWriter, h *http.Request) {
//	vars := mux.Vars(h)
//	id := vars["id"]
//
//	// Debugging: Print id to check its value
//	fmt.Printf("ID: %s\n", id)
//
//	period, ok := h.Context().Value(KeyProduct{}).(*data.ReservedPeriod)
//
//	// Debugging: Print period and ok to inspect them
//	fmt.Printf("Value from context: %#v, Conversion success: %v\n", period, ok)
//
//	if !ok {
//		log.Printf("Error: Conversion failed, value is not of type *data.AvailabilityPeriod")
//		// Handle the error appropriately
//		rw.WriteHeader(http.StatusInternalServerError)
//		return
//	}
//
//	r.repo.AddReservedPeriod(id, period)
//	rw.WriteHeader(http.StatusOK)
//}
//
//func (r *ReservationHandler) DeleteReservation(rw http.ResponseWriter, h *http.Request) {
//	vars := mux.Vars(h)
//	id := vars["id"]
//
//	r.repo.DeleteReservationById(id)
//	rw.WriteHeader(http.StatusOK)
//}
//
//func (r *ReservationHandler) DeleteReservedPeriod(rw http.ResponseWriter, h *http.Request) {
//	vars := mux.Vars(h)
//	reservationId := vars["reservationId"]
//	periodId := vars["periodId"]
//
//	err := r.repo.DeleteReservedPeriod(reservationId, periodId)
//	if err != nil {
//		rw.WriteHeader(http.StatusNotFound)
//		return
//	}
//	rw.WriteHeader(http.StatusOK)
//}
//
//func (r *ReservationHandler) MiddlewareReservationDeserialization(next http.Handler) http.Handler {
//	return http.HandlerFunc(func(rw http.ResponseWriter, h *http.Request) {
//		reservation := &data.Reservation{}
//		err := reservation.FromJSON(h.Body)
//		if err != nil {
//			http.Error(rw, "Unable to decode json", http.StatusBadRequest)
//			r.logger.Fatal(err)
//			return
//		}
//
//		ctx := context.WithValue(h.Context(), KeyProduct{}, reservation)
//		h = h.WithContext(ctx)
//
//		next.ServeHTTP(rw, h)
//	})
//}
//
//func (r *ReservationHandler) MiddlewareReservedPeriodDeserialization(next http.Handler) http.Handler {
//	return http.HandlerFunc(func(rw http.ResponseWriter, h *http.Request) {
//		avaiablePeriod := &data.ReservedPeriod{}
//		err := avaiablePeriod.FromJSON(h.Body)
//		if err != nil {
//			http.Error(rw, "Unable to decode json", http.StatusBadRequest)
//			r.logger.Fatal(err)
//			return
//		}
//
//		ctx := context.WithValue(h.Context(), KeyProduct{}, avaiablePeriod)
//		h = h.WithContext(ctx)
//
//		println(avaiablePeriod)
//
//		next.ServeHTTP(rw, h)
//	})
//}
//
