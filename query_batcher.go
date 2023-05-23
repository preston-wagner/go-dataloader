package dataloader

import (
	"github.com/preston-wagner/unicycle"
)

// A getter function accepts a list of de-duplicated keys, and returns a pair of maps from keys to values (for successful lookups) and keys to errors (for unsuccessful lookups)
type Getter[KEY_TYPE comparable, VALUE_TYPE any] func([]KEY_TYPE) (map[KEY_TYPE]VALUE_TYPE, map[KEY_TYPE]error)

type QueryBatcher[KEY_TYPE comparable, VALUE_TYPE any] struct {
	incoming chan query[KEY_TYPE, VALUE_TYPE]
	ready    chan batch[KEY_TYPE, VALUE_TYPE]
}

func NewQueryBatcher[KEY_TYPE comparable, VALUE_TYPE any](getter Getter[KEY_TYPE, VALUE_TYPE], maxConcurrentBatches, maxBatchSize int) *QueryBatcher[KEY_TYPE, VALUE_TYPE] {
	batcher := QueryBatcher[KEY_TYPE, VALUE_TYPE]{
		incoming: make(chan query[KEY_TYPE, VALUE_TYPE]),
		ready:    make(chan batch[KEY_TYPE, VALUE_TYPE]),
	}
	go batcher.batchRequests(maxBatchSize)
	go batcher.makeRequests(getter, maxConcurrentBatches)
	return &batcher
}

func (batcher *QueryBatcher[KEY_TYPE, VALUE_TYPE]) Load(key KEY_TYPE) (VALUE_TYPE, error) {
	promise := unicycle.NewPromise[VALUE_TYPE]()
	batcher.incoming <- query[KEY_TYPE, VALUE_TYPE]{
		key:     key,
		promise: promise,
	}
	return promise.Await()
}

func (batcher *QueryBatcher[KEY_TYPE, VALUE_TYPE]) batchRequests(maxBatchSize int) {
	if maxBatchSize == 0 {
		panic("maxBatchSize must be > 0!")
	}
	pendingBatch := batch[KEY_TYPE, VALUE_TYPE]{}
	for {
		if len(pendingBatch) == 0 {
			// if current batch is empty, just wait on new queries
			incomingQuery := <-batcher.incoming
			pendingBatch.addToBatch(incomingQuery)
		} else if len(pendingBatch) < maxBatchSize {
			// add new queries to pending or send pending to be executed as available
			select {
			case incomingQuery := <-batcher.incoming:
				pendingBatch.addToBatch(incomingQuery)
			case batcher.ready <- pendingBatch:
				pendingBatch = batch[KEY_TYPE, VALUE_TYPE]{}
			}
		} else {
			// if current batch is at capacity, just wait for a current query to finish before starting a new one
			batcher.ready <- pendingBatch
			pendingBatch = batch[KEY_TYPE, VALUE_TYPE]{}
		}
	}
}

func (batcher *QueryBatcher[KEY_TYPE, VALUE_TYPE]) makeRequests(getter Getter[KEY_TYPE, VALUE_TYPE], maxConcurrentBatches int) {
	unicycle.ChannelForEachMultithread(batcher.ready, func(btch batch[KEY_TYPE, VALUE_TYPE]) {
		defer func() {
			if r := recover(); r != nil {
				btch.rejectAll(GetterPanicError{recovered: r})
			}
		}()
		btch.resolveAll(getter(unicycle.Keys(btch)))
	}, maxConcurrentBatches)
}

// func (batcher *QueryBatcher[KEY_TYPE, VALUE_TYPE]) Close() {
// 	close(batcher.incoming)
// 	close(batcher.ready)
// }
