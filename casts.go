package cli

import (
	"fmt"
	"strconv"
	"strings"
)

func toInt(ptr *int) StoreFunc {
	return func(value string, count int) error {
		i, err := strconv.ParseInt(value, 10, 32)
		if err != nil {
			return fmt.Errorf("as an integer: %s", err)
		}
		*ptr = int(i)
		return nil
	}
}

func toBool(ptr *bool) StoreFunc {
	return func(value string, count int) error {
		if count != 0 {
			*ptr = true
		}
		return nil
	}
}

func toStoreBool(ptr *bool) StoreFunc {
	return func(value string, count int) error {
		b, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("as a boolean: %s", err)
		}
		*ptr = b
		return nil
	}
}

func toString(ptr *string) StoreFunc {
	return func(value string, count int) error {
		*ptr = value
		return nil
	}
}

func toCount(ptr *int) StoreFunc {
	return func(value string, count int) error {
		*ptr = count
		return nil
	}
}

func toStringSlice(ptr []string) StoreFunc {
	return func(value string, count int) error {
		ptr = StringToSlice(value, strings.TrimSpace)
		return nil
	}
}

func toIntSlice(ptr []int) StoreFunc {
	return func(value string, count int) error {
		result := strings.Split(value, ",")
		for _, item := range result {
			i, err := strconv.ParseInt(strings.TrimSpace(item), 10, 32)
			if err != nil {
				return fmt.Errorf("as an integer '%s' in slice %s", item, err)
			}
			ptr = append(ptr, int(i))
		}
		return nil
	}
}
