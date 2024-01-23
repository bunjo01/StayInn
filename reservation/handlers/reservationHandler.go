package handlers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"reservation/clients"
	"reservation/data"
)

type KeyProduct struct{}

type ReservationHandler struct {
	repo          *data.ReservationRepo
	notification  clients.NotificationClient
	profile       clients.ProfileClient
	accommodation clients.AccommodationClient
}

var secretKey = []byte("stayinn_secret")

func NewReservationHandler(r *data.ReservationRepo, n clients.NotificationClient,
	p clients.ProfileClient, a clients.AccommodationClient) *ReservationHandler {
	return &ReservationHandler{r, n, p, a}
}

func (r *ReservationHandler) GetAllAvailablePeriodsByAccommodation(rw http.ResponseWriter, h *http.Request) {
	vars := mux.Vars(h)
	id := vars["id"]

	availablePeriods, err := r.repo.GetAvailablePeriodsByAccommodation(id)
	if err != nil {
		log.Info(fmt.Sprintf("[res-service]rh#1 Error while finding available period by accommodation: %v", err))
	}

	if availablePeriods == nil {
		return
	}

	err = availablePeriods.ToJSON(rw)
	if err != nil {
		http.Error(rw, "Unable to convert to json:", http.StatusInternalServerError)
		log.Info(fmt.Sprintf("[res-service]rh#2 Error while converting from json: %v", err))
		return
	}

	log.Info(fmt.Sprintf("[res-service]rh#48 Successfuly retrieved available periods by accommodation"))
}

func (r *ReservationHandler) FindAvailablePeriodByIdAndByAccommodationId(rw http.ResponseWriter, h *http.Request) {
	vars := mux.Vars(h)
	periodID := vars["periodID"]
	accomodationID := vars["accomodationID"]

	availablePeriod, err := r.repo.FindAvailablePeriodById(periodID, accomodationID)
	if err != nil {
		log.Info(fmt.Sprintf("[res-service]rh#3 Error while finding period by id: %v", err))
	}

	if availablePeriod == nil {
		log.Info(fmt.Sprintf("[res-service]rh#4 Error while finding period by id: %v", err))
		return
	}

	err = availablePeriod.ToJSON(rw)
	if err != nil {
		http.Error(rw, "Unable to convert to json:", http.StatusInternalServerError)
		log.Info(fmt.Sprintf("[res-service]rh#5 Error while converting json: %v", err))
		return
	}
	log.Info(fmt.Sprintf("[res-service]rh#49 Successfuly retrieved available periods by id and accommodation"))
}

func (r *ReservationHandler) GetAllReservationByAvailablePeriod(rw http.ResponseWriter, h *http.Request) {
	vars := mux.Vars(h)
	id := vars["id"]

	reservations, err := r.repo.GetReservationsByAvailablePeriod(id)
	if err != nil {
		log.Info(fmt.Sprintf("[res-service]rh#6 Error while finding reservations by period: %v", err))
	}

	if reservations == nil {
		return
	}

	err = reservations.ToJSON(rw)
	if err != nil {
		http.Error(rw, "Unable to convert to json:", http.StatusInternalServerError)
		log.Info(fmt.Sprintf("[res-service]rh#7 Error while converting json: %v", err))
		return
	}
	log.Info(fmt.Sprintf("[res-service]rh#50 Successfuly retrieved reservation by avilable period"))
}

