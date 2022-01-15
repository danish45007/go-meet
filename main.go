package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

const (
	cacheTokenDir  = "/tmp/google_tokens"
	credentialPath = "credentials.json"
)

func getUpcomingGoogleMeeting() {
	calendarService, err := newGoogleCalenderService()
	if err != nil {
		fmt.Println(err)
		return
	}
	eventService := calendar.NewEventsService(calendarService)
	eventList := eventService.List("primary")
	eventList.SingleEvents(true)
	eventList.MaxResults(1)
	eventList.OrderBy("startTime")
	eventList.TimeMin(time.Now().Format(time.RFC3339))
	events, err := eventList.Do()
	if err != nil {
		fmt.Errorf("Unable to find event")
	}
	eventOne := events.Items[0]
	for _, entryPoint := range eventOne.ConferenceData.EntryPoints {
		if entryPoint.EntryPointType == "video" {
			fmt.Println(entryPoint.Uri)
		}
	}

}

func newGoogleCalenderService() (*calendar.Service, error) {
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
	ctx := context.Background()
	calendarService, err := calendar.NewService(ctx, option.WithTokenSource(config.TokenSource(ctx, token)))
	if err != nil {
		return nil, err
	}
	return calendarService, nil
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
