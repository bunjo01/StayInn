package handlers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"reservation/clients"
	"reservation/data"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

const UnableToConvertToJson = "Unable to convert to json:"
const FailedToReadUsernameFromToken = "Failed to read username from token"
const FailedToGetHostIDFromUsername = "Failed to get HostID from username"
const UnableToDecodeJson = "Unable to decode json"

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

	log.Info(fmt.Sprintf("[rese-handler]rh#59 Received request from '%s' for getting all available periods", h.RemoteAddr))

	availablePeriods, err := r.repo.GetAvailablePeriodsByAccommodation(id)
	if err != nil {
		log.Error(fmt.Sprintf("[rese-handler]rh#1 Error while finding available period by accommodation: %v", err))
	}

	if availablePeriods == nil {
		return
	}

	err = availablePeriods.ToJSON(rw)
	if err != nil {
		http.Error(rw, UnableToConvertToJson, http.StatusInternalServerError)
		log.Error(fmt.Sprintf("[rese-handler]rh#2 Error while converting from json: %v", err))
		return
	}

	log.Info(fmt.Sprintf("[rese-handler]rh#48 Successfuly retrieved available periods by accommodation"))
}

func (r *ReservationHandler) FindAvailablePeriodByIdAndByAccommodationId(rw http.ResponseWriter, h *http.Request) {
	vars := mux.Vars(h)
	periodID := vars["periodID"]
	accomodationID := vars["accomodationID"]

	log.Info(fmt.Sprintf("[rese-handler]rh#60 Received request from '%s' for getting period by id", h.RemoteAddr))

	availablePeriod, err := r.repo.FindAvailablePeriodById(periodID, accomodationID)
	if err != nil {
		log.Error(fmt.Sprintf("[rese-handler]rh#3 Error while finding period by id: %v", err))
	}

	if availablePeriod == nil {
		log.Error(fmt.Sprintf("[rese-handler]rh#4 Error while finding period by id: %v", err))
		return
	}

	err = availablePeriod.ToJSON(rw)
	if err != nil {
		http.Error(rw, UnableToConvertToJson, http.StatusInternalServerError)
		log.Error(fmt.Sprintf("[rese-handler]rh#5 Error while converting json: %v", err))
		return
	}
	log.Info(fmt.Sprintf("[rese-handler]rh#49 Successfuly retrieved available period by id"))
}

func (r *ReservationHandler) GetAllReservationByAvailablePeriod(rw http.ResponseWriter, h *http.Request) {
	vars := mux.Vars(h)
	id := vars["id"]

	log.Info(fmt.Sprintf("[rese-handler]rh#61 Received request from '%s' for all reservations by available period", h.RemoteAddr))

	reservations, err := r.repo.GetReservationsByAvailablePeriod(id)
	if err != nil {
		log.Error(fmt.Sprintf("[rese-handler]rh#6 Error while finding reservations by period: %v", err))
	}

	if reservations == nil {
		return
	}

	err = reservations.ToJSON(rw)
	if err != nil {
		http.Error(rw, UnableToConvertToJson, http.StatusInternalServerError)
		log.Error(fmt.Sprintf("[rese-handler]rh#7 Error while converting json: %v", err))
		return
	}
	log.Info(fmt.Sprintf("[rese-handler]rh#50 Successfuly retrieved reservations by avilable period"))
}

