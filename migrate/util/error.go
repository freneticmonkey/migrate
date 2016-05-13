package util

import "fmt"

func ErrorCheck(err error) bool {
	if err != nil {
		LogErrorf("Details: %v", err)
		return true
	}
	return false
}

func ErrorCheckf(err error, format string, context ...interface{}) bool {
	if err != nil {
		LogErrorf("Error: %v Context: %s", err, fmt.Sprintf(format, context...))
		return true
	}
	return false
}
