package main

import (
	"context"
	"log"
	"main.go/handlers"
	"net/http"
	"os"
	"os/signal"
	"time"

	gorillaHandlers "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"main.go/data"
)

func main() {
	//Reading from environment, if not set we will default it to 8080.
	//This allows flexibility in different environments (for eg. when running multiple docker api's and want to override the default port)
	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "8082"
	}

	// Initialize context
	timeoutContext, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	//Initialize the logger we are going to use, with prefix and datetime for every log
	logger := log.New(os.Stdout, "[reservation-api] ", log.LstdFlags)
	storeLogger := log.New(os.Stdout, "[reservation-store] ", log.LstdFlags)

	// NoSQL: Initialize Product Repository store
	store, err := data.New(timeoutContext, storeLogger)
	if err != nil {
		logger.Fatal(err)
	}
	defer store.CloseSession()
	store.CreateTables()

	//Initialize the handler and inject said logger
	reservationHandler := handlers.NewReservationHandler(logger, store)

	//Initialize the router and add a middleware for all the requests
	router := mux.NewRouter()
	router.Use(reservationHandler.MiddlewareContentTypeSet)

	getAvailablePeriodsByAccommodationRouter := router.Methods(http.MethodGet).Subrouter()
	getAvailablePeriodsByAccommodationRouter.HandleFunc("/{id}/periods", reservationHandler.GetAllAvailablePeriodsByAccommodation)

	getReservationsByAvailablePeriodRouter := router.Methods(http.MethodGet).Subrouter()
	getReservationsByAvailablePeriodRouter.HandleFunc("/{id}/reservations", reservationHandler.GetAllReservationByAvailablePeriod)

	postAvailablePeriodsByAccommodationRouter := router.Methods(http.MethodPost).Subrouter()
	postAvailablePeriodsByAccommodationRouter.HandleFunc("/period", reservationHandler.CreateAvailablePeriod)
	postAvailablePeriodsByAccommodationRouter.Use(reservationHandler.MiddlewareAvailablePeriodDeserialization)

	postReservationRouter := router.Methods(http.MethodPost).Subrouter()
	postReservationRouter.HandleFunc("/reservation", reservationHandler.CreateReservation)
	postReservationRouter.Use(reservationHandler.MiddlewareReservationDeserialization)

	updateAvailablePeriodsByAccommodationRouter := router.Methods(http.MethodPatch).Subrouter()
	updateAvailablePeriodsByAccommodationRouter.HandleFunc("/period", reservationHandler.UpdateAvailablePeriodByAccommodation)
	updateAvailablePeriodsByAccommodationRouter.Use(reservationHandler.MiddlewareAvailablePeriodDeserialization)
	//
	//deleteReservationById := router.Methods(http.MethodDelete).Subrouter()
	//deleteReservationById.HandleFunc("/{id}", reservationHandler.DeleteReservation)
	//
	//reservePeriodRouter := router.Methods(http.MethodPatch).Subrouter()
	//reservePeriodRouter.HandleFunc("/{id}", reservationHandler.ReservePeriod)
	//reservePeriodRouter.Use(reservationHandler.MiddlewareReservedPeriodDeserialization)
	//
	//
	////updateReservedPeriodRouter := router.Methods(http.MethodPatch).Subrouter()
	////updateReservedPeriodRouter.HandleFunc("/{reservationId}/update", reservationHandler.UpdateReservedPeriod)
	////updateReservedPeriodRouter.Use(reservationHandler.MiddlewareReservedPeriodDeserialization)
	//
	//deleteReservedPeriod := router.Methods(http.MethodDelete).Subrouter()
	//deleteReservedPeriod.HandleFunc("/{reservationId}/period/{periodId}", reservationHandler.DeleteReservedPeriod)

	cors := gorillaHandlers.CORS(gorillaHandlers.AllowedOrigins([]string{"*"}))

	//Initialize the server
	server := http.Server{
		Addr:         ":" + port,
		Handler:      cors(router),
		IdleTimeout:  120 * time.Second,
		ReadTimeout:  1 * time.Second,
		WriteTimeout: 1 * time.Second,
	}

	logger.Println("Server listening on port", port)
	//Distribute all the connections to goroutines
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

	//Try to shutdown gracefully
	if server.Shutdown(timeoutContext) != nil {
		logger.Fatal("Cannot gracefully shutdown...")
	}
	logger.Println("Server stopped")

}
