package main

import (
	"accommodation/data"
	"accommodation/handlers"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	gorillaHandlers "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
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

	accommodationsHandler := handlers.NewAccommodationsHandler(logger, store)

	// Router init
	router := mux.NewRouter()

	createAccommodationRouter := router.Methods(http.MethodPost).Path("/accommodation").Subrouter()
	createAccommodationRouter.HandleFunc("", accommodationsHandler.CreateAccommodation)
	createAccommodationRouter.Use(accommodationsHandler.AuthorizeRoles("HOST"))

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
