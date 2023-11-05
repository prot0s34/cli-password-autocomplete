package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/go-resty/resty/v2"
)

const (
	unlockTimeout = 10 * time.Second
	bwHostnameCli = "localhost"
	bwPortCli     = "3000"
)

func main() {
	serveCmd := exec.Command("bw", "serve", "--port", bwPortCli, "--hostname", bwHostnameCli)
	serveURL := "http://" + bwHostnameCli + ":" + bwPortCli

	if err := serveCmd.Start(); err != nil {
		log.Fatalf("Failed to start 'bw serve': %v", err)
	}

	defer func() {
		if serveCmd.Process != nil {
			if err := serveCmd.Process.Signal(syscall.SIGTERM); err != nil {
				log.Printf("Error sending SIGTERM to 'bw serve': %v", err)
				if err := serveCmd.Process.Signal(syscall.SIGKILL); err != nil {
					log.Fatalf("Error sending SIGKILL to 'bw serve': %v", err)
				}
			}
		}
	}()

	startTime := time.Now()
	for {
		client := resty.New()
		unlockResponse, err := client.R().
			SetHeader("Content-Type", "application/json").
			SetBody(map[string]string{"password": os.Getenv("BITWARDEN_MASTER_PASSWORD")}).
			Post(serveURL + "/unlock")

		if err == nil && unlockResponse.StatusCode() == 200 {
			break
		}

		if time.Since(startTime) >= unlockTimeout {
			log.Fatalf("API is not available after %v seconds", unlockTimeout)
		}

		time.Sleep(1 * time.Second)
	}

	masterPassword := os.Getenv("BITWARDEN_MASTER_PASSWORD")

	if masterPassword == "" {
		log.Fatal("Bitwarden master password not found in the environment variable")
	}

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
	fmt.Println("------------")

	var response ListResponse
	err = json.Unmarshal([]byte(listResponse.Body()), &response)
	if err != nil {
		log.Fatalf("Error unmarshaling JSON: %v", err)
	}

	fmt.Println("Parsed Items:")
	for _, item := range response.Data.Items {
		fmt.Printf("Name: %s\nUsername: %s\nPassword: %s\n\n", item.Name, item.Login.Username, item.Login.Password)
	}
}
