package templatex

import (
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"
	"regexp"

	"github.com/mah0x211/templatex/builtins"
)

type ReadFunc func(pathname string) ([]byte, error)

func DefaultReadFunc(pathname string) ([]byte, error) {
	return ioutil.ReadFile(pathname)
}

type xTemplate interface {
	Render(w io.Writer, pathname string, data map[string]interface{}) error
	Parse(f *File, text string, layout *File, includes map[string]*File) error
}

type Runtime struct {
	readfn ReadFunc
	cache  Cache
	funcs  map[string]interface{}
	text   xTemplate
	html   xTemplate
}

func NewEx(readfn ReadFunc, cache Cache, funcs map[string]interface{}) *Runtime {
	rt := &Runtime{
		readfn: readfn,
		cache:  cache,
		funcs:  funcs,
	}
	rt.text = NewTemplate(rt, NewText())
	rt.html = NewTemplate(rt, NewHTML())
	return rt
}

func New() *Runtime {
	return NewEx(DefaultReadFunc, NewNopCache(), builtins.FuncMap())
}

// match　{{(template|layout) "name" .}}
var reTemplateAction = regexp.MustCompile(
	`[^{]*(\{{2}\s*(template|layout)\s+"(@[^"]+)"[^}]*}{2})`,
)

func (rt *Runtime) preprocess(t xTemplate, pathname string, cref map[string]struct{}) (*File, error) {
	// get cached template
	f := rt.cache.Get(pathname)
	if f != nil {
		return f, nil
	}

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
	f = createFile(rt.cache, pathname)
	var layout *File
	var includes = make(map[string]*File)
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
		if isLayout && layout != nil {
			return nil, fmt.Errorf("'layout' action cannot be performed twice")
		}

		// parse associated template
		af, err := rt.preprocess(t, val, cref)
		if err != nil {
			return nil, fmt.Errorf("could not preprocess {{%s %q}} in %q: %v", act, val, pathname, err)
		}

		if isLayout {
			layout = af
			// remove 'layout' action
			buf = append(buf[:m[2]], buf[m[3]:]...)
			// update cursor and index
			cur = m[0]
		} else {
			includes[val] = af
			f.addChild(af)
		}
		af.addParent(f)

		m = reTemplateAction.FindSubmatchIndex(buf[cur:])
	}

	delete(cref, pathname)
	err = t.Parse(f, string(buf), layout, includes)
	if err != nil {
		return nil, err
	}
	rt.cache.Set(f.name, f)

	return f, nil
}

func (rt *Runtime) RenderText(w io.Writer, pathname string, data map[string]interface{}) error {
	return rt.text.Render(w, filepath.Clean(pathname), data)
}

func (rt *Runtime) RenderHTML(w io.Writer, pathname string, data map[string]interface{}) error {
	return rt.html.Render(w, filepath.Clean(pathname), data)
}
