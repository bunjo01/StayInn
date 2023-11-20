package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	"main.go/handlers"

	"github.com/gocql/gocql"
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

	// Initializing Cassandra DB
	db := os.Getenv("CASS_DB")
	cassandraHost := os.Getenv("CASSANDRA_HOST")
	cassandraPortStr := os.Getenv("CASSANDRA_PORT")
	cassandraPort, err := strconv.Atoi(cassandraPortStr)
	if err != nil {
		logger.Fatalf("failed to parse CASSANDRA_PORT: %v", err)
	}
	cassandraUser := os.Getenv("CASSANDRA_USER")
	cassandraPassword := os.Getenv("CASSANDRA_PASSWORD")

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
						}`, "reservation", 1)).Exec()

	if err != nil {
		session.Close()
		log.Fatalf("failed to create keyspace: %v", err)
	}

	session.Close()

	cluster = gocql.NewCluster(cassandraHost)
	cluster.Keyspace = "reservation"
	cluster.Port = cassandraPort
	cluster.Authenticator = gocql.PasswordAuthenticator{
		Username: cassandraUser,
		Password: cassandraPassword,
	}
	cluster.Consistency = gocql.One

	session, err = cluster.CreateSession()
	if err != nil {
		logger.Fatalf("failed to create Cassandra session: %v", err)
	} else {
		defer session.Close()
		logger.Println("connected to Cassandra")
	}

	// Initializing repo
	store, err := data.New(storeLogger, session)
	if err != nil {
		logger.Fatal(err)
	}

	err = store.CreateTables()
	if err != nil {
		logger.Fatal(err)
	}

	defer store.CloseSession()

	//Initialize the handler and inject said logger
	reservationHandler := handlers.NewReservationHandler(logger, store)

	//Initialize the router and add a middleware for all the requests
	router := mux.NewRouter()
	router.Use(reservationHandler.MiddlewareContentTypeSet)

	getAvailablePeriodsByAccommodationRouter := router.Methods(http.MethodGet).Subrouter()
	getAvailablePeriodsByAccommodationRouter.HandleFunc("/{id}/periods", reservationHandler.GetAllAvailablePeriodsByAccommodation)

	getReservationsByAvailablePeriodRouter := router.Methods(http.MethodGet).Subrouter()
	getReservationsByAvailablePeriodRouter.HandleFunc("/{id}/reservations", reservationHandler.GetAllReservationByAvailablePeriod)

	findAvailablePeriodByIdAndByAccommodationId := router.Methods(http.MethodGet).Subrouter()
	findAvailablePeriodByIdAndByAccommodationId.HandleFunc("/{accomodationID}/{periodID}", reservationHandler.FindAvailablePeriodByIdAndByAccommodationId)

	postAvailablePeriodsByAccommodationRouter := router.Methods(http.MethodPost).Subrouter()
	postAvailablePeriodsByAccommodationRouter.HandleFunc("/period", reservationHandler.CreateAvailablePeriod)
	postAvailablePeriodsByAccommodationRouter.Use(reservationHandler.MiddlewareAvailablePeriodDeserialization)

	postReservationRouter := router.Methods(http.MethodPost).Subrouter()
	postReservationRouter.HandleFunc("/reservation", reservationHandler.CreateReservation)
	postReservationRouter.Use(reservationHandler.MiddlewareReservationDeserialization)

	updateAvailablePeriodsByAccommodationRouter := router.Methods(http.MethodPatch).Subrouter()
	updateAvailablePeriodsByAccommodationRouter.HandleFunc("/period", reservationHandler.UpdateAvailablePeriodByAccommodation)
	updateAvailablePeriodsByAccommodationRouter.Use(reservationHandler.MiddlewareAvailablePeriodDeserialization)

	deleteReservation := router.Methods(http.MethodDelete).Subrouter()
	deleteReservation.HandleFunc("/{periodID}/{reservationID}", reservationHandler.DeleteReservation)

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
