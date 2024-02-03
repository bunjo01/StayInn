package handlers

import (
	"accommodation/cache"
	"accommodation/clients"
	"accommodation/data"
	"accommodation/storage"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/dgrijalva/jwt-go"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AccommodationHandler struct {
	repo        *data.AccommodationRepository
	reservation clients.ReservationClient
	profile     clients.ProfileClient
	imageCache  *cache.ImageCache
	images      *storage.FileStorage
}

var secretKey = []byte("stayinn_secret")

func NewAccommodationsHandler(r *data.AccommodationRepository,
	rc clients.ReservationClient, p clients.ProfileClient,
	ic *cache.ImageCache, i *storage.FileStorage) *AccommodationHandler {
	return &AccommodationHandler{r, rc, p, ic, i}
}

func (ah *AccommodationHandler) GetAllAccommodations(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	log.Info(fmt.Sprintf("[acco-handler]ach#1 Received request from '%s' for all accommodations", r.RemoteAddr))

	accommodations, err := ah.repo.GetAllAccommodations(ctx)
	if err != nil {
		http.Error(rw, "Failed to retrieve accommodations", http.StatusInternalServerError)
		log.Error(fmt.Sprintf("[acco-handler]ach#2 Failed to retrieve accommodations: %v", err))
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(rw).Encode(accommodations); err != nil {
		http.Error(rw, "Failed to encode accommodations", http.StatusInternalServerError)
		log.Error(fmt.Sprintf("[acco-handler]ach#3 Failed to encode accommodations: %v", err))
	}
	log.Info(fmt.Sprintf("[acco-handler]ach#4 Successfully fetched all accommodations"))
}

func (ah *AccommodationHandler) GetAccommodation(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := primitive.ObjectIDFromHex(vars["id"])
	if err != nil {
		http.Error(rw, "Invalid ID", http.StatusBadRequest)
		return
	}
	log.Info(fmt.Sprintf("[acco-handler]ach#5 Received request from '%s' for accommodation '%s'", r.RemoteAddr, id.Hex()))

	ctx := r.Context()
	accommodation, err := ah.repo.GetAccommodation(ctx, id)
	if err != nil {
		http.Error(rw, "Failed to retrieve accommodation", http.StatusInternalServerError)
		log.Error(fmt.Sprintf("[acco-handler]ach#6 Failed to retrieve accommodation '%s': %v", id.Hex(), err))
		return
	}

	if accommodation == nil {
		http.NotFound(rw, r)
		log.Info(fmt.Sprintf("[acco-handler]ach#7 Accommodation with id '%s' not found", id.Hex()))
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(rw).Encode(accommodation); err != nil {
		http.Error(rw, "Failed to encode accommodation", http.StatusInternalServerError)
		log.Error(fmt.Sprintf("[acco-handler]ach#8 Failed to encode accommodation '%s': %v", id.Hex(), err))
	}
	log.Info(fmt.Sprintf("[acco-handler]ach#9 Successfully fetched accommodation with id '%s'", id.Hex()))
}

func (ah *AccommodationHandler) CreateAccommodation(rw http.ResponseWriter, r *http.Request) {
	var accommodation data.Accommodation

	log.Info(fmt.Sprintf("[acco-handler]ach#10 User from '%s' creating a new accommodation", r.RemoteAddr))

	if err := json.NewDecoder(r.Body).Decode(&accommodation); err != nil {
		http.Error(rw, "Failed to decode request body", http.StatusBadRequest)
		log.Error(fmt.Sprintf("[acco-handler]ach#11 Failed to decode request body: %v", err))
		return
	}

	tokenStr := ah.extractTokenFromHeader(r)
	username, err := ah.getUsername(tokenStr)
	if err != nil {
		log.Error(fmt.Sprintf("[acco-handler]ach#12 Failed to read username from token: %v", err))
		http.Error(rw, "Failed to read username from token", http.StatusBadRequest)
		return
	}

	hostID, err := ah.profile.GetUserId(r.Context(), username, tokenStr)
	if err != nil {
		log.Error(fmt.Sprintf("[acco-handler]ach#13 Failed to get HostID from username: %v", err))
		http.Error(rw, "Failed to get HostID from username", http.StatusBadRequest)
		return
	}

	accommodation.HostID, err = primitive.ObjectIDFromHex(hostID)
	if err != nil {
		log.Error(fmt.Sprintf("[acco-handler]ach#14 Failed to set HostID for accommodation: %v", err))
		http.Error(rw, "Failed to set HostID for accommodation", http.StatusBadRequest)
		return
	}

	// Adding accommodation
	accommodation.ID = primitive.NewObjectID()
	if err := ah.repo.CreateAccommodation(r.Context(), &accommodation); err != nil {
		log.Error(fmt.Sprintf("[acco-handler]ach#15 Failed to create accommodation: %v", err))
		http.Error(rw, "Failed to create accommodation", http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(rw).Encode(accommodation); err != nil {
		log.Error(fmt.Sprintf("[acco-handler]ach#16 Failed to encode accommodation: %v", err))
		http.Error(rw, "Failed to encode accommodation", http.StatusInternalServerError)
	}

	log.Info(fmt.Sprintf("[acco-handler]ach#17 Successfully created accommodation with id '%s'", accommodation.ID.Hex()))
}

func (ah *AccommodationHandler) CreateAccommodationImages(rw http.ResponseWriter, r *http.Request) {
	var images cache.Images
	var accID string

	log.Info(fmt.Sprintf("[acco-handler]ach#18 Recieved request to create accommodation images from '%s'", r.RemoteAddr))

	if err := json.NewDecoder(r.Body).Decode(&images); err != nil {
		http.Error(rw, "Failed to decode request body", http.StatusBadRequest)
		log.Error(fmt.Sprintf("[acco-handler]ach#19 Failed to decode request body: %v", err))
		return
	}

	for _, image := range images {
		ah.images.WriteFileBytes(image.Data, image.AccID+"-image-"+image.ID)
		accID = image.AccID
	}
	ah.imageCache.PostAll(accID, images)

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusCreated)
	log.Info(fmt.Sprintf("[acco-handler]ach#20 Successfully created images for accommodation '%s'", accID))
}

func (ah *AccommodationHandler) GetAccommodationImages(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	accID := vars["id"]

	log.Info(fmt.Sprintf("[acco-handler]ach#21 Received request for accommodation images from '%s'", r.RemoteAddr))

	var images []*cache.Image

	for i := 0; i < 10; i++ {
		filename := fmt.Sprintf("%s-image-%d", accID, i)
		data, err := ah.images.ReadFileBytes(filename, false)
		if err != nil {
			break
		}
		image := &cache.Image{
			ID:   strconv.Itoa(i),
			Data: data,
		}
		images = append(images, image)
	}

	if len(images) > 0 {
		err := ah.imageCache.PostAll(accID, images)
		if err != nil {
			log.Error(fmt.Sprintf("[acco-handler]ach#22 Unable to write to cache: %v", err))
		}
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(rw).Encode(images); err != nil {
		log.Error(fmt.Sprintf("[acco-handler]ach#23 Failed to encode images: %v", err))
		http.Error(rw, "Failed to encode images", http.StatusInternalServerError)
	}
	log.Info(fmt.Sprintf("[acco-handler]ach#24 Successfully fetched images for accommodation '%s'", accID))
}

func (ah *AccommodationHandler) UpdateAccommodation(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := primitive.ObjectIDFromHex(vars["id"])
	if err != nil {
		http.Error(rw, "Invalid ID", http.StatusBadRequest)
		return
	}

	log.Info(fmt.Sprintf("[acco-handler]ach#25 Received request to update accommodation '%s' from '%s'", id.Hex(), r.RemoteAddr))

	var updatedAccommodation data.Accommodation
	if err := json.NewDecoder(r.Body).Decode(&updatedAccommodation); err != nil {
		http.Error(rw, "Failed to decode request body", http.StatusBadRequest)
		log.Error(fmt.Sprintf("[acco-handler]ach#26 Failed to decode request body: %v", err))
		return
	}

	updatedAccommodation.ID = id
	if err := ah.repo.UpdateAccommodation(r.Context(), &updatedAccommodation); err != nil {
		log.Error(fmt.Sprintf("[acco-handler]ach#27 Failed to update accommodation: %v", err))
		http.Error(rw, "Failed to update accommodation", http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(rw).Encode(updatedAccommodation); err != nil {
		log.Error(fmt.Sprintf("[acco-handler]ach#28 Failed to encode updated accommodation: %v", err))
		http.Error(rw, "Failed to encode updated accommodation", http.StatusInternalServerError)
	}
	log.Info(fmt.Sprintf("[acco-handler]ach#29 Successfully updated accommodation '%s'", updatedAccommodation.ID.Hex()))
}

func (ah *AccommodationHandler) DeleteAccommodation(rw http.ResponseWriter, r *http.Request) {
	tokenStr := ah.extractTokenFromHeader(r)
	vars := mux.Vars(r)
	id, err := primitive.ObjectIDFromHex(vars["id"])
	if err != nil {
		http.Error(rw, "Invalid ID", http.StatusBadRequest)
		return
	}

	log.Info(fmt.Sprintf("[acco-handler]ach#30 Recieved request from '%s' to delete accommodation '%s'", r.RemoteAddr, id.Hex()))

	var accIDs []primitive.ObjectID
	accIDs = append(accIDs, id)

	ctx, cancel := context.WithTimeout(r.Context(), 4000*time.Millisecond)
	defer cancel()
	_, err = ah.reservation.CheckAndDeletePeriods(ctx, accIDs, tokenStr)
	if err != nil {
		log.Error(fmt.Sprintf("[acco-handler]ach#31 Error checking and deleting periods: %v", err))
		writeResp(err, http.StatusServiceUnavailable, rw)
		return
	}

	if err := ah.repo.DeleteAccommodation(r.Context(), id); err != nil {
		log.Error(fmt.Sprintf("[acco-handler]ach#32 Failed to delete accommodation: %v", err))
		http.Error(rw, "Failed to delete accommodation", http.StatusInternalServerError)
		return
	}

	for i := 0; i < 10; i++ {
		filename := fmt.Sprintf("%s-image-%d", id.Hex(), i)
		err := ah.images.DeleteFile(filename, false)
		if err != nil {
			break
		}
		log.Info(fmt.Sprintf("[acco-handler]ach#33 %s-image-%d deleted\n", id.Hex(), i))
	}

	rw.WriteHeader(http.StatusNoContent)
	log.Info(fmt.Sprintf("[acco-handler]ach#34 Successfully deleted accommodation '%s'", id.Hex()))
}

func (ah *AccommodationHandler) GetAccommodationsForUser(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username := vars["username"]

	log.Info(fmt.Sprintf("[acco-handler]ach#35 Recieved request from '%s' for '%s' accommodations", r.RemoteAddr, username))

	tokenStr := ah.extractTokenFromHeader(r)
	userIDStr, err := ah.profile.GetUserId(r.Context(), username, tokenStr)
	if err != nil {
		log.Error(fmt.Sprintf("[acco-handler]ach#36 Failed to get UserID from username '%s': %v", username, err))
		http.Error(rw, "Failed to get UserID from username", http.StatusBadRequest)
		return
	}

	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		log.Error(fmt.Sprintf("[acco-handler]ach#37 Invalid UserID: %v", err))
		http.Error(rw, "Invalid userID", http.StatusBadRequest)
		return
	}

	accommodations, err := ah.repo.GetAccommodationsForUser(r.Context(), userID)
	if err != nil {
		log.Error(fmt.Sprintf("[acco-handler]ach#38 Failed to get accommodations for UserID '%v': %v", userID, err))
		http.Error(rw, "Failed to get accommodations for userID: "+userID.Hex(), http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(rw).Encode(accommodations); err != nil {
		log.Error(fmt.Sprintf("[acco-handler]ach#39 Failed to encode accommodations: %v", err))
		http.Error(rw, "Failed to encode accommodations", http.StatusInternalServerError)
	}
	log.Info(fmt.Sprintf("[acco-handler]ach#40 Successfully fetched accommodations from user '%s'", username))
}

func (ah *AccommodationHandler) DeleteUserAccommodations(rw http.ResponseWriter, r *http.Request) {
	tokenStr := ah.extractTokenFromHeader(r)
	vars := mux.Vars(r)
	userID, err := primitive.ObjectIDFromHex(vars["id"])
	if err != nil {
		http.Error(rw, "Invalid userID", http.StatusBadRequest)
		return
	}

	log.Info(fmt.Sprintf("[acco-handler]ach#41 Recieved request from '%s' to delete user '%v' accommodations", r.RemoteAddr, userID))

	accommodations, err := ah.repo.GetAccommodationsForUser(r.Context(), userID)
	if err != nil {
		log.Error(fmt.Sprintf("[acco-handler]ach#42 Failed to get accommodations for UserID '%v': %v", userID, err))
		http.Error(rw, "Failed to get accommodations for userID: "+userID.Hex(), http.StatusInternalServerError)
		return
	}

	var accIDs []primitive.ObjectID
	for _, accommodation := range accommodations {
		accIDs = append(accIDs, accommodation.ID)
	}

	// 4000 ms because it's second in chain of service calls
	ctx, cancel := context.WithTimeout(r.Context(), 4000*time.Millisecond)
	defer cancel()
	_, err = ah.reservation.CheckAndDeletePeriods(ctx, accIDs, tokenStr)
	if err != nil {
		log.Warning(fmt.Sprintf("[acco-handler]ach#43 Reservation service unavaible: %v", err))
		writeResp(err, http.StatusServiceUnavailable, rw)
		return
	}

	if err := ah.repo.DeleteAccommodationsForUser(r.Context(), userID); err != nil {
		log.Error(fmt.Sprintf("[acco-handler]ach#44 Failed to delete accommodations for UserID '%v': %v", userID, err))
		http.Error(rw, "Failed to delete accommodations for userID: "+userID.Hex(), http.StatusInternalServerError)
		return
	}

	for _, accID := range accIDs {
		for i := 0; i < 10; i++ {
			filename := fmt.Sprintf("%s-image-%d", accID.Hex(), i)
			err := ah.images.DeleteFile(filename, false)
			if err != nil {
				break
			}
			log.Info(fmt.Sprintf("[acco-handler]ach#45 %s-image-%d deleted\n", accID.Hex(), i))
		}
	}

	rw.WriteHeader(http.StatusNoContent)
	log.Info(fmt.Sprintf("[acco-handler]ach#46 Successfully deleted accommodations for user '%v'", userID))
}

func (ah *AccommodationHandler) getUsername(tokenString string) (string, error) {
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

func (ah *AccommodationHandler) AuthorizeRoles(allowedRoles ...string) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, rr *http.Request) {
			tokenString := ah.extractTokenFromHeader(rr)
			if tokenString == "" {
				log.Warning(fmt.Sprintf("[acco-handler]ach#47 No token found in request from '%s'", rr.RemoteAddr))
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			claims := jwt.MapClaims{}
			token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
				return secretKey, nil
			})

			if err != nil || !token.Valid {
				log.Warning(fmt.Sprintf("[acco-handler]ach#48 Invalid signature token found in request from '%s'", rr.RemoteAddr))
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			username, ok1 := claims["username"].(string)
			role, ok2 := claims["role"].(string)
			if !ok1 {
				log.Warning(fmt.Sprintf("[acco-handler]ach#49 Username not found in token in request from '%s'", rr.RemoteAddr))
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			if !ok2 {
				log.Warning(fmt.Sprintf("[acco-handler]ach#50 Role not found in token in request from '%s'", rr.RemoteAddr))
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			for _, allowedRole := range allowedRoles {
				if allowedRole == role {
					next.ServeHTTP(w, rr)
					return
				}
			}

			log.Warning(fmt.Sprintf("[acco-handler]ach#51 User '%s' from '%s' tried to do unauthorized actions", username, rr.RemoteAddr))
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
	tokenStr := ah.extractTokenFromHeader(r)
	ctx, cancel := context.WithTimeout(r.Context(), 5000*time.Millisecond)
	defer cancel()

	log.Info(fmt.Sprintf("[acco-handler]ach#52 Recieved request from '%s' to search accommodations", r.RemoteAddr))

	var accommodationIDs []primitive.ObjectID

	startDateStr := r.URL.Query().Get("startDate")
	endDateStr := r.URL.Query().Get("endDate")

	var startDate time.Time
	if startDateStr != "" {
		startDateTemp, err := time.Parse("2006-01-02T15:04:05Z", startDateStr)
		if err != nil {
			log.Error(fmt.Sprintf("[acco-handler]ach#53 Invalid StartDate format: %v", err))
			http.Error(rw, "Invalid startDate format", http.StatusBadRequest)
			return
		}
		startDate = startDateTemp
	}

	var endDate time.Time
	if endDateStr != "" {
		endDateTemp, err := time.Parse("2006-01-02T15:04:05Z", endDateStr)
		if err != nil {
			log.Error(fmt.Sprintf("[acco-handler]ach#54 Invalid EndDate format: %v", err))
			http.Error(rw, "Invalid endDate format", http.StatusBadRequest)
			return
		}
		endDate = endDateTemp
	}

	location := r.URL.Query().Get("location")
	numberOfGuests := r.URL.Query().Get("numberOfGuests")

	var numGuests int
	var err error
	if numberOfGuests != "" && numberOfGuests != "NaN" {
		numGuests, err = strconv.Atoi(numberOfGuests)
		if err != nil {
			log.Error(fmt.Sprintf("[acco-handler]ach#55 Failed to convert NumberOfGuests: %v", err))
			http.Error(rw, "Failed to convert numberOfGuests", http.StatusInternalServerError)
			return
		}
	}

	filter := make(bson.M)

	if location != "" {
		filter["location"] = location
	}

	if numGuests > 0 {
		filter["$and"] = bson.A{
			bson.M{"minGuests": bson.M{"$lte": numGuests}},
			bson.M{"maxGuests": bson.M{"$gte": numGuests}},
		}
	}

	accommodations, err := ah.repo.GetFilteredAccommodations(ctx, filter)
	if err != nil {
		log.Error(fmt.Sprintf("[acco-handler]ach#56 Failed to fetch filtered accommodations: %v", err))
		http.Error(rw, "Failed to retrieve accommodations", http.StatusInternalServerError)
		return
	}

	if startDateStr == "" && endDateStr != "" {
		log.Error(fmt.Sprintf("[acco-handler]ach#57 Missing start date in date filter"))
		http.Error(rw, "You forgot to select start date", http.StatusBadRequest)
		return
	}

	if endDateStr == "" && startDateStr != "" {
		log.Error(fmt.Sprintf("[acco-handler]ach#58 Missing end date in date filter"))
		http.Error(rw, "You forgot to select end date", http.StatusBadRequest)
		return
	}

	if endDateStr != "" && startDateStr != "" {
		if startDate.Before(time.Now()) {
			log.Error(fmt.Sprintf("[acco-handler]ach#59 Start date not in future"))
			http.Error(rw, "Start date must be in future", http.StatusBadRequest)
			return
		}

		if endDate.Before(time.Now()) {
			log.Error(fmt.Sprintf("[acco-handler]ach#60 End date not in future"))
			http.Error(rw, "End date must be in future", http.StatusBadRequest)
			return
		}

		if startDate.After(endDate) {
			log.Error(fmt.Sprintf("[acco-handler]ach#61 Start date not before end date"))
			http.Error(rw, "Start date must be before end date", http.StatusBadRequest)
			return
		}

		for _, accommodation := range accommodations {
			accommodationIDs = append(accommodationIDs, accommodation.ID)
		}

		ids, err := ah.reservation.PassDatesToReservationService(ctx, accommodationIDs, startDate, endDate, tokenStr)
		if err != nil {
			log.Warning(fmt.Sprintf("[acco-handler]ach#62 Reservation service is unavaible: %v", err))
			writeResp(err, http.StatusServiceUnavailable, rw)
			return
		}

		accommodationForReturn, err := ah.repo.FindAccommodationsByIDs(ctx, ids)
		if err != nil {
			log.Warning(fmt.Sprintf("[acco-handler]ach#63 Reservation service is unavaible: %v", err))
			writeResp(err, http.StatusServiceUnavailable, rw)
			return
		}

		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(rw).Encode(accommodationForReturn); err != nil {
			log.Error(fmt.Sprintf("[acco-handler]ach#64 Failed to encode accommodations: %v", err))
			http.Error(rw, "Failed to encode accommodations", http.StatusInternalServerError)
		}
	} else {
		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(rw).Encode(accommodations); err != nil {
			log.Error(fmt.Sprintf("[acco-handler]ach#65 Failed to encode accommodations: %v", err))
			http.Error(rw, "Failed to encode accommodations", http.StatusInternalServerError)
		}
	}
	log.Info(fmt.Sprintf("[acco-handler]ach#66 Successfully searched accommodations"))
}

func (ah *AccommodationHandler) WalkRoot(rw http.ResponseWriter, r *http.Request) {
	pathsArray := ah.images.WalkDirectories()
	paths := strings.Join(pathsArray, "\n")
	io.WriteString(rw, paths)
}

func (ah *AccommodationHandler) MiddlewareCacheHit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, h *http.Request) {
		vars := mux.Vars(h)
		accID := vars["accID"]
		imageID := vars["imageID"]

		log.Info(fmt.Sprintf("[acco-handler]ach#67 Checking cache for image %s-image-%s", accID, imageID))

		image, err := ah.imageCache.Get(accID, imageID)
		if err != nil {
			next.ServeHTTP(rw, h)
		} else {
			err = image.ToJSON(rw)
			if err != nil {
				http.Error(rw, "Unable to convert image to JSON", http.StatusInternalServerError)
				log.Fatal(fmt.Sprintf("[acco-handler]ach#68 Unable to convert image to JSON: %v", err))
				return
			}
		}
	})
}

func (ah *AccommodationHandler) MiddlewareCacheAllHit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, h *http.Request) {
		vars := mux.Vars(h)
		accID := vars["id"]

		log.Info(fmt.Sprintf("[acco-handler]ach#69 Checking cache for accommodation '%s' images", accID))

		images, err := ah.imageCache.GetAll(accID)
		if err != nil {
			log.Info(fmt.Sprintf("[acco-handler]ach#70 Cache not found: %v", err))
			next.ServeHTTP(rw, h)
		} else {
			err = images.ToJSON(rw)
			if err != nil {
				http.Error(rw, "Unable to convert image to JSON", http.StatusInternalServerError)
				log.Fatal(fmt.Sprintf("[acco-handler]ach#71 Unable to convert image to JSON: %v", err))
				return
			}
		}
	})
}
