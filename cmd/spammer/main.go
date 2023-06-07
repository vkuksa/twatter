package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"time"
)

func main() {
	destination := flag.String("destination", "http://localhost:9876/add", "Address to send HTTP requests")
	pace := flag.Duration("pace", time.Second, "Time duration between sending requests")
	flag.Parse()

	for {
		content := generateRandomString()

		err := sendHTTPRequest(*destination, content)
		if err != nil {
			log.Fatal(err)
		}

		time.Sleep(*pace)
	}
}

// generateRandomString generates a random string of length 10
func generateRandomString() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))

	b := make([]byte, 10)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func sendHTTPRequest(address, value string) error {
	formData := url.Values{}
	formData.Set("content", value)

	retryCount := 0
	maxRetries := 5

	for retryCount < maxRetries {
		resp, err := http.PostForm(address, formData)
		if err != nil {
			log.Printf("Request failed: %v", err)
		} else if resp.StatusCode != http.StatusAccepted {
			log.Printf("Unexpected status code: %d", resp.StatusCode)
		} else {
			log.Println("Request sent with value:", value)
			return nil
		}

		retryCount++
		log.Printf("Retrying in 5 seconds (retry %d/%d)...", retryCount, maxRetries)
		time.Sleep(5 * time.Second)
	}

	return fmt.Errorf("failed to send request after %d retries", maxRetries)
}
