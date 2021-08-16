package weather_repo_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestWeatherRepo(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "WeatherRepo Suite")
}
