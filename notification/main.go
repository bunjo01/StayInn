package main

import (
	"context"
	"log"
	"net/http"
	"notification/clients"
	"notification/data"
	"notification/domain"
	"notification/handlers"
	"os"
	"os/signal"
	"time"

	gorillaHandlers "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/sony/gobreaker"
)

func main() {
	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "8084"
	}

	// Context initialization
	timeoutContext, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Logger initialization
	logger := log.New(os.Stdout, "[notification-service] ", log.LstdFlags)
	storeLogger := log.New(os.Stdout, "[notification-store] ", log.LstdFlags)

	// Initializing repo for notifications
	store, err := data.New(timeoutContext, storeLogger)
	if err != nil {
		logger.Fatal(err)
	}
	defer store.Disconnect(timeoutContext)
	store.Ping()

	reservationClient := &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:        10,
			MaxIdleConnsPerHost: 10,
			MaxConnsPerHost:     10,
		},
	}

	profileClient := &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:        10,
			MaxIdleConnsPerHost: 10,
			MaxConnsPerHost:     10,
		},
	}

	reservationBreaker := gobreaker.NewCircuitBreaker(
		gobreaker.Settings{
			Name:        "reservation",
			MaxRequests: 1,
			Timeout:     10 * time.Second,
			Interval:    0,
			ReadyToTrip: func(counts gobreaker.Counts) bool {
				return counts.ConsecutiveFailures > 2
			},
			OnStateChange: func(name string, from, to gobreaker.State) {
				logger.Printf("CB '%s' changed from '%s' to '%s'\n", name, from, to)
			},
			IsSuccessful: func(err error) bool {
				if err == nil {
					return true
				}
				errResp, ok := err.(domain.ErrResp)
				return ok && errResp.StatusCode >= 400 && errResp.StatusCode < 500
			},
		},
	)

	profileBreaker := gobreaker.NewCircuitBreaker(
		gobreaker.Settings{
			Name:        "profile",
			MaxRequests: 1,
			Timeout:     10 * time.Second,
			Interval:    0,
			ReadyToTrip: func(counts gobreaker.Counts) bool {
				return counts.ConsecutiveFailures > 2
			},
			OnStateChange: func(name string, from, to gobreaker.State) {
				logger.Printf("CB '%s' changed from '%s' to '%s'\n", name, from, to)
			},
			IsSuccessful: func(err error) bool {
				if err == nil {
					return true
				}
				errResp, ok := err.(domain.ErrResp)
				return ok && errResp.StatusCode >= 400 && errResp.StatusCode < 500
			},
		},
	)

	reservation := clients.NewReservationClient(reservationClient, os.Getenv("RESERVATION_SERVICE_URI"), reservationBreaker)
	profile := clients.NewProfileClient(profileClient, os.Getenv("PROFILE_SERVICE_URI"), profileBreaker)

	notificationsHandler := handlers.NewNotificationsHandler(logger, store, reservation, profile)

	// Router init
	router := mux.NewRouter()
	router.Use(notificationsHandler.MiddlewareContentTypeSet)

	// Router methods

	// Get all accommodation ratings
	getAllAccommodationRatings := router.Methods(http.MethodGet).Path("/ratings/accommodation").Subrouter()
	getAllAccommodationRatings.HandleFunc("", notificationsHandler.GetAllAccommodationRatings)

	// Get all host ratings
	getAllHostRatings := router.Methods(http.MethodGet).Path("/ratings/host").Subrouter()
	getAllHostRatings.HandleFunc("", notificationsHandler.GetAllHostRatings)

	// Create and update accommodation rating
	createRatingForAccommodation := router.Methods(http.MethodPost).Path("/rating/accommodation").Subrouter()
	createRatingForAccommodation.HandleFunc("", notificationsHandler.AddRating)

	// Create and update host rating
	createRatingForHost := router.Methods(http.MethodPost).Path("/rating/host").Subrouter()
	createRatingForHost.HandleFunc("", notificationsHandler.AddHostRating)

	// Get all ratings for host's accommodations
	findAllAccommodationRatingsByHost := router.Methods(http.MethodGet).Path("/ratings/accommodation/byHost").Subrouter()
	findAllAccommodationRatingsByHost.HandleFunc("", notificationsHandler.GetAllAccommodationRatingsForLoggedHost)

	// Get all logged in host's ratings
	getAllHostRatingsByUser := router.Methods(http.MethodGet).Path("/ratings/hostByGuest").Subrouter()
	getAllHostRatingsByUser.HandleFunc("", notificationsHandler.GetAllHostRatingsByUser)

	// Get accommodation average rating
	getAverageAccommodationRating := router.Methods(http.MethodGet).Path("/ratings/average/{accommodationID}").Subrouter()
	getAverageAccommodationRating.HandleFunc("", notificationsHandler.GetAverageAccommodationRating)

	// Get host average rating
	getAverageHostRating := router.Methods(http.MethodPost).Path("/ratings/average/host").Subrouter()
	getAverageHostRating.HandleFunc("", notificationsHandler.GetAverageHostRating)

	// Get accommodation rating
	findRatingForAccommodationByGuest := router.Methods(http.MethodGet).Path("/rating/accommodation/{idAccommodation}/byGuest").Subrouter()
	findRatingForAccommodationByGuest.HandleFunc("", notificationsHandler.FindAccommodationRatingByGuest)

	// Get host rating
	findRatingForHostByGuest := router.Methods(http.MethodPost).Path("/rating/host/byGuest").Subrouter()
	findRatingForHostByGuest.HandleFunc("", notificationsHandler.FindHostRatingByGuest)

	// Delete host rating
	deleteRatingForHost := router.Methods(http.MethodDelete).Path("/rating/host/{id}").Subrouter()
	deleteRatingForHost.HandleFunc("", notificationsHandler.DeleteHostRating)

	// Delete accommodation rating
	deleteRatingForAccommodation := router.Methods(http.MethodDelete).Path("/rating/accommodation/{id}").Subrouter()
	deleteRatingForAccommodation.HandleFunc("", notificationsHandler.DeleteRatingAccommodationHandler)

	getAccommodationRatings := router.Methods(http.MethodGet).Path("/ratings/accommodation/{idAccommodation}").Subrouter()
	getAccommodationRatings.HandleFunc("", notificationsHandler.GetAccommodationRatings)

	getHostRatings := router.Methods(http.MethodPost).Path("/ratings/host/host-ratings").Subrouter()
	getHostRatings.HandleFunc("", notificationsHandler.GetRatingsHost)

	getAllAccommodationRatingsByLoggUser := router.Methods(http.MethodGet).Path("/ratings/accommodationByUser").Subrouter()
	getAllAccommodationRatingsByLoggUser.HandleFunc("", notificationsHandler.GetAllAccommodationRatingsByUser)

	getAllHostRatingsByLoggUser := router.Methods(http.MethodGet).Path("/ratings/hostByUser").Subrouter()
	getAllHostRatingsByLoggUser.HandleFunc("", notificationsHandler.GetAllAccommodationRatingsByUser)

	// Notify on reservation
	notifyForReservation := router.Methods(http.MethodPost).Subrouter()
	notifyForReservation.HandleFunc("/reservation", notificationsHandler.NotifyForReservation)

	// Get all notifications
	getAllNotifications := router.Methods(http.MethodGet).Path("/{username}").Subrouter()
	getAllNotifications.HandleFunc("", notificationsHandler.GetAllNotifications)

	// CORS middleware
	cors := gorillaHandlers.CORS(
		gorillaHandlers.AllowedOrigins([]string{"*"}),
		gorillaHandlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "DELETE", "OPTIONS"}),
		gorillaHandlers.AllowedHeaders([]string{"Content-Type"}),
	)

	server := http.Server{
		Addr:         ":" + port,
		Handler:      cors(router),
		IdleTimeout:  120 * time.Second,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	logger.Println("Server listening on port", port)

	go func() {
		err := server.ListenAndServe()
		if err != nil {
			logger.Fatal(err)
		}
	}()

	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, os.Interrupt)
	signal.Notify(sigCh, os.Kill)

	sig := <-sigCh
	logger.Println("Received terminate, graceful shutdown", sig)

	if err := server.Shutdown(timeoutContext); err != nil {
		logger.Fatal("Cannot gracefully shutdown...")
	}
	logger.Println("Server stopped")
}
