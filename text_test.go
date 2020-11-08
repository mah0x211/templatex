package templatex

import (
	"bytes"
	"testing"
	"text/template"

	"github.com/stretchr/testify/assert"
)

func TestText_Clone(t *testing.T) {
	r := Text{}
	tmpl := r.NewTemplate("foo", nil)

	// test that clone template
	clone, err := r.Clone(tmpl)
	assert.NoError(t, err)
	assert.NotNil(t, clone)

	// test that will be panic if incompatible value passed
	assert.Panics(t, func() {
		r.Clone(nil)
	})
}

func TestText_IsNil(t *testing.T) {
	r := Text{}

	// test that returns true if nil pointer is passed
	var tmpl *template.Template
	assert.True(t, r.IsNil(tmpl))

	// test that returns false if non-nil pointer is passed
	assert.False(t, r.IsNil(&template.Template{}))

	// test that returns true if nil is passed
	assert.True(t, r.IsNil(nil))

	// test that panic occurs if incompatible type value is passed
	assert.Panics(t, func() {
		r.IsNil(1)
	})
}

func TestText_NewTemplate(t *testing.T) {
	r := Text{}

	// test that create new Template
	tmpl := r.NewTemplate("foo", nil)
	assert.False(t, r.IsNil(tmpl))
}

func TestText_AddParseTree(t *testing.T) {
	r := Text{}
	a := r.NewTemplate("foo", nil)
	b := r.NewTemplate("bar", nil)
	c := r.NewTemplate("baz", nil)
	lookup := func(tmpl interface{}, name string) bool {
		_, ok := r.Lookup(tmpl, name)
		return ok
	}

	assert.False(t, lookup(c, "baz"))
	assert.False(t, lookup(c, "baz_content"))
	_, err := r.ParseString(c, `{{define "baz_content"}} baz {{end}}`)
	assert.NoError(t, err)
	assert.True(t, lookup(c, "baz"))
	assert.True(t, lookup(c, "baz_content"))

	assert.False(t, lookup(b, "bar"))
	assert.False(t, lookup(b, "bar_content"))
	_, err = r.ParseString(b, `{{define "bar_content"}} bar {{end}}`)
	assert.NoError(t, err)
	assert.True(t, lookup(b, "bar"))
	assert.True(t, lookup(b, "bar_content"))

	// test that add parse tree
	assert.False(t, lookup(b, "baz"))
	assert.False(t, lookup(b, "baz_content"))
	assert.NoError(t, r.AddParseTree(b, c))
	assert.True(t, lookup(b, "baz"))
	assert.True(t, lookup(b, "baz_content"))

	assert.False(t, lookup(a, "foo"))
	assert.False(t, lookup(a, "bar"))
	assert.False(t, lookup(a, "bar_content"))
	assert.False(t, lookup(a, "baz"))
	assert.False(t, lookup(a, "baz_content"))
	assert.NoError(t, r.AddParseTree(a, b))
	assert.False(t, lookup(a, "foo"))
	assert.True(t, lookup(a, "bar"))
	assert.True(t, lookup(a, "bar_content"))
	assert.True(t, lookup(a, "baz"))
	assert.True(t, lookup(a, "baz_content"))
}

func TestText_ParseString(t *testing.T) {
	r := Text{}
	tmpl := r.NewTemplate("foo", nil)

	// test that can parse template string
	_, err := r.ParseString(tmpl, `{{define "correct"}} hello {{end}}`)
	assert.NoError(t, err)
	assert.NotNil(t, tmpl.(*template.Template).Lookup("correct"))

	// test that cannot parse incorrect template string
	_, err = r.ParseString(tmpl, `{{define "incorrect"}} hello {{end}`)
	assert.Error(t, err)
	assert.Nil(t, tmpl.(*template.Template).Lookup("incorrect"))
}

func TestText_Execute(t *testing.T) {
	r := Text{}
	tmpl := r.NewTemplate("foo", nil)
	_, err := r.ParseString(tmpl, `hello {{.World}}`)
	assert.NoError(t, err)

	// test that can execute template
	b := bytes.NewBuffer(nil)
	assert.NoError(t, r.Execute(tmpl, b, map[string]string{
		"World": "<world!>",
	}))
	assert.Equal(t, []byte("hello <world!>"), b.Bytes())
}
