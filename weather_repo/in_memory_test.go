package weather_repo_test

import (
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"go.uber.org/goleak"
	"sync"
	"testing"

	. "github.com/matt-royal/carvel-weather-app/weather_repo"
)

func TestInMemory(t *testing.T) {
	spec.Run(t, "WeatherRepo", func(t *testing.T, when spec.G, it spec.S) {
		defer goleak.VerifyNone(t)
		var repo *InMemoryRepo
		g := NewGomegaWithT(t)

		it.Before(func() {
			repo = new(InMemoryRepo)
		})

		when("Create", func() {
			it("can handles concurrency", func() {
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
							g.Expect(err).NotTo(HaveOccurred())
						}
						wg.Done()
					}()
				}

				wg.Wait()

				createdRecords, err := repo.GetAll(QueryFilter{})
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(createdRecords).To(HaveLen(10))

				var ids []int64
				for _, record := range createdRecords {
					ids = append(ids, record.ID)
				}
				g.Expect(ids).To(Equal([]int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}))
			})

			when("a record already exists with the same ID", func() {
				it.Before(func() {
					g.Expect(repo.Create(buildRecord(1))).To(Succeed())
				})

				it("returns a ErrorDuplicateID", func() {
					err := repo.Create(buildRecord(1))
					g.Expect(err).To(MatchError(ErrorDuplicateID(1)))
				})

				it("doesn't create the record", func() {
					_ = repo.Create(buildRecord(1))
					records, err := repo.GetAll(QueryFilter{})
					g.Expect(err).NotTo(HaveOccurred())
					g.Expect(records).To(HaveLen(1))
				})
			})
		})

		when("GetAll", func() {
			it("returns the item in GetAll()", func() {
				record := buildRecord(1)
				g.Expect(repo.Create(record)).To(Succeed())
				records, err := repo.GetAll(QueryFilter{})
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(records).To(Equal([]Record{record}))
			})

			it("sorts records by ID in GetAll()", func() {
				record1 := buildRecord(1)
				record2 := buildRecord(2)
				record3 := buildRecord(3)
				g.Expect(repo.Create(record2)).To(Succeed())
				g.Expect(repo.Create(record3)).To(Succeed())
				g.Expect(repo.Create(record1)).To(Succeed())
				records, err := repo.GetAll(QueryFilter{})
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(records).To(Equal([]Record{record1, record2, record3}))
			})

			when("a date filter is given", func() {
				it.Pend("returns only matching records", func() {
					// TODO
				})
			})
		})

		when("DeleteAll", func() {
			it("deletes all records", func() {
				g.Expect(repo.Create(buildRecord(1))).To(Succeed())
				g.Expect(repo.Create(buildRecord(2))).To(Succeed())
				g.Expect(repo.DeleteAll()).To(Succeed())
				records, err := repo.GetAll(QueryFilter{})
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(records).To(BeEmpty())
			})
		})
	})
}

func buildRecord(id int64) Record {
	return Record{ID: id} // Only set the ID since the record isn't validated currently
}
