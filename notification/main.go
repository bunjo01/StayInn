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
	profileServiceURI := os.Getenv("PROFILE_SERVICE_URI")
	profile := clients.NewProfileClient(profileClient, profileServiceURI, profileBreaker)

	// Uncomment after adding router methods
	notificationsHandler := handlers.NewNotificationsHandler(logger, store, reservation, profile)

	// Router init
	router := mux.NewRouter()
	router.Use(notificationsHandler.MiddlewareContentTypeSet)

	// Router methods
	createRatingForAccommodation := router.Methods(http.MethodPost).Path("/rating/accommodation").Subrouter()
	createRatingForAccommodation.HandleFunc("", notificationsHandler.AddRating)

	createRatingForHost := router.Methods(http.MethodPost).Path("/rating/host").Subrouter()
	createRatingForHost.HandleFunc("", notificationsHandler.AddHostRating)

	getAllAccommodationRatings := router.Methods(http.MethodGet).Path("/ratings/accommodation").Subrouter()
	getAllAccommodationRatings.HandleFunc("", notificationsHandler.GetAllAccommodationRatings)

	getAllAccommodationRatingsByUser := router.Methods(http.MethodGet).Path("/ratings/accommodationByUser").Subrouter()
	getAllAccommodationRatingsByUser.HandleFunc("", notificationsHandler.GetAllAccommodationRatingsByUser)

	getAllHostRatings := router.Methods(http.MethodGet).Path("/ratings/host").Subrouter()
	getAllHostRatings.HandleFunc("", notificationsHandler.GetAllHostRatings)

	getAllHostRatingsByUser := router.Methods(http.MethodGet).Path("/ratings/hostByUser").Subrouter()
	getAllHostRatingsByUser.HandleFunc("", notificationsHandler.GetAllHostRatingsByUser)

	getHostRatings := router.Methods(http.MethodGet).Path("/ratings/host/{hostUsername}").Subrouter()
	getHostRatings.HandleFunc("", notificationsHandler.GetHostRatings)

	getAverageAccommodationRating := router.Methods(http.MethodGet).Path("/ratings/average/{accommodationID}").Subrouter()
	getAverageAccommodationRating.HandleFunc("", notificationsHandler.GetAverageAccommodationRating)

	getAverageHostRating := router.Methods(http.MethodGet).Path("/ratings/average/{username}").Subrouter()
	getAverageHostRating.HandleFunc("", notificationsHandler.GetAverageHostRating)

	findRatingForAccommodation := router.Methods(http.MethodGet).Path("/rating/accommodation/{id}").Subrouter()
	findRatingForAccommodation.HandleFunc("", notificationsHandler.FindRatingById)

	findRatingForHost := router.Methods(http.MethodGet).Path("/rating/host/{id}").Subrouter()
	findRatingForHost.HandleFunc("", notificationsHandler.FindHostRatingById)

	findRatingForAccommodationByGuest := router.Methods(http.MethodGet).Path("/rating/accommodation/{idAccommodation}/byGuest").Subrouter()
	findRatingForAccommodationByGuest.HandleFunc("", notificationsHandler.FindAccommodationRatingByGuest)

	findRatingForHostByGuest := router.Methods(http.MethodPost).Path("/rating/host/byGuest").Subrouter()
	findRatingForHostByGuest.HandleFunc("", notificationsHandler.FindHostRatingByGuest)

	updateRatingForHost := router.Methods(http.MethodPut).Path("/rating/host/{id}").Subrouter()
	updateRatingForHost.HandleFunc("", notificationsHandler.UpdateHostRating)

	updateRatingForAccommodation := router.Methods(http.MethodPut).Path("/rating/accommodation/{id}").Subrouter()
	updateRatingForAccommodation.HandleFunc("", notificationsHandler.UpdateAccommodationRating)

	deleteRatingForHost := router.Methods(http.MethodDelete).Path("/rating/host/{id}").Subrouter()
	deleteRatingForHost.HandleFunc("", notificationsHandler.DeleteHostRating)

	deleteRatingForAccommodation := router.Methods(http.MethodDelete).Path("/rating/accommodation/{id}").Subrouter()
	deleteRatingForAccommodation.HandleFunc("", notificationsHandler.DeleteRatingAccommodationHandler)

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
