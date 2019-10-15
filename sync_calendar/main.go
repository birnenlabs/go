package main

import (
	"birnenlabs.com/lib/oauth"
	"context"
	"flag"
	"fmt"
	"github.com/golang/glog"
	"google.golang.org/api/calendar/v3"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	// SRC calendar will be listed using these data as min and max time
	HOURS_FROM_TO_SYNC  = -1
	HOURS_UNTIL_TO_SYNC = 60

	// Events in DST calendar will be removed between now-HOURS_TO_CLEAR and now+HOURS_TO_CLEAR
	HOURS_TO_CLEAR = 365 * 24

	// SRC calendar will be printed until these data as min and max time
	MINUTES_FROM_TO_DISPLAY = -10
	HOURS_UNTIL_TO_DISPLAY  = 14
	// Printed data will use special output to mark event as red aor requiring attention
	MINUTES_TO_MARK_RED  = 5
	MINUTES_TO_MARK_ATTN = 1
)

var (
	srcCalName = flag.String("src_calendar", "primary", "Source calendar id, defaults to primary calendar.")
	dstCalName = flag.String("dst_calendar", "", "Destination calendar id, required param.")
	locPrefix  = flag.String("locations", "", "Comma separated prefixed of preffered locations.")
)

func doIAttend(event *calendar.Event) bool {
	for _, a := range event.Attendees {
		if a.Self {
			return a.ResponseStatus != "declined"
		}
	}
	return true
}

func trunc(s string, l int) string {
	if l <= 3 || len(s) <= l {
		return s
	}
	return s[0:l-2] + ".."
}

func dateTimeToTs(e *calendar.EventDateTime) time.Time {
	if e == nil {
		return time.Unix(0, 0)
	}

	if len(e.Date) > 0 {
		t, err := time.Parse("2006-01-02", e.Date)
		if err != nil {
			glog.Errorf("Could not parse %v: %v.", e.Date, err)
			return time.Unix(0, 0)
		}
		return t
	}

	if len(e.DateTime) > 0 {
		t, err := time.Parse(time.RFC3339, e.DateTime)
		if err != nil {
			glog.Errorf("Could not parse %v: %v.", e.DateTime, err)
			return time.Unix(0, 0)
		}
		return t
	}

	return time.Unix(0, 0)
}

func eventsEqual(e1 *calendar.Event, e2 *calendar.Event) bool {
	if e1.Summary != e2.Summary {
		return false
	}
	if e1.Description != e2.Description {
		return false
	}
	if e1.Location != e2.Location {
		return false
	}
	if !dateTimeToTs(e1.Start).Equal(dateTimeToTs(e2.Start)) {
		return false
	}
	if !dateTimeToTs(e1.End).Equal(dateTimeToTs(e2.End)) {
		return false
	}
	return true
}

func organizeLocation(location string) string {
	spl := strings.Split(location, ",")
	prefixes := strings.Split(*locPrefix, ",")
	if len(spl) <= 1 || len(prefixes) == 0 {
		return location
	}

	for i := range spl {
		spl[i] = strings.TrimSpace(spl[i])
	}

	insert := 0
	for _, prefix := range prefixes {
		for i := range spl {
			if i > insert && strings.HasPrefix(spl[i], prefix) {
				spl[insert], spl[i] = spl[i], spl[insert]
				insert++
			}
		}
	}

	return strings.Join(spl, ", ")
}

