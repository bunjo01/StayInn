package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"profile/data"
	"profile/handlers"
	"time"

	gorillaHandlers "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

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
	logger := log.New(os.Stdout, "[profile-api] ", log.LstdFlags)
	storeLogger := log.New(os.Stdout, "[profile-store] ", log.LstdFlags)

	// NoSQL: Initialize Profile Repository store
	store, err := data.New(timeoutContext, storeLogger)
	if err != nil {
		logger.Fatal(err)
	}
	defer store.Disconnect(timeoutContext)

	// NoSQL: Checking if the connection was established
	store.Ping()

	// Initialize the handler and inject said logger
	userHandler := handlers.NewUserHandler(logger, store)

	// Initialize the router and add a middleware for all the requests
	router := mux.NewRouter()

	// Router

	router.HandleFunc("/users", userHandler.CreateUser).Methods("POST")
	router.HandleFunc("/users", userHandler.GetAllUsers).Methods("GET")
	router.HandleFunc("/users/{username}", userHandler.GetUser).Methods("GET")
	router.HandleFunc("/api/users/check-username/{username}", userHandler.CheckUsernameAvailability).Methods("GET")
	router.HandleFunc("/users/{id}", userHandler.UpdateUser).Methods("PUT")
	router.HandleFunc("/users/{username}", userHandler.DeleteUser).Methods("DELETE")
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
