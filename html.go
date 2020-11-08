package templatex

import (
	"fmt"
	"html/template"
	"io"
)

type HTML struct{}

func NewHTML() HTML {
	return HTML{}
}

func (_ HTML) Clone(tmpl interface{}) (interface{}, error) {
	switch v := tmpl.(type) {
	case *template.Template:
		return v.Clone()
	}
	panic(fmt.Errorf("%T is not compatible with *template.Template", tmpl))
}

func (_ HTML) IsNil(tmpl interface{}) bool {
	switch v := tmpl.(type) {
	case *template.Template:
		return v == nil
	case nil:
		return true
	}
	panic(fmt.Errorf("%T is not compatible with *template.Template", tmpl))
}

func (_ HTML) NewTemplate(name string, funcs map[string]interface{}) interface{} {
	return template.New(name).Funcs(funcs)
}

func (_ HTML) AddParseTree(dst, src interface{}) error {
	dt := dst.(*template.Template)
	for _, t := range src.(*template.Template).Templates() {
		_, err := dt.AddParseTree(t.Name(), t.Tree)
		if err != nil {
			return err
		}
	}
	return nil
}

func (_ HTML) Lookup(tmpl interface{}, name string) (interface{}, bool) {
	t := tmpl.(*template.Template).Lookup(name)
	return t, t != nil
}

func (_ HTML) ParseString(tmpl interface{}, str string) (interface{}, error) {
	return tmpl.(*template.Template).Parse(str)
}

func (_ HTML) Execute(tmpl interface{}, w io.Writer, data interface{}) error {
	return tmpl.(*template.Template).Execute(w, data)
}
