package logging

import "log"

// Debug controls whether debug logs are printed.
var Debug bool

// Debugf logs a formatted debug message when Debug is enabled.
func Debugf(format string, v ...any) {
	if Debug {
		log.Printf("DEBUG: "+format, v...)
	}
}
