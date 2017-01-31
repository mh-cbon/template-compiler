package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/mh-cbon/template-compiler/demo/data"
	"github.com/mh-cbon/template-compiler/std/html/template"
	text "github.com/mh-cbon/template-compiler/std/text/template"
)

var aJitTemplate *template.Template
var bJitTemplate *template.Template
var cJitTemplate *template.Template
var dJitTemplate *template.Template
var eJitTemplate *template.Template
var fJitTemplate *template.Template

var aCompiledTemplate *text.Compiled
var bCompiledTemplate *text.Compiled
var cCompiledTemplate *text.Compiled
var dCompiledTemplate *text.Compiled
var eCompiledTemplate *text.Compiled
var fCompiledTemplate *text.Compiled

var tplData = data.MyTemplateData{}

func init() {
	var err error

	if aJitTemplate, err = template.ParseFiles("templates/a.tpl"); err != nil {
		panic(err)
	}
	if bJitTemplate, err = template.ParseFiles("templates/b.tpl"); err != nil {
		panic(err)
	}
	if cJitTemplate, err = template.ParseFiles("templates/c.tpl"); err != nil {
		panic(err)
	}
	if dJitTemplate, err = template.ParseFiles("templates/d.tpl"); err != nil {
		panic(err)
	}
	if eJitTemplate, err = template.ParseFiles("templates/e.tpl"); err != nil {
		panic(err)
	}
	if fJitTemplate, err = template.ParseFiles("templates/f.tpl"); err != nil {
		panic(err)
	}
	aCompiledTemplate = compiledTemplates.MustGet("a.tpl")
	bCompiledTemplate = compiledTemplates.MustGet("b.tpl")
	cCompiledTemplate = compiledTemplates.MustGet("c.tpl")
	dCompiledTemplate = compiledTemplates.MustGet("d.tpl")
	eCompiledTemplate = compiledTemplates.MustGet("e.tpl")
	fCompiledTemplate = compiledTemplates.MustGet("f.tpl")

	tplData.Some = "Some string"
	for i := 0; i < 20; i++ {
		tplData.Items = append(tplData.Items, fmt.Sprintf("item %v", i))
	}
}

func TestTemplatesA(t *testing.T) {
	var a bytes.Buffer
	if err := aJitTemplate.Execute(&a, tplData); err != nil {
		panic(err)
	}
	var b bytes.Buffer
	if err := aCompiledTemplate.Execute(&b, tplData); err != nil {
		panic(err)
	}
	aS := a.String()
	bS := b.String()
	if aS != bS {
		t.Errorf("nop\n%v\n%v", aS, bS)
	}
}
func TestTemplatesB(t *testing.T) {
	var a bytes.Buffer
	if err := bJitTemplate.Execute(&a, tplData); err != nil {
		panic(err)
	}
	var b bytes.Buffer
	if err := bCompiledTemplate.Execute(&b, tplData); err != nil {
		panic(err)
	}
	aS := a.String()
	bS := b.String()
	if aS != bS {
		t.Errorf("nop\n%v\n%v", aS, bS)
	}
}
func TestTemplatesC(t *testing.T) {
	var a bytes.Buffer
	if err := cJitTemplate.Execute(&a, tplData); err != nil {
		panic(err)
	}
	var b bytes.Buffer
	if err := cCompiledTemplate.Execute(&b, tplData); err != nil {
		panic(err)
	}
	aS := a.String()
	bS := b.String()
	if aS != bS {
		t.Errorf("nop\n%v\n%v", aS, bS)
	}
}
func TestTemplatesD(t *testing.T) {
	var a bytes.Buffer
	if err := dJitTemplate.Execute(&a, tplData); err != nil {
		panic(err)
	}
	var b bytes.Buffer
	if err := dCompiledTemplate.Execute(&b, tplData); err != nil {
		panic(err)
	}
	aS := a.String()
	bS := b.String()
	if aS != bS {
		t.Errorf("nop\n'%v'\n'%v'", aS, bS)
	}
}
func TestTemplatesE(t *testing.T) {
	var a bytes.Buffer
	if err := eJitTemplate.Execute(&a, tplData); err != nil {
		panic(err)
	}
	var b bytes.Buffer
	if err := eCompiledTemplate.Execute(&b, tplData); err != nil {
		panic(err)
	}
	aS := a.String()
	bS := b.String()
	if aS != bS {
		t.Errorf("nop\n'%v'\n'%v'", aS, bS)
	}
}
func TestTemplatesF(t *testing.T) {
	var a bytes.Buffer
	if err := fJitTemplate.Execute(&a, tplData); err != nil {
		panic(err)
	}
	var b bytes.Buffer
	if err := fCompiledTemplate.Execute(&b, tplData); err != nil {
		panic(err)
	}
	aS := a.String()
	bS := b.String()
	if aS != bS {
		t.Errorf("nop\n'%v'\n'%v'", aS, bS)
	}
}

func BenchmarkRenderWithCompiledTemplateA(b *testing.B) {
	for n := 0; n < b.N; n++ {
		aCompiledTemplate.Execute(ioutil.Discard, tplData)
	}
}

func BenchmarkRenderWithJitTemplateA(b *testing.B) {
	for n := 0; n < b.N; n++ {
		aJitTemplate.Execute(ioutil.Discard, tplData)
	}
}

func BenchmarkRenderWithCompiledTemplateB(b *testing.B) {
	for n := 0; n < b.N; n++ {
		bCompiledTemplate.Execute(ioutil.Discard, tplData)
	}
}

func BenchmarkRenderWithJitTemplateB(b *testing.B) {
	for n := 0; n < b.N; n++ {
		bJitTemplate.Execute(ioutil.Discard, tplData)
	}
}

func BenchmarkRenderWithCompiledTemplateC(b *testing.B) {
	for n := 0; n < b.N; n++ {
		cCompiledTemplate.Execute(ioutil.Discard, tplData)
	}
}

func BenchmarkRenderWithJitTemplateC(b *testing.B) {
	for n := 0; n < b.N; n++ {
		cJitTemplate.Execute(ioutil.Discard, tplData)
	}
}

func BenchmarkRenderWithCompiledTemplateD(b *testing.B) {
	for n := 0; n < b.N; n++ {
		dCompiledTemplate.Execute(ioutil.Discard, tplData)
	}
}

func BenchmarkRenderWithJitTemplateD(b *testing.B) {
	for n := 0; n < b.N; n++ {
		dJitTemplate.Execute(ioutil.Discard, tplData)
	}
}

func BenchmarkRenderWithCompiledTemplateE(b *testing.B) {
	for n := 0; n < b.N; n++ {
		eCompiledTemplate.Execute(ioutil.Discard, tplData)
	}
}

func BenchmarkRenderWithJitTemplateE(b *testing.B) {
	for n := 0; n < b.N; n++ {
		eJitTemplate.Execute(ioutil.Discard, tplData)
	}
}

func BenchmarkRenderWithCompiledTemplateF(b *testing.B) {
	for n := 0; n < b.N; n++ {
		fCompiledTemplate.Execute(ioutil.Discard, tplData)
	}
}

func BenchmarkRenderWithJitTemplateF(b *testing.B) {
	for n := 0; n < b.N; n++ {
		fJitTemplate.Execute(ioutil.Discard, tplData)
	}
}
