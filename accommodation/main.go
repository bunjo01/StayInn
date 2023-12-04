package main

import (
	"accommodation/clients"
	"accommodation/data"
	"accommodation/domain"
	"accommodation/handlers"
	"context"
	"log"
	"net/http"
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
		port = "8080"
	}

	// Context initialization
	timeoutContext, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Logger initialization
	logger := log.New(os.Stdout, "[accommodation-service] ", log.LstdFlags)
	storeLogger := log.New(os.Stdout, "[accommodation-store] ", log.LstdFlags)

	// Initializing repo for accommodations
	store, err := data.NewAccommodationRepository(timeoutContext, storeLogger)
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

	// TODO: Change second param accordingly after implementing method in client or set it in client method
	reservation := clients.NewReservationClient(reservationClient, os.Getenv("RESERVATION_SERVICE_URI"), reservationBreaker)

	accommodationsHandler := handlers.NewAccommodationsHandler(logger, store, reservation)

	// Router init
	router := mux.NewRouter()

	createAccommodationRouter := router.Methods(http.MethodPost).Path("/accommodation").Subrouter()
	createAccommodationRouter.HandleFunc("", accommodationsHandler.CreateAccommodation)
	// createAccommodationRouter.Use(accommodationsHandler.AuthorizeRoles("HOST"))

	getAllAccommodationRouter := router.Methods(http.MethodGet).Path("/accommodation").Subrouter()
	getAllAccommodationRouter.HandleFunc("", accommodationsHandler.GetAllAccommodations)

	getAccommodationRouter := router.Methods(http.MethodGet).Path("/accommodation/{id}").Subrouter()
	getAccommodationRouter.HandleFunc("", accommodationsHandler.GetAccommodation)

	updateAccommodationRouter := router.Methods(http.MethodPut).Path("/accommodation/{id}").Subrouter()
	updateAccommodationRouter.HandleFunc("", accommodationsHandler.UpdateAccommodation)
	updateAccommodationRouter.Use(accommodationsHandler.AuthorizeRoles("HOST"))

	deleteAccommodationRouter := router.Methods(http.MethodDelete).Path("/accommodation/{id}").Subrouter()
	deleteAccommodationRouter.HandleFunc("", accommodationsHandler.DeleteAccommodation)
	deleteAccommodationRouter.Use(accommodationsHandler.AuthorizeRoles("HOST"))

	// Search part

	searchAccommodationRouter := router.Methods(http.MethodGet).Path("/search").Subrouter()
	searchAccommodationRouter.HandleFunc("", accommodationsHandler.SearchAccommodations)

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
