package templatex

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"
)

func fmHasPrefix(s, prefix string) bool {
	return strings.HasPrefix(s, prefix)
}

func fmHasSuffix(s, suffix string) bool {
	return strings.HasSuffix(s, suffix)
}

func fmKeys(v interface{}) ([]interface{}, error) {
	ref := reflect.Indirect(reflect.ValueOf(v))
	switch ref.Kind() {
	case reflect.Slice:
		n := ref.Len()
		keys := make([]interface{}, n)
		for i := 0; i < n; i++ {
			keys[i] = i
		}
		return keys, nil

	case reflect.Map:
		n := ref.Len()
		keys := make([]interface{}, n)
		iter := ref.MapRange()
		i := 0
		for iter.Next() {
			keys[i] = iter.Key().Interface()
			i++
		}
		return keys, nil
	}

	return nil, &reflect.ValueError{
		Method: "Keys",
		Kind:   ref.Kind(),
	}
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

func fmEquals(x interface{}, v ...interface{}) bool {
	var xi interface{}
	if ref := reflect.Indirect(reflect.ValueOf(x)); ref.IsValid() {
		xi = ref.Interface()
	}

	for i, n := 0, len(v); i < n; i++ {
		var vi interface{}
		if ref := reflect.Indirect(reflect.ValueOf(v[i])); ref.IsValid() {
			vi = ref.Interface()
		}
		if reflect.DeepEqual(xi, vi) {
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
		"HasPrefix": fmHasPrefix,
		"HasSuffix": fmHasSuffix,
		"Keys":      fmKeys,
		"ToSlice":   fmToSlice,
		"Sort":      fmSort,
		"Equals":    fmEquals,
		"Sub":       fmSub,
		"JSON2Map":  fmJSON2Map,
		"ToJSON":    fmToJSON,
		"Prefix":    fmPrefix,
		"Suffix":    fmSuffix,
		// helper data structure
		"NewHashSet": fmNewHashSet,
	}
}
