package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/nlopes/slack"
	strava "github.com/strava/go.strava"
)

// Club id.
var stravaClubID int64 = 231903
var stravaAccessToken string
var stravaClient *strava.Client
var clubService *strava.ClubsService

var lastTimestamp = 0

var slackToken string
var slackChannel string
var slackUsername string
var slackIcon string
var slackClient *slack.Client

func main() {

	// Read settings from environment variables.
	stravaClubID, _ = strconv.ParseInt(os.Getenv("STRAVA_CLUBID"), 10, 32)
	stravaAccessToken = os.Getenv("STRAVA_ACCESS_TOKEN")
	slackToken = os.Getenv("SLACK_ACCESS_TOKEN")
	slackChannel = os.Getenv("SLACK_CHANNEL")
	slackUsername = os.Getenv("SLACK_USERNAME")
	slackIcon = os.Getenv("SLACK_ICON")

	// Check that variables exists.
	if stravaClubID == 0 || len(stravaAccessToken) == 0 || len(slackToken) == 0 || len(slackChannel) == 0 || len(slackUsername) == 0 || len(slackIcon) == 0 {
		fmt.Println("Please define following environment variables: STRAVA_CLUBID, STRAVA_ACCESS_TOKEN, SLACK_ACCESS_TOKEN, SLACK_CHANNEL, SLACK_USERNAME, SLACK_ICON")
	}

	// Create Strava client.
	stravaClient = strava.NewClient(stravaAccessToken)

	// Create slack client.
	slackClient = slack.New(slackToken)

	// Start watching for new Strava activities.
	for true {

		// Get new rides.
		// newRides := getNewRides(&lastTimestamp)
	}

}

func getNewRides(lastTimestamp *int) []*strava.ActivitySummary {

	// Create clubs service.
	clubService := strava.NewClubsService(stravaClient)

	// Get club activities.
	activities, err := clubService.ListActivities(stravaClubID).
		Page(1).
		PerPage(10).
		Do()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// List 'em all!
	for _, v := range activities {
		print(v.Name + "\n")
	}

	// print(len(activities))

	fmt.Printf("\n\n\n")

	return activities

}

func postToSlack(activity strava.ActivitySummary) {

	// New Slack client.

	params := slack.PostMessageParameters{}
	params.Username = "mikenbroidi"
	params.IconEmoji = "dog"
	channelID, timestamp, err := slackClient.PostMessage("C7N5FQP7X", "Some text2", params)
	if err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	}
	fmt.Printf("Message successfully sent to channel %s at %s", channelID, timestamp)

}
