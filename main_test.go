package main_test

import (
	"io"
	"io/ioutil"
	"testing"

	"github.com/mh-cbon/template-compiler/compiled"
	"github.com/mh-cbon/template-compiler/std/text/template"
	"github.com/mh-cbon/template-compiler/std/text/template/parse"
)

var aJitTemplate *template.Template
var aCompiledTemplate *template.Compiled
var compiledTemplates *compiled.Registry
var s = []byte("r")

func init() {
	var err error
	aJitTemplate, err = template.ParseFiles("demo/templates/a.tpl")
	if err != nil {
		panic(err)
	}
	compiledTemplates = compiled.NewRegistry()
	compiledTemplates.Add("a.tpl", fn0)
	compiledTemplates.Add("b.tpl", fn1)
	aCompiledTemplate = compiledTemplates.MustGet("a.tpl")
	s = []byte("Hello from a!\n")
}

func fn0(t parse.Templater, w io.Writer, data interface {
}) error {
	// io.WriteString(w, "Hello from a!\n")
	w.Write(s) // write predefined bytes is a good optimization too.
	return nil
}

func fn1(t parse.Templater, w io.Writer, data interface {
}) error {
	io.WriteString(w, "Hello from b!\n")
	return nil
}

func BenchmarkRenderWithCompiledTemplate(b *testing.B) {
	for n := 0; n < b.N; n++ {
		aCompiledTemplate.Execute(ioutil.Discard, nil)
	}
}

func BenchmarkRenderWithJitTemplate(b *testing.B) {
	for n := 0; n < b.N; n++ {
		aJitTemplate.Execute(ioutil.Discard, nil)
	}
}
