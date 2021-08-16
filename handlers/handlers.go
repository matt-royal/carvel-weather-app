package handlers

import (
	"encoding/json"
	"errors"
	"github.com/matt-royal/carvel-weather-app/weather_repo"
	"log"
	"net/http"
)

type WeatherRepo interface {
	GetAll(weather_repo.QueryFilter) ([]weather_repo.Record, error)
	Create(weather_repo.Record) error
	DeleteAll() error
}

func ListWeatherRecords(repo WeatherRepo, logger *log.Logger) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		filter := weather_repo.QueryFilter{}
		date, ok := r.URL.Query()["date"]
		if ok {
			filter.Date = date[0]
		}

		records, err := repo.GetAll(filter)
		if err != nil {
			handleInternalServerError(w, logger, "in ListWeatherRecords from repo.GetAll", err)
			return
		}

		err = json.NewEncoder(w).Encode(fromRecords(records))
		if err != nil {
			handleInternalServerError(w, logger, "in ListWeatherRecords from json.NewEncoder.Encode", err)
			return
		}
	}
}

func CreateWeatherRecord(repo WeatherRepo, logger *log.Logger) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var weatherBody WeatherBody
		err := json.NewDecoder(r.Body).Decode(&weatherBody)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Invalid JSON body"))
			return
		}

		if errors := weatherBody.Validate(); len(errors) > 0 {
			w.WriteHeader(http.StatusUnprocessableEntity)
			w.Write([]byte("Invalid record. Encountered the following errors: \n"))
			for _, err = range errors {
				w.Write([]byte(err.Error() + "\n"))
			}
			return
		}

		err = repo.Create(toRecord(weatherBody))
		if err != nil {
			var duplicateIDErr weather_repo.ErrorDuplicateID
			if errors.As(err, &duplicateIDErr) {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(duplicateIDErr.Error()))
			} else {
				handleInternalServerError(w, logger, "in CreateWeatherRecord from repo.Create", err)
			}
			return
		}
		w.WriteHeader(http.StatusCreated)
	}
}

func DeleteAllWeatherRecords(repo WeatherRepo, logger *log.Logger) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		err := repo.DeleteAll()
		if err != nil {
			handleInternalServerError(w, logger, "in DeleteAllWeatherRecords from repo.DeleteAll", err)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func toRecord(body WeatherBody) weather_repo.Record {
	return weather_repo.Record{
		ID:                body.ID,
		DateStr:           body.Date,
		Location:          weather_repo.LocationRecord{
			Lat:   body.Location.Lat,
			Lon:   body.Location.Lon,
			City:  body.Location.City,
			State: body.Location.State,
		},
		HourlyTemperature: body.Temperature,
	}
}

func fromRecords(records []weather_repo.Record) []WeatherBody {
	responseRecords := make([]WeatherBody, len(records))
	for i, record := range records {
		responseRecords[i] = WeatherBody{
			ID:   record.ID,
			Date: record.DateStr,
			Location:          LocationBody{
				Lat:   record.Location.Lat,
				Lon:   record.Location.Lon,
				City:  record.Location.City,
				State: record.Location.State,
			},
			Temperature: record.HourlyTemperature,
		}

	}
	return responseRecords
}

func handleInternalServerError(w http.ResponseWriter, logger *log.Logger, logContext string, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte("Encountered a server error"))
	logger.Printf("Request failed due to error: %v: %v", logContext, err.Error())
}

