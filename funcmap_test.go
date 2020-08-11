package templatex

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFuncMap_Keys(t *testing.T) {
	m := map[int]bool{
		1:  true,
		5:  true,
		13: true,
		6:  true,
		4:  true,
		9:  true,
		21: true,
		18: true,
		26: true,
	}

	// test that returns keys of map
	keys := fmKeys(reflect.ValueOf(m))
	assert.Equal(t, len(m), len(keys))
	for _, k := range keys {
		ival, ok := k.(int)
		assert.True(t, ok)
		assert.True(t, m[ival])
	}

	// test that panic occurs with non-map object
	assert.Panics(t, func() {
		fmKeys(reflect.ValueOf(1))
	})
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
	// test that returns true
	assert.True(t, fmEquals(421, 421))
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
