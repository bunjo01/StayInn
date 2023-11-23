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
	logger := log.New(os.Stdout, "[accommodation-api] ", log.LstdFlags)
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

	router.HandleFunc("/accommodation", accommodationsHandler.CreateAccommodation).Methods("POST")
	router.HandleFunc("/accommodation", accommodationsHandler.GetAllAccommodations).Methods("GET")
	router.HandleFunc("/accommodation/{id}", accommodationsHandler.GetAccommodation).Methods("GET")
	router.HandleFunc("/accommodation/{id}", accommodationsHandler.UpdateAccommodation).Methods("PUT")
	router.HandleFunc("/accommodation/{id}", accommodationsHandler.DeleteAccommodation).Methods("DELETE")

	// CORS middleware
	cors := gorillaHandlers.CORS(
		gorillaHandlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "DELETE", "OPTIONS"}),
		gorillaHandlers.AllowedHeaders([]string{"Content-Type"}),
	)

	server := http.Server{
		Addr:         ":" + port,
		Handler:      cors(router),
		IdleTimeout:  120 * time.Second,
		ReadTimeout:  1 * time.Second,
		WriteTimeout: 1 * time.Second,
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
