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

type NormalizedEvent struct {
	Title     string
	StartTime time.Time
	EndTime   time.Time
}

// ----------------------------
// PUBLIC FUNCTION
// ----------------------------

func ParseAndNormalizeTimeline(md string, loc *time.Location) ([]NormalizedEvent, error) {
	scanner := bufio.NewScanner(strings.NewReader(md))

	// Regex to extract | **Title** | DateText |
	tableRow := regexp.MustCompile(`^\|\s*\*\*(.+?)\*\*\s*\|\s*(.+?)\s*\|`)

	var final []NormalizedEvent

	for scanner.Scan() {
		line := scanner.Text()
		if m := tableRow.FindStringSubmatch(line); m != nil {
			title := strings.TrimSpace(m[1])
			dateText := strings.TrimSpace(m[2])

			events, err := processRow(title, dateText, loc)
			if err != nil {
				// Log error but continue processing other rows in a real app
				// Here we return for debugging
				return nil, fmt.Errorf("parsing '%s': %w", title, err)
			}
			final = append(final, events...)
		}
	}
	return final, nil
}

// ----------------------------
// CORE LOGIC
// ----------------------------

func processRow(title, dateText string, loc *time.Location) ([]NormalizedEvent, error) {
	// 1. Detect Range vs Single
	// Matches: "String – String" (Note: Input uses En-Dash '–', not Hyphen '-')
	rangeRegex := regexp.MustCompile(`(.+?)\s+–\s+(.+)`)

	if rangeRegex.MatchString(dateText) {
		// --- RANGE ---
		parts := rangeRegex.FindStringSubmatch(dateText)
		startRaw := parts[1]
		endRaw := parts[2]

		// Extract year from end date to fix start date if year is missing
		year := extractYear(endRaw)
		if year != "" && !strings.Contains(startRaw, year) {
			startRaw = fmt.Sprintf("%s, %s", startRaw, year)
		}

		startDate, err := parsePureDate(startRaw, loc)
		if err != nil {
			return nil, err
		}
		endDate, err := parsePureDate(endRaw, loc)
		if err != nil {
			return nil, err
		}

		return createRangeEvents(title, startDate, endDate), nil

	} else {
		// --- SINGLE DATE ---
		// Treat single dates as a deadline (Closes event)
		date, err := parsePureDate(dateText, loc)
		if err != nil {
			return nil, err
		}

		return createDeadlineEvent(title, date), nil
	}
}

// ----------------------------
// EVENT FACTORIES (Your Logic)
// ----------------------------

func createRangeEvents(title string, start, end time.Time) []NormalizedEvent {
	// 1. OPENS: Start Date 00:01 -> 01:00
	y1, m1, d1 := start.Date()
	openStart := time.Date(y1, m1, d1, 0, 1, 0, 0, start.Location())
	openEnd   := time.Date(y1, m1, d1, 1, 0, 0, 0, start.Location())

	ev1 := NormalizedEvent{
		Title:     fmt.Sprintf("Opens: %s", title),
		StartTime: openStart,
		EndTime:   openEnd,
	}

	// 2. CLOSES: End Date 23:00 -> 23:59
	y2, m2, d2 := end.Date()
	closeStart := time.Date(y2, m2, d2, 23, 0, 0, 0, end.Location())
	closeEnd   := time.Date(y2, m2, d2, 23, 59, 0, 0, end.Location())

	ev2 := NormalizedEvent{
		Title:     fmt.Sprintf("Closes: %s", title),
		StartTime: closeStart,
		EndTime:   closeEnd,
	}

	return []NormalizedEvent{ev1, ev2}
}

func createDeadlineEvent(title string, date time.Time) []NormalizedEvent {
	// Single Day logic: "Closes" from 23:00 -> 23:59
	y, m, d := date.Date()

	start := time.Date(y, m, d, 23, 0, 0, 0, date.Location())
	end   := time.Date(y, m, d, 23, 59, 0, 0, date.Location())

	return []NormalizedEvent{{
		Title:     fmt.Sprintf("Closes: %s", title), // Or just title if you prefer
		StartTime: start,
		EndTime:   end,
	}}
}

// ----------------------------
// DATE PARSING UTILS
// ----------------------------

func parsePureDate(raw string, loc *time.Location) (time.Time, error) {
	// 1. Clean the string. Remove stuff like "11AM PST (19:00 UTC)"
	// We only want "Wednesday, January 7, 2026"
	// Regex: Stop at the year (4 digits)
	cleanRegex := regexp.MustCompile(`^[A-Za-z]+, [A-Za-z]+ \d{1,2}, \d{4}`)

	clean := cleanRegex.FindString(raw)
	if clean == "" {
		// Fallback: maybe it doesn't have the Day name "January 7, 2026"
		cleanRegex2 := regexp.MustCompile(`^[A-Za-z]+ \d{1,2}, \d{4}`)
		clean = cleanRegex2.FindString(raw)
	}

	if clean == "" {
		// Last resort: try the raw string (might handle simple cases)
		clean = raw
	}

	layouts := []string{
		"Monday, January 2, 2006", // Full format
		"January 2, 2006",         // Short format
	}

	for _, l := range layouts {
		if t, err := time.ParseInLocation(l, clean, loc); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse date: %s", raw)
}

func extractYear(s string) string {
	re := regexp.MustCompile(`\d{4}`)
	return re.FindString(s)
}
