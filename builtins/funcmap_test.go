package builtins

import (
	"reflect"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func equalFunc(t *testing.T, exp, act interface{}) bool {
	ef := reflect.ValueOf(exp)
	if ef.Kind() != reflect.Func {
		t.Fatalf("expected value is not function")
	}

	af := reflect.ValueOf(act)
	if af.Kind() != reflect.Func {
		t.Fatalf("actual value is not function")
	}

	return ef.Pointer() == af.Pointer()
}

func Test_FuncMap(t *testing.T) {
	// test that returns built-in funcmap
	for k, v := range FuncMap() {
		switch k {
		case "Not":
			equalFunc(t, v, Not)
		case "HasPrefix":
			equalFunc(t, v, HasPrefix)
		case "HasSuffix":
			equalFunc(t, v, HasSuffix)
		case "Keys":
			equalFunc(t, v, Keys)
		case "ToSlice":
			equalFunc(t, v, ToSlice)
		case "Sort":
			equalFunc(t, v, Sort)
		case "Equals":
			equalFunc(t, v, Equals)
		case "Sub":
			equalFunc(t, v, Sub)
		case "JSON2Map":
			equalFunc(t, v, JSON2Map)
		case "ToJSON":
			equalFunc(t, v, ToJSON)
		case "Prefix":
			equalFunc(t, v, Prefix)
		case "Suffix":
			equalFunc(t, v, Suffix)
		case "NewSlice":
			equalFunc(t, v, NewSlice)
		case "NewHashSet":
			equalFunc(t, v, NewHashSet)
		default:
			t.Fatalf("unknown built-in function has been exported: %q", k)
		}
	}
}

func Test_Not(t *testing.T) {
	// test that returns true
	var (
		boolv      bool
		intv       int
		floatv     float32
		complexv   complex64
		arrayv     [1]int
		chanv      chan int
		funcv      func()
		interfacev interface{}
		mapv       map[int]int
		ptrv       *int
		slicev     []string
		strv       string
		structv    struct{ foo int }
	)
	assert.True(t, Not(
		boolv, intv, floatv, complexv,
		arrayv, chanv, funcv, interfacev,
		mapv, ptrv, slicev, strv, structv,
	))

	// test that returns false
	boolv = true
	intv = -1
	floatv = -0.1
	complexv = 1.1
	arrayv[0] = 1
	chanv = make(chan int, 0)
	funcv = func() {}
	interfacev = 1
	mapv = map[int]int{}
	ptrv = &intv
	slicev = make([]string, 0)
	strv = "foo"
	structv.foo = 1
	for _, v := range []interface{}{
		boolv, intv, floatv, complexv,
		arrayv, chanv, funcv, interfacev,
		mapv, ptrv, slicev, strv, structv,
	} {
		assert.False(t, Not(v))
	}
}

func Test_HasPrefix(t *testing.T) {
	// test that returns true
	assert.True(t, HasPrefix("foo-bar", "foo-b"))

	// test that returns false
	assert.False(t, HasPrefix("foo-bar", "bar"))
}

func Test_HasSuffix(t *testing.T) {
	// test that returns true
	assert.True(t, HasSuffix("foo-bar", "o-bar"))

	// test that returns false
	assert.False(t, HasSuffix("foo-bar", "foo"))
}

func Test_Keys(t *testing.T) {
	// test that returns keys of map
	v := map[interface{}]bool{
		"c": true,
		"e": true,
		"g": true,
		"a": true,
		"d": true,
		"f": true,
		"b": true,
	}
	keys := []interface{}{}
	for k := range v {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i].(string) < keys[j].(string)
	})

	res, err := Keys(v)
	if err != nil {
		t.Fatal(err)
	}
	sort.Slice(res, func(i, j int) bool {
		return res[i].(string) < res[j].(string)
	})
	assert.Equal(t, keys, res)
	// with pointer
	res, err = Keys(&v)
	if err != nil {
		t.Fatal(err)
	}
	sort.Slice(res, func(i, j int) bool {
		return res[i].(string) < res[j].(string)
	})
	assert.Equal(t, keys, res)

	// test that returns indexes of slice
	sv := []string{"c", "e", "g", "a", "d", "f", "b"}
	keys = keys[:0]
	for k := range sv {
		keys = append(keys, k)
	}
	res, err = Keys(sv)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, keys, res)
	// with pointer
	res, err = Keys(&sv)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, keys, res)

	// test that returns error if argument is not map or slice argument
	keys, err = Keys(1)
	assert.Nil(t, keys)
	assert.Error(t, err)
}

