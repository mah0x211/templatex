package templatex

import (
	"fmt"
	"io"
	"text/template"
)

type Text struct {
	*Runtime
	cache Cache
}

func NewText(rt *Runtime, cacheable bool) *Text {
	return &Text{
		Runtime: rt,
		cache:   NewCache(cacheable),
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
	tmpl, ok := t.cache.Get(pathname)
	if !ok {
		var err error
		tmpl, err = t.preprocess(t, pathname, make(map[string]struct{}))
		if err != nil {
			return err
		}
		t.cache.Set(pathname, tmpl)
	}
	return tmpl.(*template.Template).Execute(w, data)
}

func (t *Text) GetCache(pathname string) (interface{}, bool) {
	return t.cache.Get(pathname)
}

func (t *Text) RemoveCache(pathname string) {
	t.cache.Unset(pathname)
}