func (r *ReservationHandler) CreateAvailablePeriod(rw http.ResponseWriter, h *http.Request) {
	availablePeriod := h.Context().Value(KeyProduct{}).(*data.AvailablePeriodByAccommodation)

	tokenStr := r.extractTokenFromHeader(h)
	username, err := r.getUsername(tokenStr)
	if err != nil {
		log.Warning(fmt.Sprintf("[rese-handler]rh#8 Error while reading username from token: %v", err))
		http.Error(rw, FailedToReadUsernameFromToken, http.StatusBadRequest)
		return
	}

	log.Info(fmt.Sprintf("[rese-handler]rh#62 Received request from '%s' to create available period", h.RemoteAddr))

	userID, err := r.profile.GetUserId(h.Context(), username, tokenStr)
	if err != nil {
		log.Error(fmt.Sprintf("[rese-handler]rh#9 Error while getting hostId for username: %v", err))
		http.Error(rw, FailedToGetHostIDFromUsername, http.StatusBadRequest)
		return
	}

	availablePeriod.IDUser, err = primitive.ObjectIDFromHex(userID)
	if err != nil {
		log.Error(fmt.Sprintf("[rese-handler]rh#10 Error while getting hostId for accommodation: %v", err))
		http.Error(rw, "Failed to set HostID for accommodation", http.StatusBadRequest)
		return
	}

	_, err = r.accommodation.CheckAccommodationID(h.Context(), availablePeriod.IDAccommodation, tokenStr)
	if err != nil {
		log.Error(fmt.Sprintf("[rese-handler]rh#11 Error while getting accommodation by id: %v", err))
		http.Error(rw, "Failed to get accommodation by Id", http.StatusBadRequest)
		return
	}

	exists, err := r.accommodation.CheckAccommodationID(h.Context(), availablePeriod.IDAccommodation, tokenStr)
	if err != nil {
		log.Error(fmt.Sprintf("[rese-handler]rh#12 Error while getting accommodation by id: %v", err))
		http.Error(rw, "Failed to check accommodation existence", http.StatusInternalServerError)
		return
	}

	if !exists {
		log.Error(fmt.Sprintf("[rese-handler]rh#13 Error while getting accommodation by id: %v", err))
		http.Error(rw, "Accommodation does not exist", http.StatusBadRequest)
		return
	}

	err = r.repo.InsertAvailablePeriodByAccommodation(availablePeriod)
	if err != nil {
		log.Error(fmt.Sprintf("[rese-handler]rh#14 Error while inserting in database: %v", err))
		http.Error(rw, fmt.Sprintf("Failed to create available period: %v", err), http.StatusBadRequest)
		return
	}

	log.Info(fmt.Sprintf("[rese-handler]rh#51 User id:'%s' successfuly created available period", userID))

	rw.WriteHeader(http.StatusCreated)
}

func (r *ReservationHandler) CreateReservation(rw http.ResponseWriter, h *http.Request) {
	reservation := h.Context().Value(KeyProduct{}).(*data.ReservationByAvailablePeriod)

	tokenStr := r.extractTokenFromHeader(h)
	username, err := r.getUsername(tokenStr)
	if err != nil {
		log.Warning(fmt.Sprintf("[rese-handler]rh#15 Error while reading username from token: %v", err))
		http.Error(rw, FailedToReadUsernameFromToken, http.StatusBadRequest)
		return
	}

	log.Info(fmt.Sprintf("[rese-handler]rh#63 Received request from '%s' to create reservation", h.RemoteAddr))

	userID, err := r.profile.GetUserId(h.Context(), username, tokenStr)
	if err != nil {
		log.Error(fmt.Sprintf("[rese-handler]rh#16 Error while getting hostId for username: %v", err))
		http.Error(rw, FailedToGetHostIDFromUsername, http.StatusBadRequest)
		return
	}

	reservation.IDUser, err = primitive.ObjectIDFromHex(userID)
	if err != nil {
		log.Error(fmt.Sprintf("[rese-handler]rh#17 Error while getting hostId for accommodation: %v", err))
		http.Error(rw, "Failed to set HostID for accommodation", http.StatusBadRequest)
		return
	}

	exists, err := r.accommodation.CheckAccommodationID(h.Context(), reservation.IDAccommodation, tokenStr)
	if !exists {
		log.Error(fmt.Sprintf("[rese-handler]rh#18 Error while getting accommodation by id: %v", err))
		http.Error(rw, "Accommodation does not exist", http.StatusBadRequest)
		return
	}

	if err != nil {
		log.Error(fmt.Sprintf("[rese-handler]rh#19 Error while getting accommodation by id: %v", err))
		http.Error(rw, "Failed to check accommodation existence", http.StatusInternalServerError)
		return
	}

	err = r.repo.InsertReservationByAvailablePeriod(reservation)
	if err != nil {
		log.Error(fmt.Sprintf("[rese-handler]rh#20 Error while inserting in database: %v", err))
		http.Error(rw, fmt.Sprintf("Failed to create reservation: %v", err), http.StatusBadRequest)
		return
	}

	// Get accommodation
	accommodation, err := r.accommodation.GetAccommodationByID(h.Context(), reservation.IDAccommodation, tokenStr)
	if err != nil {
		log.Error(fmt.Sprintf("[rese-handler]rh#21 Error while finding accommodation by id: %v", err))
		http.Error(rw, "failed to get accommodation by ID", http.StatusInternalServerError)
		return
	}

	// Get host
	host, err := r.profile.GetUserById(h.Context(), accommodation.HostID, tokenStr)
	if err != nil {
		log.Error(fmt.Sprintf("[rese-handler]rh#22 Error while finding host by id: %v", err))
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
		log.Error(fmt.Sprintf("[rese-handler]rh#23 Error while trying to notify host: %v", err))
		http.Error(rw, "failed to notify host", http.StatusInternalServerError)
		return
	}

	log.Info(fmt.Sprintf("[rese-handler]rh#52 User id:'%s' successfuly created reservation", userID))

	rw.WriteHeader(http.StatusCreated)
}