func Test_ToSlice(t *testing.T) {
	args := []interface{}{
		1,
		nil,
		13,
		"a",
		0.5,
		[]string{"hello", "world"},
		map[string]string{"hello": "world"},
	}

	// test that returns slice from arguments
	for i, v := range ToSlice(args...) {
		assert.Equal(t, args[i], v)
	}
}

func Test_Sort(t *testing.T) {
	// test that returns sorted slice with bool arguments
	assert.Equal(t, []interface{}{
		true, false, true, false, true, false,
	}, Sort([]interface{}{
		true, false, true, false, true, false,
	}))

	// test that returns sorted slice with integer arguments
	assert.Equal(t, []interface{}{
		1, 5, 9, 13, 18, 26, 32,
	}, Sort([]interface{}{
		1, 13, 5, 32, 9, 18, 26,
	}))

	// test that returns sorted slice with unsigned integer arguments
	assert.Equal(t, []interface{}{
		uint(1), uint(5), uint(9), uint(13), uint(18), uint(26), uint(32),
	}, Sort([]interface{}{
		uint(1), uint(13), uint(5), uint(32), uint(9), uint(18), uint(26),
	}))

	// test that returns sorted slice with float arguments
	assert.Equal(t, []interface{}{
		0.1, 5.0, 9.3, 13.91, 18.01, 26.43, 32.98,
	}, Sort([]interface{}{
		0.1, 13.91, 5.0, 32.98, 9.3, 18.01, 26.43,
	}))

	// test that returns sorted slice with string
	assert.Equal(t, []interface{}{
		"apple", "banana", "cherry", "grape", "pineapple",
	}, Sort([]interface{}{
		"grape", "pineapple", "banana", "apple", "cherry",
	}))

	// test that cannot sort slice with various types of values
	assert.Equal(t, []interface{}{
		1, nil, 13, "a", 0.5, []string{"hello", "world"}, map[string]string{"hello": "world"},
	}, Sort([]interface{}{
		1, nil, 13, "a", 0.5, []string{"hello", "world"}, map[string]string{"hello": "world"},
	}))
}

func Test_Equals(t *testing.T) {
	// test that returns true with string value
	assert.True(t, Equals(nil, nil))
	sv := "foo/bar/baz"
	assert.True(t, Equals(sv, "foo/bar/baz"))
	assert.True(t, Equals(&sv, "foo/bar/baz"))
	nv := 421
	assert.True(t, Equals(nv, 421))
	assert.True(t, Equals(&nv, 421))
	// by interface{}
	var iv interface{} = sv
	assert.True(t, Equals(iv, sv))
	assert.True(t, Equals(iv, &sv))
	iv = &sv
	assert.True(t, Equals(iv, sv))
	assert.True(t, Equals(iv, &sv))
	iv = nv
	assert.True(t, Equals(iv, nv))
	assert.True(t, Equals(iv, &nv))
	iv = &nv
	assert.True(t, Equals(iv, nv))
	assert.True(t, Equals(iv, &nv))

	assert.True(t, Equals([]interface{}{"hello", "world"}, []interface{}{"hello", "world"}))
	assert.True(t, Equals([]string{"hello", "world"}, []string{"hello", "world"}))
	assert.True(t, Equals([]interface{}{
		1, nil, 13, "a", 0.5, []string{"hello", "world"}, map[string]string{"hello": "world"},
	}, []interface{}{
		1, nil, 13, "a", 0.5, []string{"hello", "world"}, map[string]string{"hello": "world"},
	}))
	assert.True(t, Equals([]string{"world"}, 611, "abc", 0.5, []string{"world"}))

	// test that returns false
	assert.False(t, Equals(421, int64(421)))
	assert.False(t, Equals([]string{"world", "hello"}, []string{"hello", "world"}))

	assert.False(t, Equals([]string{"world"}, 611, "abc", 0.5, []string{"hello"}))

}

func Test_Sub(t *testing.T) {
	// test that returns a value decremented
	assert.Equal(t, 4, Sub(5))

	// test that returns a value subtracted by second argument
	assert.Equal(t, 3, Sub(5, 2, 9, 3))
}

func Test_JSON2Map(t *testing.T) {
	// test that parse string as object
	data, err := JSON2Map(`{ "hello": "world!" }`)
	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{
		"hello": "world!",
	}, data)

	// test that parse string as array
	data, err = JSON2Map(`[ "hello", "world!" ]`)
	assert.NoError(t, err)
	assert.Equal(t, []interface{}{
		"hello", "world!",
	}, data)

	// test that parse string as string
	data, err = JSON2Map(`"hello world!"`)
	assert.NoError(t, err)
	assert.Equal(t, "hello world!", data)

	// test that parse string as number (float64)
	data, err = JSON2Map(`12345`)
	assert.NoError(t, err)
	assert.Equal(t, float64(12345), data)

	// test that returns parse error
	data, err = JSON2Map(`{ hello: "world!" }`)
	assert.Error(t, err)
	assert.Empty(t, data)
}

