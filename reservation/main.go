package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"

	"reservation/clients"
	"reservation/data"
	"reservation/domain"
	"reservation/handlers"

	gorillaHandlers "github.com/gorilla/handlers"
	log "github.com/sirupsen/logrus"
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
	lumberjackLogger := &lumberjack.Logger{
		Filename: "/logger/logs/rese.log",
		MaxSize:  1,  //MB
		MaxAge:   30, //days
	}

	log.SetOutput(io.MultiWriter(os.Stdout, lumberjackLogger))
	log.SetLevel(log.InfoLevel)

	// Initializing Cassandra DB
	db := os.Getenv("CASS_DB")
	cassandraHost := os.Getenv("CASSANDRA_HOST")
	cassandraPortStr := os.Getenv("CASSANDRA_PORT")
	cassandraPort, err := strconv.Atoi(cassandraPortStr)
	if err != nil {
		log.Fatal(fmt.Sprintf("[rese-service]rs#1 Failed to initialize repo: %v", err))
	}
	cassandraUser := os.Getenv("CASSANDRA_USER")
	cassandraPassword := os.Getenv("CASSANDRA_PASSWORD")

	cluster := gocql.NewCluster(db)
	cluster.Keyspace = "system"

	session, err := cluster.CreateSession()
	if err != nil {
		log.Fatal(fmt.Sprintf("[rese-service]rs#2 Failed to create session: %v", err))
	}

	err = session.Query(
		fmt.Sprintf(`CREATE KEYSPACE IF NOT EXISTS %s
						WITH replication = {
							'class' : 'SimpleStrategy',
							'replication_factor' : %d
						}`, "reseervation", 1)).Exec()

	if err != nil {
		session.Close()
		log.Fatal(fmt.Sprintf("[rese-service]rs#3 Failed to create namespaces: %v", err))
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
		log.Fatal(fmt.Sprintf("[rese-service]rs#4 Failed to create Cassandra session: %v", err))
	} else {
		defer session.Close()
		log.Info(fmt.Sprintf("[rese-service]rs#5 Connected to Cassandra: %v", err))
	}

	// Initializing repo
	store, err := data.New(session)
	if err != nil {
		log.Fatal(fmt.Sprintf("[rese-service]rs#6 Failed to initialize res handler: %v", err))
	}

	err = store.CreateTables()
	if err != nil {
		log.Fatal(fmt.Sprintf("[rese-service]rs#7 Failed to create Cassandra tables: %v", err))
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
				log.Info(fmt.Sprintf("[rese-service]rs#8 CB '%s' changed from '%s' to '%s'", name, from, to))
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
				log.Info(fmt.Sprintf("[rese-service]rs#9 CB '%s' changed from '%s' to '%s'", name, from, to))
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
				log.Info(fmt.Sprintf("[rese-service]rs#10 CB '%s' changed from '%s' to '%s'", name, from, to))
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

	notification := clients.NewNotificationClient(notificationClient, os.Getenv("NOTIFICATION_SERVICE_URI"), notificationBreaker)
	profile := clients.NewProfileClient(profileClient, os.Getenv("PROFILE_SERVICE_URI"), profileBreaker)
	accommodation := clients.NewAccommodationClient(accommodationClient, os.Getenv("ACCOMMODATION_SERVICE_URI"), accommodationBreaker)

	//Initialize the handler and inject said logger
	reservationHandler := handlers.NewReservationHandler(store, notification, profile, accommodation)

	//Initialize the router and add a middleware for all the requests
	router := mux.NewRouter()
	router.Use(reservationHandler.MiddlewareContentTypeSet)

	getAvailablePeriodsByAccommodationRouter := router.Methods(http.MethodGet).Path("/{id}/periods").Subrouter()
	getAvailablePeriodsByAccommodationRouter.HandleFunc("", reservationHandler.GetAllAvailablePeriodsByAccommodation)
	getAvailablePeriodsByAccommodationRouter.Use(reservationHandler.AuthorizeRoles("HOST", "GUEST"))

	getReservationsByAvailablePeriodRouter := router.Methods(http.MethodGet).Path("/{id}/reservations").Subrouter()
	getReservationsByAvailablePeriodRouter.HandleFunc("", reservationHandler.GetAllReservationByAvailablePeriod)
	getReservationsByAvailablePeriodRouter.Use(reservationHandler.AuthorizeRoles("HOST", "GUEST"))

	findAccommodationIdsByDates := router.Methods(http.MethodPost).Path("/search").Subrouter()
	findAccommodationIdsByDates.HandleFunc("", reservationHandler.FindAccommodationIdsByDates)
	findAccommodationIdsByDates.Use(reservationHandler.MiddlewareDatesDeserialization)

	findAvailablePeriodByIdAndByAccommodationId := router.Methods(http.MethodGet).Path("/{accommodationID}/{periodID}").Subrouter()
	findAvailablePeriodByIdAndByAccommodationId.HandleFunc("", reservationHandler.FindAvailablePeriodByIdAndByAccommodationId)
	findAvailablePeriodByIdAndByAccommodationId.Use(reservationHandler.AuthorizeRoles("HOST", "GUEST"))

	findAllReservationsByUserIDExpired := router.Methods(http.MethodGet).Path("/expired").Subrouter()
	findAllReservationsByUserIDExpired.HandleFunc("", reservationHandler.FindAllReservationsByUserIDExpired)

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

	deletePeriodsByAccommodationRouter := router.Methods(http.MethodPost).Path("/check-acc").Subrouter()
	deletePeriodsByAccommodationRouter.HandleFunc("", reservationHandler.DeletePeriodsForAccommodations)
	deletePeriodsByAccommodationRouter.Use(reservationHandler.AuthorizeRoles("HOST"))

	getReservationsByUserIdRouter := router.Methods(http.MethodGet).Path("/user/{username}/reservations").Subrouter()
	getReservationsByUserIdRouter.HandleFunc("", reservationHandler.GetAllReservationsByUser)

	checkReservationsByUserIdRouter := router.Methods(http.MethodDelete).Path("/user/{id}/reservations").Subrouter()
	checkReservationsByUserIdRouter.HandleFunc("", reservationHandler.CheckAndDeleteReservationsForUser)
	checkReservationsByUserIdRouter.Use(reservationHandler.AuthorizeRoles("GUEST"))

	cors := gorillaHandlers.CORS(gorillaHandlers.AllowedOrigins([]string{"*"}))

	//Initialize the server
	server := http.Server{
		Addr:         ":" + port,
		Handler:      cors(router),
		IdleTimeout:  120 * time.Second,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	log.Info(fmt.Sprintf("[rese-service]rs#11 Listening on port: %v", port))
	//Distribute all the connections to goroutines
	go func() {
		err := server.ListenAndServe()
		if err != nil {
			if err != nil {
				log.Fatal(fmt.Sprintf("[rese-service]rs#12 Error while serving request: %v", err))
			}
		}
	}()

	// Protecting logs from unauthorized access and modification
	dirPath := "/logger/logs"
	err = protectLogs(dirPath)
	if err != nil {
		log.Fatal(fmt.Sprintf("[rese-service]rs#16 Error while protecting logs: %v", err))
	}

	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, os.Interrupt)
	signal.Notify(sigCh, os.Kill)

	sig := <-sigCh
	log.Info(fmt.Sprintf("[rese-service]rs#13 Recieved terminate, starting gracefull shutdown %v", sig))

	//Try to shutdown gracefully
	if server.Shutdown(timeoutContext) != nil {
		log.Fatal("[rese-service]rs#14 Cannot gracefully shutdown")
	}

	log.Info("[rese-service]rs#15 Server gracefully stopped")

}

// Changes ownership and sets permissions
func protectLogs(dirPath string) error {
	// Walk through all files in the directory
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return errors.New("error accessing path " + dirPath)
		}

		// Change ownership to user
		if err := os.Chown(path, 0, 0); err != nil {
			log.Fatal(fmt.Sprintf("[rese-service]rs#17 Failed to set log ownership to root: %v", err))
			return errors.New("error changing onwership to root for " + path)
		}

		// Set read-only permissions for the owner
		if err := os.Chmod(path, 0400); err != nil {
			log.Fatal(fmt.Sprintf("[rese-service]rs#18 Failed to set read-only permissions for root: %v", err))
			return errors.New("error changing permissions for " + path)
		}

		return nil
	})
	return err
}
