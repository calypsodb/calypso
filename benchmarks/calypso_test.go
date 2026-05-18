package test_bench_test

import (
	"os"
	"testing"

	"github.com/calypsodb/calypso"
)

func TestNew(t *testing.T) {
	dbPath := "./testdata"
	defer os.RemoveAll(dbPath)

	db, err := calypso.Open(dbPath)
	if err != nil {
		t.Fatalf("New: unexpected error opening database: %v", err)
	}
	defer db.Close()

	if db == nil {
		t.Errorf("New: database is nil")
	}
}

func TestPutGet(t *testing.T) {
	dbPath := "./testdata"
	defer os.RemoveAll(dbPath)

	db, err := calypso.Open(dbPath)
	if err != nil {
		t.Fatalf("PutGet: unexpected error opening database: %v", err)
	}
	defer db.Close()

	key := "testKey"
	value := []byte("testValue")

	err = db.Put([]byte(key), value)
	if err != nil {
		t.Fatalf("PutGet: unexpected error putting key-value pair: %v", err)
	}

	retrievedValue, err := db.Get([]byte(key))
	if err != nil {
		t.Fatalf("PutGet: unexpected error getting value for key: %v", err)
	}
	if string(retrievedValue) != string(value) {
		t.Errorf("PutGet: retrieved value %s does not match expected value %s", retrievedValue, value)
	}
}
