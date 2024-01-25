package main

import (
	"auth/clients"
	"auth/data"
	"auth/domain"
	"auth/handlers"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	log "github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"

	gorillaHandlers "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/sony/gobreaker"
)

func main() {
	//Reading from environment, if not set we will default it to 8081.
	//This allows flexibility in different environments (for eg. when running multiple docker api's and want to override the default port)
	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "8081"
	}

	// Initialize context
	timeoutContext, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	//Initialize the logger we are going to use
	lumberjackLogger := &lumberjack.Logger{
		Filename: "/logger/logs/auth.log",
		MaxSize:  1,  //MB
		MaxAge:   30, //days
	}

	log.SetOutput(io.MultiWriter(os.Stdout, lumberjackLogger))
	log.SetLevel(log.InfoLevel)

	// NoSQL: Initialize Auth Repository store
	store, err := data.New(timeoutContext)
	if err != nil {
		log.Fatal(fmt.Sprintf("[auth-service]as#10 Failed to initialize repo: %v", err))
	}
	defer store.Disconnect(timeoutContext)

	// NoSQL: Checking if the connection was established
	store.Ping()

	//Creating clients for other services
	profileClient := &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:        10,
			MaxIdleConnsPerHost: 10,
			MaxConnsPerHost:     10,
		},
	}

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
				log.Info(fmt.Sprintf("[auth-service]as#1 CB '%s' changed from '%s' to '%s'", name, from, to))
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

	profile := clients.NewProfileClient(profileClient, os.Getenv("PROFILE_SERVICE_URI"), profileBreaker)

	//Initialize the handler and inject logger and other services clients
	credentialsHandler := handlers.NewCredentialsHandler(store, profile)

	//Initialize the router and add a middleware for all the requests
	router := mux.NewRouter()

	// TODO Router

	router.HandleFunc("/login", credentialsHandler.Login).Methods("POST")
	router.HandleFunc("/register", credentialsHandler.Register).Methods("POST")
	router.HandleFunc("/change-password", credentialsHandler.ChangePassword).Methods("POST")
	router.HandleFunc("/activate/{activationUUID}", credentialsHandler.ActivateAccount).Methods("GET")
	router.HandleFunc("/recover-password", credentialsHandler.SendRecoveryEmail).Methods("POST")
	router.HandleFunc("/recovery-password", credentialsHandler.UpdatePasswordWithRecoveryUUID).Methods("POST")
	router.HandleFunc("/getAllUsers", credentialsHandler.GetAllUsers).Methods("GET")
	router.HandleFunc("/update-username/{oldUsername}/{username}", credentialsHandler.UpdateUsername).Methods("PUT")
	router.HandleFunc("/update-email/{oldEmail}/{email}", credentialsHandler.UpdateEmail).Methods("PUT")

	deleteUserRouter := router.Methods(http.MethodDelete).Path("/delete/{username}").Subrouter()
	deleteUserRouter.HandleFunc("", credentialsHandler.DeleteUser)
	deleteUserRouter.Use(credentialsHandler.AuthorizeRoles("HOST", "GUEST"))

	cors := gorillaHandlers.CORS(gorillaHandlers.AllowedOrigins([]string{"*"}))

	//Initialize the server
	server := http.Server{
		Addr:         ":" + port,
		Handler:      cors(router),
		IdleTimeout:  120 * time.Second,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	log.Info(fmt.Sprintf("[auth-service]as#2 Server listening on port %s", port))
	//Distribute all the connections to goroutines
	go func() {
		err := server.ListenAndServe()
		if err != nil {
			log.Fatal(fmt.Sprintf("[auth-service]as#3 Error while serving request: %v", err))
		}
	}()

	// Protecting logs from unauthorized access and modification
	dirPath := "/logger/logs"
	err = protectLogs(dirPath)
	if err != nil {
		log.Fatal(fmt.Sprintf("[auth-service]as#9 Error while protecting logs: %v", err))
	}

	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, os.Interrupt)
	signal.Notify(sigCh, os.Kill)

	sig := <-sigCh
	log.Info(fmt.Sprintf("[auth-service]as#4 Recieved terminate, starting gracefull shutdown %v", sig))

	//Try to shutdown gracefully
	if server.Shutdown(timeoutContext) != nil {
		log.Fatal("[auth-service]as#5 Cannot gracefully shutdown")
	}
	log.Info("[auth-service]as#6 Server gracefully stopped")
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
			log.Fatal(fmt.Sprintf("[auth-service]as#7 Failed to set log ownership to root: %v", err))
			return errors.New("error changing onwership to root for " + path)
		}

		// Set read-only permissions for the owner
		if err := os.Chmod(path, 0400); err != nil {
			log.Fatal(fmt.Sprintf("[auth-service]as#8 Failed to set read-only permissions for root: %v", err))
			return errors.New("error changing permissions for " + path)
		}

		return nil
	})
	return err
}
