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

	// vrati sve
	getAllAccommodationRatings := router.Methods(http.MethodGet).Path("/ratings/accommodation").Subrouter()
	getAllAccommodationRatings.HandleFunc("", notificationsHandler.GetAllAccommodationRatings)

	// vrati sve
	getAllHostRatings := router.Methods(http.MethodGet).Path("/ratings/host").Subrouter()
	getAllHostRatings.HandleFunc("", notificationsHandler.GetAllHostRatings)

	//oceni smestaj, promena ocene
	createRatingForAccommodation := router.Methods(http.MethodPost).Path("/rating/accommodation").Subrouter()
	createRatingForAccommodation.HandleFunc("", notificationsHandler.AddRating)

	//oceni hosta, promena ocene
	createRatingForHost := router.Methods(http.MethodPost).Path("/rating/host").Subrouter()
	createRatingForHost.HandleFunc("", notificationsHandler.AddHostRating)

	//pronadji sve ocene za accommodacije jednog hosta
	findAllAccommodationRatingsByHost := router.Methods(http.MethodGet).Path("/ratings/accommodation/byHost").Subrouter()
	findAllAccommodationRatingsByHost.HandleFunc("", notificationsHandler.GetAllAccommodationRatingsForLoggedHost)

	//vrati ocene od ulogovanog hosta
	getAllHostRatingsByUser := router.Methods(http.MethodGet).Path("/ratings/hostByGuest").Subrouter()
	getAllHostRatingsByUser.HandleFunc("", notificationsHandler.GetAllHostRatingsByUser)

	//vrati srednju ocenu za acc
	getAverageAccommodationRating := router.Methods(http.MethodGet).Path("/ratings/average/{accommodationID}").Subrouter()
	getAverageAccommodationRating.HandleFunc("", notificationsHandler.GetAverageAccommodationRating)

	//vrati srednju ocenu za hosta
	getAverageHostRating := router.Methods(http.MethodPost).Path("/ratings/average/host").Subrouter()
	getAverageHostRating.HandleFunc("", notificationsHandler.GetAverageHostRating)

	//vrati ocenu za tu acc
	findRatingForAccommodationByGuest := router.Methods(http.MethodGet).Path("/rating/accommodation/{idAccommodation}/byGuest").Subrouter()
	findRatingForAccommodationByGuest.HandleFunc("", notificationsHandler.FindAccommodationRatingByGuest)

	//vrati ocenu za tog hosta
	findRatingForHostByGuest := router.Methods(http.MethodPost).Path("/rating/host/byGuest").Subrouter()
	findRatingForHostByGuest.HandleFunc("", notificationsHandler.FindHostRatingByGuest)

	// obrisi rating za hosta
	deleteRatingForHost := router.Methods(http.MethodDelete).Path("/rating/host/{id}").Subrouter()
	deleteRatingForHost.HandleFunc("", notificationsHandler.DeleteHostRating)

	// obrisi rating za acc
	deleteRatingForAccommodation := router.Methods(http.MethodDelete).Path("/rating/accommodation/{id}").Subrouter()
	deleteRatingForAccommodation.HandleFunc("", notificationsHandler.DeleteRatingAccommodationHandler)

	getAccommodationRatings := router.Methods(http.MethodGet).Path("/ratings/accommodation/{idAccommodation}").Subrouter()
	getAccommodationRatings.HandleFunc("", notificationsHandler.GetAccommodationRatings)

	//
	//getAllAccommodationRatingsByUser := router.Methods(http.MethodGet).Path("/ratings/accommodationByUser").Subrouter()
	//getAllAccommodationRatingsByUser.HandleFunc("", notificationsHandler.GetAllAccommodationRatingsByUser)

	//getHostRatings := router.Methods(http.MethodGet).Path("/ratings/host/{hostUsername}").Subrouter()
	//getHostRatings.HandleFunc("", notificationsHandler.GetHostRatings)

	//findRatingForHost := router.Methods(http.MethodGet).Path("/rating/host/{id}").Subrouter()
	//findRatingForHost.HandleFunc("", notificationsHandler.FindHostRatingById)

	//updateRatingForHost := router.Methods(http.MethodPut).Path("/rating/host/{id}").Subrouter()
	//updateRatingForHost.HandleFunc("", notificationsHandler.UpdateHostRating)
	//
	//updateRatingForAccommodation := router.Methods(http.MethodPut).Path("/rating/accommodation/{id}").Subrouter()
	//updateRatingForAccommodation.HandleFunc("", notificationsHandler.UpdateAccommodationRating)

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
