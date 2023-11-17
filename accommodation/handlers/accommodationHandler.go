package handlers

import (
	"accommodation/data"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AccommodationHandler struct {
	logger *log.Logger
	repo   *data.AccommodationRepository
}

func NewAccommodationsHandler(logger *log.Logger, repo *data.AccommodationRepository) *AccommodationHandler {
	return &AccommodationHandler{logger: logger, repo: repo}
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
