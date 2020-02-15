package templatex

import (
	"fmt"
	template_html "html/template"
	"io"
	"io/ioutil"
	"path/filepath"
	"regexp"
	template_text "text/template"
)

type ReadFunc func(pathname string) ([]byte, error)

func DefaultReadFunc(pathname string) ([]byte, error) {
	return ioutil.ReadFile(pathname)
}

type Template struct {
	readfn ReadFunc
	funcs  map[string]interface{}
	cache  map[string]map[string]interface{}
}

func NewEx(readfn ReadFunc, funcs map[string]interface{}) *Template {
	return &Template{
		readfn: readfn,
		funcs:  funcs,
		cache:  make(map[string]map[string]interface{}),
	}
}

func New() *Template {
	return NewEx(DefaultReadFunc, DefaultFuncMap())
}

func (t *Template) setCache(v interface{}, pathname, asa string) {
	cache, ok := t.cache[asa]
	if !ok {
		cache = make(map[string]interface{})
		t.cache[asa] = cache
	}
	cache[pathname] = v
}

func (t *Template) unsetCache(pathname, asa string) bool {
	if cache, ok := t.cache[asa]; ok {
		if _, ok := cache[pathname]; ok {
			delete(cache, pathname)
			return true
		}
	}
	return false
}

func (t *Template) getCache(pathname, asa string) interface{} {
	cache, ok := t.cache[asa]
	if !ok {
		return nil
	}
	tmpl, ok := cache[pathname]
	if !ok {
		return nil
	}

	return tmpl
}

func attach(tmpl interface{}, name string, val interface{}) error {
	switch v := tmpl.(type) {
	case *template_text.Template:
		clone, err := val.(*template_text.Template).Clone()
		if err != nil {
			return err
		}
		_, err = v.AddParseTree(name, clone.Lookup(name).Tree)
		return err

	case *template_html.Template:
		clone, err := val.(*template_html.Template).Clone()
		if err != nil {
			return err
		}
		_, err = v.AddParseTree(name, clone.Lookup(name).Tree)
		return err

	default:
		panic(fmt.Errorf("unknown template type %T passed", v))
	}
}

// match　{{(template|layout) "name" .}}
var reTemplateAction = regexp.MustCompile(
	`[^{](\{{2}\s*(template|layout)\s+"(@[^"]+)"[^}]*}{2})`,
)

func (t *Template) parse(pathname, asa string, cref map[string]struct{}) (interface{}, error) {
	// refuse recursive parsing
	if _, exists := cref[pathname]; exists {
		return nil, fmt.Errorf("cannot parse %q recursively", pathname)
	}
	cref[pathname] = struct{}{}

	// read file
	buf, err := t.readfn(pathname)
	if err != nil {
		return nil, err
	}

	// create template
	var tmpl interface{}
	switch asa {
	case "text":
		tmpl = template_text.New(pathname).Funcs(t.funcs)
	case "html":
		tmpl = template_html.New(pathname).Funcs(t.funcs)
	default:
		panic(fmt.Errorf("unknown template type %q passed", asa))
	}

	// lookup associated templates
	var hasLayout bool
	var layout interface{}
	var includes = make(map[string]interface{})
	var cur int
	m := reTemplateAction.FindSubmatchIndex(buf)
	for m != nil {
		// manipulate index
		m[0], m[1], m[2], m[3], m[4], m[5], m[6], m[7] = m[0]+cur, m[1]+cur, m[2]+cur, m[3]+cur, m[4]+cur, m[5]+cur, m[6]+cur, m[7]+cur
		// update cursor
		cur = m[1]
		// extract action "value" pair
		act := string(buf[m[4]:m[5]])
		val := filepath.Clean(string(buf[m[6]:m[7]]))

		// load layout template
		isLayout := act == "layout"
		if isLayout {
			// layout already defined
			if hasLayout {
				return nil, fmt.Errorf("'layout' action cannot be performed twice")
			}
			hasLayout = true
			// remove 'layout' action
			buf = append(buf[:m[2]], buf[m[3]:]...)
			// update cursor and index
			cur = m[0]
		}

		// parse associated template
		vtmpl := t.getCache(val, asa)
		if vtmpl == (interface{})(nil) {
			if vtmpl, err = t.parse(val[1:], asa, cref); err != nil {
				if isLayout {
					return nil, fmt.Errorf("could not parse action {{%s %q}} of %q: %v", act, val, pathname, err)
				}
				return nil, fmt.Errorf("could not parse action {{%s %q}} of %q: %v", act, val, pathname, err)
			} else if !isLayout {
				// add non-layout template into cache
				t.setCache(vtmpl, val, asa)
			}
		}

		if isLayout {
			layout = vtmpl
		} else {
			includes[val] = vtmpl
		}

		m = reTemplateAction.FindSubmatchIndex(buf[cur:])
	}

	// use layout template as the base template
	if hasLayout {
		tmpl = layout
	}
	// attach associated templates
	for name, t := range includes {
		if err = attach(tmpl, name, t); err != nil {
			return nil, fmt.Errorf("could not attach associated template %q to %q: %v", name, pathname, err)
		}
	}

	switch v := tmpl.(type) {
	case *template_text.Template:
		return v.Parse(string(buf))

	case *template_html.Template:
		return v.Parse(string(buf))

	default:
		panic(fmt.Errorf("unknown template type %T passed", v))
	}
}

func (t *Template) render(w io.Writer, pathname, asa string, data map[string]interface{}) error {
	var err error

	pathname = filepath.Clean(pathname)
	v := t.getCache(pathname, asa)
	if v == (interface{})(nil) {
		v, err = t.parse(pathname, asa, make(map[string]struct{}))
		if err != nil {
			return err
		}
		t.setCache(v, pathname, asa)
	}

	switch tmpl := v.(type) {
	case *template_text.Template:
		err = tmpl.Execute(w, data)
	case *template_html.Template:
		err = tmpl.Execute(w, data)
	default:
		panic(fmt.Errorf("unknown template type %q passed", asa))
	}

	return err
}

func (t *Template) RenderText(w io.Writer, pathname string, data map[string]interface{}) error {
	return t.render(w, pathname, "text", data)
}

func (t *Template) RenderHTML(w io.Writer, pathname string, data map[string]interface{}) error {
	return t.render(w, pathname, "html", data)
}

func (t *Template) RemoveCacheText(pathname string) bool {
	return t.unsetCache(pathname, "text")
}

func (t *Template) RemoveCacheHTML(pathname string) bool {
	return t.unsetCache(pathname, "html")
}
