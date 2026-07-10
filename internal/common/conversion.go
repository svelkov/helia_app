package common

import (
	"strconv"
	"strings"
	"time"
)

var (
	defaultInt     int
	defaultInt64   int64
	defaultFloat64 float64
	defaultFloat32 float32
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
func AnyToFloat64(value any) float64 {
	switch v := value.(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int:
		return float64(v)
	case int64:
		return float64(v)
	case string:
		return StringToFloat64(v)
	default:
		return defaultFloat64
	}
}

func StringToFloat64(str string) float64 {
	str = strings.TrimSpace(str)
	commas := strings.Count(str, ",")
	dots := strings.Count(str, ".")

	switch {
	case commas > 0 && dots > 0:
		// Both present — last one is the decimal separator
		if strings.LastIndex(str, ",") > strings.LastIndex(str, ".") {
			// European: 1.234,56 → remove dots, comma becomes dot
			str = strings.ReplaceAll(str, ".", "")
			str = strings.ReplaceAll(str, ",", ".")
		} else {
			// US: 1,234.56 → remove commas
			str = strings.ReplaceAll(str, ",", "")
		}
	case commas == 1:
		// Only comma → decimal separator (e.g. "150,00")
		str = strings.ReplaceAll(str, ",", ".")
	}
	// Only dot or plain integer → already valid for ParseFloat

	val, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return defaultFloat64
	}
	return val
}
func StringToInt(str string) int {
	str = strings.ReplaceAll(str, ".", "")
	str = strings.ReplaceAll(str, ",", "")

	val, err := strconv.Atoi(str)
	if err != nil {
		return defaultInt
	}
	return val
}
func StringToInt64(str string) int64 {
	str = strings.ReplaceAll(str, ".", "")
	str = strings.ReplaceAll(str, ",", "")

	val, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return defaultInt64
	}
	return val
}
func StringToFloat32(str string) float32 {
	str = strings.TrimSpace(str)
	commas := strings.Count(str, ",")
	dots := strings.Count(str, ".")

	switch {
	case commas > 0 && dots > 0:
		// Both present — last one is the decimal separator
		if strings.LastIndex(str, ",") > strings.LastIndex(str, ".") {
			// European: 1.234,56 → remove dots, comma becomes dot
			str = strings.ReplaceAll(str, ".", "")
			str = strings.ReplaceAll(str, ",", ".")
		} else {
			// US: 1,234.56 → remove commas
			str = strings.ReplaceAll(str, ",", "")
		}
	case commas == 1:
		// Only comma → decimal separator (e.g. "150,00")
		str = strings.ReplaceAll(str, ",", ".")
	}
	// Only dot or plain integer → already valid for ParseFloat

	val, err := strconv.ParseFloat(str, 32)
	if err != nil {
		return defaultFloat32
	}
	return float32(val)
}
