package templatex

import (
	"fmt"
	"io"
	"sync"
	"text/template"
)

type Text struct {
	sync.Mutex
	*Runtime
	cache map[string]*template.Template
}

func NewText(rt *Runtime) *Text {
	return &Text{
		Runtime: rt,
		cache:   make(map[string]*template.Template),
	}
}

func (t *Text) Parse(pathname, text string, layout interface{}, includes map[string]interface{}) (interface{}, error) {
	// use layout template as the base template if not nil
	tmpl, ok := layout.(*template.Template)
	if !ok {
		tmpl = template.New(pathname).Funcs(t.funcs)
	}
	// attach associated templates
	for name, inc := range includes {
		clone, err := inc.(*template.Template).Clone()
		if err == nil {
			_, err = tmpl.AddParseTree(name, clone.Lookup(name).Tree)
		}
		if err != nil {
			return nil, fmt.Errorf("could not attach associated template %q to %q: %v", name, pathname, err)
		}
	}

	return tmpl.Parse(text)
}

func (t *Text) Render(w io.Writer, pathname string, data map[string]interface{}) error {
	t.Lock()
	tmpl := t.cache[pathname]
	t.Unlock()
	if tmpl == nil {
		v, err := t.preprocess(t, pathname, make(map[string]struct{}))
		if err != nil {
			return err
		}
		tmpl = v.(*template.Template)
		t.Lock()
		t.cache[pathname] = tmpl
		t.Unlock()
	}
	return tmpl.Execute(w, data)
}

func (t *Text) GetCache(pathname string) (interface{}, bool) {
	t.Lock()
	v, ok := t.cache[pathname]
	t.Unlock()
	return v, ok
}

func (t *Text) RemoveCache(pathname string) {
	t.Lock()
	delete(t.cache, pathname)
	t.Unlock()
}
