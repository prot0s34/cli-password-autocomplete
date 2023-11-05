package main

import (
	"fmt"
	"log"
	"os"

	"github.com/go-resty/resty/v2"
)

func main() {
	// Define the Bitwarden CLI serve URL
	serveURL := "http://localhost:3000"

	// Replace 'YOUR_MASTER_PASSWORD' with your Bitwarden master password
	masterPassword := os.Getenv("BITWARDEN_MASTER_PASSWORD")

	if masterPassword == "" {
		log.Fatal("Bitwarden master password not found in the environment variable")
	}

	// Unlock the Bitwarden vault
	unlockData := map[string]string{
		"password": masterPassword,
	}

	client := resty.New()
	unlockResponse, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(unlockData).
		Post(serveURL + "/unlock")

	if err != nil {
		log.Fatalf("Error during unlocking: %v", err)
	}

	sessionKey := unlockResponse.String()
	os.Setenv("BW_SESSION", sessionKey)

	listResponse, err := client.R().
		SetHeader("Authorization", "Bearer "+sessionKey).
		Get(serveURL + "/list/object/items")

	if err != nil {
		log.Fatalf("Error retrieving item list: %v", err)
	}

	if listResponse.StatusCode() != 200 {
		log.Fatalf("Item list retrieval failed with status code: %d", listResponse.StatusCode())
	}

	fmt.Println("Raw Response:")
	fmt.Println(listResponse)

}