func (r *ReservationHandler) FindAccommodationIdsByDates(rw http.ResponseWriter, h *http.Request) {
	log.Info(fmt.Sprintf("[rese-handler]rh#64 Received request from '%s' for finding accommodation ids by dates", h.RemoteAddr))

	dates := h.Context().Value(KeyProduct{}).(data.Dates)
	ids, err := r.repo.FindAccommodationIdsByDates(&dates)
	if err != nil {
		log.Error(fmt.Sprintf("[rese-handler]rh#24 Error while finding accommodation by id and dates: %v", err))
		http.Error(rw, fmt.Sprintf("Failed to find accommodation ids: %v", err), http.StatusBadRequest)
		return
	}
	err = ids.ToJSON(rw)
	if err != nil {
		log.Error(fmt.Sprintf("[rese-handler]rh#25 Error while trying to convert json: %v", err))
		http.Error(rw, UnableToConvertToJson, http.StatusInternalServerError)
		return
	}

	log.Info(fmt.Sprintf("[rese-handler]rh#53 Successfuly retrieved accommodation ids by dates"))

	rw.WriteHeader(http.StatusOK)
}

func (r *ReservationHandler) FindAllReservationsByUserIDExpired(rw http.ResponseWriter, h *http.Request) {
	tokenStr := r.extractTokenFromHeader(h)
	username, err := r.getUsername(tokenStr)
	if err != nil {
		log.Warning(fmt.Sprintf("[rese-handler]rh#26 Error while reading username from token: %v", err))
		http.Error(rw, FailedToReadUsernameFromToken, http.StatusBadRequest)
		return
	}

	log.Info(fmt.Sprintf("[rese-handler]rh#65 Received request from '%s' for all expired reservations for user '%s'", h.RemoteAddr, username))

	guestID, err := r.profile.GetUserId(h.Context(), username, tokenStr)
	if err != nil {
		log.Error(fmt.Sprintf("[rese-handler]rh#27 Error while getting hostId for username: %v", err))
		http.Error(rw, "Failed to get guestID from username", http.StatusBadRequest)
		return
	}

	reservations, err := r.repo.FindAllReservationsByUserIDExpired(guestID)
	if err != nil {
		log.Error(fmt.Sprintf("[rese-handler]rh#28 Error while finding expired reservations by user: %v", err))
		http.Error(rw, "Failed to fetch expired reservations", http.StatusBadRequest)
		return
	}

	err = reservations.ToJSON(rw)
	if err != nil {
		log.Error(fmt.Sprintf("[rese-handler]rh#29 Error while converting json: %v", err))
		http.Error(rw, "Unable to convert to JSON", http.StatusBadRequest)
		return
	}

	log.Info(fmt.Sprintf("[rese-handler]rh#54 Successfully retrieved expired reservations for user '%s'", username))

	rw.WriteHeader(http.StatusOK)
}

