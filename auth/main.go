package main

import (
	"auth/data"
	"auth/handlers"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	gorillaHandlers "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func seedData() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	store, err := data.New(ctx, log.New(os.Stdout, "[data] ", log.LstdFlags))
	if err != nil {
		log.Fatal(err)
	}
	defer store.Disconnect(ctx)

	// Test data
	testCredentials := data.Credentials{
		Username: "testUser",
		Password: "testPassword",
	}

	testCredentials2 := data.Credentials{
		Username: "admin",
		Password: "admin",
	}

	if err := store.AddCredentials(testCredentials.Username, testCredentials.Password); err != nil {
		log.Fatal(err)
	}

	if err := store.AddCredentials(testCredentials2.Username, testCredentials2.Password); err != nil {
		log.Fatal(err)
	}
}

func main() {
	seedData()
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
	credentialsHandler := handlers.NewCredentialsHandler(logger, store)

	//Initialize the router and add a middleware for all the requests
	router := mux.NewRouter()

	// TODO Router

	router.HandleFunc("/login", credentialsHandler.Login).Methods("POST")
	router.HandleFunc("/register", credentialsHandler.Register).Methods("POST")

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
