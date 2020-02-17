package templatex

import (
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"
	"regexp"
)

type ReadFunc func(pathname string) ([]byte, error)

func DefaultReadFunc(pathname string) ([]byte, error) {
	return ioutil.ReadFile(pathname)
}

type xTemplate interface {
	Render(w io.Writer, pathname string, data map[string]interface{}) error
	Parse(pathname, text string, layout interface{}, includes map[string]interface{}) (interface{}, error)
	GetCache(key string) (interface{}, bool)
	RemoveCache(key string)
}

type Runtime struct {
	readfn ReadFunc
	funcs  map[string]interface{}
	text   xTemplate
	html   xTemplate
}

func NewEx(readfn ReadFunc, funcs map[string]interface{}, cacheable bool) *Runtime {
	rt := &Runtime{
		readfn: readfn,
		funcs:  funcs,
	}
	rt.text = NewTemplate(rt, cacheable, NewText())
	rt.html = NewTemplate(rt, cacheable, NewHTML())
	return rt
}

func New() *Runtime {
	return NewEx(DefaultReadFunc, DefaultFuncMap(), true)
}

// match　{{(template|layout) "name" .}}
var reTemplateAction = regexp.MustCompile(
	`[^{](\{{2}\s*(template|layout)\s+"(@[^"]+)"[^}]*}{2})`,
)

func (rt *Runtime) preprocess(t xTemplate, pathname string, cref map[string]struct{}) (interface{}, error) {
	// refuse recursive parsing
	if _, exists := cref[pathname]; exists {
		return nil, fmt.Errorf("cannot parse %q recursively", pathname)
	}
	cref[pathname] = struct{}{}

	// read file
	buf, err := rt.readfn(pathname)
	if err != nil {
		return nil, err
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
		vtmpl, exists := t.GetCache(val)
		if !exists {
			if vtmpl, err = rt.preprocess(t, val[1:], cref); err != nil {
				if isLayout {
					return nil, fmt.Errorf("could not parse action {{%s %q}} of %q: %v", act, val, pathname, err)
				}
				return nil, fmt.Errorf("could not parse action {{%s %q}} of %q: %v", act, val, pathname, err)
			}
		}

		if isLayout {
			layout = vtmpl
		} else {
			includes[val] = vtmpl
		}

		m = reTemplateAction.FindSubmatchIndex(buf[cur:])
	}

	delete(cref, pathname)
	return t.Parse(pathname, string(buf), layout, includes)
}

func (rt *Runtime) RenderText(w io.Writer, pathname string, data map[string]interface{}) error {
	return rt.text.Render(w, filepath.Clean(pathname), data)
}

func (rt *Runtime) RenderHTML(w io.Writer, pathname string, data map[string]interface{}) error {
	return rt.html.Render(w, filepath.Clean(pathname), data)
}

func (rt *Runtime) RemoveCacheText(pathname string) {
	rt.text.RemoveCache(filepath.Clean(pathname))
}

func (rt *Runtime) RemoveCacheHTML(pathname string) {
	rt.html.RemoveCache(filepath.Clean(pathname))
}
