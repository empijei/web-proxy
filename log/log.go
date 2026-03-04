package l

import "log"

func Debugf(f string, vals ...any) {
	log.Printf("D: "+f, vals...)
}

func Infof(f string, vals ...any) {
	log.Printf("I: "+f, vals...)
}

func Warnf(f string, vals ...any) {
	log.Printf("W: "+f, vals...)
}

func Errorf(f string, vals ...any) {
	log.Printf("E: "+f, vals...)
}

func Fatalf(f string, vals ...any) {
	log.Fatalf("F: "+f, vals...)
}
