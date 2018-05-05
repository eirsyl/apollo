package test_cases

import (
	"errors"
	"fmt"

	"math/rand"

	"sync"

	"github.com/go-redis/redis"
	log "github.com/sirupsen/logrus"
)

type WriteData struct{}

func NewWriteData() (*WriteData, error) {
	return &WriteData{}, nil
}

func (wd *WriteData) GetName() string {
	return "write-data"
}

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
				_, err := client.Set(key, RandStringBytesMask(8), 0).Result()
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
		n -= 1
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

func RandStringBytesMask(n int) string {
	b := make([]byte, n)
	for i := 0; i < n; {
		if idx := int(rand.Int63() & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i++
		}
	}
	return string(b)
}
