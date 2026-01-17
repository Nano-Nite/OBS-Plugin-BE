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

func TTLUntilMidnight() time.Duration {
	midnight := time.Date(
		time.Now().Year(),
		time.Now().Month(),
		time.Now().Day()+1,
		0, 0, 0, 0,
		time.Now().Location(),
	)
	return time.Until(midnight)
}
