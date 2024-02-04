package main

import (
	"accommodation/cache"
	"accommodation/clients"
	"accommodation/data"
	"accommodation/domain"
	"accommodation/handlers"
	"accommodation/storage"
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

const accommodationPath = "/accommodation/{id}"

func main() {
	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "8080"
	}

	// Context initialization
	timeoutContext, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Logger initialization
	lumberjackLogger := &lumberjack.Logger{
		Filename: "/logger/logs/acco.log",
		MaxSize:  1,  //MB
		MaxAge:   30, //days
	}

	log.SetOutput(io.MultiWriter(os.Stdout, lumberjackLogger))
	log.SetLevel(log.InfoLevel)

	// Initializing repo for accommodations
	store, err := data.NewAccommodationRepository(timeoutContext)
	if err != nil {
		log.Fatal(fmt.Sprintf("[acco-service]acs#1 Failed to initialize repo: %v", err))
	}
	defer store.Disconnect(timeoutContext)
	store.Ping()

	// Redis
	imageCache := cache.New()
	imageCache.Ping()

	// HDFS
	images, err := storage.New()
	if err != nil {
		log.Fatal(fmt.Sprintf("[acco-service]acs#2 Failed to initialize HDFS: %v", err))
	}

	defer images.Close()

	_ = images.CreateDirectories()

	// CBs
	reservationClient := &http.Client{
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
				log.Info(fmt.Sprintf("[acco-service]acs#3 CB '%s' changed from '%s' to '%s'\n", name, from, to))
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
				log.Info(fmt.Sprintf("[acco-service]acs#4 CB '%s' changed from '%s' to '%s'\n", name, from, to))
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

	reservation := clients.NewReservationClient(reservationClient, os.Getenv("RESERVATION_SERVICE_URI"), reservationBreaker)
	profile := clients.NewProfileClient(profileClient, os.Getenv("PROFILE_SERVICE_URI"), profileBreaker)

	accommodationsHandler := handlers.NewAccommodationsHandler(store, reservation, profile, imageCache, images)

	// Router init
	router := mux.NewRouter()

	createAccommodationRouter := router.Methods(http.MethodPost).Path("/accommodation").Subrouter()
	createAccommodationRouter.HandleFunc("", accommodationsHandler.CreateAccommodation)
	createAccommodationRouter.Use(accommodationsHandler.AuthorizeRoles("HOST"))

	createAccommodationImagesRouter := router.Methods(http.MethodPost).Path("/accommodation/images").Subrouter()
	createAccommodationImagesRouter.HandleFunc("", accommodationsHandler.CreateAccommodationImages)
	createAccommodationImagesRouter.Use(accommodationsHandler.AuthorizeRoles("HOST"))

	getAccommodationImagesRouter := router.Methods(http.MethodGet).Path(accommodationPath + "/images").Subrouter()
	getAccommodationImagesRouter.HandleFunc("", accommodationsHandler.GetAccommodationImages)
	getAccommodationImagesRouter.Use(accommodationsHandler.MiddlewareCacheAllHit)

	getAllAccommodationRouter := router.Methods(http.MethodGet).Path("/accommodation").Subrouter()
	getAllAccommodationRouter.HandleFunc("", accommodationsHandler.GetAllAccommodations)

	getAccommodationRouter := router.Methods(http.MethodGet).Path(accommodationPath).Subrouter()
	getAccommodationRouter.HandleFunc("", accommodationsHandler.GetAccommodation)

	getAccommodationsForUserRouter := router.Methods(http.MethodGet).Path("/user/{username}/accommodations").Subrouter()
	getAccommodationsForUserRouter.HandleFunc("", accommodationsHandler.GetAccommodationsForUser)

	updateAccommodationRouter := router.Methods(http.MethodPut).Path(accommodationPath).Subrouter()
	updateAccommodationRouter.HandleFunc("", accommodationsHandler.UpdateAccommodation)
	updateAccommodationRouter.Use(accommodationsHandler.AuthorizeRoles("HOST"))

	deleteAccommodationRouter := router.Methods(http.MethodDelete).Path(accommodationPath).Subrouter()
	deleteAccommodationRouter.HandleFunc("", accommodationsHandler.DeleteAccommodation)
	deleteAccommodationRouter.Use(accommodationsHandler.AuthorizeRoles("HOST"))

	deleteUserAccommodationsRouter := router.Methods(http.MethodDelete).Path("/user/{id}/accommodations").Subrouter()
	deleteUserAccommodationsRouter.HandleFunc("", accommodationsHandler.DeleteUserAccommodations)
	deleteUserAccommodationsRouter.Use(accommodationsHandler.AuthorizeRoles("HOST"))

	// Search part

	searchAccommodationRouter := router.Methods(http.MethodGet).Path("/search").Subrouter()
	searchAccommodationRouter.HandleFunc("", accommodationsHandler.SearchAccommodations)

	// CORS middleware
	cors := gorillaHandlers.CORS(
		gorillaHandlers.AllowedOrigins([]string{"*"}),
		gorillaHandlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "DELETE", "OPTIONS"}),
		gorillaHandlers.AllowedHeaders([]string{"Content-Type"}),
	)

	server := http.Server{
		Addr:         ":" + port,
		Handler:      cors(router),
		IdleTimeout:  120 * time.Second,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	log.Info(fmt.Sprintf("[acco-service]acs#5 Server listening on port %s", port))

	go func() {
		err := server.ListenAndServe()
		if err != nil {
			log.Fatal(fmt.Sprintf("[acco-service]acs#6 Error while serving request: %v", err))
		}
	}()

	// Protecting logs from unauthorized access and modification
	dirPath := "/logger/logs"
	err = protectLogs(dirPath)
	if err != nil {
		log.Fatal(fmt.Sprintf("[acco-service]acs#7 Error while protecting logs: %v", err))
	}

	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, os.Interrupt)
	signal.Notify(sigCh, os.Kill)

	sig := <-sigCh
	log.Info(fmt.Sprintf("[acco-service]acs#8 Recieved terminate, starting gracefull shutdown %v", sig))

	if err := server.Shutdown(timeoutContext); err != nil {
		log.Fatal("[acco-service]acs#9 Cannot gracefully shutdown")
	}
	log.Info("[acco-service]acs#10 Server gracefully stopped")
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
			log.Fatal(fmt.Sprintf("[acco-service]acs#11 Failed to set log ownership to root: %v", err))
			return errors.New("error changing onwership to root for " + path)
		}

		// Set read-only permissions for the owner
		if err := os.Chmod(path, 0400); err != nil {
			log.Fatal(fmt.Sprintf("[acco-service]acs#12 Failed to set read-only permissions for root: %v", err))
			return errors.New("error changing permissions for " + path)
		}

		return nil
	})
	return err
}
