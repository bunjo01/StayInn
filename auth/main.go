package main

import (
	"auth/clients"
	"auth/data"
	"auth/domain"
	"auth/handlers"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	gorillaHandlers "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/sony/gobreaker"
)

func main() {
	//Reading from environment, if not set we will default it to 8080.
	//This allows flexibility in different environments (for eg. when running multiple docker api's and want to override the default port)
	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "8081"
	}

	// Initialize context
	timeoutContext, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	//Initialize the logger we are going to use, with prefix and datetime for every log
	logger := log.New(os.Stdout, "[auth-service] ", log.LstdFlags)
	storeLogger := log.New(os.Stdout, "[auth-store] ", log.LstdFlags)

	// NoSQL: Initialize Auth Repository store
	store, err := data.New(timeoutContext, storeLogger)
	if err != nil {
		logger.Fatal(err)
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

	profile := clients.NewProfileClient(profileClient, os.Getenv("PROFILE_SERVICE_URI")+"/users", profileBreaker)

	//Initialize the handler and inject logger and other services clients
	credentialsHandler := handlers.NewCredentialsHandler(logger, store, profile)

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
	//deleteUserRouter.Use(credentialsHandler.AuthorizeRoles("HOST", "GUEST"))

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
