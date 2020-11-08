package templatex

import (
	"fmt"
	"io"
)

type xRenderer interface {
	Clone(tmpl interface{}) (interface{}, error)
	IsNil(tmpl interface{}) bool
	NewTemplate(name string, funcs map[string]interface{}) interface{}
	AddParseTree(dst, src interface{}) error
	Lookup(tmpl interface{}, name string) (interface{}, bool)
	ParseString(tmpl interface{}, str string) (interface{}, error)
	Execute(tmpl interface{}, w io.Writer, data interface{}) error
}

type Template struct {
	*Runtime
	renderer xRenderer
}

func NewTemplate(rt *Runtime, renderer xRenderer) *Template {
	return &Template{
		Runtime:  rt,
		renderer: renderer,
	}
}

func (t *Template) Parse(f *File, text string, layout *File, includes map[string]*File) error {
	var err error

	// NOTE: layout template will be the root template but it cannot be
	// parsed more than twice. so, it must use the cloned template.
	if layout == nil {
		f.tmpl = t.renderer.NewTemplate(f.name, t.funcs)
	} else if f.tmpl, err = t.renderer.Clone(layout.tmpl); err != nil {
		return err
	} else {
		for _, c := range layout.child {
			c.addParent(f)
		}
	}

	// attach associated templates
	for name, inc := range includes {
		if err := t.renderer.AddParseTree(f.tmpl, inc.tmpl); err != nil {
			return fmt.Errorf("could not attach associated template %q to %q: %v", name, f.name, err)
		}
	}

	f.tmpl, err = t.renderer.ParseString(f.tmpl, text)
	return err
}

func (t *Template) Render(w io.Writer, pathname string, data map[string]interface{}) error {
	f, err := t.preprocess(t, pathname, make(map[string]struct{}))
	if err != nil {
		return err
	}
	return t.renderer.Execute(f.tmpl, w, data)
}
