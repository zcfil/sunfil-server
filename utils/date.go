package utils

import "time"

const (
	YYMMDDHMS = "2006-01-02 15:04:05"
	YYMMDD    = "2006-01-02"
)

// UnixToTime Convert timestamp to date
func UnixToTime(timestamp int) time.Time {
	return time.Unix(int64(timestamp), 0)
}

// DateToUnix Convert date to timestamp
func DateToUnix(str string) int64 {
	template := "2006-01-02 15:04:05"
	t, err := time.ParseInLocation(template, str, time.Local)
	if err != nil {
		return 0
	}
	return t.Unix()
}

// StrToTime String conversion time
func StrToTime(data string) time.Time {
	loc, _ := time.LoadLocation("Local")
	t, _ := time.ParseInLocation("2006-01-02 15:04:05", data, loc)
	return t
}

func TimeToStr(t time.Time, timeType string) string {
	return t.Format(timeType)
}

func GetNowStr() string {
	return TimeToStr(time.Now(), YYMMDDHMS)
}
