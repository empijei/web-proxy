// Package l is a work in progress.
package l

import "log"

// Debugf logs at debug level.
func Debugf(f string, vals ...any) {
	log.Printf("D: "+f, vals...)
}

// Infof logs at info level.
func Infof(f string, vals ...any) {
	log.Printf("I: "+f, vals...)
}

// Warnf logs at warning level.
func Warnf(f string, vals ...any) {
	log.Printf("W: "+f, vals...)
}

// Errorf logs at error level.
func Errorf(f string, vals ...any) {
	log.Printf("E: "+f, vals...)
}

// Fatalf logs at fatal level and closes the program.
func Fatalf(f string, vals ...any) {
	log.Fatalf("F: "+f, vals...)
}
