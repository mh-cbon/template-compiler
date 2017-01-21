package main

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/mh-cbon/template-compiler/std/html/template"
	text "github.com/mh-cbon/template-compiler/std/text/template"
)

var aJitTemplate *template.Template
var bJitTemplate *template.Template
var cJitTemplate *template.Template
var dJitTemplate *template.Template
var aCompiledTemplate *text.Compiled
var bCompiledTemplate *text.Compiled
var cCompiledTemplate *text.Compiled
var dCompiledTemplate *text.Compiled

func init() {
	var err error
	aJitTemplate, err = template.ParseFiles("templates/a.tpl")
	if err != nil {
		panic(err)
	}
	bJitTemplate, err = template.ParseFiles("templates/b.tpl")
	if err != nil {
		panic(err)
	}
	cJitTemplate, err = template.ParseFiles("templates/c.tpl")
	if err != nil {
		panic(err)
	}
	dJitTemplate, err = template.ParseFiles("templates/d.tpl")
	if err != nil {
		panic(err)
	}
	aCompiledTemplate = compiledTemplates.MustGet("a.tpl")
	bCompiledTemplate = compiledTemplates.MustGet("b.tpl")
	cCompiledTemplate = compiledTemplates.MustGet("c.tpl")
	dCompiledTemplate = compiledTemplates.MustGet("d.tpl")
}

func TestTemplatesA(t *testing.T) {
	var a bytes.Buffer
	aJitTemplate.Execute(&a, nil)
	var b bytes.Buffer
	aCompiledTemplate.Execute(&b, nil)
	aS := a.String()
	bS := b.String()
	if aS != bS {
		t.Errorf("nop\n%v\n%v", aS, bS)
	}
}
func TestTemplatesB(t *testing.T) {
	var a bytes.Buffer
	bJitTemplate.Execute(&a, nil)
	var b bytes.Buffer
	bCompiledTemplate.Execute(&b, nil)
	aS := a.String()
	bS := b.String()
	if aS != bS {
		t.Errorf("nop\n%v\n%v", aS, bS)
	}
}
func TestTemplatesC(t *testing.T) {
	var a bytes.Buffer
	cJitTemplate.Execute(&a, nil)
	var b bytes.Buffer
	cCompiledTemplate.Execute(&b, nil)
	aS := a.String()
	bS := b.String()
	if aS != bS {
		t.Errorf("nop\n%v\n%v", aS, bS)
	}
}
func TestTemplatesD(t *testing.T) {
	var a bytes.Buffer
	dJitTemplate.Execute(&a, nil)
	var b bytes.Buffer
	if err := dCompiledTemplate.Execute(&b, nil); err != nil {
		panic(err)
	}
	aS := a.String()
	bS := b.String()
	if aS != bS {
		t.Errorf("nop\n'%v'\n'%v'", aS, bS)
	}
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

func BenchmarkRenderWithCompiledTemplateC(b *testing.B) {
	for n := 0; n < b.N; n++ {
		cCompiledTemplate.Execute(ioutil.Discard, nil)
	}
}

func BenchmarkRenderWithJitTemplateC(b *testing.B) {
	for n := 0; n < b.N; n++ {
		cJitTemplate.Execute(ioutil.Discard, nil)
	}
}

func BenchmarkRenderWithCompiledTemplateD(b *testing.B) {
	for n := 0; n < b.N; n++ {
		dCompiledTemplate.Execute(ioutil.Discard, nil)
	}
}

func BenchmarkRenderWithJitTemplateD(b *testing.B) {
	for n := 0; n < b.N; n++ {
		dJitTemplate.Execute(ioutil.Discard, nil)
	}
}