func Test_ToJSON(t *testing.T) {
	// test that returns json string
	for cmp, v := range map[string]interface{}{
		"{\n\"hello\": \"world!\"\n}": map[string]interface{}{
			"hello": "world!",
		},
		"[\n\"hello\",\n\"world!\"\n]": []interface{}{
			"hello", "world!",
		},
		`"hello world!"`: "hello world!",
		`12345`:          12345,
		`true`:           true,
		`false`:          false,
	} {
		s, err := ToJSON(v)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, cmp, s)
	}

	// test that returns json string with indent
	s, err := ToJSON(map[string]interface{}{
		"foo": "bar",
	}, "  ")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "{\n  \"foo\": \"bar\"\n}", s)

	// test that returns error
	_, err = ToJSON(`"hello"`, "", "")
	assert.Error(t, err)
}

func Test_Prefix(t *testing.T) {
	// test that returns the first 3 characters
	assert.Equal(t, "foo", Prefix("foo/bar/baz", 3))

	// test that returns the first 0 characters
	assert.Equal(t, "", Prefix("foo/bar/baz", 0))

	// test that returns the all characters if n is greater than length of
	// specified string
	assert.Equal(t, "foo/bar/baz", Prefix("foo/bar/baz", 12))
}

func Test_Suffix(t *testing.T) {
	// test that returns the last 3 characters
	assert.Equal(t, "baz", Suffix("foo/bar/baz", 3))

	// test that returns the last 0 characters
	assert.Equal(t, "", Suffix("foo/bar/baz", 0))

	// test that returns the all characters if n is greater than length of
	// specified string
	assert.Equal(t, "foo/bar/baz", Suffix("foo/bar/baz", 12))
}

func Test_Slice(t *testing.T) {
	// test that returns new instance of Slice
	s := NewSlice()
	assert.Equal(t, &Slice{}, s)

	// test that returns new instance of Slice with arguments
	s = NewSlice("foo", "bar", "baz")
	assert.Equal(t, []interface{}{"foo", "bar", "baz"}, s.Value())

	// test that returns length of slice
	assert.Equal(t, 3, s.Len())

	// test that returns head value of slice
	assert.Equal(t, "foo", s.Head())

	// test that returns tail value of slice
	assert.Equal(t, "baz", s.Tail())

	// test that append values
	assert.Equal(t, s, s.Append("quu", "qux"))
	assert.Equal(t, []interface{}{"foo", "bar", "baz", "quu", "qux"}, s.Value())

	// test that clear value
	s.Clear()
	assert.Equal(t, 0, s.Len())
	assert.Nil(t, s.Value())
	assert.Nil(t, s.Head())
	assert.Nil(t, s.Tail())

	// test that push value
	for i, v := range []string{"foo", "bar", "baz"} {
		assert.Equal(t, s, s.Push(v))
		// test that returns pushed value
		assert.Equal(t, v, s.Tail())
		assert.Equal(t, i+1, s.Len())
	}

	// test that pop value
	n := s.Len()
	for _, v := range []string{"baz", "bar", "foo"} {
		assert.Equal(t, v, s.Pop())
		n--
		assert.Equal(t, n, s.Len())
	}
	assert.Nil(t, s.Pop())

	// test that unshift value
	for i, v := range []string{"quu", "qux", "quux"} {
		assert.Equal(t, s, s.Unshift(v))
		assert.Equal(t, v, s.Head())
		assert.Equal(t, i+1, s.Len())
	}

	// test that unshift value
	n = s.Len()
	for i, v := range []string{"quux", "qux", "quu"} {
		assert.Equal(t, v, s.Shift())
		assert.Equal(t, n-(i+1), s.Len())
	}
	assert.Nil(t, s.Shift())
}

func Test_HashSet(t *testing.T) {
	// test that returns new instance fnHashSet data structure
	s := NewHashSet()
	assert.Equal(t, &HashSet{data: map[interface{}]bool{}}, s)

	// test that returns true if value is stored in hashset
	assert.True(t, s.Set("foo"))
	assert.True(t, s.Set("bar"))

	// test that returns false if value is already stored in hashset
	assert.False(t, s.Set("foo"))
	assert.False(t, s.Set("bar"))

	// test that returns true if value has been removed from hashset
	assert.True(t, s.Unset("foo"))

	// test that returns false if value does not exists
	assert.False(t, s.Unset("foo"))

	// test that returns true if value is stored in hashset
	assert.True(t, s.Set("foo"))
}
