package main

import (
	"encoding/csv"
	"fmt"
	"math/rand"
	"os"
	"time"
)

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func main() {
	file, err := os.Create("output.csv")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()
	headers := []string{"key", "value"}
	if err := writer.Write(headers); err != nil {
		fmt.Println("Error writing headers:", err)
		return
	}

	for i := 0; i < 1000000; i++ {
		keyLength := rand.Intn(36) + 15
		valueLength := rand.Intn(271) + 30

		key := randomString(keyLength)
		value := randomString(valueLength)

		record := []string{key, value}
		if err := writer.Write(record); err != nil {
			fmt.Println("Error writing record:", err)
			return
		}
	}

	fmt.Println("CSV file successfully created.")
}
