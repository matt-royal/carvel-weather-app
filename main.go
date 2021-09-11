package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	gorillaHandlers "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/matt-royal/carvel-weather-app/handlers"
	"github.com/matt-royal/carvel-weather-app/weather_repo"
)

func main() {
	router := mux.NewRouter()

	weatherRepo := new(weather_repo.InMemoryRepo)
	logger := log.New(os.Stderr, "", log.LstdFlags)

	router.Path("/weather").Methods(http.MethodGet).HandlerFunc(handlers.ListWeatherRecords(weatherRepo, logger))
	router.Path("/weather").Methods(http.MethodPost).HandlerFunc(handlers.CreateWeatherRecord(weatherRepo, logger))
	router.Path("/erase").Methods(http.MethodDelete).HandlerFunc(handlers.DeleteAllWeatherRecords(weatherRepo, logger))

	http.Handle("/", gorillaHandlers.CombinedLoggingHandler(os.Stderr, router))

	port := parsePort()
	logger.Printf("Listening on port %d", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}

func parsePort() int {
	portStr := os.Getenv("PORT")

	if portStr == "" {
		return 8080
	}

	port, err := strconv.Atoi(portStr)

	if err != nil {
		panic(fmt.Sprintf("Invalid PORT %q", portStr))
	}
	return port
}
