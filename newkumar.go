package main

import (
	"fmt"
	"log"
	"os"
	"time"

	// Update with your actual path

	ical "github.com/arran4/golang-ical"
	"github.com/bupd/mentoring-calendar/timeline"
	"github.com/google/uuid"
)

func main() {
	// 1. Read the Markdown file
	data, err := os.ReadFile("events.md")
	if err != nil {
		log.Fatalf("Error reading events.md: %v", err)
	}
	markdownContent := string(data)

	// 2. Setup Calendar
	cal := ical.NewCalendar()
	cal.SetMethod(ical.MethodPublish)
	cal.SetProdId("-//LFX Mentorship//Timeline//EN")

	// 3. Define Location (e.g., "Asia/Kolkata" or time.UTC)
	// This determines the timezone for "00:01" and "23:00"
	loc, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		log.Println("Warning: Could not load timezone, falling back to UTC")
		loc = time.UTC
	}

	// 4. Parse Timeline
	events, err := timeline.ParseAndNormalizeTimeline(markdownContent, loc)
	if err != nil {
		log.Fatalf("Failed to parse timeline: %v", err)
	}

	// 5. Generate iCal Events
	now := time.Now()

	for _, evt := range events {
		uid := uuid.New().String()
		event := cal.AddEvent(uid)

		event.SetDtStampTime(now)
		event.SetCreatedTime(now)
		event.SetModifiedAt(now)

		event.SetStartAt(evt.StartTime)
		event.SetEndAt(evt.EndTime)

		event.SetSummary(evt.Title)
		event.SetDescription("Generated from LFX Timeline")
	}

	// 6. Print to Stdout (or you could write to a file)
	fmt.Println(cal.Serialize())
}
