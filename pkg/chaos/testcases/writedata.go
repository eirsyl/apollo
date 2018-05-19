package testcases

import (
	"errors"
	"fmt"

	"crypto/rand"

	"sync"

	"math"
	"math/big"

	"github.com/go-redis/redis"
	log "github.com/sirupsen/logrus"
)

// WriteData represents the data writer used before test cases
type WriteData struct{}

// NewWriteData creates a new data writer
func NewWriteData() (*WriteData, error) {
	return &WriteData{}, nil
}

// GetName returns the name of the write data task
func (wd *WriteData) GetName() string {
	return "write-data"
}

// Run performs the data writing to the cluster
func (wd *WriteData) Run(args []string) error {
	if len(args) < 1 {
		return errors.New("please provide a server url as an argument")
	}

	client := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs: args,
	})

	var keys = make(chan string, 10000000)
	var wg = sync.WaitGroup{}
	workers := 1000

	// Run worker pool
	for i := workers; i > 0; i-- {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for key := range keys {
				_, err := client.Set(key, randStringBytesMask(8), 0).Result()
				if err != nil {
					log.Error(err)
				}
			}
		}()
	}

	n := 10000000
	for n > 0 {
		key := fmt.Sprintf("test-case-%d", n)
		keys <- key
		n--
	}
	log.Info("Done with key buffering")

	close(keys)
	wg.Wait()
	return nil
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
)

func randStringBytesMask(n int) string {
	b := make([]byte, n)
	for i := 0; i < n; {
		randInt, err := rand.Int(rand.Reader, big.NewInt(math.MaxInt64))
		if err != nil {
			panic(err)
		}
		if idx := int(randInt.Int64() & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i++
		}
	}
	return string(b)
}
