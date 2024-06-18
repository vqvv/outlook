package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/fatih/color"
)

type Payload struct {
	IncludeSuggestions bool   `json:"includeSuggestions"`
	SignInName         string `json:"signInName"`
	Uiflvr             int    `json:"uiflvr"`
	Scid               int    `json:"scid"`
	Uaid               string `json:"uaid"`
	Hpgid              int    `json:"hpgid"`
}

type Response struct {
	IsAvailable bool `json:"isAvailable"`
}

func checkDomain(email string) (bool, bool) {
	url := "https://signup.live.com/API/CheckAvailableSigninNames"

	payload := Payload{
		IncludeSuggestions: true,
		SignInName:         email,
		Uiflvr:             1001,
		Scid:               100118,
		Uaid:               "2b7f382ac4e44b1db3ed5e4be959802d",
		Hpgid:              200225,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		fmt.Printf("Error creating JSON payload: %v\n", err)
		return false, false
	}

	client := &http.Client{}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		return false, false
	}

	cookies := "Your Cookies Here"
        canary := "Your  Canary Thingy Here"
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.36")
	req.Header.Set("Canary", canary)
	req.Header.Set("Cookie", cookies)

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error making request: %v\n", err)
		return false, false
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response body: %v\n", err)
		return false, false
	}

	var response Response
	err = json.Unmarshal(body, &response)
	if err != nil {
		fmt.Printf("Error parsing JSON response: %v\n", err)
		return false, false
	}

	return response.IsAvailable, !response.IsAvailable
}

func main() {
	fmt.Println("Reading usernames from emails.txt...")

	filePath := "emails.txt"
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		fmt.Printf("Error reading file %s: %v\n", filePath, err)
		os.Exit(1)
	}

	usernames := strings.Split(string(content), "\n")

	availableFile, err := os.Create("available.txt")
	if err != nil {
		fmt.Printf("Error creating file: %v\n", err)
		os.Exit(1)
	}
	defer availableFile.Close()

	notAvailableFile, err := os.Create("taken.txt")
	if err != nil {
		fmt.Printf("Error creating file: %v\n", err)
		os.Exit(1)
	}
	defer notAvailableFile.Close()

	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()

	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Choose the domain for the email addresses:")
	fmt.Println("1. hotmail.com")
	fmt.Println("2. outlook.com")
	fmt.Print("Enter 1 or 2: ")
	domainChoice, _ := reader.ReadString('\n')
	domainChoice = strings.TrimSpace(domainChoice)

	var domain string
	switch domainChoice {
	case "1":
		domain = "hotmail.com"
	case "2":
		domain = "outlook.com"
	default:
		fmt.Println("Invalid choice. Exiting.")
		os.Exit(1)
	}

	usernameChan := make(chan string, 10)
	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for username := range usernameChan {
				email := fmt.Sprintf("%s@%s", username, domain)
				isAvailable, isNotAvailable := checkDomain(email)
				if isAvailable {
					fmt.Printf("[+] AVAILABLE %s\n", green(email))
					fmt.Fprintln(availableFile, email)
				} else if isNotAvailable {
					fmt.Printf("[-] TAKEN %s\n", red(email))
					fmt.Fprintln(notAvailableFile, email)
				}
			}
		}()
	}
	for _, username := range usernames {
		username = strings.TrimSpace(username)
		if username != "" {
			usernameChan <- username
		}
	}
	close(usernameChan)
	wg.Wait()
}
