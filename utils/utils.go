package utils

import (
	"encoding/json"
	"fmt"
	"math"
	"time"
)

func GetDateStringFromTimestamp(timestampInSeconds int64) string {
	return time.Unix(int64(timestampInSeconds), 0).Format(time.RFC850)
}

func GetWeekdayFromTimestamp(timestampInSeconds int64) string {
	return time.Unix(int64(timestampInSeconds), 0).Weekday().String()
}

func GetRFC3339Date(year string, month string, date string, hour string, minute string, second string) (time.Time, error) {
	return time.Parse(time.RFC3339, year+"-"+month+"-"+date+"T"+hour+":"+minute+":"+second+"+05:30")
}

func BruteStringify(v any) string {
	asString, err := json.Marshal(v)
	if err != nil {
		return "N/A"
	}
	return string(asString)
}

func FloatCompare(a float64, b float64) int {
	const EPS float64 = 1e-9
	if math.Abs(a-b) < EPS {
		return 0
	} else if a > b {
		return 1
	} else {
		return -1
	}
}

func RoundFloat(num float64) string {
	cmp := FloatCompare(num, 0.)
	if cmp > 0 {
		return fmt.Sprintf("+%.2f", num)
	} else if cmp < 0 {
		return fmt.Sprintf("-%.2f", -num)
	} else {
		return fmt.Sprintf("+%.2f", 0.)
	}
}

func Max(x, y int) int {
	if x > y {
		return x
	}
	return y
}
func Min(x, y int) int {
	if x < y {
		return x
	}
	return y
}
