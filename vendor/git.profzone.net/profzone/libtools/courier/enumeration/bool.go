package enumeration

import (
	"encoding"
	"encoding/json"
	"errors"
	"strings"
)

// swagger:enum
type Bool uint8

var _ interface {
	json.Unmarshaler
	json.Marshaler
} = (*Bool)(nil)

const (
	BOOL_UNKNOWN Bool = iota
	BOOL__TRUE        // true
	BOOL__FALSE       // false
)

func (v Bool) True() bool {
	return v == BOOL__TRUE
}

func (Bool) Enums() map[int][]string {
	return map[int][]string{
		int(BOOL__TRUE):  {"TRUE", "true"},
		int(BOOL__FALSE): {"FALSE", "false"},
	}
}

func BoolFromBool(b bool) Bool {
	if b {
		return BOOL__TRUE
	}
	return BOOL__FALSE
}

var InvalidBool = errors.New("invalid Bool")

func ParseBoolFromString(s string) (Bool, error) {
	switch s {
	case "":
		return BOOL_UNKNOWN, nil
	case "FALSE":
		return BOOL__FALSE, nil
	case "TRUE":
		return BOOL__TRUE, nil
	}
	return BOOL_UNKNOWN, InvalidBool
}

func ParseBoolFromLabelString(s string) (Bool, error) {
	switch s {
	case "":
		return BOOL_UNKNOWN, nil
	case "false":
		return BOOL__FALSE, nil
	case "true":
		return BOOL__TRUE, nil
	}
	return BOOL_UNKNOWN, InvalidBool
}

func (v Bool) String() string {
	switch v {
	case BOOL_UNKNOWN:
		return ""
	case BOOL__FALSE:
		return "FALSE"
	case BOOL__TRUE:
		return "TRUE"
	}
	return "UNKNOWN"
}

func (v Bool) Label() string {
	switch v {
	case BOOL_UNKNOWN:
		return ""
	case BOOL__FALSE:
		return "false"
	case BOOL__TRUE:
		return "true"
	}
	return "UNKNOWN"
}

var _ interface {
	encoding.TextMarshaler
	encoding.TextUnmarshaler
	json.Marshaler
	json.Unmarshaler
} = (*Bool)(nil)

func (v Bool) MarshalText() ([]byte, error) {
	str := v.Label()
	if str == "" {
		return []byte("null"), nil
	}
	if str == "UNKNOWN" {
		return nil, InvalidBool
	}
	return []byte(str), nil
}

func (v *Bool) UnmarshalText(data []byte) (err error) {
	s := strings.ToLower(strings.Trim(string(data), `"`))
	if s == "null" {
		*v = BOOL_UNKNOWN
		return
	}
	*v, err = ParseBoolFromLabelString(s)
	return
}

func (v Bool) MarshalJSON() ([]byte, error) {
	return v.MarshalText()
}

func (v *Bool) UnmarshalJSON(data []byte) (err error) {
	return v.UnmarshalText(data)
}
