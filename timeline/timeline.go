package timeline

import (
	"bufio"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// ----------------------------
// MODELS
// ----------------------------

type RawEvent struct {
	Title    string
	DateText string
}

type NormalizedEvent struct {
	Title     string
	StartTime time.Time
	EndTime   time.Time
}

// ----------------------------
// PUBLIC META FUNCTION
// ----------------------------

// ParseAndNormalizeTimeline reads the event.md text → extracts rows → creates events.
func ParseAndNormalizeTimeline(md string, loc *time.Location) ([]NormalizedEvent, error) {
	raw := extractRawEvents(md)

	// fmt.Println("raw = ", raw)  // debug if needed

	var final []NormalizedEvent

	for _, r := range raw {
		events, err := convertRawToEvents(r, loc)
		if err != nil {
			return nil, err
		}
		final = append(final, events...)
	}
	return final, nil
}

// ----------------------------
// STEP 1 — Extract rows from timeline table
// ----------------------------

func extractRawEvents(md string) []RawEvent {
	scanner := bufio.NewScanner(strings.NewReader(md))
	fmt.Println("scanner = ", scanner)

	tableRow := regexp.MustCompile(`^\|\s*\*\*(.+?)\*\*\s*\|\s*(.+?)\s*\|`)

	var events []RawEvent

	for scanner.Scan() {
		line := scanner.Text()

		if m := tableRow.FindStringSubmatch(line); m != nil {
			events = append(events, RawEvent{
				Title:    strings.TrimSpace(m[1]),
				DateText: strings.TrimSpace(m[2]),
			})
		}
	}

	return events
}

// ----------------------------
// STEP 2 — Convert human date → normalized events
// ----------------------------

// convertRawToEvents handles:
//   - "Jan 1 – Jan 10, 2026"   ⇒ opens + closes events
//   - "Feb 12, 2026"           ⇒ single-day event
//
// The location (PST/UTC/etc) is passed from caller.
func convertRawToEvents(r RawEvent, loc *time.Location) ([]NormalizedEvent, error) {

	// Pattern for ranges:  "Monday, January 7 – Tuesday, January 20, 2026"
	rangeRegex := regexp.MustCompile(`(.+?)\s+–\s+(.+)`)
	singleRegex := regexp.MustCompile(`^[A-Za-z]+.*\d{4}$`)

	txt := r.DateText

	switch {
	case rangeRegex.MatchString(txt):
		// MULTI-DAY RANGE → create "opens" and "closes" events
		parts := rangeRegex.FindStringSubmatch(txt)
		startStr := parts[1]
		endStr := parts[2]

		startDate, err := parseHumanDate(startStr, loc)
		if err != nil {
			return nil, fmt.Errorf("start date parse failed: %w", err)
		}

		endDate, err := parseHumanDate(endStr, loc)
		if err != nil {
			return nil, fmt.Errorf("end date parse failed: %w", err)
		}

		return createOpenCloseEvents(r.Title, startDate, endDate), nil

	case singleRegex.MatchString(txt):
		// SINGLE DATE
		d, err := parseHumanDate(txt, loc)
		if err != nil {
			return nil, err
		}

		return []NormalizedEvent{
			{
				Title:     r.Title,
				StartTime: time.Date(d.Year(), d.Month(), d.Day(), 0, 1, 0, 0, loc),
				EndTime:   time.Date(d.Year(), d.Month(), d.Day(), 1, 0, 0, 0, loc),
			},
		}, nil
	}

	return nil, fmt.Errorf("unrecognized date format: %s", txt)
}

// ----------------------------
// STEP 3 — Parsing human-readable dates
// ----------------------------

var dateLayouts = []string{
	"Monday, January 2, 2006",
	"Monday, Jan 2, 2006",
	"January 2, 2006",
	"Jan 2, 2006",
}

func parseHumanDate(s string, loc *time.Location) (time.Time, error) {
	s = strings.TrimSpace(s)

	for _, layout := range dateLayouts {
		if t, err := time.ParseInLocation(layout, s, loc); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("cannot parse date: %s", s)
}

// ----------------------------
// STEP 4 — Create open/close events for range
// ----------------------------

func createOpenCloseEvents(title string, start, end time.Time) []NormalizedEvent {

	opens := NormalizedEvent{
		Title:     fmt.Sprintf("%s — Opens", title),
		StartTime: time.Date(start.Year(), start.Month(), start.Day(), 0, 1, 0, 0, start.Location()),
		EndTime:   time.Date(start.Year(), start.Month(), start.Day(), 1, 0, 0, 0, start.Location()),
	}

	closes := NormalizedEvent{
		Title:     fmt.Sprintf("%s — Closes", title),
		StartTime: time.Date(end.Year(), end.Month(), end.Day(), 23, 0, 0, 0, end.Location()),
		EndTime:   time.Date(end.Year(), end.Month(), end.Day(), 23, 59, 0, 0, end.Location()),
	}

	return []NormalizedEvent{opens, closes}
}