func (r *ReservationHandler) UpdateAvailablePeriodByAccommodation(rw http.ResponseWriter, h *http.Request) {
	availablePeriod := h.Context().Value(KeyProduct{}).(*data.AvailablePeriodByAccommodation)
	tokenStr := r.extractTokenFromHeader(h)
	username, err := r.getUsername(tokenStr)
	if err != nil {
		log.Warning(fmt.Sprintf("[rese-handler]rh#30 Error while reading username from token: %v", err))
		http.Error(rw, FailedToReadUsernameFromToken, http.StatusBadRequest)
		return
	}

	log.Info(fmt.Sprintf("[rese-handler]rh#66 Received request from '%s' to update available period '%s'", h.RemoteAddr, availablePeriod.ID.String()))

	userID, err := r.profile.GetUserId(h.Context(), username, tokenStr)
	if err != nil {
		log.Error(fmt.Sprintf("[rese-handler]rh#31 Error while reading host id from username: %v", err))
		http.Error(rw, FailedToGetHostIDFromUsername, http.StatusBadRequest)
		return
	}

	if availablePeriod.IDUser.Hex() != userID {
		log.Error(fmt.Sprintf("[rese-handler]rh#32 User is not owner of period: %v", err))
		http.Error(rw, "You are not the owner of available period", http.StatusBadRequest)
		return
	}

	err = r.repo.UpdateAvailablePeriodByAccommodation(availablePeriod)
	if err != nil {
		log.Error(fmt.Sprintf("[rese-handler]rh#33 Error while updating period by accommodation: %v", err))
		rw.WriteHeader(http.StatusBadRequest)
		return
	}
	log.Info(fmt.Sprintf("[rese-handler]rh#55 Successfully updated available period by accommodation id '%s'", availablePeriod.ID.String()))
	rw.WriteHeader(http.StatusCreated)
}

func (r *ReservationHandler) DeletePeriodsForAccommodations(rw http.ResponseWriter, h *http.Request) {
	log.Info(fmt.Sprintf("[rese-handler]rh#67 Received request from '%s' to delete periods for accommodations", h.RemoteAddr))

	if h.Context().Value(KeyProduct{}) != nil {
		accIDs := h.Context().Value(KeyProduct{}).([]primitive.ObjectID)
		if (accIDs != nil) && len(accIDs) > 0 {
			err := r.repo.DeletePeriodsForAccommodations(accIDs)
			if err != nil {
				log.Error(fmt.Sprintf("[rese-handler]rh#34 Error while deleting periods by accommodation: %v", err))
				rw.WriteHeader(http.StatusBadRequest)
				return
			}
			for _, accID := range accIDs {
				log.Info(fmt.Sprintf("[rese-handler]rh#72 Deleted all available periods for accommodation '%s'", accID.Hex()))
			}
		}
	}

	log.Info(fmt.Sprintf("[rese-handler]rh#71 Successfully deleted all available periods for requested accommodations"))
	rw.WriteHeader(http.StatusNoContent)
}

func (r *ReservationHandler) GetAllReservationsByUser(rw http.ResponseWriter, h *http.Request) {
	tokenStr := r.extractTokenFromHeader(h)
	vars := mux.Vars(h)
	username := vars["username"]

	log.Info(fmt.Sprintf("[rese-handler]rh#68 Received request from '%s' for all reservations for user '%s'", h.RemoteAddr, username))

	userID, err := r.profile.GetUserId(h.Context(), username, tokenStr)
	if err != nil {
		log.Warning(fmt.Sprintf("[rese-handler]rh#35 Error while reading username from token: %v", err))
		http.Error(rw, "Failed to get UserID from username", http.StatusBadRequest)
		return
	}

	reservations, err := r.repo.FindAllReservationsByUserID(userID)
	if err != nil {
		log.Error(fmt.Sprintf("[rese-handler]rh#36 Error while finding reservations by userId: %v", err))
		rw.WriteHeader(http.StatusBadRequest)
	}

	err = reservations.ToJSON(rw)
	if err != nil {
		log.Error(fmt.Sprintf("[rese-handler]rh#37 Error while converting json: %v", err))
		http.Error(rw, UnableToConvertToJson, http.StatusInternalServerError)
		return
	}
	log.Info(fmt.Sprintf("[rese-handler]rh#56 Successfully retrieved reservations for user '%s'", username))
	rw.WriteHeader(http.StatusOK)
}

