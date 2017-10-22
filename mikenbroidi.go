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

// Club id.
var stravaClubID int64
var stravaAccessToken string
var stravaClient *strava.Client
var clubService *strava.ClubsService

var lastTimestamp time.Time

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

		fmt.Println("Get new rides.")

		// Get new rides.
		newActivities := getNewActivities()
		// newActivities = getNewActivities()

		// List 'em all!
		for _, activity := range newActivities {
			print(activity.Name + "\n")
			fmt.Println(activity.StartDateLocal)
			postToSlack(activity)
		}

		time.Sleep(60000 * time.Millisecond)

	}

}

func getNewActivities() []strava.ActivitySummary {

	var newLastTimestamp time.Time
	activityLimit := 10

	// Create clubs service.
	clubService := strava.NewClubsService(stravaClient)

	// Get club activities.
	activities, err := clubService.ListActivities(stravaClubID).
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
		lastTimestamp = newLastTimestamp
		return []strava.ActivitySummary{}
	}

	// Pick activities newer than given timestamp.
	newActivities := make([]strava.ActivitySummary, 0)
	i := 0
	for _, activity := range activities {
		if activity.StartDateLocal.After(lastTimestamp) {
			fmt.Println("new activity")
			newActivities = append(newActivities, *activity)
			i++
		}
	}

	// Grab timestamp.
	lastTimestamp = newLastTimestamp

	return newActivities

}

func postToSlack(activity strava.ActivitySummary) {

	// Setup message parameters.
	params := slack.PostMessageParameters{}
	params.Username = slackUsername
	params.IconEmoji = slackIcon

	// Generate message.
	message := fmt.Sprintf(
		"%s just finished a %d km %s. Check it out: https://www.strava.com/activities/%d",
		activity.Athlete.FirstName,
		int(activity.Distance/1000),
		strings.ToLower(fmt.Sprintf("%s", activity.Type)),
		activity.Id)

	// Send message.
	_, _, err := slackClient.PostMessage(slackChannel, message, params)

	if err != nil {
		fmt.Printf("%s\n", err)
	} else {
		fmt.Printf(message)
	}

}
