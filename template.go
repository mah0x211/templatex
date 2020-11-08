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
	f.tmpl = t.renderer.NewTemplate(f.name, t.funcs)
	f.root = f.tmpl
	var child map[string]*File
	if layout != nil {
		// NOTE: layout template will be the root template but it cannot be
		// parsed more than twice. so, it must use the cloned template.
		if tmpl, err := t.renderer.Clone(layout.tmpl); err != nil {
			return err
		} else if err = t.renderer.AddParseTree(f.tmpl, tmpl); err != nil {
			return err
		}

		root, ok := t.renderer.Lookup(f.tmpl, layout.Name())
		if !ok {
			// layout template name
			panic(fmt.Errorf(
				"layout template %q not found: template name must be same as filename",
				layout.Name(),
			))
		}
		f.root = root
		child = layout.child
	}

	// attach associated templates
	for name, inc := range includes {
		if err := t.renderer.AddParseTree(f.tmpl, inc.tmpl); err != nil {
			return fmt.Errorf("could not attach associated template %q to %q: %v", name, f.name, err)
		}
	}

	if _, err := t.renderer.ParseString(f.tmpl, text); err != nil {
		return err
	}
	for _, c := range child {
		c.addParent(f)
	}

	return nil
}

func (t *Template) Render(w io.Writer, pathname string, data map[string]interface{}) error {
	f, err := t.preprocess(t, pathname, make(map[string]struct{}))
	if err != nil {
		return err
	}
	return t.renderer.Execute(f.root, w, data)
}
