package cli

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// Methods here store the parsed values in the pointer provided, while preforming type conversions
// from the 3 basic kinds (ScalarKind, SliceKind, MapKind) as provided by a `Store.Get()`.
// They do not preform any assertion checks since the parser does this when retrieving values
// from the store.

func toSet(ptr *bool) StoreFunc {
	return func(value interface{}, count int) error {
		if count != 0 {
			*ptr = true
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

// String

func toString(o interface{}) StoreFunc {
	ptr := o.(*string)
	return func(value interface{}, count int) error {
		*ptr = value.(string)
		return nil
	}
}

func toStringSlice(o interface{}) StoreFunc {
	ptr := o.(*[]string)
	return func(value interface{}, count int) error {
		*ptr = value.([]string)
		return nil
	}
}

func toStringMap(o interface{}) StoreFunc {
	ptr := o.(*map[string]string)
	return func(value interface{}, count int) error {
		*ptr = value.(map[string]string)
		return nil
	}
}

// Bool

func toBool(o interface{}) StoreFunc {
	ptr := o.(*bool)
	return func(value interface{}, count int) error {
		b, err := ToBool(value.(string))
		if err != nil {
			return err
		}
		*ptr = b
		return nil
	}
}

func toBoolSlice(o interface{}) StoreFunc {
	ptr := o.(*[]bool)
	return func(value interface{}, count int) error {
		var r []bool
		for _, item := range value.([]string) {
			b, err := ToBool(strings.TrimSpace(item))
			if err != nil {
				return err
			}
			r = append(r, b)
		}
		*ptr = r
		return nil
	}
}

func toBoolMap(o interface{}) StoreFunc {
	ptr := o.(*map[string]bool)
	return func(value interface{}, count int) error {
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
		return nil
	}
}

// Int

func toInt(o interface{}) StoreFunc {
	ptr := o.(*int)
	return func(value interface{}, count int) error {
		i, err := strconv.ParseInt(value.(string), 0, strconv.IntSize)
		if err != nil {
			return fmt.Errorf("'%s' is not an integer", value.(string))
		}
		*ptr = int(i)
		return nil
	}
}

func toIntSlice(o interface{}) StoreFunc {
	ptr := o.(*[]int)
	return func(value interface{}, count int) error {
		var r []int
		for _, item := range value.([]string) {
			i, err := strconv.ParseInt(strings.TrimSpace(item), 0, strconv.IntSize)
			if err != nil {
				return fmt.Errorf("'%s' is not an integer", item)
			}
			r = append(r, int(i))
		}
		*ptr = r
		return nil
	}
}

func toIntMap(o interface{}) StoreFunc {
	ptr := o.(*map[string]int)
	return func(value interface{}, count int) error {
		strMap := value.(map[string]string)
		result := make(map[string]int, len(strMap))
		for k, v := range strMap {
			i, err := strconv.ParseInt(strings.TrimSpace(v), 0, strconv.IntSize)
			if err != nil {
				return fmt.Errorf("'%s' is not an integer", v)
			}
			result[k] = int(i)
		}
		*ptr = result
		return nil
	}
}

// Uint

func toUint(o interface{}) StoreFunc {
	ptr := o.(*uint)
	return func(value interface{}, count int) error {
		i, err := strconv.ParseInt(value.(string), 0, strconv.IntSize)
		if err != nil {
			return fmt.Errorf("'%s' is not an integer", value.(string))
		}
		*ptr = uint(i)
		return nil
	}
}

func toUintSlice(o interface{}) StoreFunc {
	ptr := o.(*[]uint)
	return func(value interface{}, count int) error {
		var r []uint
		for _, item := range value.([]string) {
			i, err := strconv.ParseInt(strings.TrimSpace(item), 0, strconv.IntSize)
			if err != nil {
				return fmt.Errorf("'%s' is not an integer", item)
			}
			r = append(r, uint(i))
		}
		*ptr = r
		return nil
	}
}

func toUintMap(o interface{}) StoreFunc {
	ptr := o.(*map[string]uint)
	return func(value interface{}, count int) error {
		strMap := value.(map[string]string)
		result := make(map[string]uint, len(strMap))
		for k, v := range strMap {
			i, err := strconv.ParseInt(strings.TrimSpace(v), 0, strconv.IntSize)
			if err != nil {
				return fmt.Errorf("'%s' is not an integer", v)
			}
			result[k] = uint(i)
		}
		*ptr = result
		return nil
	}
}

// Int64

func toInt64(o interface{}) StoreFunc {
	ptr := o.(*int64)
	return func(value interface{}, count int) error {
		i, err := strconv.ParseInt(value.(string), 0, 64)
		if err != nil {
			return fmt.Errorf("'%s' is not an integer", value.(string))
		}
		*ptr = i
		return nil
	}
}

func toInt64Slice(o interface{}) StoreFunc {
	ptr := o.(*[]int64)
	return func(value interface{}, count int) error {
		var r []int64
		for _, item := range value.([]string) {
			i, err := strconv.ParseInt(strings.TrimSpace(item), 0, 64)
			if err != nil {
				return fmt.Errorf("'%s' is not an integer", item)
			}
			r = append(r, i)
		}
		*ptr = r
		return nil
	}
}

func toInt64Map(o interface{}) StoreFunc {
	ptr := o.(*map[string]int64)
	return func(value interface{}, count int) error {
		strMap := value.(map[string]string)
		result := make(map[string]int64, len(strMap))
		for k, v := range strMap {
			i, err := strconv.ParseInt(strings.TrimSpace(v), 0, 64)
			if err != nil {
				return fmt.Errorf("'%s' is not an integer", v)
			}
			result[k] = i
		}
		*ptr = result
		return nil
	}
}

// Uint64

func toUint64(o interface{}) StoreFunc {
	ptr := o.(*uint64)
	return func(value interface{}, count int) error {
		i, err := strconv.ParseInt(value.(string), 0, 64)
		if err != nil {
			return fmt.Errorf("'%s' is not an integer", value.(string))
		}
		*ptr = uint64(i)
		return nil
	}
}

func toUint64Slice(o interface{}) StoreFunc {
	ptr := o.(*[]uint64)
	return func(value interface{}, count int) error {
		var r []uint64
		for _, item := range value.([]string) {
			i, err := strconv.ParseInt(strings.TrimSpace(item), 0, 64)
			if err != nil {
				return fmt.Errorf("'%s' is not an integer", item)
			}
			r = append(r, uint64(i))
		}
		*ptr = r
		return nil
	}
}

func toUint64Map(o interface{}) StoreFunc {
	ptr := o.(*map[string]uint64)
	return func(value interface{}, count int) error {
		strMap := value.(map[string]string)
		result := make(map[string]uint64, len(strMap))
		for k, v := range strMap {
			i, err := strconv.ParseInt(strings.TrimSpace(v), 0, 64)
			if err != nil {
				return fmt.Errorf("'%s' is not an integer", v)
			}
			result[k] = uint64(i)
		}
		*ptr = result
		return nil
	}
}

// Float64

func toFloat64(o interface{}) StoreFunc {
	ptr := o.(*float64)
	return func(value interface{}, count int) error {
		i, err := strconv.ParseFloat(value.(string), 64)
		if err != nil {
			return fmt.Errorf("'%s' is not an integer", value.(string))
		}
		*ptr = i
		return nil
	}
}

func toFloat64Slice(o interface{}) StoreFunc {
	ptr := o.(*[]float64)
	return func(value interface{}, count int) error {
		var r []float64
		for _, item := range value.([]string) {
			i, err := strconv.ParseFloat(strings.TrimSpace(item), 64)
			if err != nil {
				return fmt.Errorf("'%s' is not an integer", item)
			}
			r = append(r, i)
		}
		*ptr = r
		return nil
	}
}

func toFloat64Map(o interface{}) StoreFunc {
	ptr := o.(*map[string]float64)
	return func(value interface{}, count int) error {
		strMap := value.(map[string]string)
		result := make(map[string]float64, len(strMap))
		for k, v := range strMap {
			i, err := strconv.ParseFloat(strings.TrimSpace(v), 64)
			if err != nil {
				return fmt.Errorf("'%s' is not an integer", v)
			}
			result[k] = i
		}
		*ptr = result
		return nil
	}
}

// Duration

func toDuration(o interface{}) StoreFunc {
	ptr := o.(*time.Duration)
	return func(value interface{}, count int) error {
		i, err := time.ParseDuration(value.(string))
		if err != nil {
			return fmt.Errorf("'%s' is not a valid duration", value.(string))
		}
		*ptr = i
		return nil
	}
}

func toDurationSlice(o interface{}) StoreFunc {
	ptr := o.(*[]time.Duration)
	return func(value interface{}, count int) error {
		var r []time.Duration
		for _, item := range value.([]string) {
			i, err := time.ParseDuration(strings.TrimSpace(item))
			if err != nil {
				return fmt.Errorf("'%s' is not a valid duration", item)
			}
			r = append(r, i)
		}
		*ptr = r
		return nil
	}
}

func toDurationMap(o interface{}) StoreFunc {
	ptr := o.(*map[string]time.Duration)
	return func(value interface{}, count int) error {
		strMap := value.(map[string]string)
		result := make(map[string]time.Duration, len(strMap))
		for k, v := range strMap {
			i, err := time.ParseDuration(strings.TrimSpace(v))
			if err != nil {
				return fmt.Errorf("'%s' is not a valid duration", v)
			}
			result[k] = i
		}
		*ptr = result
		return nil
	}
}

var scalars = map[reflect.Type]func(interface{}) StoreFunc{
	reflect.TypeOf(""):               toString,
	reflect.TypeOf(true):             toBool,
	reflect.TypeOf(int(0)):           toInt,
	reflect.TypeOf(uint(0)):          toUint,
	reflect.TypeOf(int64(0)):         toInt64,
	reflect.TypeOf(uint64(0)):        toUint64,
	reflect.TypeOf(float64(0)):       toFloat64,
	reflect.TypeOf(time.Duration(0)): toDuration,
}

var slices = map[reflect.Type]func(interface{}) StoreFunc{
	reflect.TypeOf(""):               toStringSlice,
	reflect.TypeOf(true):             toBoolSlice,
	reflect.TypeOf(int(0)):           toIntSlice,
	reflect.TypeOf(uint(0)):          toUintSlice,
	reflect.TypeOf(int64(0)):         toInt64Slice,
	reflect.TypeOf(uint64(0)):        toUint64Slice,
	reflect.TypeOf(float64(0)):       toFloat64Slice,
	reflect.TypeOf(time.Duration(0)): toDurationSlice,
}

var maps = map[reflect.Type]func(interface{}) StoreFunc{
	reflect.TypeOf(""):               toStringMap,
	reflect.TypeOf(true):             toBoolMap,
	reflect.TypeOf(int(0)):           toIntMap,
	reflect.TypeOf(uint(0)):          toUintMap,
	reflect.TypeOf(int64(0)):         toInt64Map,
	reflect.TypeOf(uint64(0)):        toUint64Map,
	reflect.TypeOf(float64(0)):       toFloat64Map,
	reflect.TypeOf(time.Duration(0)): toDurationMap,
}

func newStoreFunc(r *rule, dest interface{}) error {
	// If the dest conforms to the SetValue interface
	if sv, ok := dest.(SetValue); ok {
		fn := func(value interface{}, count int) error {
			values := value.([]string)
			for _, v := range values {
				if err := sv.Set(v); err != nil {
					return err
				}
			}
			return nil
		}
		r.SetFlag(SliceKind, true)
		r.StoreFuncs = append(r.StoreFuncs, fn)
		r.Usage = "<string>"
		return nil
	}

	d := reflect.ValueOf(dest)
	if d.Kind() != reflect.Ptr {
		return fmt.Errorf("cannot use non pointer type '%s'; must provide a pointer", reflect.TypeOf(dest))
	}

	// Dereference the pointer
	d = reflect.Indirect(d)

	// Determine if it's a scalar, slice or map
	switch d.Kind() {
	case reflect.Slice:
		elem := reflect.TypeOf(dest).Elem().Elem()
		fn, ok := slices[elem]
		if !ok {
			return fmt.Errorf("cannot store '[]%s'; only "+
				"'%s' currently supported", elem.Kind(), supportedSlices())
		}

		r.SetFlag(SliceKind, true)
		r.StoreFuncs = append(r.StoreFuncs, fn(dest))
		r.Usage = fmt.Sprintf("<%[1]s>,<%[1]s>", elem.String())
		return nil
	case reflect.Map:
		elem := d.Type().Elem()
		key := d.Type().Key()

		if key.Kind() != reflect.String {
			return fmt.Errorf("cannot use 'map[%s]%s'; only "+
				"'%s' currently supported", key.Kind(), elem.Kind(), supportedMaps())
		}

		fn, ok := maps[elem]
		if !ok {
			return fmt.Errorf("cannot use 'map[%s]%s'; only "+
				"'%s' currently supported", key.Kind(), elem.Kind(), supportedMaps())
		}

		r.SetFlag(MapKind, true)
		r.StoreFuncs = append(r.StoreFuncs, fn(dest))
		r.Usage = fmt.Sprintf("<string>=<%s>", elem.String())
		return nil
	}

	// Handle Scalars
	fn, ok := scalars[d.Type()]
	if !ok {
		// Slightly less confusing error for those attempting to use arrays
		if d.Kind() == reflect.Array {
			return fmt.Errorf("cannot store '%s'; only slices supported", d.Type().String())
		}
		return fmt.Errorf("cannot store '%s'; type not supported", d.Type().String())
	}

	r.SetFlag(ScalarKind, true)
	r.StoreFuncs = append(r.StoreFuncs, fn(dest))
	r.Usage = fmt.Sprintf("<%s>", d.Kind().String())
	return nil
}

func supportedMaps() string {
	var results []string
	for k := range maps {
		results = append(results, fmt.Sprintf("map[string]%s", k.String()))
	}
	return strings.Join(results, ", ")
}

func supportedSlices() string {
	var results []string
	for k := range maps {
		results = append(results, fmt.Sprintf("[]%s", k.String()))
	}
	return strings.Join(results, ", ")
}
