package templatex

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
)

func indirectInterface(v reflect.Value) reflect.Value {
	if v.Kind() != reflect.Interface {
		return v
	} else if v.IsNil() {
		return reflect.Value{}
	}
	return v.Elem()
}

func fmKeys(arg reflect.Value) []interface{} {
	v := indirectInterface(arg)
	vk := v.Kind()
	if !v.IsValid() || vk != reflect.Map {
		panic(&reflect.ValueError{
			Method: "Keys",
			Kind:   vk,
		})
	}
	n := v.Len()
	keys := make([]interface{}, 0, n)
	iter := v.MapRange()
	for iter.Next() {
		k := iter.Key()
		keys = append(keys, k.Interface())
	}

	return keys
}

func fmToSlice(v ...interface{}) []interface{} {
	return v
}

func fmSort(arg []interface{}) []interface{} {
	sort.Slice(arg, func(i, j int) bool {
		iv := reflect.ValueOf(arg[i])
		jv := reflect.ValueOf(arg[j])
		if !iv.IsValid() || !jv.IsValid() {
			return iv.IsValid() == jv.IsValid()
		} else if iv.Type() == jv.Type() {
			switch iv.Kind() {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				return iv.Int() < jv.Int()

			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
				return iv.Uint() < jv.Uint()

			case reflect.Float32, reflect.Float64:
				return iv.Float() < jv.Float()

			case reflect.String:
				return iv.String() < jv.String()
			}
		}
		return false
	})

	return arg
}

func fmEquals(x interface{}, y ...interface{}) bool {
	for i, n := 0, len(y); i < n; i++ {
		if reflect.DeepEqual(x, y[i]) {
			return true
		}
	}
	return false
}

func fmSub(c ...int) int {
	if len(c) > 1 {
		return c[0] - c[1]
	}
	return c[0] - 1
}

// JSON2Map is helper function for web user interface prototyping
func fmJSON2Map(src string) (interface{}, error) {
	var data interface{}
	return data, json.Unmarshal([]byte(src), &data)
}

func fmToJSON(v interface{}, opts ...string) (string, error) {
	indent := ""
	switch len(opts) {
	case 0:
	case 1:
		indent = opts[0]
	default:
		return "", fmt.Errorf("too many arguments: %#v", opts)
	}

	b, err := json.MarshalIndent(v, "", indent)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// extract first n characters from src
func fmPrefix(src string, n int) string {
	r := []rune(src)
	if n <= 0 {
		return ""
	} else if n > len(r) {
		return src
	}
	return string(r[:n])
}

// extract last n characters from src
func fmSuffix(src string, n int) string {
	r := []rune(src)
	l := len(r)
	if n <= 0 {
		return ""
	} else if n > l {
		return src
	}
	return string(r[l-n:])
}

/*
	helper data structures
*/
type fmHashSet struct {
	data map[interface{}]bool
}

func fmNewHashSet() *fmHashSet {
	return &fmHashSet{
		data: map[interface{}]bool{},
	}
}

func (s *fmHashSet) Set(v interface{}) bool {
	if _, exists := s.data[v]; exists {
		return false
	}
	s.data[v] = true
	return true
}

func (s *fmHashSet) Unset(v interface{}) bool {
	if _, exists := s.data[v]; exists {
		delete(s.data, v)
		return true
	}
	return false
}

func DefaultFuncMap() map[string]interface{} {
	return map[string]interface{}{
		// functions
		"Keys":     fmKeys,
		"ToSlice":  fmToSlice,
		"Sort":     fmSort,
		"Equal":    fmEquals,
		"Sub":      fmSub,
		"JSON2Map": fmJSON2Map,
		"ToJSON":   fmToJSON,
		"Prefix":   fmPrefix,
		"Suffix":   fmSuffix,
		// helper data structure
		"NewHashSet": fmNewHashSet,
	}
}