func (r *ReservationHandler) CreateAvailablePeriod(rw http.ResponseWriter, h *http.Request) {
	availablePeriod := h.Context().Value(KeyProduct{}).(*data.AvailablePeriodByAccommodation)

	tokenStr := r.extractTokenFromHeader(h)
	username, err := r.getUsername(tokenStr)
	if err != nil {
		log.Warning(fmt.Sprintf("[res-service]rh#8 Error while reading username from token: %v", err))
		http.Error(rw, "Failed to read username from token", http.StatusBadRequest)
		return
	}

	userID, err := r.profile.GetUserId(h.Context(), username, tokenStr)
	if err != nil {
		log.Info(fmt.Sprintf("[res-service]rh#9 Error while getting hostId for username: %v", err))
		http.Error(rw, "Failed to get HostID from username", http.StatusBadRequest)
		return
	}

	availablePeriod.IDUser, err = primitive.ObjectIDFromHex(userID)
	if err != nil {
		log.Info(fmt.Sprintf("[res-service]rh#10 Error while getting hostId for accommodation: %v", err))
		http.Error(rw, "Failed to set HostID for accommodation", http.StatusBadRequest)
		return
	}

	_, err = r.accommodation.CheckAccommodationID(h.Context(), availablePeriod.IDAccommodation, tokenStr)
	if err != nil {
		log.Info(fmt.Sprintf("[res-service]rh#11 Error while getting accommodation by id: %v", err))
		http.Error(rw, "Failed to get accommodation by Id", http.StatusBadRequest)
		return
	}

	exists, err := r.accommodation.CheckAccommodationID(h.Context(), availablePeriod.IDAccommodation, tokenStr)
	if err != nil {
		log.Info(fmt.Sprintf("[res-service]rh#12 Error while getting accommodation by id: %v", err))
		http.Error(rw, "Failed to check accommodation existence", http.StatusInternalServerError)
		return
	}

	if !exists {
		log.Info(fmt.Sprintf("[res-service]rh#13 Error while getting accommodation by id: %v", err))
		http.Error(rw, "Accommodation does not exist", http.StatusBadRequest)
		return
	}

	err = r.repo.InsertAvailablePeriodByAccommodation(availablePeriod)
	if err != nil {
		log.Info(fmt.Sprintf("[res-service]rh#14 Error while inserting in database: %v", err))
		http.Error(rw, fmt.Sprintf("Failed to create available period: %v", err), http.StatusBadRequest)
		return
	}

	log.Info(fmt.Sprintf("[res-service]rh#51 User id:'%s' successfuly created available period", userID))

	rw.WriteHeader(http.StatusCreated)
}

func (r *ReservationHandler) CreateReservation(rw http.ResponseWriter, h *http.Request) {
	reservation := h.Context().Value(KeyProduct{}).(*data.ReservationByAvailablePeriod)

	tokenStr := r.extractTokenFromHeader(h)
	username, err := r.getUsername(tokenStr)
	if err != nil {
		log.Warning(fmt.Sprintf("[res-service]rh#15 Error while reading username from token: %v", err))
		http.Error(rw, "Failed to read username from token", http.StatusBadRequest)
		return
	}

	userID, err := r.profile.GetUserId(h.Context(), username, tokenStr)
	if err != nil {
		log.Info(fmt.Sprintf("[res-service]rh#16 Error while getting hostId for username: %v", err))
		http.Error(rw, "Failed to get HostID from username", http.StatusBadRequest)
		return
	}

	reservation.IDUser, err = primitive.ObjectIDFromHex(userID)
	if err != nil {
		log.Info(fmt.Sprintf("[res-service]rh#17 Error while getting hostId for accommodation: %v", err))
		http.Error(rw, "Failed to set HostID for accommodation", http.StatusBadRequest)
		return
	}

	exists, err := r.accommodation.CheckAccommodationID(h.Context(), reservation.IDAccommodation, tokenStr)
	if !exists {
		log.Info(fmt.Sprintf("[res-service]rh#18 Error while getting accommodation by id: %v", err))
		http.Error(rw, "Accommodation does not exist", http.StatusBadRequest)
		return
	}

	if err != nil {
		log.Info(fmt.Sprintf("[res-service]rh#19 Error while getting accommodation by id: %v", err))
		http.Error(rw, "Failed to check accommodation existence", http.StatusInternalServerError)
		return
	}

	err = r.repo.InsertReservationByAvailablePeriod(reservation)
	if err != nil {
		log.Info(fmt.Sprintf("[res-service]rh#20 Error while inserting in database: %v", err))
		http.Error(rw, fmt.Sprintf("Failed to create reservation: %v", err), http.StatusBadRequest)
		return
	}

	// Get accommodation
	accommodation, err := r.accommodation.GetAccommodationByID(h.Context(), reservation.IDAccommodation, tokenStr)
	if err != nil {
		log.Info(fmt.Sprintf("[res-service]rh#21 Error while finding accommodation by id: %v", err))
		http.Error(rw, "failed to get accommodation by ID", http.StatusInternalServerError)
		return
	}

	// Get host
	host, err := r.profile.GetUserById(h.Context(), accommodation.HostID, tokenStr)
	if err != nil {
		log.Info(fmt.Sprintf("[res-service]rh#22 Error while finding host by id: %v", err))
		http.Error(rw, "failed to get host by ID", http.StatusInternalServerError)
		return
	}

	// Notify host
	notification := data.Notification{
		HostID:       host.ID,
		HostUsername: host.Username,
		HostEmail:    host.Email,
		Text:         fmt.Sprintf("Reservation created for %s, by user %s", accommodation.Name, username),
		Time:         time.Now(),
	}

	notified, err := r.notification.NotifyReservation(h.Context(), notification, tokenStr)
	if !notified {
		log.Info(fmt.Sprintf("[res-service]rh#23 Error while trying to notify host: %v", err))
		http.Error(rw, "failed to notify host", http.StatusInternalServerError)
		return
	}

	log.Info(fmt.Sprintf("[res-service]rh#52 User id:'%s' successfuly created reservation", userID))

	rw.WriteHeader(http.StatusCreated)
}

