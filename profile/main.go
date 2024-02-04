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
	"profile/clients"
	"profile/data"
	"profile/domain"
	"profile/handlers"
	"time"

	log "github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"

	gorillaHandlers "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/sony/gobreaker"
)

const usersUsername = "/users/{username}"

func main() {
	//Reading from environment, if not set we will default it to 8083.
	//This allows flexibility in different environments (for eg. when running multiple docker api's and want to override the default port)
	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "8083"
	}

	// Initialize context
	timeoutContext, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	//Initialize the logger we are going to use, with prefix and datetime for every log
	lumberjackLogger := &lumberjack.Logger{
		Filename: "/logger/logs/prof.log",
		MaxSize:  1,  //MB
		MaxAge:   30, //days
	}

	log.SetOutput(io.MultiWriter(os.Stdout, lumberjackLogger))
	log.SetLevel(log.InfoLevel)

	// NoSQL: Initialize Profile Repository store
	store, err := data.New(timeoutContext)
	if err != nil {
		log.Fatal(fmt.Sprintf("[prof-service]ps#10 Failed to initialize repo: %v", err))
	}
	defer store.Disconnect(timeoutContext)

	// NoSQL: Checking if the connection was established
	store.Ping()

	// Creating clients for other services
	accommodationClient := &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:        10,
			MaxIdleConnsPerHost: 10,
			MaxConnsPerHost:     10,
		},
	}

	authClient := &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:        10,
			MaxIdleConnsPerHost: 10,
			MaxConnsPerHost:     10,
		},
	}

	reservationClient := &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:        10,
			MaxIdleConnsPerHost: 10,
			MaxConnsPerHost:     10,
		},
	}

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
				log.Info(fmt.Sprintf("[prof-service]ps#1 CB '%s' changed from '%s' to '%s'\n", name, from, to))
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

	authBreaker := gobreaker.NewCircuitBreaker(
		gobreaker.Settings{
			Name:        "auth",
			MaxRequests: 1,
			Timeout:     10 * time.Second,
			Interval:    0,
			ReadyToTrip: func(counts gobreaker.Counts) bool {
				return counts.ConsecutiveFailures > 2
			},
			OnStateChange: func(name string, from, to gobreaker.State) {
				log.Info(fmt.Sprintf("[prof-service]ps#2 CB '%s' changed from '%s' to '%s'\n", name, from, to))
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

	reservationBreaker := gobreaker.NewCircuitBreaker(
		gobreaker.Settings{
			Name:        "reservation",
			MaxRequests: 1,
			Timeout:     10 * time.Second,
			Interval:    0,
			ReadyToTrip: func(counts gobreaker.Counts) bool {
				return counts.ConsecutiveFailures > 2
			},
			OnStateChange: func(name string, from, to gobreaker.State) {
				log.Info(fmt.Sprintf("[prof-service]ps#3 CB '%s' changed from '%s' to '%s'\n", name, from, to))
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

	accommodation := clients.NewAccommodationClient(accommodationClient, os.Getenv("ACCOMMODATION_SERVICE_URI"), accommodationBreaker)
	auth := clients.NewAuthClient(authClient, os.Getenv("AUTH_SERVICE_URI"), authBreaker)
	reservation := clients.NewReservationClient(reservationClient, os.Getenv("RESERVATION_SERVICE_URI"), reservationBreaker)

	// Initialize the handler and inject said logger
	userHandler := handlers.NewUserHandler(store, accommodation, auth, reservation)

	// Initialize the router and add a middleware for all the requests
	router := mux.NewRouter()

	// Router

	router.HandleFunc("/users", userHandler.CreateUser).Methods("POST")
	router.HandleFunc("/users", userHandler.GetAllUsers).Methods("GET")
	router.HandleFunc(usersUsername, userHandler.GetUser).Methods("GET")
	router.HandleFunc("/users/get-user-by-id", userHandler.GetUserById).Methods("POST")
	router.HandleFunc("/api/users/check-username/{username}", userHandler.CheckUsernameAvailability).Methods("GET")
	router.HandleFunc(usersUsername, userHandler.UpdateUser).Methods("PUT")
	router.HandleFunc(usersUsername, userHandler.DeleteUser).Methods("DELETE")
	router.Use(userHandler.AuthorizeRoles("HOST", "GUEST"))

	cors := gorillaHandlers.CORS(
		gorillaHandlers.AllowedOrigins([]string{"*"}),
		gorillaHandlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "DELETE", "OPTIONS"}),
		gorillaHandlers.AllowedHeaders([]string{"Content-Type"}),
	)

	// Initialize the server
	server := http.Server{
		Addr:         ":" + port,
		Handler:      cors(router),
		IdleTimeout:  120 * time.Second,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	log.Info(fmt.Sprintf("[prof-service]ps#4 Server listening on port %s", port))
	//Distribute all the connections to goroutines
	go func() {
		err := server.ListenAndServe()
		if err != nil {
			log.Fatal(fmt.Sprintf("[prof-service]ps#5 Error while serving request: %v", err))
		}
	}()

	// Protecting logs from unauthorized access and modification
	dirPath := "/logger/logs"
	err = protectLogs(dirPath)
	if err != nil {
		log.Fatal(fmt.Sprintf("[prof-service]ps#9 Error while protecting logs: %v", err))
	}

	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, os.Interrupt)
	signal.Notify(sigCh, os.Kill)

	sig := <-sigCh
	log.Info(fmt.Sprintf("[prof-service]ps#6 Recieved terminate, starting gracefull shutdown %v", sig))

	//Try to shutdown gracefully
	if server.Shutdown(timeoutContext) != nil {
		log.Fatal("[prof-service]ps#7 Cannot gracefully shutdown...")
	}
	log.Info("[prof-service]ps#8 Server gracefully stopped")
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
			log.Fatal(fmt.Sprintf("[prof-service]ps#10 Failed to set log ownership to root: %v", err))
			return errors.New("error changing onwership to root for " + path)
		}

		// Set read-only permissions for the owner
		if err := os.Chmod(path, 0400); err != nil {
			log.Fatal(fmt.Sprintf("[prof-service]ps#11 Failed to set read-only permissions for root: %v", err))
			return errors.New("error changing permissions for " + path)
		}

		return nil
	})
	return err
}
