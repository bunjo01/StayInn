package main

import (
	"accommodation/data"
	"accommodation/handlers"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	gorillaHandlers "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func main() {
	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "8080"
	}

	// Inicijalizacija konteksta
	timeoutContext, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Inicijalizacija loggera
	logger := log.New(os.Stdout, "[accommodation-api] ", log.LstdFlags)
	storeLogger := log.New(os.Stdout, "[accommodation-store] ", log.LstdFlags)

	// Čitanje okoline za Cassandra povezivanje
	// cassandraHost := os.Getenv("CASSANDRA_HOST")
	// cassandraPortStr := os.Getenv("CASSANDRA_PORT")
	// cassandraPort, err := strconv.Atoi(cassandraPortStr)
	// if err != nil {
	// 	logger.Fatalf("Failed to parse CASSANDRA_PORT: %v", err)
	// }
	// cassandraUser := os.Getenv("CASSANDRA_USER")
	// cassandraPassword := os.Getenv("CASSANDRA_PASSWORD")

	// // Inicijalizacija sesije za Cassandra bazu
	// cluster := gocql.NewCluster(cassandraHost)
	// cluster.Keyspace = "accommodation"
	// cluster.Port = cassandraPort
	// cluster.Authenticator = gocql.PasswordAuthenticator{
	// 	Username: cassandraUser,
	// 	Password: cassandraPassword,
	// }

	// session, err := cluster.CreateSession()
	// if err != nil {
	// 	logger.Fatal(err)
	// }
	// defer session.Close()

	// Inicijalizacija repozitorijuma za smeštaj
	store, err := data.NewAccommodationRepository(storeLogger)
	if err != nil {
		logger.Fatal(err)
	}
	defer store.CloseSession()

	// newAccommodation := &data.Accommodation{
    //     ID:         gocql.TimeUUID(),
    //     Name:       "Primer smeštaja",
    //     Location:   "Primer lokacije",
    //     Amenities:  []data.AmenityEnum{0, 1, 2, 3, 4, 5},
    //     MinGuests:  2,
    //     MaxGuests:  4,
    // }

	// store.CreateAccommodationTable()

	// err = store.CreateAccommodation(context.Background(), newAccommodation)
    // if err != nil {
    //     logger.Fatal(err)
    // }

    // logger.Println("Smeštaj kreiran uspešno.")

	accommodationsHandler := handlers.NewAccommodationsHandler(logger, store)

	// Inicijalizacija router-a
	router := mux.NewRouter()


	router.HandleFunc("/accommodation", accommodationsHandler.CreateAccommodation).Methods("POST")
	router.HandleFunc("/accommodation", accommodationsHandler.GetAllAccommodations).Methods("GET")
	router.HandleFunc("/accommodation/{id}",accommodationsHandler.GetAccommodation).Methods("GET")
	router.HandleFunc("/accommodation/{id}",accommodationsHandler.UpdateAccommodation).Methods("PUT")
	router.HandleFunc("/accommodation/{id}",accommodationsHandler.DeleteAccommodation).Methods("DELETE")

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
