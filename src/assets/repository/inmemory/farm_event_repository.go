package inmemory

import (
	"github.com/Tanibox/tania-server/src/assets/repository"
	"github.com/Tanibox/tania-server/src/assets/storage"
	uuid "github.com/satori/go.uuid"
)

type FarmEventRepositoryInMemory struct {
	Storage *storage.FarmEventStorage
}

func NewFarmEventRepositoryInMemory(s *storage.FarmEventStorage) repository.FarmEventRepository {
	return &FarmEventRepositoryInMemory{Storage: s}
}

// Save is to save
func (f *FarmEventRepositoryInMemory) Save(uid uuid.UUID, latestVersion int, events []interface{}) <-chan error {
	result := make(chan error)

	go func() {
		f.Storage.Lock.Lock()
		defer f.Storage.Lock.Unlock()

		for _, v := range events {
			latestVersion++
			f.Storage.FarmEvents = append(f.Storage.FarmEvents, storage.FarmEvent{
				FarmUID: uid,
				Version: latestVersion,
				Event:   v,
			})
		}

		result <- nil

		close(result)
	}()

	return result
}
