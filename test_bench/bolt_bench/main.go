package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/boltdb/bolt"
)

func main() {
	// Open the BoltDB database
	db, err := bolt.Open("mydb.db", 0666, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Start measuring time for writes
	startWrite := time.Now()
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("key%d", i)
		value := fmt.Sprintf("value%d", rand.Intn(1000)) // Random value for each write
		err := db.Update(func(tx *bolt.Tx) error {
			bucket, err := tx.CreateBucketIfNotExists([]byte("MyBucket"))
			if err != nil {
				return err
			}
			return bucket.Put([]byte(key), []byte(value))
		})
		if err != nil {
			log.Fatalf("Write operation failed: %v", err)
		}
	}
	endWrite := time.Now()

	// Start measuring time for reads
	startRead := time.Now()
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("key%d", i)
		err := db.View(func(tx *bolt.Tx) error {
			bucket := tx.Bucket([]byte("MyBucket"))
			if bucket == nil {
				return fmt.Errorf("Bucket not found")
			}
			value := bucket.Get([]byte(key))
			if value == nil {
				return fmt.Errorf("Key not found")
			}
			return nil
		})
		if err != nil {
			log.Fatalf("Read operation failed: %v", err)
		}
	}
	endRead := time.Now()

	// Calculate and print the time taken for writes and reads
	writeDuration := endWrite.Sub(startWrite)
	readDuration := endRead.Sub(startRead)
	fmt.Printf("Time taken for 1000 writes: %v\n", writeDuration)
	fmt.Printf("Time taken for 1000 reads: %v\n", readDuration)
}
