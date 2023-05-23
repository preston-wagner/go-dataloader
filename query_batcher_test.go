package dataloader

import (
	"errors"
	"testing"
	// "time"
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
	batcher := NewQueryBatcher(alwaysSucceedGetter, 3, 10)
	// defer batcher.Close()

	key := "lorem"
	result, err := batcher.Load(key)
	if err != nil {
		t.Fatal(err)
	}
	if reverseString(result) != key {
		t.Fatal("batcher did not return the expected result for the query")
	}
}

func TestQueryBatcherFail(t *testing.T) {
	batcher := NewQueryBatcher(alwaysFailGetter, 3, 10)
	// defer batcher.Close()

	key := "lorem"
	_, err := batcher.Load(key)
	if reverseString(err.Error()) != key {
		t.Fatal("batcher did not return the expected error for the query")
	}
}
