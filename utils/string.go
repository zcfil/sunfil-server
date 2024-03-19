package utils

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"time"
)

var (
	PointLength = math.Pow(10, 18)
	TwoBit      = "2"
	FourBit     = "4"
)

const (
	StandardFormat = "2006-01-02 15:04:05"
	FileCoinFormat = "2020-08-25 06:00:00"
	Height         = 30
	DayHeight      = 28800
)

func Strval(value interface{}) string {
	var key string
	if value == nil {
		return key
	}

	switch value.(type) {
	case float64:
		ft := value.(float64)
		key = strconv.FormatFloat(ft, 'f', -1, 64)
	case float32:
		ft := value.(float32)
		key = strconv.FormatFloat(float64(ft), 'f', -1, 64)
	case int:
		it := value.(int)
		key = strconv.Itoa(it)
	case uint:
		it := value.(uint)
		key = strconv.Itoa(int(it))
	case int8:
		it := value.(int8)
		key = strconv.Itoa(int(it))
	case uint8:
		it := value.(uint8)
		key = strconv.Itoa(int(it))
	case int16:
		it := value.(int16)
		key = strconv.Itoa(int(it))
	case uint16:
		it := value.(uint16)
		key = strconv.Itoa(int(it))
	case int32:
		it := value.(int32)
		key = strconv.Itoa(int(it))
	case uint32:
		it := value.(uint32)
		key = strconv.Itoa(int(it))
	case int64:
		it := value.(int64)
		key = strconv.FormatInt(it, 10)
	case uint64:
		it := value.(uint64)
		key = strconv.FormatUint(it, 10)
	case string:
		key = value.(string)
	case []byte:
		key = string(value.([]byte))
	default:
		newValue, _ := json.Marshal(value)
		key = string(newValue)
	}

	return key
}

func BoolToByte(b bool) []byte {

	buf := bytes.NewBuffer([]byte{})
	binary.Write(buf, binary.BigEndian, b)
	return buf.Bytes()
}

func StrToFloat64(str, bit string) float64 {

	b, _ := strconv.ParseFloat(str, 64)
	b1, _ := strconv.ParseFloat(fmt.Sprintf("%."+bit+"f", b/PointLength), 64)
	return b1
}

func FloatAccurateBit(str float64, bit string) float64 {
	b1, _ := strconv.ParseFloat(fmt.Sprintf("%."+bit+"f", str), 64)
	return b1
}

func BlockHeightToStr(height int64) string {

	times, _ := time.Parse(StandardFormat, FileCoinFormat)
	tsum := height*Height + times.Unix() - DayHeight
	timeStr := time.Unix(tsum, 0).Format(StandardFormat)
	return timeStr
}
