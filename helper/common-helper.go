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
	return time.Now()
}

func AddDaysFromNextMidnight(baseTime time.Time, days int) time.Time {
	return time.Now().AddDate(0, 0, days+1)
}
