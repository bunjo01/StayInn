package main

import (
	"accommodation/data"
	"accommodation/handlers"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	gocql "github.com/gocql/gocql"
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

	// Reading enviroment for Cassandra
	cassandraHost := os.Getenv("CASSANDRA_HOST")
	cassandraPortStr := os.Getenv("CASSANDRA_PORT")
	cassandraPort, err := strconv.Atoi(cassandraPortStr)
	if err != nil {
		logger.Fatalf("Failed to parse CASSANDRA_PORT: %v", err)
	}
	cassandraUser := os.Getenv("CASSANDRA_USER")
	cassandraPassword := os.Getenv("CASSANDRA_PASSWORD")

	// Initializing Cassandra session
	cluster := gocql.NewCluster(cassandraHost)
	cluster.Keyspace = "accommodation"
	cluster.Port = cassandraPort
	cluster.Authenticator = gocql.PasswordAuthenticator{
		Username: cassandraUser,
		Password: cassandraPassword,
	}

	// Set consistency level if needed
	cluster.Consistency = gocql.One

	session, err := cluster.CreateSession()
	if err != nil {
		logger.Fatalf("Failed to create Cassandra session: %v", err)
	} else {
		defer session.Close()
		logger.Println("Connected to Cassandra")
	}

	// Initializing repo for accommodations
	store, err := data.NewAccommodationRepository(storeLogger)
	if err != nil {
		logger.Fatal(err)
	}
	defer store.CloseSession()

	accommodationsHandler := handlers.NewAccommodationsHandler(logger, store)

	// Router init
	router := mux.NewRouter()

	router.HandleFunc("/accommodation", accommodationsHandler.CreateAccommodation).Methods("POST")
	router.HandleFunc("/accommodation", accommodationsHandler.GetAllAccommodations).Methods("GET")
	router.HandleFunc("/accommodation/{id}", accommodationsHandler.GetAccommodation).Methods("GET")
	router.HandleFunc("/accommodation/{id}", accommodationsHandler.UpdateAccommodation).Methods("PUT")
	router.HandleFunc("/accommodation/{id}", accommodationsHandler.DeleteAccommodation).Methods("DELETE")

	cors := gorillaHandlers.CORS(gorillaHandlers.AllowedOrigins([]string{"*"}))

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
