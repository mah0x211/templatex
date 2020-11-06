package builtins

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"
)

func Not(args ...interface{}) bool {
	for _, a := range args {
		v := reflect.ValueOf(a)
		if v.IsValid() && !v.IsZero() {
			return false
		}
	}
	return true
}

func HasPrefix(s, prefix string) bool {
	return strings.HasPrefix(s, prefix)
}

func HasSuffix(s, suffix string) bool {
	return strings.HasSuffix(s, suffix)
}

func Keys(v interface{}) ([]interface{}, error) {
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

func ToSlice(v ...interface{}) []interface{} {
	return v
}

func Sort(arg []interface{}) []interface{} {
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

func Equals(x interface{}, v ...interface{}) bool {
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

func Sub(c ...int) int {
	if len(c) > 1 {
		return c[0] - c[1]
	}
	return c[0] - 1
}

// JSON2Map is helper function for web user interface prototyping
func JSON2Map(src string) (interface{}, error) {
	var data interface{}
	return data, json.Unmarshal([]byte(src), &data)
}

func ToJSON(v interface{}, opts ...string) (string, error) {
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
func Prefix(src string, n int) string {
	r := []rune(src)
	if n <= 0 {
		return ""
	} else if n > len(r) {
		return src
	}
	return string(r[:n])
}

// extract last n characters from src
func Suffix(src string, n int) string {
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
type HashSet struct {
	data map[interface{}]bool
}

func NewHashSet() *HashSet {
	return &HashSet{
		data: map[interface{}]bool{},
	}
}

func (s *HashSet) Set(v interface{}) bool {
	if _, exists := s.data[v]; exists {
		return false
	}
	s.data[v] = true
	return true
}

func (s *HashSet) Unset(v interface{}) bool {
	if _, exists := s.data[v]; exists {
		delete(s.data, v)
		return true
	}
	return false
}

func FuncMap() map[string]interface{} {
	return map[string]interface{}{
		// functions
		"Not":       Not,
		"HasPrefix": HasPrefix,
		"HasSuffix": HasSuffix,
		"Keys":      Keys,
		"ToSlice":   ToSlice,
		"Sort":      Sort,
		"Equals":    Equals,
		"Sub":       Sub,
		"JSON2Map":  JSON2Map,
		"ToJSON":    ToJSON,
		"Prefix":    Prefix,
		"Suffix":    Suffix,
		// helper data structure
		"NewHashSet": NewHashSet,
	}
}
