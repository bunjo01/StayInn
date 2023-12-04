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

	"reservation/clients"
	"reservation/data"
	"reservation/domain"
	"reservation/handlers"

	gorillaHandlers "github.com/gorilla/handlers"
	"github.com/sony/gobreaker"

	"github.com/gocql/gocql"
	"github.com/gorilla/mux"
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
	logger := log.New(os.Stdout, "[reservation-service] ", log.LstdFlags)
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

	notificationClient := &http.Client{
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

	accommodationClient := &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:        10,
			MaxIdleConnsPerHost: 10,
			MaxConnsPerHost:     10,
		},
	}

	notificationBreaker := gobreaker.NewCircuitBreaker(
		gobreaker.Settings{
			Name:        "notification",
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

	accommodationBreaker := gobreaker.NewCircuitBreaker(
		gobreaker.Settings{
			Name:        "accommodation",
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

	// TODO: Change second param accordingly after implementing methods on notification service
	notification := clients.NewNotificationClient(notificationClient, os.Getenv("NOTIFICATION_SERVICE_URI"), notificationBreaker)
	// TODO: Change second param in methods when sending request
	profile := clients.NewProfileClient(profileClient, os.Getenv("PROFILE_SERVICE_URI")+"/users", profileBreaker)
	// TODO: Change second param or add it in method
	accommodation := clients.NewAccommodationClient(accommodationClient, os.Getenv("ACCOMMODATION_SERVICE_URI"), accommodationBreaker)

	//Initialize the handler and inject said logger
	reservationHandler := handlers.NewReservationHandler(logger, store, notification, profile, accommodation)

	//Initialize the router and add a middleware for all the requests
	router := mux.NewRouter()
	router.Use(reservationHandler.MiddlewareContentTypeSet)

	getAvailablePeriodsByAccommodationRouter := router.Methods(http.MethodGet).Path("/{id}/periods").Subrouter()
	getAvailablePeriodsByAccommodationRouter.HandleFunc("", reservationHandler.GetAllAvailablePeriodsByAccommodation)
	getAvailablePeriodsByAccommodationRouter.Use(reservationHandler.AuthorizeRoles("HOST", "GUEST"))

	getReservationsByAvailablePeriodRouter := router.Methods(http.MethodGet).Path("/{id}/reservations").Subrouter()
	getReservationsByAvailablePeriodRouter.HandleFunc("", reservationHandler.GetAllReservationByAvailablePeriod)
	getReservationsByAvailablePeriodRouter.Use(reservationHandler.AuthorizeRoles("HOST", "GUEST"))

	findAvailablePeriodByIdAndByAccommodationId := router.Methods(http.MethodGet).Path("/{accommodationID}/{periodID}").Subrouter()
	findAvailablePeriodByIdAndByAccommodationId.HandleFunc("", reservationHandler.FindAvailablePeriodByIdAndByAccommodationId)
	findAvailablePeriodByIdAndByAccommodationId.Use(reservationHandler.AuthorizeRoles("HOST", "GUEST"))

	postAvailablePeriodsByAccommodationRouter := router.Methods(http.MethodPost).Path("/period").Subrouter()
	postAvailablePeriodsByAccommodationRouter.HandleFunc("", reservationHandler.CreateAvailablePeriod)
	postAvailablePeriodsByAccommodationRouter.Use(reservationHandler.MiddlewareAvailablePeriodDeserialization)
	postAvailablePeriodsByAccommodationRouter.Use(reservationHandler.AuthorizeRoles("HOST"))

	postReservationRouter := router.Methods(http.MethodPost).Path("/reservation").Subrouter()
	postReservationRouter.HandleFunc("", reservationHandler.CreateReservation)
	postReservationRouter.Use(reservationHandler.AuthorizeRoles("GUEST"))
	postReservationRouter.Use(reservationHandler.MiddlewareReservationDeserialization)

	updateAvailablePeriodsByAccommodationRouter := router.Methods(http.MethodPatch).Path("/period").Subrouter()
	updateAvailablePeriodsByAccommodationRouter.HandleFunc("", reservationHandler.UpdateAvailablePeriodByAccommodation)
	updateAvailablePeriodsByAccommodationRouter.Use(reservationHandler.MiddlewareAvailablePeriodDeserialization)
	updateAvailablePeriodsByAccommodationRouter.Use(reservationHandler.AuthorizeRoles("HOST"))

	deleteReservation := router.Methods(http.MethodDelete).Path("/{periodID}/{reservationID}").Subrouter()
	deleteReservation.HandleFunc("", reservationHandler.DeleteReservation)
	deleteReservation.Use(reservationHandler.AuthorizeRoles("GUEST"))

	cors := gorillaHandlers.CORS(gorillaHandlers.AllowedOrigins([]string{"*"}))

	//Initialize the server
	server := http.Server{
		Addr:         ":" + port,
		Handler:      cors(router),
		IdleTimeout:  120 * time.Second,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
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
