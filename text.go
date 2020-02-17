package templatex

import (
	"fmt"
	"io"
	"text/template"
)

type Text struct{}

func NewText() Text {
	return Text{}
}

func (_ Text) IsNil(tmpl interface{}) bool {
	switch v := tmpl.(type) {
	case *template.Template:
		return v == nil
	case nil:
		return true
	}
	panic(fmt.Errorf("%T is not compatible with *template.Template", tmpl))
}

func (_ Text) NewTemplate(name string, funcs map[string]interface{}) interface{} {
	return template.New(name).Funcs(funcs)
}

func (_ Text) AddParseTree(dst, src interface{}, name string) error {
	clone, err := src.(*template.Template).Clone()
	if err == nil {
		_, err = dst.(*template.Template).AddParseTree(name, clone.Lookup(name).Tree)
	}
	return err
}

func (_ Text) ParseString(tmpl interface{}, str string) (interface{}, error) {
	return tmpl.(*template.Template).Parse(str)
}

func (_ Text) Execute(tmpl interface{}, w io.Writer, data interface{}) error {
	return tmpl.(*template.Template).Execute(w, data)
}
