package main

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"go.mills.io/bitcask/v2"
)

const (
	dbPath = "./prol_test"
)

func insertData(db *bitcask.Bitcask, keySize, dataSize int) time.Duration {
	start := time.Now()

	for i := 0; i < dataSize; i++ {
		key := generateRandomString(keySize)
		value := generateRandomString(100) // Fixed data size for simplicity
		db.Put([]byte(key), []byte(value))
	}

	return time.Since(start)
}

func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	seed := rand.NewSource(time.Now().UnixNano())
	randGen := rand.New(seed)

	b := make([]byte, length)
	for i := range b {
		b[i] = charset[randGen.Intn(len(charset))]
	}

	return string(b)
}

func BenchmarkInsert(b *testing.B) {
	for keySize := 10; keySize <= 100; keySize += 10 {
		for _, dataSize := range []int{100, 1000, 10000} {
			db, err := bitcask.Open(dbPath)
			if err != nil {
				fmt.Println("Error opening Bitcask DB:", err)
				return
			}
			defer db.Close()

			b.Run(fmt.Sprintf("KeySize%dDataSize%d", keySize, dataSize), func(b *testing.B) {
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					insertData(db, keySize, dataSize)
				}
			})
		}
	}
}

func main() {
	// Run the benchmark
	testing.Benchmark(BenchmarkInsert)
}