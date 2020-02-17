package templatex

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultReadFunc(t *testing.T) {
	// setup
	pathname := "example.txt"
	f, err := os.OpenFile(pathname, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	assert.NoError(t, err)
	defer func() {
		f.Close()
		os.Remove(pathname)
	}()
	_, err = f.WriteAt([]byte("hello world!"), 0)
	assert.NoError(t, err)

	// test that DefaultReadFunc read file contents
	b, err := DefaultReadFunc(pathname)
	assert.NoError(t, err)
	assert.Equal(t, []byte("hello world!"), b)
}

func TestNew(t *testing.T) {
	tpl := New()

	// test that readfn is equal to DefaultReadFunc
	assert.Equal(t, fmt.Sprintf("%p", DefaultReadFunc), fmt.Sprintf("%p", tpl.readfn))

	// test that funcs is equal to returns of DefaultFuncMap()
	assert.Equal(t, fmt.Sprintf("%#v", DefaultFuncMap()), fmt.Sprintf("%#v", tpl.funcs))
}

func TestNewEx(t *testing.T) {
	// setup
	readfn := func(pathname string) ([]byte, error) {
		return nil, syscall.ENOENT
	}
	funcs := map[string]interface{}{
		"foo": "bar",
	}
	tpl := NewEx(readfn, funcs, false)

	// test that readfn is equal to readfn
	assert.Equal(t, fmt.Sprintf("%p", readfn), fmt.Sprintf("%p", tpl.readfn))
	// test that funcs is equal to funcs
	assert.Equal(t, fmt.Sprintf("%#v", funcs), fmt.Sprintf("%#v", tpl.funcs))
}

func TestRuntime_RenderHTML(t *testing.T) {
	// setup
	rootdir := "/root/dir/"
	files := map[string]string{}
	readfn := func(pathname string) ([]byte, error) {
		if pathname == "/" {
			pathname = "/index.html"
		}

		if s, ok := files[filepath.Join(rootdir, pathname)]; ok {
			return []byte(s), nil
		}

		return nil, syscall.ENOENT
	}
	b := bytes.NewBuffer(nil)
	create := func() *Runtime { return NewEx(readfn, DefaultFuncMap(), false) }

	// test that return syscall.ENOENT error
	err := create().RenderHTML(b, "index.html", map[string]interface{}{
		"World": "world!",
	})
	assert.Equal(t, syscall.ENOENT, err)

	// test that render the file formatted as html/template
	files = map[string]string{
		"/root/dir/index.html": `hello {{.World}}`,
	}
	assert.NoError(t, create().RenderHTML(b, "/index.html", map[string]interface{}{
		"World": "<world!>",
	}))
	assert.Equal(t, []byte("hello &lt;world!&gt;"), b.Bytes())

	// test that return template parse error
	files = map[string]string{
		"/root/dir/invalid.html": `hello {{.World}`,
	}
	err = create().RenderHTML(b, "invalid.html", map[string]interface{}{
		"World": "world!",
	})
	assert.Error(t, err)
	assert.Regexp(t, "template: invalid.html:.+ unexpected ", err)

	// test that render with sub template
	b.Reset()
	files = map[string]string{
		"/root/dir/with_include.html": `
			hello {{.World}} {{template "@include.html" .}}
			{{template "@include.html" .}} is included twice
		`,
		"/root/dir/include.html": `
			{{define "@include.html"}}
			with {{.SubMessage}}
			{{end}}
		`,
	}
	err = create().RenderHTML(b, "with_include.html", map[string]interface{}{
		"World":      "world",
		"SubMessage": "sub template!",
	})
	assert.NoError(t, err)
	assert.Regexp(t, `\s*hello world\s+with sub template!\s*with sub template!\s*is included twice`, string(b.Bytes()))

	// test that render with nested sub templates
	b.Reset()
	files = map[string]string{
		"/root/dir/with_include_nested.html": `
			hello {{.World}}
			{{template "@include.html" .}}
			{{template "@nested_include.html" .}}
		`,
		"/root/dir/include.html": `
			{{define "@include.html"}}
			with {{.SubMessage}}
			{{end}}
		`,
		"/root/dir/nested_include.html": `
			{{define "@nested_include.html"}}
			{{template "@include.html" .}}
			{{end}}
		`,
	}
	err = create().RenderHTML(b, "with_include_nested.html", map[string]interface{}{
		"World":      "world",
		"SubMessage": "sub template!",
	})
	assert.NoError(t, err)
	assert.Regexp(t, `\s*hello world\s+with sub template!\s+with sub template!\s*`, string(b.Bytes()))

	// that returns error if the template is included recursively
	b.Reset()
	files = map[string]string{
		"/root/dir/with_include_recursively.html": `
			hello {{.World}} {{template "@include_recursively.html" .}}
		`,
		"/root/dir/include_recursively.html": `
			{{define "@include_recursively.html"}}
			{{template "@include_recursively.html" .}}
			{{end}}
		`,
	}
	err = create().RenderHTML(b, "with_include_recursively.html", map[string]interface{}{
		"World": "world",
	})
	assert.Error(t, err)
	assert.Regexp(t, `cannot parse "include_recursively.html" recursively`, err)

	// test that insert sub template error as html comments
	b.Reset()
	files = map[string]string{
		"/root/dir/with_invalid.html": `
			hello {{.World}} {{template "@invalid.html" .}}
		`,
		"/root/dir/invalid.html": `hello {{.World}`,
	}
	err = create().RenderHTML(b, "with_invalid.html", map[string]interface{}{
		"World":      "world",
		"SubMessage": "sub template!",
	})
	assert.Regexp(t, `could not parse action {{template "@invalid.html"}} of "with_invalid.html".+ invalid.html:.+ unexpected `, err)

	// test that render with layout template
	b.Reset()
	files = map[string]string{
		"/root/dir/layout.html": `
			layout
			----------------------
			{{template "content" .}}
			----------------------
		`,
		"/root/dir/include.html": `
			{{define "@include.html"}}
			with {{.SubMessage}}
			{{end}}
		`,
		"/root/dir/with_layout.html": `
			{{define "content"}}
			hello {{.World}} {{template "@include.html" .}}
			{{end}}
			{{layout "@layout.html"}}
		`,
	}
	err = create().RenderHTML(b, "with_layout.html", map[string]interface{}{
		"World":      "world",
		"SubMessage": "sub template!",
	})
	assert.NoError(t, err)
	assert.Regexp(t, `layout\s+[-]+\s+hello world\s+with sub template!\s+[-]+\s+`, string(b.Bytes()))

	// test that can specify only one layout template
	b.Reset()
	files = map[string]string{
		"/root/dir/layout.html": `
			layout
			----------------------
			{{template "content" .}}
			----------------------
		`,
		"/root/dir/with_two_layout.html": `
			{{layout "@layout.html"}}
			{{define "content"}}
			hello {{.World}}
			{{end}}
			{{layout "@layout2.html"}}
		`,
	}
	err = create().RenderHTML(b, "with_two_layout.html", map[string]interface{}{
		"World": "world!",
	})
	assert.Regexp(t, `'layout' action cannot be performed twice`, err)

	// test that returns error if failed to parse layout template
	b.Reset()
	files = map[string]string{
		"/root/dir/invalid_layout.html": `
			layout
			----------------------
			{{template "content" .}
			----------------------
		`,
		"/root/dir/with_invalid_layout.html": `
			{{define "content"}}
			hello {{.World}}
			{{end}}
			{{layout "@invalid_layout.html"}}
		`,
	}
	err = create().RenderHTML(b, "with_invalid_layout.html", map[string]interface{}{
		"World": "world!",
	})
	assert.Regexp(t, `could not parse action {{layout "@invalid_layout.html"}} of "with_invalid_layout.html".+ invalid_layout.html:.+ unexpected`, err)
}

func TestRuntime_RenderText(t *testing.T) {
	// setup
	rootdir := "/root/dir/"
	files := map[string]string{}
	readfn := func(pathname string) ([]byte, error) {
		if pathname == "/" {
			pathname = "/index.html"
		}

		if s, ok := files[filepath.Join(rootdir, pathname)]; ok {
			return []byte(s), nil
		}

		return nil, syscall.ENOENT
	}
	b := bytes.NewBuffer(nil)
	create := func() *Runtime { return NewEx(readfn, DefaultFuncMap(), false) }

	// test that render the file formatted as text/template
	files["/root/dir/with_include.html"] = `hello {{.World}} {{template "@include.html" .}}`
	files["/root/dir/include.html"] = `{{define "@include.html"}}with {{.SubMessage}}{{end}}`
	assert.NoError(t, create().RenderText(b, "with_include.html", map[string]interface{}{
		"World":      "<world!>",
		"SubMessage": "sub template!",
	}))
	assert.Equal(t, []byte("hello <world!> with sub template!"), b.Bytes())
}

func TestRuntime_RemoveCacheText(t *testing.T) {
	// setup
	rootdir := "/root/dir/"
	files := map[string]string{}
	readfn := func(pathname string) ([]byte, error) {
		if pathname == "/" {
			pathname = "/index.html"
		}

		if s, ok := files[filepath.Join(rootdir, pathname)]; ok {
			return []byte(s), nil
		}

		return nil, syscall.ENOENT
	}
	b := bytes.NewBuffer(nil)
	tpl := NewEx(readfn, DefaultFuncMap(), true)

	// rendered template is to be cached
	files["/root/dir/index.html"] = `hello {{.World}}`
	assert.NoError(t, tpl.RenderText(b, "/index.html", map[string]interface{}{
		"World": "<world!>",
	}))

	_, ok := tpl.text.GetCache("/index.html")
	assert.True(t, ok)

	// test that remove cached template
	tpl.RemoveCacheText("/index.html")
	_, ok = tpl.text.GetCache("/index.html")
	assert.False(t, ok)
}

func TestRuntime_RemoveCacheHTML(t *testing.T) {
	// setup
	rootdir := "/root/dir/"
	files := map[string]string{}
	readfn := func(pathname string) ([]byte, error) {
		if pathname == "/" {
			pathname = "/index.html"
		}

		if s, ok := files[filepath.Join(rootdir, pathname)]; ok {
			return []byte(s), nil
		}

		return nil, syscall.ENOENT
	}
	b := bytes.NewBuffer(nil)
	tpl := NewEx(readfn, DefaultFuncMap(), true)

	// rendered template is to be cached
	files["/root/dir/index.html"] = `hello {{.World}}`
	assert.NoError(t, tpl.RenderHTML(b, "/index.html", map[string]interface{}{
		"World": "&lt;world!&gt;",
	}))

	_, ok := tpl.html.GetCache("/index.html")
	assert.True(t, ok)

	// test that remove cached template
	tpl.RemoveCacheHTML("/index.html")
	_, ok = tpl.html.GetCache("/index.html")
	assert.False(t, ok)
}
