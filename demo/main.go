package main

import (
	"fmt"
	"io"
	"os"

	"github.com/mh-cbon/template-compiler/compiled"
	"github.com/mh-cbon/template-compiler/std/text/template"
	"github.com/mh-cbon/template-compiler/std/text/template/parse"
)

//go:generate template-compiler -html -tpl "templates/*.tpl" -out "gen.go" -var compiledTemplates -data github.com/mh-cbon/template-compiler/demo/data:MyTemplateData
var compiledTemplates *compiled.Registry

func main() {

	c := template.NewCompiled("c", func(t parse.Templater, w io.Writer, data interface{}) error {
		w.Write([]byte("hello this is a compiled template\n"))
		return nil
	})

	c.Execute(os.Stdout, nil)

	fmt.Println("more to do")

	compiledTemplate := compiledTemplates.MustGet("a.tpl")
	compiledTemplate.Execute(os.Stdout, nil)
}
