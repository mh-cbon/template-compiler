package main

import (
	"fmt"
	"io"
	"os"

	"github.com/mh-cbon/template-compiler/compiled"
	"github.com/mh-cbon/template-compiler/demo/data"
	"github.com/mh-cbon/template-compiler/std/text/template"
	"github.com/mh-cbon/template-compiler/std/text/template/parse"
)

//go:generate template-compiler -var compiledTemplates
var compiledTemplates = compiled.New(
	"gen.go",
	[]compiled.TemplateConfiguration{
		compiled.TemplateConfiguration{
			HTML:          true,
			TemplatesPath: "templates/*.tpl",
			TemplatesData: map[string]interface{}{
				"*": data.MyTemplateData{},
			},
		},
		compiled.TemplateConfiguration{
			TemplateName:    "notafile",
			TemplateContent: `hello!{{define "embed"}}{{.}}{{end}}`,
			TemplatesData: map[string]interface{}{
				"*":     nil,
				"embed": data.MyTemplateData{},
				// "embed": OtherTemplateData{},
			},
		},
	},
).SetPkg("main")

// later, re arrange the demo to not use main,
// as its not a good case to run.
// example:
// - the loader is not able to load such main package,
// - the bootstraper can t import such package to consume the data
// https://godoc.org/golang.org/x/tools/go/loader#hdr-CONCEPTS_AND_TERMINOLOGY
// type OtherTemplateData struct{}

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
