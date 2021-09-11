package weather_repo

import "sync"

type InMemoryRepo struct {
	mutex   sync.RWMutex
	records []Record
}

func (w *InMemoryRepo) Create(record Record) error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	for i, r := range w.records {
		if r.ID > record.ID {
			// keep the records in asc order of ID
			w.records = append(w.records[:i+1], w.records[i:]...)
			w.records[i] = record
			return nil
		} else if r.ID == record.ID {
			return ErrorDuplicateID(r.ID)
		}
	}

	// the record has the highest ID, put it at the end
	w.records = append(w.records, record)
	return nil
}

func (w *InMemoryRepo) GetAll(filter QueryFilter) ([]Record, error) {
	w.mutex.RLock()
	defer w.mutex.RUnlock()

	records := w.records

	if filter.Date != "" {
		var filtered []Record
		for _, record := range records {
			if record.DateStr == filter.Date {
				filtered = append(filtered, record)
			}
		}
		records = filtered
	}
	return records, nil
}

func (w *InMemoryRepo) DeleteAll() error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	w.records = []Record{}
	return nil
}
