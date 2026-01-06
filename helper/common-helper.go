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
