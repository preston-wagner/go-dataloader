package dataloader

import (
	"errors"
	"testing"
	"time"
)

func reverseString(str string) string {
	result := ""
	for _, v := range str {
		result = string(v) + result
	}
	return result
}

func alwaysSucceedGetter(input []string) (map[string]string, map[string]error) {
	result := map[string]string{}
	for _, value := range input {
		result[value] = reverseString(value)
	}
	return result, nil
}

func alwaysFailGetter(input []string) (map[string]string, map[string]error) {
	result := map[string]error{}
	for _, value := range input {
		result[value] = errors.New(reverseString(value))
	}
	return nil, result
}

func TestQueryBatcherSuccess(t *testing.T) {
	batcher := NewQueryBatcher(alwaysSucceedGetter, 1, 1)
	defer batcher.Close()

	key := "lorem"
	result, err := batcher.Load(key)
	if err != nil {
		t.Fatal(err)
	}
	if reverseString(result) != key {
		t.Fatal("QueryBatcher did not return the expected result for the query")
	}
}

func TestQueryBatcherFail(t *testing.T) {
	batcher := NewQueryBatcher(alwaysFailGetter, 1, 1)
	defer batcher.Close()

	key := "lorem"
	_, err := batcher.Load(key)
	if reverseString(err.Error()) != key {
		t.Fatal("QueryBatcher did not return the expected error for the query")
	}
}

func TestQueryBatcherSuccessMany(t *testing.T) {
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

	batcher := NewQueryBatcher(countCallsGetter, 1, 10)
	defer batcher.Close()

	maxCalls := 30
	for i := 1; i < maxCalls; i++ {
		go batcher.Load(i)
	}
	batcher.Load(maxCalls)

	time.Sleep(time.Second * 5)

	if calls > (maxCalls / 5) { // 6
		t.Fatal("QueryBatcher did not batch the queries, made", calls, "calls")
	}

	if keysCount != maxCalls {
		t.Fatal("QueryBatcher did not call the getter with all keys, used", keysCount, "keys")
	}
}

func TestQueryBatcherSuccessMultithread(t *testing.T) {
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

	batcher := NewQueryBatcher(countCallsGetter, 3, 10)
	defer batcher.Close()

	maxCalls := 50
	for i := 1; i < maxCalls; i++ {
		go batcher.Load(i)
	}
	batcher.Load(maxCalls)

	time.Sleep(time.Second * 5)

	if calls > (maxCalls / 4) {
		t.Fatal("QueryBatcher did not batch the queries, made", calls, "calls")
	}

	if keysCount != maxCalls {
		t.Fatal("QueryBatcher did not call the getter with all keys, used", keysCount, "keys")
	}
}
