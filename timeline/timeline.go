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

	// Regex matches rows like: | **Bold Title** | Date Text |
	// It will implicitly ignore lines like "### Timeline" or normal text.
	tableRow := regexp.MustCompile(`^\|\s*\*\*(.+?)\*\*\s*\|\s*(.+?)\s*\|`)

	var final []NormalizedEvent

	for scanner.Scan() {
		line := scanner.Text()

		// 1. Check if line matches table format
		if m := tableRow.FindStringSubmatch(line); m != nil {
			title := strings.TrimSpace(m[1])
			dateText := strings.TrimSpace(m[2])

			// 2. FILTER: Skip the table header row
			if strings.EqualFold(title, "Activity") {
				continue
			}

			// 3. Process the valid row
			events, err := processRow(title, dateText, loc)
			if err != nil {
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
	// Range Splitter (Note: uses En Dash '–')
	rangeRegex := regexp.MustCompile(`(.+?)\s+–\s+(.+)`)

	if rangeRegex.MatchString(dateText) {
		// --- RANGE EVENT ---
		parts := rangeRegex.FindStringSubmatch(dateText)
		startRaw := parts[1]
		endRaw := parts[2]

		// Smart Fix: If start date missing year, grab it from end date
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
		// --- SINGLE DATE EVENT ---
		date, err := parsePureDate(dateText, loc)
		if err != nil {
			return nil, err
		}
		// Single days are treated as deadlines
		return createDeadlineEvent(title, date), nil
	}
}

// ----------------------------
// EVENT FACTORIES
// ----------------------------

func createRangeEvents(title string, start, end time.Time) []NormalizedEvent {
	// OPENS: Start Date 00:01 -> 01:00
	y1, m1, d1 := start.Date()
	openStart := time.Date(y1, m1, d1, 0, 1, 0, 0, start.Location())
	openEnd := time.Date(y1, m1, d1, 1, 0, 0, 0, start.Location())

	ev1 := NormalizedEvent{
		Title:     fmt.Sprintf("Opens: %s", title),
		StartTime: openStart,
		EndTime:   openEnd,
	}

	// CLOSES: End Date 23:00 -> 23:59
	y2, m2, d2 := end.Date()
	closeStart := time.Date(y2, m2, d2, 23, 0, 0, 0, end.Location())
	closeEnd := time.Date(y2, m2, d2, 23, 59, 0, 0, end.Location())

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

	start := time.Date(y, m, d, 0, 1, 0, 0, date.Location())
	end := time.Date(y, m, d, 01, 0, 0, 0, date.Location())

	return []NormalizedEvent{{
		Title:     fmt.Sprintf("Closes: %s", title),
		StartTime: start,
		EndTime:   end,
	}}
}

// ----------------------------
// DATE PARSING UTILS
// ----------------------------

func parsePureDate(raw string, loc *time.Location) (time.Time, error) {
	// Remove specific times like "11AM PST" or "(19:00 UTC)"
	// We want to extract just the date part: "Month Day, Year"

	// Regex looks for Month + Day + Year (e.g., "January 20, 2026")
	// It deliberately ignores the text after the year
	cleanRegex := regexp.MustCompile(`[A-Za-z]+ \d{1,2}, \d{4}`)

	clean := cleanRegex.FindString(raw)
	if clean == "" {
		return time.Time{}, fmt.Errorf("unable to find valid date in: %s", raw)
	}

	// Try parsing
	t, err := time.ParseInLocation("January 2, 2006", clean, loc)
	if err != nil {
		return time.Time{}, err
	}
	return t, nil
}

func extractYear(s string) string {
	re := regexp.MustCompile(`\d{4}`)
	return re.FindString(s)
}
