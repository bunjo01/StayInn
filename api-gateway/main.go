package main

import (
	"context"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func main() {
	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "8090"
	}

	timeoutContext, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	logger := log.New(os.Stdout, "[api-gateway] ", log.LstdFlags)

	// Reverse proxies for services
	services := map[string]string{
		"accommodation": "http://localhost:8080",
		"auth":          "http://localhost:8081",
		"reservation":   "http://localhost:8082",
	}

	proxies := make(map[string]*httputil.ReverseProxy)
	for serviceName, serviceURL := range services {
		u, _ := url.Parse(serviceURL)
		proxies[serviceName] = httputil.NewSingleHostReverseProxy(u)
	}

	// Handler function for proxying requests to services
	proxyHandler := func(serviceName string) http.HandlerFunc {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			proxy := proxies[serviceName]
			proxy.ServeHTTP(w, r)
		})
	}

	router := mux.NewRouter()

	router.PathPrefix("/service/accommodation").Handler(http.StripPrefix("/service/accommodation", proxyHandler("accommodation")))
	router.PathPrefix("/service/auth").Handler(http.StripPrefix("/service/auth", proxyHandler("auth")))
	router.PathPrefix("/service/reservation").Handler(http.StripPrefix("/service/reservation", proxyHandler("reservation")))

	// CORS middleware
	cors := handlers.CORS(handlers.AllowedOrigins([]string{"*"}))

	// Initialize the server
	server := http.Server{
		Addr:         ":" + port,
		Handler:      cors(router),
		IdleTimeout:  120 * time.Second,
		ReadTimeout:  1 * time.Second,
		WriteTimeout: 1 * time.Second,
	}

	logger.Println("API Gateway listening on port", port)

	go func() {
		err := server.ListenAndServe()
		if err != nil {
			logger.Fatal(err)
		}
	}()

	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, os.Interrupt, os.Kill)
	sig := <-sigCh
	logger.Println("Received termination signal, attempting graceful shutdown", sig)

	if err := server.Shutdown(timeoutContext); err != nil {
		logger.Fatal("Error during graceful shutdown:", err)
	}

	logger.Println("API Gateway server stopped gracefully")
}
