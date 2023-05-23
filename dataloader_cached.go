package dataloader

// import (
// 	"sync"
// )

// type DataLoaderCached[KEY_TYPE comparable, VALUE_TYPE any] struct {
// 	loader     DataLoader[KEY_TYPE, VALUE_TYPE]
// 	valueCache map[KEY_TYPE]VALUE_TYPE
// 	lock       *sync.RWMutex
// }

// func (dataLoaderCached *DataLoaderCached[KEY_TYPE, VALUE_TYPE]) Load(key KEY_TYPE) (VALUE_TYPE, error) {
// 	dataLoaderCached.lock.RLock()
// 	value, ok := dataLoaderCached.valueCache[key]
// 	dataLoaderCached.lock.RUnlock()
// 	if ok {
// 		return value, nil
// 	}
// 	value, err := dataLoaderCached.loader.Load(key)
// 	if err == nil {
// 		dataLoaderCached.lock.Lock()
// 		dataLoaderCached.valueCache[key] = value
// 		dataLoaderCached.lock.Unlock()
// 	}
// 	return value, err
// }