func main() {
	ctx := context.Background()
	flag.Parse()
	if len(*dstCalName) == 0 {
		glog.Fatalf("dst_calendar flag is required")
	}

	oauthClientSrc, err := oauth.Create("sync_calendar-src")
	if err != nil {
		glog.Fatalf("cannot create oauth src: %v", err)
	}

	oauthClientDst, err := oauth.Create("sync_calendar-dst")
	if err != nil {
		glog.Fatalf("cannot create oauth dst: %v", err)
	}

	if !oauthClientSrc.HasToken() {
		fmt.Printf("SOURCE calendar: ")
		oauthClientSrc.VerifyToken(ctx)
	}

	if !oauthClientDst.HasToken() {
		fmt.Printf("DESTINATION calendar: ")
		oauthClientDst.VerifyToken(ctx)
	}

	clientSrc, err := oauthClientSrc.CreateAuthenticatedHttpClient(ctx)
	if err != nil {
		glog.Fatalf("clientSrc OAuth: %v", err)
	}

	clientDst, err := oauthClientDst.CreateAuthenticatedHttpClient(ctx)
	if err != nil {
		glog.Fatalf("clientDst OAuth: %v", err)
	}

	calSrc, err := calendar.New(clientSrc)
	if err != nil {
		glog.Fatalf("calSrc %v", err)
	}

	calDst, err := calendar.New(clientDst)
	if err != nil {
		glog.Fatalf("calDst %v", err)
	}

	now := time.Now()

	glog.Infof("Listing src events")
	eventsSrc, err := calSrc.Events.List(*srcCalName).ShowDeleted(false).SingleEvents(true).TimeMin(now.Add(HOURS_FROM_TO_SYNC * time.Hour).Format(time.RFC3339)).TimeMax(now.Add(HOURS_UNTIL_TO_SYNC * time.Hour).Format(time.RFC3339)).OrderBy("startTime").Do()
	if err != nil {
		glog.Fatalf("Unable to list src events: %v", err)
	}
	glog.Infof("Listed %v events in src cal", len(eventsSrc.Items))

	glog.Infof("Listing dst events")
	eventsDst, err := calDst.Events.List(*dstCalName).ShowDeleted(false).SingleEvents(true).TimeMin(now.Add(-HOURS_TO_CLEAR * time.Hour).Format(time.RFC3339)).TimeMax(now.Add(HOURS_TO_CLEAR * time.Hour).Format(time.RFC3339)).OrderBy("startTime").Do()
	if err != nil {
		glog.Fatalf("Unable to list dst events: %v", err)
	}
	glog.Infof("Listed %v events in dst cal", len(eventsDst.Items))

	fakeId := 0
	existingDstEvents := make(map[string]*calendar.Event)
	for _, i := range eventsDst.Items {
		if i.Source == nil || len(i.Source.Title) == 0 {
			existingDstEvents[strconv.Itoa(fakeId)] = i
			fakeId++
		} else {
			existingDstEvents[i.Source.Title] = i
		}
	}

	var eventsToPrint []string
	var eventsToAdd []*calendar.Event
	var firstEvent *calendar.Event

	for _, i := range eventsSrc.Items {
		if len(i.Id) == 0 {
			glog.Fatalf("No id for src event: %v", i)
		}
		if !doIAttend(i) {
			glog.Infof("Ignoring: not attending: %q", i.Summary)
			continue
		}
		if len(i.Attendees) == 0 && len(i.Location) == 0 {
			glog.Infof("Ignoring: attendees and location empty: %q", i.Summary)
			continue
		}
		i.Location = organizeLocation(i.Location)

		if existingDstEvents[i.Id] == nil || !eventsEqual(existingDstEvents[i.Id], i) {
			// If does not exist or not equal, let's add
			eventsToAdd = append(eventsToAdd, i)
		} else {
			// If they exists and are equal, no need to add or delete so removing from map, which will be used for deletion
			delete(existingDstEvents, i.Id)
		}

		if len(i.Start.DateTime) == 0 {
			glog.Infof("All day event, sync but not print: %q", i.Summary)
			continue
		}
		start := dateTimeToTs(i.Start)
		if start.Before(now.Add(MINUTES_FROM_TO_DISPLAY*time.Minute)) || start.After(now.Add(HOURS_UNTIL_TO_DISPLAY*time.Hour)) {
			glog.Infof("Too far back or future, sync but not print: %q", i.Summary)
			continue
		}
		glog.Infof("Sync and print: %q", i.Summary)
		if firstEvent == nil {
			firstEvent = i
		}
		var summary string
		if len(eventsToPrint) == 0 {
			summary = i.Summary
		} else {
			summary = trunc(i.Summary, 15)
		}
		eventsToPrint = append(eventsToPrint, start.Format("15:04: ")+summary)
	}

	if len(eventsToPrint) > 0 {
		fmt.Printf(strings.Join(eventsToPrint, ", ") + "\n")
		fmt.Printf(eventsToPrint[0] + "\n")
		fTime := dateTimeToTs(firstEvent.Start)
		if fTime.After(now.Add(-MINUTES_TO_MARK_RED*time.Minute)) && fTime.Before(now.Add(MINUTES_TO_MARK_RED*time.Minute)) {
			fmt.Printf("#ff5555\n")
			if fTime.After(now.Add(-MINUTES_TO_MARK_ATTN*time.Minute)) && fTime.Before(now.Add(MINUTES_TO_MARK_ATTN*time.Minute)) {
				os.Exit(33)
			}
		}
	} else {
		fmt.Printf("No events\n")
	}

	// Now let's sync calendars
	for _, e := range existingDstEvents {
		// Existing events that were not matched needs to be removed from the calendar
		glog.Infof("Removing %q from dst calendar", e.Summary)
		err = calDst.Events.Delete(*dstCalName, e.Id).Do()
		if err != nil {
			glog.Errorf("Could not delete %#v from %v: %v.", e, *dstCalName, err)
		}
	}
	for _, e := range eventsToAdd {
		glog.Infof("Adding %q to dst calendar", e.Summary)
		newEvent := calendar.Event{
			Summary:     e.Summary,
			Description: e.Description,
			Start:       e.Start,
			End:         e.End,
			Location:    e.Location,
			Source: &calendar.EventSource{
				Title: e.Id,
				Url:   "http://calendar.google.com",
			},
		}
		_, err = calDst.Events.Insert(*dstCalName, &newEvent).Do()
		if err != nil {
			glog.Errorf("Could not insert %#v to %v: %v.", newEvent, *dstCalName, err)
		}
	}
}
