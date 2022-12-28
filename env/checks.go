package env

import (
	"fmt"
	"os"
)

func RunChecks() {
	checkEnv("APP_URL")
}

func checkEnv(key string) {
	if os.Getenv(key) == "" {
		panic(fmt.Sprintf("Environment variable %s is not set", key))
	}
}
