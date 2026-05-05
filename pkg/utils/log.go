package utils

import (
	"fmt"
	"os"
	"time"
)

// Log writes a debug message to /tmp/wibble.log if the WIBBLE_DEBUG environment
// variable is set to "1".
func Log(msg string) {
	if os.Getenv("WIBBLE_DEBUG") != "1" {
		return
	}

	f, err := os.OpenFile("/tmp/wibble.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer f.Close()

	fmt.Fprintf(f, "%s %s\n", time.Now().Format("2006-01-02 15:04:05.000000"), msg)
}
