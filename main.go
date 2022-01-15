package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
)

const (
	cacheTokenDir  = "/tmp/google_tokens"
	credentialPath = "credentials.json"
)

func getUpcomingGoogleMeeting() {
	client, err := newGoogleClient()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("client--->", client)
}

func newGoogleClient() (*http.Client, error) {
	credByte, err := ioutil.ReadFile(credentialPath)
	if err != nil {
		return nil, err
	}
	config, err := google.ConfigFromJSON(credByte, calendar.CalendarReadonlyScope)
	if err != nil {
		return nil, err
	}
	token, err := getToken(config)
	if err != nil {
		return nil, err
	}
	return config.Client(context.Background(), token), nil
}

func getToken(config *oauth2.Config) (*oauth2.Token, error) {
	cachedToken, err := getCachedToken()
	if err != nil {
		return generateNewToken(config)
	}
	return cachedToken, nil
}

func getCachedToken() (*oauth2.Token, error) {
	tokenByte, err := ioutil.ReadFile(cacheTokenDir)
	if err != nil {
		return &oauth2.Token{}, err
	}
	if len(tokenByte) == 0 {
		return &oauth2.Token{}, fmt.Errorf("token file empty")
	}
	var token *oauth2.Token
	err = json.Unmarshal(tokenByte, &token)
	if err != nil {
		return &oauth2.Token{}, err
	}
	return token, nil
}

func cacheToken(token *oauth2.Token) error {
	tokenByte, err := json.Marshal(token)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(cacheTokenDir, tokenByte, 0644)
}

// TODO: remove print statements and add logging package
func generateNewToken(config *oauth2.Config) (*oauth2.Token, error) {
	fmt.Println("fetching new google token")
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the authorization code: \n%v\n", authURL)
	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	token, err := config.Exchange(context.TODO(), authCode)
	fmt.Println(token)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	cacheToken(token)
	return token, nil

}

func main() {
	getUpcomingGoogleMeeting()
}
