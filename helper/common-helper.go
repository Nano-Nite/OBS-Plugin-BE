package helper

import (
	"os"
	"time"
)

func GetEnv(p string) string {
	if os.Getenv(p) == "" {
		return ""
	}
	return os.Getenv(p)
}

// getCurrentTime returns current time in +7 timezone
func getCurrentTime() time.Time {
	loc, _ := time.LoadLocation("Asia/Bangkok") // UTC+7
	return time.Now().In(loc)
}

func convertStringTimetoTime(s string) time.Time {
	loc, _ := time.LoadLocation("Asia/Bangkok") // UTC+7
	t, err := time.ParseInLocation(
		"2006-01-02T15:04:05",
		s,
		loc,
	)
	if err != nil {
		return time.Now()
	}
	return t
}

func AddDaysFromNextMidnight(baseTime time.Time, days int) time.Time {
	loc := baseTime.Location()

	// Normalize to midnight
	midnight := time.Date(
		baseTime.Year(),
		baseTime.Month(),
		baseTime.Day(),
		0, 0, 0, 0,
		loc,
	)

	// If time is after midnight, move to next day
	if baseTime.After(midnight) {
		midnight = midnight.AddDate(0, 0, 1)
	}

	// Add N days
	return midnight.AddDate(0, 0, days)
}
