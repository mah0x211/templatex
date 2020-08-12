package templatex

import (
	"fmt"
	"io"
)

type xRenderer interface {
	IsNil(tmpl interface{}) bool
	NewTemplate(name string, funcs map[string]interface{}) interface{}
	AddParseTree(dst, src interface{}, name string) error
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

func (t *Template) Parse(pathname, text string, layout interface{}, includes map[string]interface{}) (interface{}, error) {
	// use layout template as the base template if not nil
	tmpl := layout
	if t.renderer.IsNil(tmpl) {
		tmpl = t.renderer.NewTemplate(pathname, t.funcs)
	}
	// attach associated templates
	for name, inc := range includes {
		if err := t.renderer.AddParseTree(tmpl, inc, name); err != nil {
			return nil, fmt.Errorf("could not attach associated template %q to %q: %v", name, pathname, err)
		}
	}

	return t.renderer.ParseString(tmpl, text)
}

func (t *Template) Render(w io.Writer, pathname string, data map[string]interface{}) error {
	tmpl, err := t.preprocess(t, pathname, make(map[string]struct{}))
	if err != nil {
		return err
	}
	return t.renderer.Execute(tmpl, w, data)
}