func (r *ReservationHandler) CheckAndDeleteReservationsForUser(rw http.ResponseWriter, h *http.Request) {
	vars := mux.Vars(h)
	userID, err := primitive.ObjectIDFromHex(vars["id"])
	if err != nil {
		http.Error(rw, "Invalid ID", http.StatusBadRequest)
		return
	}

	log.Info(fmt.Sprintf("[rese-handler]rh#69 Received request from '%s' to check and delete reservations for user '%s'", h.RemoteAddr, userID.Hex()))

	err = r.repo.CheckAndDeleteReservationsByUserID(userID)
	if err != nil {
		log.Error(fmt.Sprintf("[rese-handler]rh#38 Error while checking and deleting reservations by user id : %v", err))
		rw.WriteHeader(http.StatusBadRequest)
	}
	log.Info(fmt.Sprintf("[rese-handler]rh#57 Successfully deleted reservtions for user id:'%s'", userID))

	rw.WriteHeader(http.StatusNoContent)
}

func (r *ReservationHandler) DeleteReservation(rw http.ResponseWriter, h *http.Request) {
	vars := mux.Vars(h)
	periodID := vars["periodID"]
	reservationID := vars["reservationID"]
	tokenStr := r.extractTokenFromHeader(h)
	username, err := r.getUsername(tokenStr)
	if err != nil {
		log.Warning(fmt.Sprintf("[rese-handler]rh#39 Error while reading username from token: %v", err))
		http.Error(rw, FailedToReadUsernameFromToken, http.StatusBadRequest)
		return
	}

	log.Info(fmt.Sprintf("[rese-handler]rh#70 Received request from '%s' to delete reservation '%s'", h.RemoteAddr, reservationID))

	userID, err := r.profile.GetUserId(h.Context(), username, tokenStr)
	if err != nil {
		log.Error(fmt.Sprintf("[rese-handler]rh#40 Error while reading host id from username: %v", err))
		http.Error(rw, FailedToGetHostIDFromUsername, http.StatusBadRequest)
		return
	}

	// Get reservation
	reservation, err := r.repo.FindReservationByIdAndAvailablePeriod(reservationID, periodID)
	if err != nil {
		log.Error(fmt.Sprintf("[rese-handler]rh#41 Error while finding reservation by id and period: %v", err))
		http.Error(rw, "failed to get reservation by ID", http.StatusInternalServerError)
		return
	}

	err = r.repo.DeleteReservationByIdAndAvailablePeriodID(reservationID, periodID, userID)
	if err != nil {
		log.Error(fmt.Sprintf("[rese-handler]rh#42 Error while reading host id from usenrame: %v", err))
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	// Get period
	period, err := r.repo.FindAvailablePeriodsByAccommodationId(reservation.IDAccommodation.Hex())
	if err != nil {
		log.Error(fmt.Sprintf("[rese-handler]rh#43 Error while finding period by accommodation id: %v", err))
		http.Error(rw, "failed to get period by accommodation ID", http.StatusInternalServerError)
		return
	}

	// Get host
	host, err := r.profile.GetUserById(h.Context(), period[0].IDUser, tokenStr)
	if err != nil {
		log.Error(fmt.Sprintf("[rese-handler]rh#44 Error while finding user by id: %v", err))
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
		log.Error(fmt.Sprintf("[rese-handler]rh#44 Error while trying to notify host: %v", err))
		http.Error(rw, "failed to notify host", http.StatusInternalServerError)
		return
	}
	log.Info(fmt.Sprintf("[rese-handler]rh#58 Successfully deleted reservation '%s'", reservationID))

	rw.WriteHeader(http.StatusAccepted)
}

func (r *ReservationHandler) MiddlewareAvailablePeriodDeserialization(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, h *http.Request) {
		availablePeriod := &data.AvailablePeriodByAccommodation{}
		err := availablePeriod.FromJSON(h.Body)
		if err != nil {
			log.Error(fmt.Sprintf("[rese-handler]rh#45 Error while trying to convert json: %v", err))
			http.Error(rw, UnableToDecodeJson, http.StatusBadRequest)
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
			log.Error(fmt.Sprintf("[rese-handler]rh#46 Error while trying to convert json: %v", err))
			http.Error(rw, UnableToDecodeJson, http.StatusBadRequest)
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
			log.Error(fmt.Sprintf("[rese-handler]rh#47 Error while trying to convert json: %v", err))
			http.Error(rw, UnableToDecodeJson, http.StatusBadRequest)
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
