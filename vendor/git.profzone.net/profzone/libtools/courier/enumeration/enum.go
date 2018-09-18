package enumeration

import (
	"fmt"
	"strconv"
)

type EnumOption struct {
	Val   interface{} `json:"val"`
	Value interface{} `json:"value"`
	Label string      `json:"label"`
}

type Enum []EnumOption

var EnumMap = map[string]Enum{}

func RegisterEnums(enumType string, values map[string]string) {
	if _, ok := EnumMap[enumType]; ok {
		panic(fmt.Errorf("`%s` is already defined, please make enum unqiue in one service", enumType))
	}
	for value, label := range values {
		EnumMap[enumType] = append(EnumMap[enumType], EnumOption{Value: value, Label: label})
	}
}

//@deprecated
func RegisterEnum(enumType string, optionValue string, label string) {
	if _, ok := EnumMap[enumType]; !ok {
		EnumMap[enumType] = []EnumOption{}
	}
	EnumMap[enumType] = append(EnumMap[enumType], EnumOption{Value: optionValue, Label: label})
}

func GetEnumValueList(enumType string) (enumList []EnumOption, found bool) {
	enumList, found = EnumMap[enumType]
	return
}

func AsInt64(v interface{}, defaultInteger int64) (int64, error) {
	switch s := v.(type) {
	case []byte:
		if len(s) > 0 {
			i, err := strconv.ParseInt(string(s), 10, 64)
			if err != nil {
				return defaultInteger, err
			}
			return i, err
		}
		return defaultInteger, nil
	case string:
		if s != "" {
			i, err := strconv.ParseInt(s, 10, 64)
			if err != nil {
				return defaultInteger, err
			}
			return i, err
		}
		return defaultInteger, nil
	case int:
		return int64(s), nil
	case int8:
		return int64(s), nil
	case int16:
		return int64(s), nil
	case int32:
		return int64(s), nil
	case int64:
		return int64(s), nil
	case uint:
		return int64(s), nil
	case uint8:
		return int64(s), nil
	case uint16:
		return int64(s), nil
	case uint32:
		return int64(s), nil
	case uint64:
		return int64(s), nil
	case nil:
		return defaultInteger, nil
	default:
		return defaultInteger, nil
	}
}
