package templatex

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/mah0x211/templatex/builtins"
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

	// test that funcs is equal to returns of builtins.FuncMap()
	assert.Equal(t, fmt.Sprintf("%#v", builtins.FuncMap()), fmt.Sprintf("%#v", tpl.funcs))
}

func TestNewEx(t *testing.T) {
	// setup
	readfn := func(pathname string) ([]byte, error) {
		return nil, syscall.ENOENT
	}
	funcs := map[string]interface{}{
		"foo": "bar",
	}
	tpl := NewEx(readfn, NewNopCache(), funcs)

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
	create := func() *Runtime { return NewEx(readfn, NewNopCache(), builtins.FuncMap()) }

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
		"/root/dir/@include.html": `
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
	assert.Regexp(t, `\s*hello world\s+with sub template!\s*with sub template!\s*is included twice`, b.String())

	// test that render with nested sub templates
	b.Reset()
	files = map[string]string{
		"/root/dir/with_include_nested.html": `
			hello {{.World}}
			{{template "@include.html" .}}
			{{template "@nested_include.html" .}}
		`,
		"/root/dir/@include.html": `
			{{define "@include.html"}}
			with {{.SubMessage}}
			{{end}}
		`,
		"/root/dir/@nested_include.html": `
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
	assert.Regexp(t, `\s*hello world\s+with sub template!\s+with sub template!\s*`, b.String())

	// that returns error if the template is included recursively
	b.Reset()
	files = map[string]string{
		"/root/dir/with_include_recursively.html": `
			hello {{.World}} {{template "@include_recursively.html" .}}
		`,
		"/root/dir/@include_recursively.html": `
			{{define "@include_recursively.html"}}
			{{template "@include_recursively.html" .}}
			{{end}}
		`,
	}
	err = create().RenderHTML(b, "with_include_recursively.html", map[string]interface{}{
		"World": "world",
	})
	assert.Error(t, err)
	assert.Regexp(t, `cannot parse "@include_recursively.html" recursively`, err)

	// test that insert sub template error as html comments
	b.Reset()
	files = map[string]string{
		"/root/dir/with_invalid.html": `
			hello {{.World}} {{template "@invalid.html" .}}
		`,
		"/root/dir/@invalid.html": `hello {{.World}`,
	}
	err = create().RenderHTML(b, "with_invalid.html", map[string]interface{}{
		"World":      "world",
		"SubMessage": "sub template!",
	})
	assert.Regexp(t, `could not preprocess {{template "@invalid.html"}} in "with_invalid.html".+ @invalid.html:.+ unexpected `, err)

	// test that render with layout template
	b.Reset()
	files = map[string]string{
		"/root/dir/@layout.html": `
			layout
			----------------------
			{{template "content" .}}
			----------------------
		`,
		"/root/dir/@include.html": `
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
	assert.Regexp(t, `layout\s+[-]+\s+hello world\s+with sub template!\s+[-]+\s+`, b.String())

	// test that can specify only one layout template
	b.Reset()
	files = map[string]string{
		"/root/dir/@layout.html": `
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
		"/root/dir/@invalid_layout.html": `
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
	assert.Regexp(t, `could not preprocess {{layout "@invalid_layout.html"}} in "with_invalid_layout.html".+ @invalid_layout.html:.+ unexpected`, err)

	// test that rendered templates are cached
	b.Reset()
	files = map[string]string{
		"/root/dir/@footer.html":      `{{define "@footer.html"}}with footer{{end}}`,
		"/root/dir/@layout.html":      `layout: {{template "content" .}} {{template "@footer.html"}}`,
		"/root/dir/with_layout.html":  `{{define "content"}}hello {{.World}}{{end}}{{layout "@layout.html"}}`,
		"/root/dir/with_layout2.html": `{{define "content"}}hello 2 {{.World}}{{end}}{{layout "@layout.html"}}`,
	}
	cache := NewMapCache()
	rt := NewEx(readfn, cache, builtins.FuncMap())
	err = rt.RenderHTML(b, "with_layout.html", map[string]interface{}{
		"World": "world!",
	})
	assert.NoError(t, err)
	assert.Equal(t, `layout: hello world! with footer`, b.String())
	assert.NotNil(t, cache.Get("@layout.html"))
	assert.NotNil(t, cache.Get("with_layout.html"))
	assert.Equal(t, "@layout.html", cache.Get("@layout.html").Name())

	b.Reset()
	err = rt.RenderHTML(b, "with_layout2.html", map[string]interface{}{
		"World": "world!",
	})
	assert.NoError(t, err)
	assert.Equal(t, `layout: hello 2 world! with footer`, b.String())
	assert.NotNil(t, cache.Get("@layout.html"))
	assert.NotNil(t, cache.Get("with_layout.html"))
	assert.NotNil(t, cache.Get("with_layout2.html"))

	// test that rendered templates are cached
	b.Reset()
	cache.Get("with_layout.html").Uncache()
	err = rt.RenderHTML(b, "with_layout.html", map[string]interface{}{
		"World": "world!",
	})
	assert.NoError(t, err)
	assert.Equal(t, `layout: hello world! with footer`, b.String())
	assert.NotNil(t, cache.Get("with_layout2.html"))
}

func TestRuntime_RenderText(t *testing.T) {
	// setup
	rootdir := "/root/dir/"
	files := map[string]string{
		"/root/dir/with_include.html": `hello {{.World}} {{template "@include.html" .}}`,
		"/root/dir/@include.html":     `{{define "@include.html"}}with {{.SubMessage}}{{end}}`,

		// error in @err_include.html:
		"/root/dir/@err_include.html": `
			{{define "@err_include.html"}}
			with {{UnknownFunc .}}
			{{end}}
		`,
		"/root/dir/with_err_include.html": `
			hello {{.World}}
			{{template "@err_include.html" .}}
		`,

		"/root/dir/@err_layout.html": `
			head|
			{{UnknownFunc .}}
			|tail
		`,
		"/root/dir/with_err_layout.html": `
			{{define "content"}}
			{{layout "@err_layout.html"}}
				hello
			{{end}}
		`,

		"/root/dir/@layout.html": `
			head|
			{{template "content" .}}
			|tail
		`,
		"/root/dir/with_err_in_content_in_layout.html": `
			{{layout "@layout.html"}}
			{{define "content"}}
			hello
			{{UnknownFunc .}}
			{{end}}
		`,

		"/root/dir/with_err_include_in_content_in_layout.html": `
			{{layout "@layout.html"}}
			{{define "content"}}
			hello
			{{template "@err_include.html" .}}
			{{end}}
		`,

		"/root/dir/@layout_with_err_include.html": `
			head|
			{{template "content" .}}
			{{template "@err_include.html" .}}
			|tail
		`,
		"/root/dir/with_err_include_in_layout.html": `
			{{layout "@layout_with_err_include.html"}}
			{{define "content"}}
			hello
			{{end}}
		`,
	}
	nread := 0
	readfn := func(pathname string) ([]byte, error) {
		nread++
		if pathname == "/" {
			pathname = "/index.html"
		}

		if s, ok := files[filepath.Join(rootdir, pathname)]; ok {
			return []byte(s), nil
		}

		return nil, syscall.ENOENT
	}
	b := bytes.NewBuffer(nil)
	cache := NewMapCache()
	rt := NewEx(readfn, cache, builtins.FuncMap())

	// test that render the file formatted as text/template
	assert.NoError(t, rt.RenderText(b, "with_include.html", map[string]interface{}{
		"World":      "<world!>",
		"SubMessage": "sub template!",
	}))
	assert.Equal(t, []byte("hello <world!> with sub template!"), b.Bytes())
	assert.Equal(t, 2, nread)

	// test that render the cached template
	b.Reset()
	assert.NoError(t, rt.RenderText(b, "with_include.html", map[string]interface{}{
		"World":      "<world!>",
		"SubMessage": "sub template!",
	}))
	assert.Equal(t, []byte("hello <world!> with sub template!"), b.Bytes())
	assert.Equal(t, 2, nread)

	// test that render the file again
	b.Reset()
	cache.Get("with_include.html").Uncache()
	cache.Get("@include.html").Uncache()
	assert.NoError(t, rt.RenderText(b, "with_include.html", map[string]interface{}{
		"World":      "<world!>",
		"SubMessage": "sub template!",
	}))
	assert.Equal(t, []byte("hello <world!> with sub template!"), b.Bytes())
	assert.Equal(t, 4, nread)

	// test that returns error with err_include error
	b.Reset()
	err := rt.RenderText(b, "with_err_include.html", map[string]interface{}{
		"World":      "<world!>",
		"SubMessage": "sub template!",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), `in "with_err_include.html"`)
	assert.Contains(t, err.Error(), `@err_include.html:3: function "UnknownFunc" not defined`)

	// test that returns error with err_layout error
	b.Reset()
	err = rt.RenderText(b, "with_err_layout.html", map[string]interface{}{
		"World":      "<world!>",
		"SubMessage": "sub template!",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), `in "with_err_layout.html"`)
	assert.Contains(t, err.Error(), `@err_layout.html:3: function "UnknownFunc" not defined`)

	// test that returns error with err_layout error
	b.Reset()
	err = rt.RenderText(b, "with_err_in_content_in_layout.html", map[string]interface{}{
		"World":      "<world!>",
		"SubMessage": "sub template!",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), `with_err_in_content_in_layout.html:5: function "UnknownFunc" not defined`)

	// test that returns error with err_layout error
	b.Reset()
	err = rt.RenderText(b, "with_err_include_in_content_in_layout.html", map[string]interface{}{
		"World":      "<world!>",
		"SubMessage": "sub template!",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), `in "with_err_include_in_content_in_layout.html"`)
	assert.Contains(t, err.Error(), `@err_include.html:3: function "UnknownFunc" not defined`)

	// test that returns error with err_layout error
	b.Reset()
	err = rt.RenderText(b, "with_err_include_in_layout.html", map[string]interface{}{
		"World":      "<world!>",
		"SubMessage": "sub template!",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), `in "with_err_include_in_layout.html"`)
	assert.Contains(t, err.Error(), `@err_include.html:3: function "UnknownFunc" not defined`)

}
