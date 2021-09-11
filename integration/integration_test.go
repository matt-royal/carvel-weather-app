package integration_test

import (
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/sclevine/spec"
)

func TestIntegration(t *testing.T) {
	g := NewGomegaWithT(t)

	buildDir, err := gexec.Build("../.")
	g.Expect(err).NotTo(HaveOccurred())
	defer gexec.CleanupBuildArtifacts()

	serverCmd := filepath.Join(buildDir, "carvel-weather-app")

	spec.Run(t, "", func(t *testing.T, when spec.G, it spec.S) {
		var (
			serverSession   *gexec.Session
			weatherEndpoint string
			eraseEndpoint   string
		)

		it.Before(func() {
			serverSession = nil
			port := 9876
			command := exec.Command(serverCmd)
			command.Env = []string{fmt.Sprintf("PORT=%d", port)}

			var err error
			serverSession, err = gexec.Start(command, it.Out(), it.Out())
			g.Expect(err).NotTo(HaveOccurred())

			appURL := fmt.Sprintf("http://localhost:%d", port)
			weatherEndpoint = appURL + "/weather"
			eraseEndpoint = appURL + "/erase"

			// Quick and dirty way to wait for the server to start listening
			g.Eventually(func() error {
				_, err := http.Get(appURL)
				return err
			}).ShouldNot(HaveOccurred())
		})

		it.After(func() {
			if serverSession != nil {
				serverSession.Kill()
			}
		})

		const weather1JSON = `
		  {
			"id":1,
			"date":"1985-01-01",
			"location": {
			  "lat": 36.1189,
			  "lon": -86.6892,
			  "city":"Palo Alto",
			  "state":"California"
			},
			"temperature": [
			  37.3, 36.8, 36.4, 36.0, 35.6, 35.3, 35.0, 34.9, 35.8, 38.0, 40.2, 42.3,
			  43.8, 44.9, 45.5, 45.7, 44.9, 43.0, 41.7, 40.8, 39.9, 39.2, 38.6, 38.1
			]
		  }`
		const weather2JSON = `
		  {
			"id":2,
			"date":"1985-01-01",
			"location": {
			  "lat": 37.7818,
			  "lon": -122.4061635,
			  "city":"San Francisco",
			  "state":"California"
			},
			"temperature": [
			  36.3, 35.8, 35.4, 35.0, 34.6, 34.3, 34.0, 33.9, 34.8, 37.0, 39.2, 41.3,
			  42.8, 43.9, 44.5, 44.7, 43.9, 42.0, 40.7, 39.8, 38.9, 38.2, 37.6, 37.1
			]
		  }`
		const weather3JSON = `
		  {
			"id":3,
			"date":"1985-01-02",
			"location": {
			  "lat": 36.1189,
			  "lon": -86.6892,
			  "city":"Palo Alto",
			  "state":"California"
			},
			"temperature": [
			  38.3, 37.8, 37.4, 37.0, 36.6, 36.3, 36.0, 35.9, 36.8, 39.0, 41.2, 43.3,
			  44.8, 45.9, 46.5, 46.7, 45.9, 44.0, 42.7, 41.8, 40.9, 40.2, 39.6, 39.1
			]
		  }`

		it("succeeds at the happy path", func() {
			// Confirm no records exist
			resp, err := http.Get(weatherEndpoint)
			g.Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			g.Expect(resp.StatusCode).To(Equal(200))
			body, err := io.ReadAll(resp.Body)
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(body).To(MatchJSON("[]"))

			// Create 1 record
			resp, err = http.Post(weatherEndpoint, "application/json", strings.NewReader(weather3JSON))
			g.Expect(err).NotTo(HaveOccurred())

			g.Expect(resp.StatusCode).To(Equal(201))

			// Confirm the record was stored
			resp, err = http.Get(weatherEndpoint)
			g.Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			g.Expect(resp.StatusCode).To(Equal(200))
			body, err = io.ReadAll(resp.Body)
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(body).To(MatchJSON(fmt.Sprintf("[%s]", weather3JSON)))

			// Create 2 more records
			resp, err = http.Post(weatherEndpoint, "application/json", strings.NewReader(weather2JSON))
			g.Expect(err).NotTo(HaveOccurred())
			resp.Body.Close()
			g.Expect(resp.StatusCode).To(Equal(201))

			resp, err = http.Post(weatherEndpoint, "application/json", strings.NewReader(weather1JSON))
			g.Expect(err).NotTo(HaveOccurred())
			resp.Body.Close()
			g.Expect(resp.StatusCode).To(Equal(201))

			// Confirm the records were stored and returned in order
			resp, err = http.Get(weatherEndpoint)
			g.Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			g.Expect(resp.StatusCode).To(Equal(200))
			body, err = io.ReadAll(resp.Body)
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(body).To(MatchJSON(fmt.Sprintf("[%s, %s, %s]", weather1JSON, weather2JSON, weather3JSON)))

			// Confirm the records can be filtered by date
			resp, err = http.Get(fmt.Sprintf("%s?date=%s", weatherEndpoint, "1985-01-01"))
			g.Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			g.Expect(resp.StatusCode).To(Equal(200))
			body, err = io.ReadAll(resp.Body)
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(body).To(MatchJSON(fmt.Sprintf("[%s, %s]", weather1JSON, weather2JSON)))

			resp, err = http.Get(fmt.Sprintf("%s?date=%s", weatherEndpoint, "1985-04-04"))
			g.Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			g.Expect(resp.StatusCode).To(Equal(200))
			body, err = io.ReadAll(resp.Body)
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(body).To(MatchJSON("[]"))

			// Erase all records
			deleteReq, err := http.NewRequest(http.MethodDelete, eraseEndpoint, nil)
			g.Expect(err).NotTo(HaveOccurred())
			resp, err = http.DefaultClient.Do(deleteReq)
			g.Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			g.Expect(resp.StatusCode).To(Equal(200))

			// Confirm all records were destroyed
			resp, err = http.Get(weatherEndpoint)
			g.Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			g.Expect(resp.StatusCode).To(Equal(200))
			body, err = io.ReadAll(resp.Body)
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(body).To(MatchJSON("[]"))
		})

		when("a record exists", func() {
			const differentWeather1JSON = `
			  {
				"id":1,
				"date":"1985-01-02",
				"location": {
				  "lat": 37.7818,
				  "lon": -122.4061635,
				  "city":"San Francisco",
				  "state":"California"
				},
				"temperature": [
				  36.3, 35.8, 35.4, 35.0, 34.6, 34.3, 34.0, 33.9, 34.8, 37.0, 39.2, 41.3,
				  42.8, 43.9, 44.5, 44.7, 43.9, 42.0, 40.7, 39.8, 38.9, 38.2, 37.6, 37.1
				]
			  }`

			it("rejects the creation of a record with the same ID", func() {
				// Confirm no records exist
				resp, err := http.Get(weatherEndpoint)
				g.Expect(err).NotTo(HaveOccurred())
				defer resp.Body.Close()
				g.Expect(resp.StatusCode).To(Equal(200))
				body, err := io.ReadAll(resp.Body)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(body).To(MatchJSON("[]"))

				// Create a record
				resp, err = http.Post(weatherEndpoint, "application/json", strings.NewReader(weather1JSON))
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(resp.StatusCode).To(Equal(201))

				// Confirm a duplicate record cannot be created
				resp, err = http.Post(weatherEndpoint, "application/json", strings.NewReader(differentWeather1JSON))
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(resp.StatusCode).To(Equal(400))
			})
		})

		when("the JSON is invalid", func() {
			it("returns 400", func() {
				resp, err := http.Post(weatherEndpoint, "application/json", strings.NewReader("{"))
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(resp.StatusCode).To(Equal(400))
			})
		})

		when("the JSON is valid, but the data is invalid", func() {
			const invalidWeatherJSON = `
			  {
				"date":"1985-01",
				"location": {
				  "lat": 91,
				  "lon": 181,
				  "city":"",
				  "state":""
				},
				"temperature": [ 37.3, 36.8 ]
			  }`

			it("returns 422 and makes no record", func() {
				resp, err := http.Post(weatherEndpoint, "application/json", strings.NewReader(invalidWeatherJSON))
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(resp.StatusCode).To(Equal(422))

				resp, err = http.Get(weatherEndpoint)
				g.Expect(err).NotTo(HaveOccurred())
				defer resp.Body.Close()
				g.Expect(resp.StatusCode).To(Equal(200))
				body, err := io.ReadAll(resp.Body)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(body).To(MatchJSON("[]"))
			})
		})
	})
}