func (r *ReservationHandler) FindAccommodationIdsByDates(rw http.ResponseWriter, h *http.Request) {
	dates := h.Context().Value(KeyProduct{}).(data.Dates)
	ids, err := r.repo.FindAccommodationIdsByDates(&dates)
	if err != nil {
		log.Info(fmt.Sprintf("[res-service]rh#24 Error while finding accommodation by id and dates: %v", err))
		http.Error(rw, fmt.Sprintf("Failed to find accommodation ids: %v", err), http.StatusBadRequest)
		return
	}
	err = ids.ToJSON(rw)
	if err != nil {
		log.Info(fmt.Sprintf("[res-service]rh#25 Error while trying to convert json: %v", err))
		http.Error(rw, "Unable to convert to json:", http.StatusInternalServerError)
		return
	}

	log.Info(fmt.Sprintf("[res-service]rh#53 Successfuly retrieved accommodation ids by dates"))

	rw.WriteHeader(http.StatusOK)
}

func (r *ReservationHandler) FindAllReservationsByUserIDExpired(rw http.ResponseWriter, h *http.Request) {
	tokenStr := r.extractTokenFromHeader(h)
	username, err := r.getUsername(tokenStr)
	if err != nil {
		log.Warning(fmt.Sprintf("[res-service]rh#26 Error while reading username from token: %v", err))
		http.Error(rw, "Failed to read username from token", http.StatusBadRequest)
		return
	}

	guestID, err := r.profile.GetUserId(h.Context(), username, tokenStr)
	if err != nil {
		log.Info(fmt.Sprintf("[res-service]rh#27 Error while getting hostId for username: %v", err))
		http.Error(rw, "Failed to get guestID from username", http.StatusBadRequest)
		return
	}

	reservations, err := r.repo.FindAllReservationsByUserIDExpired(guestID)
	if err != nil {
		log.Info(fmt.Sprintf("[res-service]rh#28 Error while finding expired reservations by user: %v", err))
		http.Error(rw, "Failed to fetch expired reservations", http.StatusBadRequest)
		return
	}

	err = reservations.ToJSON(rw)
	if err != nil {
		log.Info(fmt.Sprintf("[res-service]rh#29 Error while converting json: %v", err))
		http.Error(rw, "Unable to convert to JSON", http.StatusBadRequest)
		return
	}

	log.Info(fmt.Sprintf("[res-service]rh#54 User id:'%s' successfully retrieved expired reservations by user", guestID))

	rw.WriteHeader(http.StatusOK)
}

