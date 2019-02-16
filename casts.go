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
				return fmt.Errorf("'%s' is not an integer", t)
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

func toSet(ptr *bool) StoreFunc {
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
			b, err := ToBool(t)
			if err != nil {
				return err
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
			*ptr = ToSlice(t, strings.TrimSpace)
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

func toStringMap(ptr *map[string]string) StoreFunc {
	return func(value interface{}, count int) error {
		var err error
		switch t := value.(type) {
		case string:
			*ptr, err = ToStringMap(t)
			if err != nil {
				return err
			}
		case map[string]string:
			*ptr = t
		}
		return nil
	}
}

func toIntMap(ptr *map[string]int) StoreFunc {
	return func(value interface{}, count int) error {
		var err error
		switch t := value.(type) {
		case string:
			*ptr, err = ToIntMap(t)
			if err != nil {
				return err
			}
		case map[string]string:
			strMap := value.(map[string]string)
			result := make(map[string]int, len(strMap))
			for k, v := range strMap {
				i, err := strconv.ParseInt(strings.TrimSpace(v), 10, 32)
				if err != nil {
					return fmt.Errorf("'%s' is not an integer", v)
				}
				result[k] = int(i)
			}
			*ptr = result
		}
		return nil
	}
}

func toBoolMap(ptr *map[string]bool) StoreFunc {
	return func(value interface{}, count int) error {
		var err error
		switch t := value.(type) {
		case string:
			*ptr, err = ToBoolMap(t)
			if err != nil {
				return err
			}
		case map[string]string:
			strMap := value.(map[string]string)
			result := make(map[string]bool, len(strMap))
			for k, v := range strMap {
				b, err := ToBool(strings.TrimSpace(v))
				if err != nil {
					return err
				}
				result[k] = b
			}
			*ptr = result
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
				return nil, fmt.Errorf("'%s' is not an integer", item)
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

func toBoolSlice(ptr *[]bool) StoreFunc {
	strListToBoolList := func(slice []string) ([]bool, error) {
		var r []bool
		for _, item := range slice {
			b, err := ToBool(strings.TrimSpace(item))
			if err != nil {
				return nil, err
			}
			r = append(r, b)
		}
		return r, nil
	}

	return func(value interface{}, count int) error {
		switch t := value.(type) {
		case string:
			r, err := strListToBoolList(strings.Split(t, ","))
			if err != nil {
				return err
			}
			*ptr = r
		case []string:
			r, err := strListToBoolList(t)
			if err != nil {
				return err
			}
			*ptr = r
		}
		return nil
	}
}
