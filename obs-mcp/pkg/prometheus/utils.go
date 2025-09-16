package prometheus

import (
	"fmt"
	"strconv"
	"time"
)

func ParseDuration(duration string) (time.Duration, error) {
	if duration == "" {
		return 0, fmt.Errorf("empty duration")
	}

	// Handle weeks (not supported by time.ParseDuration)
	if len(duration) > 1 && duration[len(duration)-1] == 'w' {
		weeks, err := strconv.Atoi(duration[:len(duration)-1])
		if err != nil {
			return 0, fmt.Errorf("invalid weeks format: %s", duration)
		}
		return time.Duration(weeks) * 7 * 24 * time.Hour, nil
	}

	// Handle days (not supported by time.ParseDuration)
	if len(duration) > 1 && duration[len(duration)-1] == 'd' {
		days, err := strconv.Atoi(duration[:len(duration)-1])
		if err != nil {
			return 0, fmt.Errorf("invalid days format: %s", duration)
		}
		return time.Duration(days) * 24 * time.Hour, nil
	}

	return time.ParseDuration(duration)
}

func ParseTimestamp(timestamp string) (time.Time, error) {
	// Try parsing as RFC3339 first
	if t, err := time.Parse(time.RFC3339, timestamp); err == nil {
		return t, nil
	}

	// Try parsing as Unix timestamp
	if unixTime, err := strconv.ParseInt(timestamp, 10, 64); err == nil {
		return time.Unix(unixTime, 0), nil
	}

	return time.Time{}, fmt.Errorf("timestamp must be RFC3339 format or Unix timestamp")
}
