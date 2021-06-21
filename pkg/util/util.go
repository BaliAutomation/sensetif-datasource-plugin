package util

import (
	"os"
	"strings"
)

var developmentMode = 0

func IsDevelopmentMode() bool {
	if developmentMode == 0 {
		devMode, ok := os.LookupEnv("SENSETIF_DEV_MODE")
		if ok && strings.ToLower(devMode) == "true" {
			developmentMode = 1
		} else {
			developmentMode = 2
		}
	}
	return developmentMode == 1
}
