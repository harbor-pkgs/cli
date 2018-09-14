package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"regexp"
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

// Returns a curl command representation of the passed http.Request
func CurlString(req *http.Request, payload *[]byte) string {
	parts := []string{"curl", "-i", "-X", req.Method, req.URL.String()}
	for key, value := range req.Header {
		parts = append(parts, fmt.Sprintf("-H \"%s: %s\"", key, value[0]))
	}

	if payload != nil {
		parts = append(parts, fmt.Sprintf(" -d '%s'", string(*payload)))
	}
	return strings.Join(parts, " ")
}

// Given an indented string remove any common leading whitespace from every
// line in text. Works much like python's `textwrap.dedent()` function.
// (Mixing Spaces and Tabs will have undesired effects)
func Dedent(input string) string {
	text := []byte(input)

	// find the first \n::space:: combo
	leadingWhitespace := regexp.MustCompile(`(?m)^[ \t]+`)
	idx := leadingWhitespace.FindIndex(text)
	if idx == nil {
		fmt.Printf("Unable to find \\n::space:: combo\n")
		return input
	}
	//fmt.Printf("idx: '%d:%d'\n", idx[0], idx[1])

	// Create a regex to match any the number of spaces we first found
	gobbleRegex := fmt.Sprintf("(?m)^[ \t]{%d}?", (idx[1] - idx[0]))
	//fmt.Printf("gobbleRegex: '%s'\n", gobbleRegex)
	gobbleIndents := regexp.MustCompile(gobbleRegex)
	// Find any identical spaces and remove them
	dedented := gobbleIndents.ReplaceAll(text, []byte{})
	return string(dedented)
}

// Exactly like `Dedent()` but trims trailing `cutset` characters
func DedentTrim(cutset, input string) string {
	return strings.Trim(Dedent(input), cutset)
}

// Formats the text `msg` wrapping the text to the character `length` specified.
// Indenting the following lines `indent` number of spaces
func WordWrap(msg string, indent int, wordWrap int) string {
	// Remove any previous formatting
	regex, _ := regexp.Compile(` {2,}|\n|\t`)
	msg = regex.ReplaceAllString(msg, " ")
	if (wordWrap - indent) <= 0 {
		panic(fmt.Sprintf("indent spacing '%d' exceeds wordwrap length '%d'\n", indent, wordWrap))
	}

	if len(msg) < wordWrap {
		return msg
	}

	indentWord := strings.Repeat(" ", indent)
	remaining := wordWrap

	var words []string
	for i, word := range strings.Fields(msg) {
		if len(word)+1 > remaining {
			// Add a new line, our indent, word and the space
			words = append(words, "\n"+indentWord+word+" ")
			remaining = wordWrap - (len(word) + indent)

			// Since this word should be on the next line,
			// Trim the previous word of any space (if there is a prev word)
			if i > 0 {
				words[i-1] = strings.TrimSuffix(words[i-1], " ")
			}
		} else {
			// Regular word, just add a space
			words = append(words, word+" ")
			remaining = remaining - (len(word) + 1)
		}
	}
	return strings.TrimSuffix(strings.Join(words, ""), " ")
}

// Returns true if the file has ModeCharDevice set. This is useful when determining if
// a CLI is receiving piped data.
//
//   var contents []byte
//   var inputFile string
//   var err error
//
//   // If stdin is getting piped data, read from stdin
//   if args.IsCharDevice(os.Stdin) {
//       contents, err = ioutil.ReadAll(os.Stdin)
//   } else {
//       // load from file given instead
//       contents, err = ioutil.ReadFile(inputFile)
//   }
func IsCharDevice(file *os.File) bool {
	stat, err := file.Stat()
	if err != nil {
		return false
	}
	return (stat.Mode() & os.ModeCharDevice) == 0
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
// 	args.SetDefault(&config.Foo, "default")
// 	args.SetDefault(&config.Bar, 200)
//
// Supply additional default values and SetDefault will
// choose the first default that is not of zero value
//  args.SetDefault(&config.Foo, os.Getenv("FOO"), "default")
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
// 	args.IsZero(thingy) == true
func IsZero(value interface{}) bool {
	return IsZeroValue(reflect.ValueOf(value))
}

// Returns true if 'value' is zero (the default golang value)
//	var count int64
// 	args.IsZeroValue(reflect.ValueOf(count)) == true
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
