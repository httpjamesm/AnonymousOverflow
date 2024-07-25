package utils

import "os"

func GetThemeFromEnv() string {
	theme := os.Getenv("THEME")
	if theme == "" {
		theme = "auto"
	}
	return theme
}
