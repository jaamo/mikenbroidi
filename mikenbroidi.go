package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/nlopes/slack"
	strava "github.com/strava/go.strava"
)

type stravaStruct struct {
	clubID      int64
	accessToken string
	client      *strava.Client
	clubService *strava.ClubsService
}

type slackStruct struct {
	token    string
	channel  string
	username string
	icon     string
	client   *slack.Client
}

func main() {

	var lastTimestamp time.Time
	var strv stravaStruct
	var slck slackStruct

	// Read settings from environment variables.
	strv.clubID, _ = strconv.ParseInt(os.Getenv("STRAVA_CLUBID"), 10, 32)
	strv.accessToken = os.Getenv("STRAVA_ACCESS_TOKEN")
	slck.token = os.Getenv("SLACK_ACCESS_TOKEN")
	slck.channel = os.Getenv("SLACK_CHANNEL")
	slck.username = os.Getenv("SLACK_USERNAME")
	slck.icon = os.Getenv("SLACK_ICON")

	// Check that variables exists.
	if strv.clubID == 0 || len(strv.accessToken) == 0 || len(slck.token) == 0 || len(slck.channel) == 0 || len(slck.username) == 0 || len(slck.icon) == 0 {
		fmt.Println("Please define following environment variables: STRAVA_CLUBID, STRAVA_ACCESS_TOKEN, SLACK_ACCESS_TOKEN, SLACK_CHANNEL, SLACK_USERNAME, SLACK_ICON")
	}

	// Create Strava client.
	strv.client = strava.NewClient(strv.accessToken)

	// Create slack client.
	slck.client = slack.New(slck.token)

	// for making things recurring
	ticker := time.NewTicker(time.Duration(60000) * time.Millisecond)

	// Start watching for new Strava activities.
	for {

		fmt.Println("Get new rides.")

		// Get new rides.
		newActivities := getNewActivities(strv, &lastTimestamp)

		// List 'em all!
		for _, activity := range newActivities {
			print(activity.Name + "\n")
			fmt.Println(activity.StartDateLocal)
			postToSlack(slck, activity)
		}

		<-ticker.C

	}

}

func getNewActivities(strv stravaStruct, lastTimestamp *time.Time) (newActivities []strava.ActivitySummary) {

	var newLastTimestamp time.Time
	activityLimit := 10

	// Create clubs service.
	clubService := strava.NewClubsService(strv.client)

	// Get club activities.
	activities, err := clubService.ListActivities(strv.clubID).
		Page(1).
		PerPage(activityLimit).
		Do()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Get the timestamp of the first.
	if len(activities) > 0 {
		newLastTimestamp = activities[0].StartDateLocal
	}

	// First time this function is called. Just save the latest timestamp and quit.
	if lastTimestamp.IsZero() {
		*lastTimestamp = newLastTimestamp
		return nil
	}

	// Pick activities newer than given timestamp.
	newActivities = make([]strava.ActivitySummary, 0)
	i := 0
	for _, activity := range activities {
		if activity.StartDateLocal.After(*lastTimestamp) {
			fmt.Println("new activity")
			newActivities = append(newActivities, *activity)
			i++
		}
	}

	// Grab timestamp.
	*lastTimestamp = newLastTimestamp

	return

}

func postToSlack(slck slackStruct, activity strava.ActivitySummary) {

	// Setup message parameters.
	params := slack.PostMessageParameters{}
	params.Username = slck.username
	params.IconEmoji = slck.icon

	// Generate message.
	message := fmt.Sprintf(
		"%s just finished a %d km %s. Check it out: https://www.strava.com/activities/%d",
		activity.Athlete.FirstName,
		int(activity.Distance/1000),
		strings.ToLower(fmt.Sprintf("%s", activity.Type)),
		activity.Id)

	// Send message.
	_, _, err := slck.client.PostMessage(slck.channel, message, params)

	if err != nil {
		fmt.Printf("%s\n", err)
	} else {
		fmt.Printf(message)
	}

}
