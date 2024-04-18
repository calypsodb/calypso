package main

import (
	"fmt"
	"github.com/calypsodb/calypso"
	"log"
	"time"
)

func main() {
	// Open or create a new Bitcask database
	db, err := bitcask.Open("./db_test")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Benchmark insert time
	startInsert := time.Now()
	for i := 0; i < 1000; i++ {
		key := []byte(fmt.Sprintf("key%d", i))
		value := []byte(fmt.Sprintf("value%d", i))
		if err := db.Put(key, value); err != nil {
			log.Fatal(err)
		}
	}
	insertTime := time.Since(startInsert)
	fmt.Printf("Insertion time for 1000 records: %s\n", insertTime)

	// Benchmark read time
	startRead := time.Now()
	for i := 0; i < 1000; i++ {
		key := []byte(fmt.Sprintf("key%d", i))
		value, err := db.Get(key)
		if err != nil {
			log.Fatal(err)
		}
		_ = value // Ignore value for benchmarking purposes
	}
	readTime := time.Since(startRead)
	fmt.Printf("Read time for 1000 records: %s\n", readTime)
}