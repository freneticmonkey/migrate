package util

import "log"

func ErrorCheck(err error) {
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
}
