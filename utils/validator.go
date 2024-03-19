package utils

import (
	"errors"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

type Rules map[string][]string

type RulesMap map[string]Rules

var CustomizeMap = make(map[string]Rules)

//@function: RegisterRule
//@description: It is recommended to register custom rule schemes at the router initialization layer
//@param: key string, rule Rules
//@return: err error

func RegisterRule(key string, rule Rules) (err error) {
	if CustomizeMap[key] != nil {
		return errors.New(key + "Registered, cannot duplicate registration")
	} else {
		CustomizeMap[key] = rule
		return nil
	}
}

//@function: NotEmpty
//@description: Non empty cannot be a 0 value for its corresponding type
//@return: string

func NotEmpty() string {
	return "notEmpty"
}

// @function: RegexpMatch
// @description: Regular validation verifies whether the input item satisfies a regular expression
// @param:  rule string
// @return: string

func RegexpMatch(rule string) string {
	return "regexp=" + rule
}

//@function: Lt
//@description: Less than the input parameter (<). If it is a string array Slice, it is a length comparison. If it is an int uint float, it is a numerical comparison
//@param: mark string
//@return: string

func Lt(mark string) string {
	return "lt=" + mark
}

//@function: Le
//@description: Less than or equal to the input parameter (<=). If it is a string array Slice, it is a length comparison. If it is an int uint float, it is a numerical comparison
//@param: mark string
//@return: string

func Le(mark string) string {
	return "le=" + mark
}

//@function: Eq
//@description: Equal to the input parameter (==). If it is a string array Slice, it is a length comparison. If it is an int uint float, it is a numerical comparison
//@param: mark string
//@return: string

func Eq(mark string) string {
	return "eq=" + mark
}

//@function: Ne
//@description: Not equal to input parameter (!=). If it is a string array Slice, it is a length comparison. If it is an int uint float, it is a numerical comparison
//@param: mark string
//@return: string

func Ne(mark string) string {
	return "ne=" + mark
}

//@function: Ge
//@description: Greater than or equal to the input parameter (>=). If it is a string array Slice, it is a length comparison. If it is an int uint float, it is a numerical comparison
//@param: mark string
//@return: string

func Ge(mark string) string {
	return "ge=" + mark
}

//@function: Gt
//@description: Greater than the input parameter (>). If it is a string array Slice, it is a length comparison. If it is an int uint float, it is a numerical comparison
//@param: mark string
//@return: string

func Gt(mark string) string {
	return "gt=" + mark
}

//@function: Verify
//@description: Verification method
//@param: st interface{}, roleMap Rules(Enter parameter instance, rule map)
//@return: err error

func Verify(st interface{}, roleMap Rules) (err error) {
	compareMap := map[string]bool{
		"lt": true,
		"le": true,
		"eq": true,
		"ne": true,
		"ge": true,
		"gt": true,
	}

	typ := reflect.TypeOf(st)
	val := reflect.ValueOf(st) // get reflect.Type

	kd := val.Kind() // Obtain the category corresponding to st
	if kd != reflect.Struct {
		return errors.New("expect struct")
	}
	num := val.NumField()
	for i := 0; i < num; i++ {
		tagVal := typ.Field(i)
		val := val.Field(i)
		if len(roleMap[tagVal.Name]) > 0 {
			for _, v := range roleMap[tagVal.Name] {
				switch {
				case v == "notEmpty":
					if isBlank(val) {
						return errors.New(tagVal.Name + "value cannot be empty")
					}
				case strings.Split(v, "=")[0] == "regexp":
					if !regexpMatch(strings.Split(v, "=")[1], val.String()) {
						return errors.New(tagVal.Name + "format verification failed")
					}
				case compareMap[strings.Split(v, "=")[0]]:
					if !compareVerify(val, v) {
						return errors.New(tagVal.Name + "the length or value is not within the legal range," + v)
					}
				}
			}
		}
	}
	return nil
}

//@function: compareVerify
//@description: The verification method for length and numbers is automatically verified based on the type
//@param: value reflect.Value, VerifyStr string
//@return: bool

func compareVerify(value reflect.Value, VerifyStr string) bool {
	switch value.Kind() {
	case reflect.String:
		return compare(len([]rune(value.String())), VerifyStr)
	case reflect.Slice, reflect.Array:
		return compare(value.Len(), VerifyStr)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return compare(value.Uint(), VerifyStr)
	case reflect.Float32, reflect.Float64:
		return compare(value.Float(), VerifyStr)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return compare(value.Int(), VerifyStr)
	default:
		return false
	}
}

//@function: isBlank
//@description: Non null validation
//@param: value reflect.Value
//@return: bool

func isBlank(value reflect.Value) bool {
	switch value.Kind() {
	case reflect.String, reflect.Slice:
		return value.Len() == 0
	case reflect.Bool:
		return !value.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return value.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return value.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return value.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return value.IsNil()
	}
	return reflect.DeepEqual(value.Interface(), reflect.Zero(value.Type()).Interface())
}

//@function: compare
//@description: Comparison function
//@param: value interface{}, VerifyStr string
//@return: bool

func compare(value interface{}, VerifyStr string) bool {
	VerifyStrArr := strings.Split(VerifyStr, "=")
	val := reflect.ValueOf(value)
	switch val.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		VInt, VErr := strconv.ParseInt(VerifyStrArr[1], 10, 64)
		if VErr != nil {
			return false
		}
		switch {
		case VerifyStrArr[0] == "lt":
			return val.Int() < VInt
		case VerifyStrArr[0] == "le":
			return val.Int() <= VInt
		case VerifyStrArr[0] == "eq":
			return val.Int() == VInt
		case VerifyStrArr[0] == "ne":
			return val.Int() != VInt
		case VerifyStrArr[0] == "ge":
			return val.Int() >= VInt
		case VerifyStrArr[0] == "gt":
			return val.Int() > VInt
		default:
			return false
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		VInt, VErr := strconv.Atoi(VerifyStrArr[1])
		if VErr != nil {
			return false
		}
		switch {
		case VerifyStrArr[0] == "lt":
			return val.Uint() < uint64(VInt)
		case VerifyStrArr[0] == "le":
			return val.Uint() <= uint64(VInt)
		case VerifyStrArr[0] == "eq":
			return val.Uint() == uint64(VInt)
		case VerifyStrArr[0] == "ne":
			return val.Uint() != uint64(VInt)
		case VerifyStrArr[0] == "ge":
			return val.Uint() >= uint64(VInt)
		case VerifyStrArr[0] == "gt":
			return val.Uint() > uint64(VInt)
		default:
			return false
		}
	case reflect.Float32, reflect.Float64:
		VFloat, VErr := strconv.ParseFloat(VerifyStrArr[1], 64)
		if VErr != nil {
			return false
		}
		switch {
		case VerifyStrArr[0] == "lt":
			return val.Float() < VFloat
		case VerifyStrArr[0] == "le":
			return val.Float() <= VFloat
		case VerifyStrArr[0] == "eq":
			return val.Float() == VFloat
		case VerifyStrArr[0] == "ne":
			return val.Float() != VFloat
		case VerifyStrArr[0] == "ge":
			return val.Float() >= VFloat
		case VerifyStrArr[0] == "gt":
			return val.Float() > VFloat
		default:
			return false
		}
	default:
		return false
	}
}

func regexpMatch(rule, matchStr string) bool {
	return regexp.MustCompile(rule).MatchString(matchStr)
}

func VerifyParamType(param interface{}, kind reflect.Kind) (err error) {
	val := reflect.ValueOf(param) // get reflect.Type
	kd := val.Kind()              // Obtain the category corresponding to st
	if kd != kind {
		return errors.New("not expect params type")
	}
	if isBlank(val) {
		return errors.New("value is empty !")
	}
	return nil
}
