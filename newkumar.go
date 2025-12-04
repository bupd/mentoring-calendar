package main

import (
	"fmt"
	"os"
	"time"

	ical "github.com/arran4/golang-ical"
	"github.com/bupd/mentoring-calendar/timeline"
)

func main() {
	// Load events from markdown
	loc, _ := time.LoadLocation("UTC") // You can use local timezone if you like

	events, err := timeline.ParseAndNormalizeTimeline("./events.md", loc)
	if err != nil {
		panic(err)
	}

	// Create calendar
	cal := ical.NewCalendar()
	cal.SetName("Program Timeline")

	// Add events
	for i, e := range events {
		ev := cal.AddEvent(fmt.Sprintf("event-%d", i))
		ev.SetSummary(e.Title)
		ev.SetStartAt(e.StartTime)
		ev.SetEndAt(e.EndTime)
		ev.SetDtStampTime(time.Now().UTC())
		ev.SetDescription("Generated from events.md")
	}

	// Serialize to ICS file
	err = os.WriteFile("timeline.ics", []byte(cal.Serialize()), 0644)
	if err != nil {
		panic(err)
	}

	fmt.Println("✅ timeline.ics generated successfully")
}