func (r *ReservationHandler) UpdateAvailablePeriodByAccommodation(rw http.ResponseWriter, h *http.Request) {
	availablePeriod := h.Context().Value(KeyProduct{}).(*data.AvailablePeriodByAccommodation)
	tokenStr := r.extractTokenFromHeader(h)
	username, err := r.getUsername(tokenStr)
	if err != nil {
		log.Warning(fmt.Sprintf("[res-service]rh#30 Error while reading username from token: %v", err))
		http.Error(rw, "Failed to read username from token", http.StatusBadRequest)
		return
	}

	userID, err := r.profile.GetUserId(h.Context(), username, tokenStr)
	if err != nil {
		log.Info(fmt.Sprintf("[res-service]rh#31 Error while reading host id from username: %v", err))
		http.Error(rw, "Failed to get HostID from username", http.StatusBadRequest)
		return
	}

	if availablePeriod.IDUser.Hex() != userID {
		log.Info(fmt.Sprintf("[res-service]rh#32 User is not owner of period: %v", err))
		http.Error(rw, "You are not the owner of available period", http.StatusBadRequest)
		return
	}

	err = r.repo.UpdateAvailablePeriodByAccommodation(availablePeriod)
	if err != nil {
		log.Info(fmt.Sprintf("[res-service]rh#33 Error while updating period by accommodation: %v", err))
		rw.WriteHeader(http.StatusBadRequest)
		return
	}
	log.Info(fmt.Sprintf("[res-service]rh#55 User id:'%s' successfully updated available period by accommodation id:'%s'", userID, availablePeriod.ID.String()))
	rw.WriteHeader(http.StatusCreated)
}

func (r *ReservationHandler) DeletePeriodsForAccommodations(rw http.ResponseWriter, h *http.Request) {
	if h.Context().Value(KeyProduct{}) != nil {
		accIDs := h.Context().Value(KeyProduct{}).([]primitive.ObjectID)
		if (accIDs != nil) && len(accIDs) > 0 {
			err := r.repo.DeletePeriodsForAccommodations(accIDs)
			if err != nil {
				log.Info(fmt.Sprintf("[res-service]rh#34 Error while deleting periods by accommodation: %v", err))
				rw.WriteHeader(http.StatusBadRequest)
				return
			}
		}
	}

	log.Info(fmt.Sprintf("[res-service]rh#56 Successfully deleted available period by accommodations"))
	rw.WriteHeader(http.StatusNoContent)
}

func (r *ReservationHandler) GetAllReservationsByUser(rw http.ResponseWriter, h *http.Request) {
	tokenStr := r.extractTokenFromHeader(h)
	vars := mux.Vars(h)
	username := vars["username"]

	userID, err := r.profile.GetUserId(h.Context(), username, tokenStr)
	if err != nil {
		log.Warning(fmt.Sprintf("[res-service]rh#35 Error while reading username from token: %v", err))
		http.Error(rw, "Failed to get UserID from username", http.StatusBadRequest)
		return
	}

	reservations, err := r.repo.FindAllReservationsByUserID(userID)
	if err != nil {
		log.Info(fmt.Sprintf("[res-service]rh#36 Error while finding reservations by userId: %v", err))
		rw.WriteHeader(http.StatusBadRequest)
	}

	err = reservations.ToJSON(rw)
	if err != nil {
		log.Info(fmt.Sprintf("[res-service]rh#37 Error while converting json: %v", err))
		http.Error(rw, "Unable to convert to json:", http.StatusInternalServerError)
		return
	}
	log.Info(fmt.Sprintf("[res-service]rh#56 Successfully retrieved reservtions by user"))
	rw.WriteHeader(http.StatusOK)
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
		log.Info(fmt.Sprintf("[res-service]rh#38 Error while checking and deleting reservations by user id : %v", err))
		rw.WriteHeader(http.StatusBadRequest)
	}
	log.Info(fmt.Sprintf("[res-service]rh#57 Successfully deleted reservtions for user id:'%s'", userID))

	rw.WriteHeader(http.StatusNoContent)
}

