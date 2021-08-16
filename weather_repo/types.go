package weather_repo

import "fmt"

type Record struct {
	ID                int64
	DateStr           string
	Location          LocationRecord
	HourlyTemperature []float64
}

type LocationRecord struct {
	Lat   float64
	Lon   float64
	City  string
	State string
}

type QueryFilter struct {
	Date string
}

type ErrorDuplicateID int

func (e ErrorDuplicateID) Error() string {
	return fmt.Sprintf("a record already exists with ID %d", int(e))
}
