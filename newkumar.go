package main

import (
	"fmt"
	"log"
	"time"

	// Update with your actual path

	ical "github.com/arran4/golang-ical"
	"github.com/bupd/mentoring-calendar/timeline"
	"github.com/google/uuid"
)

const markdownData = `
| **Project Proposals Open** | Wednesday, January 7 – Tuesday, January 20, 2026 |
| **Mentee Applications Open** | Monday, January 26 – Tuesday, February 10, 2026, 11AM PST (19:00 UTC) |
| **Selection Notifications** | Thursday, February 26, 2026 |
| **Mentorship Program Begins** | Monday, March 2, 2026 |
`

func main() {
	// 1. Setup Calendar
	cal := ical.NewCalendar()
	cal.SetMethod(ical.MethodPublish)
	cal.SetProductId("-//LFX Mentorship//Timeline//EN")
	cal.SetName("LFX Mentorship Timeline")

	// 2. Set Timezone (Used for the 00:01 and 23:00 calculations)
	// Using Local or specific timezone ensures the 23:00 is relevant to that region.
	loc, _ := time.LoadLocation("Asia/Kolkata")

	// 3. Parse
	events, err := timeline.ParseAndNormalizeTimeline(markdownData, loc)
	if err != nil {
		log.Fatal(err)
	}

	// 4. Generate iCal Events
	now := time.Now()

	for _, evt := range events {
		// Create unique ID
		uid := uuid.New().String()
		event := cal.AddEvent(uid)

		// Metadata
		event.SetDtStampTime(now)
		event.SetCreatedTime(now)
		event.SetModifiedAt(now)

		// Times
		event.SetStartAt(evt.StartTime)
		event.SetEndAt(evt.EndTime)

		// Content
		event.SetSummary(evt.Title)
		event.SetDescription("LFX Mentorship Program Event")
	}

	fmt.Println(cal.Serialize())
}
