package main

import (
	"accommodation/data"
	"accommodation/handlers"
	"context"
	"fmt"
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
	db := os.Getenv("CASS_DB")
	cassandraHost := os.Getenv("CASSANDRA_HOST")
	cassandraPortStr := os.Getenv("CASSANDRA_PORT")
	cassandraPort, err := strconv.Atoi(cassandraPortStr)
	if err != nil {
		logger.Fatalf("Failed to parse CASSANDRA_PORT: %v", err)
	}
	cassandraUser := os.Getenv("CASSANDRA_USER")
	cassandraPassword := os.Getenv("CASSANDRA_PASSWORD")

	// Initializing Cassandra session
	cluster := gocql.NewCluster(db)
	cluster.Keyspace = "system"

	session, err := cluster.CreateSession()
	if err != nil {
		log.Fatalf("failed to create session: %v", err)
	}

	err = session.Query(
		fmt.Sprintf(`CREATE KEYSPACE IF NOT EXISTS %s
					WITH replication = {
						'class' : 'SimpleStrategy',
						'replication_factor' : %d
					}`, "accommodation", 1)).Exec()
	if err != nil {
		session.Close()
		log.Fatalf("failed to create keyspace: %v", err)
	}

	// Close the session after keyspace creation
	session.Close()

	cluster = gocql.NewCluster(cassandraHost)
	cluster.Keyspace = "accommodation"
	cluster.Port = cassandraPort
	cluster.Authenticator = gocql.PasswordAuthenticator{
		Username: cassandraUser,
		Password: cassandraPassword,
	}
	cluster.Consistency = gocql.One
	session, err = cluster.CreateSession()
	if err != nil {
		logger.Fatalf("Failed to create Cassandra session: %v", err)
	} else {
		defer session.Close()
		logger.Println("Connected to Cassandra")
	}

	// Initializing repo for accommodations
	store, err := data.NewAccommodationRepository(storeLogger, session)
	if err != nil {
		logger.Fatal(err)
	}
	err = store.CreateAccommodationTable()
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
