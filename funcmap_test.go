package templatex

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFuncMap_HasPrefix(t *testing.T) {
	// test that returns true
	assert.True(t, fmHasPrefix("foo-bar", "foo-b"))

	// test that returns false
	assert.False(t, fmHasPrefix("foo-bar", "bar"))
}

func TestFuncMap_Keys(t *testing.T) {
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

	res, err := fmKeys(v)
	if err != nil {
		t.Fatal(err)
	}
	sort.Slice(res, func(i, j int) bool {
		return res[i].(string) < res[j].(string)
	})
	assert.Equal(t, keys, res)
	// with pointer
	res, err = fmKeys(&v)
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
	res, err = fmKeys(sv)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, keys, res)
	// with pointer
	res, err = fmKeys(&sv)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, keys, res)

	// test that returns error if argument is not map or slice argument
	keys, err = fmKeys(1)
	assert.Nil(t, keys)
	assert.Error(t, err)
}

func TestFuncMap_ToSlice(t *testing.T) {
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
	for i, v := range fmToSlice(args...) {
		assert.Equal(t, args[i], v)
	}
}

func TestFuncMap_Sort(t *testing.T) {
	// test that returns sorted slice with bool arguments
	assert.Equal(t, []interface{}{
		true, false, true, false, true, false,
	}, fmSort([]interface{}{
		true, false, true, false, true, false,
	}))

	// test that returns sorted slice with integer arguments
	assert.Equal(t, []interface{}{
		1, 5, 9, 13, 18, 26, 32,
	}, fmSort([]interface{}{
		1, 13, 5, 32, 9, 18, 26,
	}))

	// test that returns sorted slice with unsigned integer arguments
	assert.Equal(t, []interface{}{
		uint(1), uint(5), uint(9), uint(13), uint(18), uint(26), uint(32),
	}, fmSort([]interface{}{
		uint(1), uint(13), uint(5), uint(32), uint(9), uint(18), uint(26),
	}))

	// test that returns sorted slice with float arguments
	assert.Equal(t, []interface{}{
		0.1, 5.0, 9.3, 13.91, 18.01, 26.43, 32.98,
	}, fmSort([]interface{}{
		0.1, 13.91, 5.0, 32.98, 9.3, 18.01, 26.43,
	}))

	// test that returns sorted slice with string
	assert.Equal(t, []interface{}{
		"apple", "banana", "cherry", "grape", "pineapple",
	}, fmSort([]interface{}{
		"grape", "pineapple", "banana", "apple", "cherry",
	}))

	// test that cannot sort slice with various types of values
	assert.Equal(t, []interface{}{
		1, nil, 13, "a", 0.5, []string{"hello", "world"}, map[string]string{"hello": "world"},
	}, fmSort([]interface{}{
		1, nil, 13, "a", 0.5, []string{"hello", "world"}, map[string]string{"hello": "world"},
	}))
}

func TestFuncMap_Equals(t *testing.T) {
	// test that returns true with string value
	assert.True(t, fmEquals(nil, nil))
	sv := "foo/bar/baz"
	assert.True(t, fmEquals(sv, "foo/bar/baz"))
	assert.True(t, fmEquals(&sv, "foo/bar/baz"))
	nv := 421
	assert.True(t, fmEquals(nv, 421))
	assert.True(t, fmEquals(&nv, 421))
	// by interface{}
	var iv interface{} = sv
	assert.True(t, fmEquals(iv, sv))
	assert.True(t, fmEquals(iv, &sv))
	iv = &sv
	assert.True(t, fmEquals(iv, sv))
	assert.True(t, fmEquals(iv, &sv))
	iv = nv
	assert.True(t, fmEquals(iv, nv))
	assert.True(t, fmEquals(iv, &nv))
	iv = &nv
	assert.True(t, fmEquals(iv, nv))
	assert.True(t, fmEquals(iv, &nv))

	assert.True(t, fmEquals([]interface{}{"hello", "world"}, []interface{}{"hello", "world"}))
	assert.True(t, fmEquals([]string{"hello", "world"}, []string{"hello", "world"}))
	assert.True(t, fmEquals([]interface{}{
		1, nil, 13, "a", 0.5, []string{"hello", "world"}, map[string]string{"hello": "world"},
	}, []interface{}{
		1, nil, 13, "a", 0.5, []string{"hello", "world"}, map[string]string{"hello": "world"},
	}))
	assert.True(t, fmEquals([]string{"world"}, 611, "abc", 0.5, []string{"world"}))

	// test that returns false
	assert.False(t, fmEquals(421, int64(421)))
	assert.False(t, fmEquals([]string{"world", "hello"}, []string{"hello", "world"}))

	assert.False(t, fmEquals([]string{"world"}, 611, "abc", 0.5, []string{"hello"}))

}

func TestFuncMap_Sub(t *testing.T) {
	// test that returns a value decremented
	assert.Equal(t, 4, fmSub(5))

	// test that returns a value subtracted by second argument
	assert.Equal(t, 3, fmSub(5, 2, 9, 3))
}

func TestFuncMap_JSON2Map(t *testing.T) {
	// test that parse string as object
	data, err := fmJSON2Map(`{ "hello": "world!" }`)
	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{
		"hello": "world!",
	}, data)

	// test that parse string as array
	data, err = fmJSON2Map(`[ "hello", "world!" ]`)
	assert.NoError(t, err)
	assert.Equal(t, []interface{}{
		"hello", "world!",
	}, data)

	// test that parse string as string
	data, err = fmJSON2Map(`"hello world!"`)
	assert.NoError(t, err)
	assert.Equal(t, "hello world!", data)

	// test that parse string as number (float64)
	data, err = fmJSON2Map(`12345`)
	assert.NoError(t, err)
	assert.Equal(t, float64(12345), data)

	// test that returns parse error
	data, err = fmJSON2Map(`{ hello: "world!" }`)
	assert.Error(t, err)
	assert.Empty(t, data)
}

func TestFuncMap_ToJSON(t *testing.T) {
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
		s, err := fmToJSON(v)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, cmp, s)
	}

	// test that returns json string with indent
	s, err := fmToJSON(map[string]interface{}{
		"foo": "bar",
	}, "  ")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "{\n  \"foo\": \"bar\"\n}", s)

	// test that returns error
	_, err = fmToJSON(`"hello"`, "", "")
	assert.Error(t, err)
}

func TestFuncMap_Prefix(t *testing.T) {
	// test that returns the first 3 characters
	assert.Equal(t, "foo", fmPrefix("foo/bar/baz", 3))

	// test that returns the first 0 characters
	assert.Equal(t, "", fmPrefix("foo/bar/baz", 0))

	// test that returns the all characters if n is greater than length of
	// specified string
	assert.Equal(t, "foo/bar/baz", fmPrefix("foo/bar/baz", 12))
}

func TestFuncMap_Suffix(t *testing.T) {
	// test that returns the last 3 characters
	assert.Equal(t, "baz", fmSuffix("foo/bar/baz", 3))

	// test that returns the last 0 characters
	assert.Equal(t, "", fmSuffix("foo/bar/baz", 0))

	// test that returns the all characters if n is greater than length of
	// specified string
	assert.Equal(t, "foo/bar/baz", fmSuffix("foo/bar/baz", 12))
}

func TestFuncMap_HashSet(t *testing.T) {
	// test that returns new instance fnHashSet data structure
	s := fmNewHashSet()
	assert.Equal(t, &fmHashSet{data: map[interface{}]bool{}}, s)

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
