package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"github.com/go-resty/resty/v2"
)

const (
	unlockTimeout = 10 * time.Second
	bwHostnameCli = "localhost"
	bwPortCli     = "3000"
)

var targetUsername string
var targetPassword string

func main() {

	if len(os.Args) != 2 {
		log.Fatal("Usage: sshbw user@host")
	}

	arg := os.Args[1]
	argParts := strings.Split(arg, "@")
	if len(argParts) != 2 {
		log.Fatal("Invalid argument format. Use 'user@host'.")
	}

	user := argParts[0]
	host := argParts[1]

	if user == "" || host == "" {
		log.Fatal("Both user and host must be provided and not empty.")
	}

	killCmd := exec.Command("pkill", "bw")
	killCmd.Stdout = os.Stdout
	killCmd.Stderr = os.Stderr
	if err := killCmd.Run(); err != nil {
		log.Printf("Failed to kill existing Bitwarden server: %v", err)
	}

	syncCmd := exec.Command("bw", "sync")
	syncCmd.Stdout = os.Stdout
	syncCmd.Stderr = os.Stderr

	if err := syncCmd.Run(); err != nil {
		log.Fatalf("Failed to run 'bw sync': %v", err)
	}

	serveCmd := exec.Command("bw", "serve", "--port", bwPortCli, "--hostname", bwHostnameCli)
	serveURL := "http://" + bwHostnameCli + ":" + bwPortCli
	masterPassword := os.Getenv("BITWARDEN_MASTER_PASSWORD")

	if masterPassword == "" {
		log.Fatal("Bitwarden master password not found in the environment variable")
	}

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

	fmt.Println("user & hostname", user, host)

	for _, item := range response.Data.Items {
		if item.Login.Username == user {
			targetUsername = item.Login.Username
			targetPassword = item.Login.Password
			break
		}
	}

	if targetUsername != "" {
		fmt.Printf("Matched User: %s\nPassword: %s\n", targetUsername, targetPassword)
	} else {
		fmt.Println("User not found in Bitwarden items.")
	}

	if targetUsername != "" {
		fmt.Printf("Matched User: %s\nPassword: %s\n", targetUsername, targetPassword)

		sshCmd := exec.Command("sshpass", "-p", targetPassword, "ssh", "-o", "StrictHostKeyChecking=no", targetUsername+"@"+host)
		sshCmd.Stdout = os.Stdout
		sshCmd.Stderr = os.Stderr
		sshCmd.Stdin = os.Stdin

		if err := sshCmd.Run(); err != nil {
			log.Fatalf("Failed to execute SSH command: %v", err)
		}
	} else {
		fmt.Println("User not found in Bitwarden items.")
	}

}
