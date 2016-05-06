package util

import "log"

func ErrorCheck(err error) {
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
}

func ErrorCheckf(err error, context string) {
	if err != nil {
		log.Fatalf("Error: %v Context: %s", err, context)
	}
}
