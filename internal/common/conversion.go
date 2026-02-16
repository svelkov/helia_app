package common

import (
	"strconv"
	"time"
)

var (
	defaultInt     int
	defaultInt64   int64
	defaultFloat64 float64
	defaultTime    time.Time
	defaultBool    bool
)

// StringToBool converts string to bool
func StringToBool(boolStr string) bool {
	val, err := strconv.ParseBool(boolStr)
	if err != nil {
		return defaultBool
	}
	return val
}

// StringToDate converts string to time.Time (date only)
func StringToDate(dateStr string) time.Time {
	val, err := time.Parse(HtmlLayout, dateStr)
	if err != nil {
		return defaultTime
	}
	return val
}

// StringToDateTime converts string to time.Time (date and time)
func StringToDateTime(datetimeStr string) time.Time {
	val, err := time.Parse("2006-01-02 15:04:05", datetimeStr)
	if err != nil {
		return defaultTime
	}
	return val
}

// StringToTime converts string to time.Time (time only)
func StringToTime(timeStr string) time.Time {
	val, err := time.Parse("15:04:05", timeStr)
	if err != nil {
		return defaultTime
	}
	return val
}

// StringToTimeWithLayout converts string to time.Time with custom layout
func StringToTimeWithLayout(timeStr, layout string) (time.Time, error) {
	return time.Parse(layout, timeStr)
}

func StringToFloat64(str string) float64 {
	val, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return defaultFloat64
	}
	return val
}
func StringToInt(str string) int {
	val, err := strconv.Atoi(str)
	if err != nil {
		return defaultInt
	}
	return val
}
func StringToInt64(str string) int64 {
	val, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return defaultInt64
	}
	return val
}
