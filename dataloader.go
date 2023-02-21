package dataloader

import (
	"sync"
	"time"

	"github.com/preston-wagner/unicycle"
)

type batch[KEY_TYPE comparable, VALUE_TYPE any] struct {
	queries map[KEY_TYPE]unicycle.Promise[VALUE_TYPE]
	start   time.Time
}

// A getter function accepts a list of de-duplicated keys, and returns a pair of maps from keys to values (for successful lookups) and keys to errors (for unsuccessful lookups)
type Getter[KEY_TYPE comparable, VALUE_TYPE any] func([]KEY_TYPE) (map[KEY_TYPE]VALUE_TYPE, map[KEY_TYPE]error)

type DataLoader[KEY_TYPE comparable, VALUE_TYPE any] struct {
	channelCache map[KEY_TYPE][]chan VALUE_TYPE
	valueCache   map[KEY_TYPE]VALUE_TYPE
	lock         *sync.RWMutex

	getter       Getter[KEY_TYPE, VALUE_TYPE]
	maxBatchSize int
	maxBatchWait time.Duration
}

func NewDataLoader[KEY_TYPE comparable, VALUE_TYPE any](getter Getter[KEY_TYPE, VALUE_TYPE]) *DataLoader[KEY_TYPE, VALUE_TYPE] {
	dataloader := DataLoader[KEY_TYPE, VALUE_TYPE]{
		getter: getter,
	}
	return &dataloader
}

func (dataloader *DataLoader[KEY_TYPE, VALUE_TYPE]) QueueRequest(key KEY_TYPE) VALUE_TYPE {

	// for {
	// 	select {
	// 	case <-ticker.C:
	// 		go wrapped()
	// 	case <-ctx.Done():
	// 		return
	// 	}
	// }
}

func (dataloader *DataLoader[KEY_TYPE, VALUE_TYPE]) Load(key KEY_TYPE) (VALUE_TYPE, error) {

	// for {
	// 	select {
	// 	case <-ticker.C:
	// 		go wrapped()
	// 	case <-ctx.Done():
	// 		return
	// 	}
	// }
}

type DataLoaderCached[KEY_TYPE comparable, VALUE_TYPE any] struct {
	loader     DataLoader[KEY_TYPE, VALUE_TYPE]
	valueCache map[KEY_TYPE]VALUE_TYPE
	lock       *sync.RWMutex
}

func (dataLoaderCached *DataLoaderCached[KEY_TYPE, VALUE_TYPE]) Load(key KEY_TYPE) (VALUE_TYPE, error) {
	dataLoaderCached.lock.RLock()
	value, ok := dataLoaderCached.valueCache[key]
	dataLoaderCached.lock.RUnlock()
	if ok {
		return value, nil
	}
	value, err := dataLoaderCached.loader.Load(key)
	if err == nil {
		dataLoaderCached.lock.Lock()
		dataLoaderCached.valueCache[key] = value
		dataLoaderCached.lock.Unlock()
	}
	return value, err
}