func (r *ReservationHandler) DeleteReservation(rw http.ResponseWriter, h *http.Request) {
	vars := mux.Vars(h)
	periodID := vars["periodID"]
	reservationID := vars["reservationID"]
	tokenStr := r.extractTokenFromHeader(h)
	username, err := r.getUsername(tokenStr)
	if err != nil {
		log.Warning(fmt.Sprintf("[res-service]rh#39 Error while reading username from token: %v", err))
		http.Error(rw, "Failed to read username from token", http.StatusBadRequest)
		return
	}

	userID, err := r.profile.GetUserId(h.Context(), username, tokenStr)
	if err != nil {
		log.Info(fmt.Sprintf("[res-service]rh#40 Error while reading host id from username: %v", err))
		http.Error(rw, "Failed to get HostID from username", http.StatusBadRequest)
		return
	}

	// Get reservation
	reservation, err := r.repo.FindReservationByIdAndAvailablePeriod(reservationID, periodID)
	if err != nil {
		log.Info(fmt.Sprintf("[res-service]rh#41 Error while finding reservation by id and period: %v", err))
		http.Error(rw, "failed to get reservation by ID", http.StatusInternalServerError)
		return
	}

	err = r.repo.DeleteReservationByIdAndAvailablePeriodID(reservationID, periodID, userID)
	if err != nil {
		log.Info(fmt.Sprintf("[res-service]rh#42 Error while reading host id from usenrame: %v", err))
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	// Get period
	period, err := r.repo.FindAvailablePeriodsByAccommodationId(reservation.IDAccommodation.Hex())
	if err != nil {
		log.Info(fmt.Sprintf("[res-service]rh#43 Error while finding period by accommodation id: %v", err))
		http.Error(rw, "failed to get period by accommodation ID", http.StatusInternalServerError)
		return
	}

	// Get host
	host, err := r.profile.GetUserById(h.Context(), period[0].IDUser, tokenStr)
	if err != nil {
		log.Info(fmt.Sprintf("[res-service]rh#44 Error while finding user by id: %v", err))
		http.Error(rw, "failed to get host by ID", http.StatusInternalServerError)
		return
	}

	startDate := reservation.StartDate.Format("02. January 2006.")
	endDate := reservation.EndDate.Format("02. January 2006.")

	// Notify host
	notification := data.Notification{
		HostID:       host.ID,
		HostUsername: host.Username,
		HostEmail:    host.Email,
		Text:         fmt.Sprintf("Reservation from %s to %s deleted by user %s", startDate, endDate, username),
		Time:         time.Now(),
	}

	notified, err := r.notification.NotifyReservation(h.Context(), notification, tokenStr)
	if !notified {
		log.Info(fmt.Sprintf("[res-service]rh#44 Error while trying to notify host: %v", err))
		http.Error(rw, "failed to notify host", http.StatusInternalServerError)
		return
	}
	log.Info(fmt.Sprintf("[res-service]rh#58 User id:'%s' successfully delete reservation id:'%s'", userID, reservationID))

	rw.WriteHeader(http.StatusAccepted)
}

func (r *ReservationHandler) MiddlewareAvailablePeriodDeserialization(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, h *http.Request) {
		availablePeriod := &data.AvailablePeriodByAccommodation{}
		err := availablePeriod.FromJSON(h.Body)
		if err != nil {
			log.Info(fmt.Sprintf("[res-service]rh#45 Error while trying to convert json: %v", err))
			http.Error(rw, "Unable to decode json", http.StatusBadRequest)
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
			log.Info(fmt.Sprintf("[res-service]rh#46 Error while trying to convert json: %v", err))
			http.Error(rw, "Unable to decode json", http.StatusBadRequest)
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
			log.Info(fmt.Sprintf("[res-service]rh#47 Error while trying to convert json: %v", err))
			http.Error(rw, "Unable to decode json", http.StatusBadRequest)
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

func (r *ReservationHandler) getUsername(tokenString string) (string, error) {
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
