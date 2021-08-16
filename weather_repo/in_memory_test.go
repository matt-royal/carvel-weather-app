package weather_repo_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"sync"

	. "github.com/matt-royal/carvel-weather-app/weather_repo"
)

var _ = Describe("InMemory WeatherRepo", func() {
	var repo *InMemoryRepo

	BeforeEach(func() {
		repo = new(InMemoryRepo)
	})

	Describe("Create", func() {
		It("can handles concurrency", func() {
			recordChan := make(chan Record)

			go func() {
				for id := int64(1); id <= 10; id++ {
					recordChan <- buildRecord(id)
				}
				close(recordChan)
			}()

			var wg sync.WaitGroup

			for i := 0; i < 3; i++ {
				wg.Add(1)
				go func() {
					for record := range recordChan {
						err := repo.Create(record)
						Expect(err).NotTo(HaveOccurred())
					}
					wg.Done()
				}()
			}

			wg.Wait()

			createdRecords, err := repo.GetAll(QueryFilter{})
			Expect(err).NotTo(HaveOccurred())
			Expect(createdRecords).To(HaveLen(10))

			var ids []int64
			for _, record := range createdRecords {
				ids = append(ids, record.ID)
			}
			Expect(ids).To(Equal([]int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}))
		})

		When("a record already exists with the same ID", func () {
			BeforeEach(func() {
				Expect(repo.Create(buildRecord(1))).To(Succeed())
			})

			It("returns a ErrorDuplicateID", func () {
				err := repo.Create(buildRecord(1))
				Expect(err).To(MatchError(ErrorDuplicateID(1)))
			})

			It("doesn't create the record", func() {
				_ = repo.Create(buildRecord(1))
				records, err := repo.GetAll(QueryFilter{})
				Expect(err).NotTo(HaveOccurred())
				Expect(records).To(HaveLen(1))
			})
		})
	})

	Describe("GetAll", func () {
		It("returns the item in GetAll()", func() {
			record := buildRecord(1)
			Expect(repo.Create(record)).To(Succeed())
			records, err := repo.GetAll(QueryFilter{})
			Expect(err).NotTo(HaveOccurred())
			Expect(records).To(Equal([]Record{record}))
		})

		It("sorts records by ID in GetAll()", func() {
			record1 := buildRecord(1)
			record2 := buildRecord(2)
			record3 := buildRecord(3)
			Expect(repo.Create(record2)).To(Succeed())
			Expect(repo.Create(record3)).To(Succeed())
			Expect(repo.Create(record1)).To(Succeed())
			records, err := repo.GetAll(QueryFilter{})
			Expect(err).NotTo(HaveOccurred())
			Expect(records).To(Equal([]Record{record1, record2, record3}))
		})

		When("a date filter is given", func () {
			XIt("returns only matching records")
		})
	})

	Describe("DeleteAll", func () {
		It("deletes all records", func () {
			Expect(repo.Create(buildRecord(1))).To(Succeed())
			Expect(repo.Create(buildRecord(2))).To(Succeed())
			Expect(repo.DeleteAll()).To(Succeed())
			records, err := repo.GetAll(QueryFilter{})
			Expect(err).NotTo(HaveOccurred())
			Expect(records).To(BeEmpty())
		})
	})
})

func buildRecord(id int64) Record{
  return Record{ID: id} // Only set the ID since the record isn't validated currently
}

