package main

import (
	"context"
	"github.com/gorilla/mux"
	"log"
	"main.go/data"
	"main.go/handlers"
	"net/http"
	"os"
	"os/signal"
	"time"

	gorillaHandlers "github.com/gorilla/handlers"
)

func main() {
	//Reading from environment, if not set we will default it to 8080.
	//This allows flexibility in different environments (for eg. when running multiple docker api's and want to override the default port)
	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "8080"
	}

	// Initialize context
	timeoutContext, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	//Initialize the logger we are going to use, with prefix and datetime for every log
	logger := log.New(os.Stdout, "[product-api] ", log.LstdFlags)
	storeLogger := log.New(os.Stdout, "[patient-store] ", log.LstdFlags)

	// NoSQL: Initialize Product Repository store
	store, err := data.New(timeoutContext, storeLogger)
	if err != nil {
		logger.Fatal(err)
	}
	defer store.Disconnect(timeoutContext)

	// NoSQL: Checking if the connection was established
	store.Ping()

	//Initialize the handler and inject said logger
	reservationHandler := handlers.NewReservationHandler(logger, store)

	//Initialize the router and add a middleware for all the requests
	router := mux.NewRouter()
	router.Use(reservationHandler.MiddlewareContentTypeSet)

	//TODO : middleWare

	getRouter := router.Methods(http.MethodGet).Subrouter()
	getRouter.HandleFunc("/", reservationHandler.GetAllReservations)

	postRouter := router.Methods(http.MethodPost).Subrouter()
	postRouter.HandleFunc("/", reservationHandler.PostReservation)
	postRouter.Use(reservationHandler.MiddlewareReservationDeserialization)

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
