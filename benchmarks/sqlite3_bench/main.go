package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// Open SQLite3 database
	log.Println("Starting...")
	db, err := sql.Open("sqlite3", "./test.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create a test table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS test (
			id INTEGER PRIMARY KEY,
			key TEXT,
			value TEXT
		)
	`)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Created table")
	// Benchmark write time
	startWrite := time.Now()
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("key%d", i)
		value := fmt.Sprintf("value%d", i)
		_, err := db.Exec("INSERT INTO test(key, value) VALUES (?, ?)", key, value)
		if err != nil {
			log.Fatal(err)
		}
	}
	writeTime := time.Since(startWrite)
	fmt.Printf("Write time for 1000 records: %s\n", writeTime)

	// Benchmark read time
	startRead := time.Now()
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("key%d", i)
		var value string
		err := db.QueryRow("SELECT value FROM test WHERE key = ?", key).Scan(&value)
		if err != nil {
			log.Fatal(err)
		}
	}
	readTime := time.Since(startRead)
	fmt.Printf("Read time for 1000 records: %s\n", readTime)
}
