package templatex

import (
	"bytes"
	"fmt"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
	tpl := NewEx(readfn, funcs)

	// test that readfn is equal to readfn
	assert.Equal(t, fmt.Sprintf("%p", readfn), fmt.Sprintf("%p", tpl.readfn))
	// test that funcs is equal to funcs
	assert.Equal(t, fmt.Sprintf("%#v", funcs), fmt.Sprintf("%#v", tpl.funcs))
}

func TestTemplate_RenderHTML(t *testing.T) {
	// setup
	// layf, err := testutil.Mkfile("@layout.html")

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
	tpl := NewEx(readfn, DefaultFuncMap())

	// test that return syscall.ENOENT error
	err := tpl.RenderText(b, "index.html", map[string]interface{}{
		"World": "world!",
	})
	assert.Equal(t, syscall.ENOENT, err)

	// test that render the file formatted as html/template
	files["/root/dir/index.html"] = `hello {{.World}}`
	assert.NoError(t, tpl.RenderHTML(b, "/index.html", map[string]interface{}{
		"World": "world!",
	}))
	assert.Equal(t, []byte("hello world!"), b.Bytes())

	// test that uses "index.html" as filename if pathname ended with slash
	b.Reset()
	assert.NoError(t, tpl.RenderHTML(b, "/", map[string]interface{}{
		"World": "world!",
	}))
	assert.Equal(t, []byte("hello world!"), b.Bytes())

	// test that return template parse error
	files["/root/dir/invalid.html"] = `hello {{.World}`
	err = tpl.RenderHTML(b, "invalid.html", map[string]interface{}{
		"World": "world!",
	})
	assert.Error(t, err)
	assert.Regexp(t, "template: invalid.html:.+ unexpected ", err)

	// test that render with sub template
	b.Reset()
	files["/root/dir/with_include.html"] = `hello {{.World}} {{template "@include.html" .}}`
	files["/root/dir/include.html"] = `{{define "@include.html"}} with {{.SubMessage}} {{end}}`
	err = tpl.RenderHTML(b, "with_include.html", map[string]interface{}{
		"World":      "world",
		"SubMessage": "sub template!",
	})
	assert.NoError(t, err)
	assert.Regexp(t, `\s*hello world\s+with sub template!\s*`, string(b.Bytes()))

	// test that insert sub template error as html comments
	b.Reset()
	files["/root/dir/with_invalid.html"] = `hello {{.World}} {{template "@invalid.html" .}}`
	err = tpl.RenderHTML(b, "with_invalid.html", map[string]interface{}{
		"World":      "world",
		"SubMessage": "sub template!",
	})
	assert.Regexp(t, `could not parse action {{template "@invalid.html"}} of "with_invalid.html".+ invalid.html:.+ unexpected `, err)

	// test that render with layout template
	b.Reset()
	files["/root/dir/layout.html"] = `
	layout
	----------------------
	{{template "content" .}}
	----------------------
	`
	files["/root/dir/with_layout.html"] = `
	{{layout "@layout.html"}}
	{{define "content"}}
	hello {{.World}} {{template "@include.html" .}}
	{{end}}
	`
	err = tpl.RenderHTML(b, "with_layout.html", map[string]interface{}{
		"World":      "world",
		"SubMessage": "sub template!",
	})
	assert.NoError(t, err)
	assert.Regexp(t, `layout\s+[-]+\s+hello world\s+with sub template!\s+[-]+\s+`, string(b.Bytes()))

	// test that can specify only one layout template
	b.Reset()
	files["/root/dir/with_two_layout.html"] = `
	{{layout "@layout.html"}}
	{{layout "@layout2.html"}}
	{{define "content"}}
	hello {{.World}}
	{{end}}
	`
	err = tpl.RenderHTML(b, "with_two_layout.html", map[string]interface{}{
		"World": "world!",
	})
	assert.Regexp(t, `'layout' action cannot be performed twice`, err)

	// test that returns error if failed to parse layout template
	b.Reset()
	files["/root/dir/invalid_layout.html"] = `
	layout
	----------------------
	{{template "content" .}
	----------------------
	`
	files["/root/dir/with_invalid_layout.html"] = `
	{{layout "@invalid_layout.html"}}
	{{define "content"}}
	hello {{.World}}
	{{end}}
	`
	err = tpl.RenderHTML(b, "with_invalid_layout.html", map[string]interface{}{
		"World": "world!",
	})
	assert.Regexp(t, `could not parse action {{layout "@invalid_layout.html"}} of "with_invalid_layout.html".+ invalid_layout.html:.+ unexpected`, err)
}
