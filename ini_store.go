package cli

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
)

type keyValue struct {
	Key   string
	Value string
}

type INIStore struct {
	values map[string][]keyValue
}

func NewIniStore(r io.Reader) (FromStore, error) {
	contents, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	values := make(map[string][]keyValue)
	lines := strings.Split(string(contents), "\n")
	var count int
	for _, line := range lines {
		// Skip comments or malformed lines
		if len(line) == 0 || line[0] == '#' || line[0] == ' ' || line[0] == '\n' {
			continue
		}

		// Split the line into key=value or "key"="value"
		var inQuote bool
		var eRune rune
		parts := strings.FieldsFunc(line, func(c rune) bool {
			switch {
			case c == '"' || c == '\'':
				// If we are inside a quote
				if inQuote {
					// And we match the same rune that began the quote
					if c == eRune {
						inQuote = !inQuote
					}
				} else {
					// If we are outside a quote, record the start rune
					inQuote = !inQuote
					eRune = c
				}
				return false
			case inQuote:
				return false
			default:
				switch c {
				case '=':
					return true
				case '\n':
					return true
				default:
					return false
				}
			}
		})

		if len(parts) == 0 {
			return nil, fmt.Errorf("unknown parsing error on line '%d'", count)
		}

		// Determine if the key has a value
		var key, value string
		if len(parts) > 1 {
			key = strings.Trim(parts[0], `"'`)
			value = strings.Trim(parts[1], `"'`)
		} else {
			key = parts[0]
		}

		// Append or set the value
		if _, ok := values[key]; ok {
			values[key] = append(values[key], keyValue{Key: key, Value: value})
		} else {
			values[key] = []keyValue{{Key: key, Value: value}}
		}
	}
	fmt.Printf("Values: %+v\n", values)
	return &INIStore{values: values}, nil
}

func (kv *INIStore) Get(ctx context.Context, key string, kind Kind) (interface{}, int, error) {
	kvs, ok := kv.values[key]
	if !ok {
		return "", 0, nil
	}

	// Collect the values if any from our keys
	var values []string
	var count int
	for _, kv := range kvs {
		count++
		values = append(values, kv.Value)
	}

	if len(values) == 0 {
		return nil, count, nil
	}

	return sliceToKind(values, kind, count)
}

func (kv *INIStore) Source() string {
	// TODO: Make this something users can reference
	return "key-value-store"
}
