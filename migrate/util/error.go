package util

import "fmt"

func ErrorCheck(err error) bool {
	if err != nil {
		LogError("Details: %v", err)
		return true
	}
	return false
}

func ErrorCheckf(err error, context ...interface{}) bool {
	if err != nil {
		LogError("Error: %v Context: %s", err, fmt.Sprintln(context...))
		return true
	}
	return false
}
