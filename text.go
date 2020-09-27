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

func (_ Text) Clone(tmpl interface{}) (interface{}, error) {
	switch v := tmpl.(type) {
	case *template.Template:
		return v.Clone()
	}
	panic(fmt.Errorf("%T is not compatible with *template.Template", tmpl))
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

func (_ Text) AddParseTree(dst, src interface{}) error {
	dt := dst.(*template.Template)
	for _, t := range src.(*template.Template).Templates() {
		_, err := dt.AddParseTree(t.Name(), t.Tree)
		if err != nil {
			return err
		}
	}
	return nil
}

func (_ Text) ParseString(tmpl interface{}, str string) (interface{}, error) {
	return tmpl.(*template.Template).Parse(str)
}

func (_ Text) Execute(tmpl interface{}, w io.Writer, data interface{}) error {
	return tmpl.(*template.Template).Execute(w, data)
}
