package helper

import "os"

func GetEnv(p string) string {
	if os.Getenv(p) == "" {
		return ""
	}
	return os.Getenv(p)
}
