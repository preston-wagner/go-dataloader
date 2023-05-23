package dataloader

import (
	"testing"
	"time"
)

func TestDataLoaderSuccess(t *testing.T) {
	batcher := NewDataLoader(alwaysSucceedGetter, 1, 1)
	defer batcher.Close()

	key := "lorem"
	result, err := batcher.Load(key)
	if err != nil {
		t.Fatal(err)
	}
	if reverseString(result) != key {
		t.Fatal("DataLoader did not return the expected result for the query")
	}
}

func TestDataLoaderFail(t *testing.T) {
	batcher := NewDataLoader(alwaysFailGetter, 1, 1)
	defer batcher.Close()

	key := "lorem"
	_, err := batcher.Load(key)
	if reverseString(err.Error()) != key {
		t.Fatal("DataLoader did not return the expected error for the query")
	}
}

func TestDataLoaderSuccessMany(t *testing.T) {
	calls := 0
	keysCount := 0
	countCallsGetter := func(input []int) (map[int]int, map[int]error) {
		calls += 1

		time.Sleep(time.Second)

		result := map[int]int{}
		for _, value := range input {
			keysCount++
			result[value] = -value
		}
		return result, nil
	}

	batcher := NewDataLoader(countCallsGetter, 1, 10)
	defer batcher.Close()

	maxCalls := 30
	for i := 1; i < maxCalls; i++ {
		go batcher.Load(i)
	}
	batcher.Load(maxCalls)

	time.Sleep(time.Second * 5)

	// due to the intricacies of goroutines and channels, as well as the speed of the actual hardware, the theoretical best-case performance of 4 calls may not always be reached
	if calls > (maxCalls / 5) { // 6
		t.Fatal("DataLoader did not batch the queries, made", calls, "calls")
	}

	if keysCount != maxCalls {
		t.Fatal("DataLoader did not call the getter with all keys, used", keysCount, "keys")
	}
}

func TestDataLoaderSuccessMultithread(t *testing.T) {
	calls := 0
	keysCount := 0
	countCallsGetter := func(input []int) (map[int]int, map[int]error) {
		calls += 1

		time.Sleep(time.Second)

		result := map[int]int{}
		for _, value := range input {
			keysCount++
			result[value] = -value
		}
		return result, nil
	}

	batcher := NewDataLoader(countCallsGetter, 3, 10)
	defer batcher.Close()

	maxCalls := 50
	for i := 1; i < maxCalls; i++ {
		go batcher.Load(i)
	}
	batcher.Load(maxCalls)

	time.Sleep(time.Second * 5)

	if calls > (maxCalls / 2) {
		t.Fatal("DataLoader did not batch the queries, made", calls, "calls")
	}

	if keysCount != maxCalls {
		t.Fatal("DataLoader did not call the getter with all keys, used", keysCount, "keys")
	}
}
