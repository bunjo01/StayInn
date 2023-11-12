package handlers

import (
	"accommodation/data"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gocql/gocql"
	"github.com/gorilla/mux"
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
	id, err := gocql.ParseUUID(vars["id"])
	if err != nil {
		http.Error(rw, "Invalid UUID", http.StatusBadRequest)
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
    // Dekodiraj JSON telo zahteva u objekat Accommodation
    var accommodation data.Accommodation
    if err := json.NewDecoder(r.Body).Decode(&accommodation); err != nil {
        http.Error(rw, "Failed to decode request body", http.StatusBadRequest)
        return
    }

    // Generiši jedinstveni ID za smeštaj
    accommodation.ID = gocql.TimeUUID()

    // Kreiraj smeštaj u bazi podataka
    if err := ah.repo.CreateAccommodation(r.Context(), &accommodation); err != nil {
        ah.logger.Println("Failed to create accommodation:", err)
        http.Error(rw, "Failed to create accommodation", http.StatusInternalServerError)
        return
    }

    // Postavi odgovarajući status odgovora i pošaljite odgovor
    rw.Header().Set("Content-Type", "application/json")
    rw.WriteHeader(http.StatusCreated)
    if err := json.NewEncoder(rw).Encode(accommodation); err != nil {
        ah.logger.Println("Failed to encode accommodation:", err)
        http.Error(rw, "Failed to encode accommodation", http.StatusInternalServerError)
    }
}


func (ah *AccommodationHandler) UpdateAccommodation(rw http.ResponseWriter, r *http.Request) {
    // Izvuci ID smeštaja iz putanje
    vars := mux.Vars(r)
    id, err := gocql.ParseUUID(vars["id"])
    if err != nil {
        http.Error(rw, "Invalid UUID", http.StatusBadRequest)
        return
    }

    // Dekodiraj JSON telo zahteva u objekat Accommodation
    var updatedAccommodation data.Accommodation
    if err := json.NewDecoder(r.Body).Decode(&updatedAccommodation); err != nil {
        http.Error(rw, "Failed to decode request body", http.StatusBadRequest)
        return
    }

    // Postavi ID smeštaja na vrednost iz putanje
    updatedAccommodation.ID = id

    // Ažuriraj smeštaj u bazi podataka
    if err := ah.repo.UpdateAccommodation(r.Context(), &updatedAccommodation); err != nil {
        ah.logger.Println("Failed to update accommodation:", err)
        http.Error(rw, "Failed to update accommodation", http.StatusInternalServerError)
        return
    }

    // Vrati ažurirani smeštaj kao odgovor
    rw.Header().Set("Content-Type", "application/json")
    rw.WriteHeader(http.StatusOK)
    if err := json.NewEncoder(rw).Encode(updatedAccommodation); err != nil {
        ah.logger.Println("Failed to encode updated accommodation:", err)
        http.Error(rw, "Failed to encode updated accommodation", http.StatusInternalServerError)
    }
}


func (ah *AccommodationHandler) DeleteAccommodation(rw http.ResponseWriter, r *http.Request) {
    // Izvuci ID smeštaja iz putanje
    vars := mux.Vars(r)
    id, err := gocql.ParseUUID(vars["id"])
    if err != nil {
        http.Error(rw, "Invalid UUID", http.StatusBadRequest)
        return
    }

    // Obriši smeštaj iz baze podataka
    if err := ah.repo.DeleteAccommodation(r.Context(), id); err != nil {
        ah.logger.Println("Failed to delete accommodation:", err)
        http.Error(rw, "Failed to delete accommodation", http.StatusInternalServerError)
        return
    }

    // Postavi odgovarajući status odgovora
    rw.WriteHeader(http.StatusNoContent)
}
