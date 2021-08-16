package integration_test

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"io"
	"net/http"
	"os/exec"
	"path/filepath"
	"strings"
)

var _ = Describe("Integration", func() {
	var (
		serverCmd       string
		serverSession   *gexec.Session
		weatherEndpoint string
		eraseEndpoint   string
	)

	BeforeSuite(func() {
		buildDir, err := gexec.Build("../.")
		Expect(err).NotTo(HaveOccurred())
		serverCmd = filepath.Join(buildDir, "carvel-weather-app")
	})

	AfterSuite(func() {
		gexec.CleanupBuildArtifacts()
	})

	BeforeEach(func() {
		serverSession = nil
		port := 9876
		command := exec.Command(serverCmd)
		command.Env = []string{fmt.Sprintf("PORT=%d", port)}

		var err error
		serverSession, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

		appURL := fmt.Sprintf("http://localhost:%d", port)
		weatherEndpoint = appURL + "/weather"
		eraseEndpoint = appURL + "/erase"

		// Quick and dirty way to wait for the server to start listening
		Eventually(func() error {
			_, err := http.Get(appURL)
			return err
		}).ShouldNot(HaveOccurred())
	})

	AfterEach(func() {
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

	It("succeeds at the happy path", func() {
		By("Confirm no records exist")
		resp, err := http.Get(weatherEndpoint)
		Expect(err).NotTo(HaveOccurred())
		defer resp.Body.Close()
		Expect(resp.StatusCode).To(Equal(200))
		body, err := io.ReadAll(resp.Body)
		Expect(err).NotTo(HaveOccurred())
		Expect(body).To(MatchJSON("[]"))

		By("Create 1 record")
		resp, err = http.Post(weatherEndpoint, "application/json", strings.NewReader(weather3JSON))
		Expect(err).NotTo(HaveOccurred())

		Expect(resp.StatusCode).To(Equal(201))

		By("Confirm the record was stored")
		resp, err = http.Get(weatherEndpoint)
		Expect(err).NotTo(HaveOccurred())
		defer resp.Body.Close()
		Expect(resp.StatusCode).To(Equal(200))
		body, err = io.ReadAll(resp.Body)
		Expect(err).NotTo(HaveOccurred())
		Expect(body).To(MatchJSON(fmt.Sprintf("[%s]", weather3JSON)))

		By("Create 2 more records")
		resp, err = http.Post(weatherEndpoint, "application/json", strings.NewReader(weather2JSON))
		Expect(err).NotTo(HaveOccurred())
		resp.Body.Close()
		Expect(resp.StatusCode).To(Equal(201))

		resp, err = http.Post(weatherEndpoint, "application/json", strings.NewReader(weather1JSON))
		Expect(err).NotTo(HaveOccurred())
		resp.Body.Close()
		Expect(resp.StatusCode).To(Equal(201))

		By("Confirm the records were stored and returned in order")
		resp, err = http.Get(weatherEndpoint)
		Expect(err).NotTo(HaveOccurred())
		defer resp.Body.Close()
		Expect(resp.StatusCode).To(Equal(200))
		body, err = io.ReadAll(resp.Body)
		Expect(err).NotTo(HaveOccurred())
		Expect(body).To(MatchJSON(fmt.Sprintf("[%s, %s, %s]", weather1JSON, weather2JSON, weather3JSON)))

		By("Confirm the records can be filtered by date")
		resp, err = http.Get(fmt.Sprintf("%s?date=%s", weatherEndpoint, "1985-01-01"))
		Expect(err).NotTo(HaveOccurred())
		defer resp.Body.Close()
		Expect(resp.StatusCode).To(Equal(200))
		body, err = io.ReadAll(resp.Body)
		Expect(err).NotTo(HaveOccurred())
		Expect(body).To(MatchJSON(fmt.Sprintf("[%s, %s]", weather1JSON, weather2JSON)))

		resp, err = http.Get(fmt.Sprintf("%s?date=%s", weatherEndpoint, "1985-04-04"))
		Expect(err).NotTo(HaveOccurred())
		defer resp.Body.Close()
		Expect(resp.StatusCode).To(Equal(200))
		body, err = io.ReadAll(resp.Body)
		Expect(err).NotTo(HaveOccurred())
		Expect(body).To(MatchJSON("[]"))

		By("Erase all records")
		deleteReq, err := http.NewRequest(http.MethodDelete, eraseEndpoint, nil)
		Expect(err).NotTo(HaveOccurred())
		resp, err = http.DefaultClient.Do(deleteReq)
		Expect(err).NotTo(HaveOccurred())
		defer resp.Body.Close()
		Expect(resp.StatusCode).To(Equal(200))

		By("Confirm all records were destroyed")
		resp, err = http.Get(weatherEndpoint)
		Expect(err).NotTo(HaveOccurred())
		defer resp.Body.Close()
		Expect(resp.StatusCode).To(Equal(200))
		body, err = io.ReadAll(resp.Body)
		Expect(err).NotTo(HaveOccurred())
		Expect(body).To(MatchJSON("[]"))
	})

	When("a record exists", func() {
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

		It("rejects the creation of a record with the same ID", func() {
			By("Confirm no records exist")
			resp, err := http.Get(weatherEndpoint)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(200))
			body, err := io.ReadAll(resp.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(body).To(MatchJSON("[]"))

			By("Create a record")
			resp, err = http.Post(weatherEndpoint, "application/json", strings.NewReader(weather1JSON))
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(201))

			By("Confirm a duplicate record cannot be created")
			resp, err = http.Post(weatherEndpoint, "application/json", strings.NewReader(differentWeather1JSON))
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(400))
		})
	})

	When("the JSON is invalid", func() {
		It("returns 400", func() {
			resp, err := http.Post(weatherEndpoint, "application/json", strings.NewReader("{"))
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(400))
		})
	})

	When("the JSON is valid, but the data is invalid", func() {
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

		It("returns 422 and makes no record", func() {
			resp, err := http.Post(weatherEndpoint, "application/json", strings.NewReader(invalidWeatherJSON))
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(422))

			resp, err = http.Get(weatherEndpoint)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(200))
			body, err := io.ReadAll(resp.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(body).To(MatchJSON("[]"))
		})
	})
})
