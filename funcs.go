package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
)

type ErrorFunc func(string)

func panicFunc(msg string) {
	panic(msg)
}

// ContainsString checks if a given slice of strings contains the provided string.
// If a modifier func is provided, it is called with the slice item before the comparation.
//      haystack := []string{"one", "Two", "Three"}
//	if slice.ContainsString(haystack, "two", strings.ToLower) {
//		// Do thing
// 	}
func ContainsString(s string, slice []string, modifier func(s string) string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
		if modifier != nil && modifier(item) == s {
			return true
		}
	}
	return false
}

// Given a comma separated string, return a slice of string items.
// Return the entire string as the first item if no comma is found.
func StringToSlice(value string, modifiers ...func(s string) string) []string {
	result := strings.Split(value, ",")
	// Apply the modifiers
	for _, modifier := range modifiers {
		for idx, item := range result {
			result[idx] = modifier(item)
		}
	}
	return result
}

// Given a comma separated string of key values in the form `key=value`.
// Return a map of key values as strings, Also excepts JSON for more complex
// quoted or escaped data.
func StringToMap(value string) (map[string]string, error) {
	tokenizer := newKeyValueTokenizer(value)
	result := make(map[string]string)

	var lvalue, rvalue, expression string
	for {
		lvalue = tokenizer.Next()
		if lvalue == "" {
			return result, fmt.Errorf("expected key at pos '%d' but found none; "+
				"map values should be 'key=value' separated by commas", tokenizer.Pos)
		}
		if strings.HasPrefix(lvalue, "{") {
			// Assume this is JSON format and attempt to un-marshal
			return jsonToMap(value)
		}

		expression = tokenizer.Next()
		if expression != "=" {
			return result, fmt.Errorf("expected '=' after '%s' but found '%s'; "+
				"map values should be 'key=value' separated by commas", lvalue, expression)
		}
		rvalue = tokenizer.Next()
		if rvalue == "" {
			return result, fmt.Errorf("expected value after '%s' but found none; "+
				"map values should be 'key=value' separated by commas", expression)
		}
		result[lvalue] = rvalue

		// Are there anymore tokens?
		delimiter := tokenizer.Next()
		if delimiter == "" {
			break
		}

		// Should be a comma next
		if delimiter != "," {
			return result, fmt.Errorf("expected ',' after '%s' but found '%s'; "+
				"map values should be 'key=value' separated by commas", rvalue, delimiter)
		}
	}
	return result, nil
}

func jsonToMap(value string) (map[string]string, error) {
	result := make(map[string]string)
	err := json.Unmarshal([]byte(value), &result)
	if err != nil {
		return result, errors.New(fmt.Sprintf("JSON map decoding for '%s' failed with '%s'; "+
			`JSON map values should be in form '{"key":"value", "foo":"bar"}'`, value, err))
	}
	return result, nil
}

// If 'value' is empty or of zero value, assign the default value.
// This panics if the value is not a pointer or if value and
// default value are not of the same type.
//      var config struct {
//		Verbose *bool
//		Foo string
//		Bar int
//	}
// 	cli.SetDefault(&config.Foo, "default")
// 	cli.SetDefault(&config.Bar, 200)
//
// Supply additional default values and SetDefault will
// choose the first default that is not of zero value
//  cli.SetDefault(&config.Foo, os.Getenv("FOO"), "default")
func SetDefault(dest interface{}, defaultValue ...interface{}) {
	d := reflect.ValueOf(dest)
	if d.Kind() != reflect.Ptr {
		panic("holster.SetDefault: Expected first argument to be of type reflect.Ptr")
	}
	d = reflect.Indirect(d)
	if IsZeroValue(d) {
		// Use the first non zero default value we find
		for _, value := range defaultValue {
			v := reflect.ValueOf(value)
			if !IsZeroValue(v) {
				d.Set(reflect.ValueOf(value))
				return
			}
		}
	}
}

// Returns true if 'value' is zero (the default golang value)
//	var thingy string
// 	holster.IsZero(thingy) == true
func IsZero(value interface{}) bool {
	return IsZeroValue(reflect.ValueOf(value))
}

// Returns true if 'value' is zero (the default golang value)
//	var count int64
// 	holster.IsZeroValue(reflect.ValueOf(count)) == true
func IsZeroValue(value reflect.Value) bool {
	switch value.Kind() {
	case reflect.Array, reflect.String:
		return value.Len() == 0
	case reflect.Bool:
		return !value.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return value.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return value.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return value.Float() == 0
	case reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		return value.IsNil()
	}
	return false
}

// Returns true if the error was because help flag was found during parsing
func IsHelpError(err error) bool {
	obj, ok := err.(isHelpError)
	return ok && obj.IsHelpError()
}

type isHelpError interface {
	IsHelpError() bool
}

type HelpError struct{}

func (e *HelpError) Error() string {
	return "user asked for help; inspect this error with cli.isHelpError()"
}

func (e *HelpError) IsHelpError() bool {
	return true
}
