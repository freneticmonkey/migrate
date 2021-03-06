package mysql

import "time"

// TimeFormat The go time format string for DB times.
var TimeFormat = `2006-01-02 15:04:05`

// GetTimeNow Return the current time as a string in MySQL time format
func GetTimeNow() string {
	return FormatTime(time.Now())
}

// FormatTime Return the current time as a string in MySQL time format
func FormatTime(tm time.Time) string {
	return tm.UTC().Format(TimeFormat)
}
