package templatex

import (
	"bytes"
	"html/template"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHTML_IsNil(t *testing.T) {
	r := HTML{}

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

func TestHTML_NewTemplate(t *testing.T) {
	r := HTML{}

	// test that create new Template
	tmpl := r.NewTemplate("foo", nil)
	assert.False(t, r.IsNil(tmpl))
}

func TestHTML_Clone(t *testing.T) {
	r := HTML{}
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

func TestHTML_AddParseTree(t *testing.T) {
	r := HTML{}
	a := r.NewTemplate("foo", nil)
	b := r.NewTemplate("bar", nil)

	assert.Nil(t, b.(*template.Template).Lookup("content"))
	_, err := r.ParseString(b, `{{define "content"}} hello {{end}}`)
	assert.NoError(t, err)
	assert.NotNil(t, b.(*template.Template).Lookup("content"))

	// test that create new Template
	assert.Nil(t, a.(*template.Template).Lookup("content"))
	assert.NoError(t, r.AddParseTree(a, b, "content"))
	assert.NotNil(t, a.(*template.Template).Lookup("content"))
}

func TestHTML_ParseString(t *testing.T) {
	r := HTML{}
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

func TestHTML_Execute(t *testing.T) {
	r := HTML{}
	tmpl := r.NewTemplate("foo", nil)
	_, err := r.ParseString(tmpl, `hello {{.World}}`)
	assert.NoError(t, err)

	// test that can execute template
	b := bytes.NewBuffer(nil)
	assert.NoError(t, r.Execute(tmpl, b, map[string]string{
		"World": "<world!>",
	}))
	assert.Equal(t, []byte("hello &lt;world!&gt;"), b.Bytes())
}
