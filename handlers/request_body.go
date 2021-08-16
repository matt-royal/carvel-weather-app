package handlers

import (
	"fmt"
	"regexp"
)

type InvalidFieldError struct {
	Field string
	Message string
}

func (e InvalidFieldError) Error() string {
	return fmt.Sprintf("%s %s", e.Field, e.Message)
}

type WeatherBody struct {
	ID          int64        `json:"id"`
	Date        string       `json:"date"`
	Location    LocationBody `json:"location"`
	Temperature []float64    `json:"temperature"`
}

func (b WeatherBody) Validate() []error {
	var errs []error

	if b.ID <= 0 {
		errs = append(errs, InvalidFieldError{Field: "id", Message: "must be greater than 0"})
	}

	dateRegex := regexp.MustCompile("^\\d{4}-\\d{2}-\\d{2}$")
	if !dateRegex.MatchString(b.Date) {
		errs = append(errs, InvalidFieldError{Field: "date", Message: "must be in the format \"YYYY-MM-DD\""})
	}

	if len(b.Temperature) != 24 {
		errs = append(errs, InvalidFieldError{Field: "temperature", Message: "must have exactly 24 entries"})
	}

	errs = append(errs, b.Location.Validate()...)
	return errs
}

type LocationBody struct {
	Lat   float64 `json:"lat"`
	Lon   float64 `json:"lon"`
	City  string  `json:"city"`
	State string  `json:"state"`
}

func (l LocationBody) Validate() []error {
	var errs []error

	if l.Lat > 90 || l.Lat < -90 {
		errs = append(errs, InvalidFieldError{Field: "location.lat", Message: "is invalid"})
	}

	if l.Lon > 180 || l.Lon < -180 {
		errs = append(errs, InvalidFieldError{Field: "location.lon", Message: "is invalid"})
	}

	if l.City == "" {
		errs = append(errs, InvalidFieldError{Field: "location.city", Message: "cannot be blank"})
	}

	if l.State == "" {
		errs = append(errs, InvalidFieldError{Field: "location.state", Message: "cannot be blank"})
	}

	return errs
}
