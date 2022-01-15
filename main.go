package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
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

type GoogleMeeting struct {
	Title              string
	StartTime          time.Time
	ConferenceNumber   string
	ConferencePassword string
}

func getUpcomingGoogleMeeting() (gm GoogleMeeting, err error) {
	calendarService, err := newGoogleCalenderService()
	if err != nil {
		return gm, err
	}
	eventService := calendar.NewEventsService(calendarService)
	eventList := eventService.List("primary")
	eventList.SingleEvents(true)
	eventList.MaxResults(1)
	eventList.OrderBy("startTime")
	eventList.TimeMin(time.Now().Format(time.RFC3339))
	eventList.TimeMax(time.Now().Add(time.Minute * 30).Format(time.RFC3339))
	events, err := eventList.Do()
	if err != nil {
		return gm, fmt.Errorf("unable to find event")
	}
	eventOne := events.Items[0]
	gm.Title = eventOne.Summary
	meetingTime, _ := time.Parse(time.RFC3339, eventOne.Start.DateTime)
	gm.StartTime = meetingTime
	for _, entryPoint := range eventOne.ConferenceData.EntryPoints {
		if entryPoint.EntryPointType == "video" {
			gm.ConferenceNumber = entryPoint.MeetingCode
			gm.ConferencePassword = entryPoint.Passcode
			break
		}
	}
	return gm, nil

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

func generateZoomMtgLink(confNumber, confPassword string) string {
	zoomMtgLink := fmt.Sprintf("zoommtg://zoom.us/join?confno=%s&pwd=%s", confNumber, confPassword)
	return zoomMtgLink
}

func main() {
	meeting, err := getUpcomingGoogleMeeting()
	if err != nil {
		fmt.Errorf("error while getting the meeting")
		return
	}
	fmt.Printf("Here is your meeting title %s scheduled for %s \n", meeting.Title, meeting.StartTime.Format(time.Kitchen))
	fmt.Printf("Enter this passcode on zoom client %s \n", meeting.ConferencePassword)
	zoomLink := generateZoomMtgLink(meeting.ConferenceNumber, meeting.ConferencePassword)
	exec.Command("open", zoomLink).Run()

}
