package app

import (
	"log"
	"os"
)

// debugf logs debug messages only when DEBUG environment variable is set
func debugf(format string, v ...interface{}) {
	if os.Getenv("DEBUG") == "1" {
		log.Printf("[DEBUG] "+format, v...)
	}
}
