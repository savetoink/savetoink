package task

import (
	"errors"
	"fmt"
)

// ParamValue represents a task parameter value.
type ParamValue struct {
	stringVal *string
}

// StringParam creates a ParamValue from a string.
func StringParam(v string) ParamValue {
	return ParamValue{stringVal: &v}
}

// String returns the string value, or an error if nil.
func (p ParamValue) String() (string, error) {
	if p.stringVal == nil {
		return "", errors.New("parameter is nil")
	}
	return *p.stringVal, nil
}

// RequireString returns the required string parameter with the given key,
// or an error if the parameter is not found.
func RequireString(params map[string]ParamValue, key string) (string, error) {
	val, ok := params[key]
	if !ok {
		return "", fmt.Errorf("required parameter '%s' not found", key)
	}
	return val.String()
}

// OptionalString returns the optional string parameter with the given key,
// or the default value if the parameter is not found.
func OptionalString(params map[string]ParamValue, key, def string) (string, error) {
	val, ok := params[key]
	if !ok {
		return def, nil
	}
	return val.String()
}
