package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/calypsodb/calypso"
)

func main() {
	file, err := os.Open("output.csv")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	db, err := calypso.Open("million_db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	reader := csv.NewReader(file)
	headers, err := reader.Read()
	if err != nil {
		log.Fatal(err)
	}
	if headers[0] != "key" || headers[1] != "value" {
		log.Fatal("CSV headers do not match expected format 'key,value'")
	}

	start := time.Now()

	count := 0
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		key := record[0]
		value := record[1]

		if err := db.Put([]byte(key), []byte(value)); err != nil {
			log.Fatal(err)
		}

		count++
		if count%100000 == 0 {
			fmt.Printf("Inserted %d records\n", count)
		}
	}

	elapsed := time.Since(start)
	fmt.Printf("Time taken to insert 1 million records: %s\n", elapsed)
}
