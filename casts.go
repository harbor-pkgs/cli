package cli

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

func toInt(ptr *int) StoreFunc {
	return func(value interface{}, count int) error {
		switch t := value.(type) {
		case string:
			i, err := strconv.ParseInt(t, 10, 32)
			if err != nil {
				return fmt.Errorf("as an integer: %s", err)
			}
			*ptr = int(i)
			return nil
		case []string:
			*ptr = len(t)
		case map[string]string:
			*ptr = len(t)
		}
		return nil
	}
}

func toExists(ptr *bool) StoreFunc {
	return func(value interface{}, count int) error {
		if count != 0 {
			*ptr = true
		}
		return nil
	}
}

func toBool(ptr *bool) StoreFunc {
	return func(value interface{}, count int) error {
		switch t := value.(type) {
		case string:
			b, err := strconv.ParseBool(t)
			if err != nil {
				return fmt.Errorf("as a boolean: %s", err)
			}
			*ptr = b
		case []string, map[string]string:
			*ptr = true
		}
		return nil
	}
}

func toString(ptr *string) StoreFunc {
	return func(value interface{}, count int) error {
		switch t := value.(type) {
		case string:
			*ptr = t
		case []string, map[string]string:
			b, err := json.Marshal(t)
			if err != nil {
				return fmt.Errorf("while marshalling '%s' into JSON: %s", t, err)
			}
			*ptr = string(b)
		}
		return nil
	}
}

func toCount(ptr *int) StoreFunc {
	return func(value interface{}, count int) error {
		*ptr = count
		return nil
	}
}

func toStringSlice(ptr *[]string) StoreFunc {
	return func(value interface{}, count int) error {
		switch t := value.(type) {
		case string:
			*ptr = StringToSlice(t, strings.TrimSpace)
		case []string:
			*ptr = t
		case map[string]string:
			var r []string
			for k, v := range t {
				r = append(r, fmt.Sprintf("%s=%s", k, v))
			}
			*ptr = r
		}
		return nil
	}
}

func toIntSlice(ptr *[]int) StoreFunc {
	strListToIntList := func(slice []string) ([]int, error) {
		var r []int
		for _, item := range slice {
			i, err := strconv.ParseInt(strings.TrimSpace(item), 10, 32)
			if err != nil {
				return nil, fmt.Errorf("as an integer '%s' in slice %s", item, err)
			}
			r = append(r, int(i))
		}
		return r, nil
	}

	return func(value interface{}, count int) error {
		switch t := value.(type) {
		case string:
			r, err := strListToIntList(strings.Split(t, ","))
			if err != nil {
				return err
			}
			*ptr = r
		case []string:
			r, err := strListToIntList(t)
			if err != nil {
				return err
			}
			*ptr = r
		}
		return nil
	}
}
