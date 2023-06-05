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

// sendHTTPRequest sends an HTTP POST request with the given value
func sendHTTPRequest(address, value string) error {
	formData := url.Values{}
	formData.Set("content", value)

	resp, err := http.PostForm(address, formData)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check the response status code and handle accordingly
	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	fmt.Println("Request sent with value:", value)
	return nil
}
