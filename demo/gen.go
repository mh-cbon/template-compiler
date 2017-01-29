package main

//golint:ignore

import (
	"io"
	"github.com/mh-cbon/template-compiler/std/text/template/parse"
	aliasdata "github.com/mh-cbon/template-compiler/demo/data"
	"text/template"
	aliastemplate "github.com/mh-cbon/template-compiler/std/html/template"
)

var builtin2 = []byte("\n")
var builtin3 = []byte(" World!\n")
var builtin4 = []byte("Hello")
var builtin0 = []byte("Hello from a!\n")
var builtin1 = []byte("Hello from b!\n")


func init () {
  compiledTemplates.Add("a.tpl", fnaTpl)
  compiledTemplates.Add("b.tpl", fnbTpl)
  compiledTemplates.Add("c.tpl", fncTpl)
  compiledTemplates.Add("d.tpl", fndTpl)
  compiledTemplates.Add("tt", fndTplTt)
  tpl3X0 := compiledTemplates.MustGet("d.tpl")
  tpl3Y0 := compiledTemplates.MustGet("tt")
  tpl3X0, _ = tpl3X0.Compiled(tpl3Y0)
  compiledTemplates.Set("d.tpl", tpl3X0)
}

func fnaTpl(t parse.Templater, w io.Writer, indata interface{}) error {
	var writeErr error
	_, writeErr = w.Write(builtin0)
	if writeErr != nil {
		return writeErr
	}
	return nil
}

func fnbTpl(t parse.Templater, w io.Writer, indata interface{}) error {
	var writeErr error
	_, writeErr = w.Write(builtin1)
	if writeErr != nil {
		return writeErr
	}
	return nil
}

func fncTpl(t parse.Templater, w io.Writer, indata interface{}) error {
	var data aliasdata.MyTemplateData
	if d, ok := indata.(aliasdata.MyTemplateData); ok {
		data = d
	}
	var writeErr error
	var tplZ int = 4
	_, writeErr = w.Write(builtin2)
	if writeErr != nil {
		return writeErr
	}
	var tplA string = template.HTMLEscaper(tplZ)
	_, writeErr = w.Write(builtin2)
	if writeErr != nil {
		return writeErr
	}
	var tplY string = data.Some
	_, writeErr = w.Write(builtin2)
	if writeErr != nil {
		return writeErr
	}
	var var0 string = aliastemplate.HTMLEscaper(tplZ)
	_, writeErr = io.WriteString(w, var0)
	if writeErr != nil {
		return writeErr
	}
	_, writeErr = w.Write(builtin2)
	if writeErr != nil {
		return writeErr
	}
	var var1 string = aliastemplate.HTMLEscaper(tplA)
	_, writeErr = io.WriteString(w, var1)
	if writeErr != nil {
		return writeErr
	}
	_, writeErr = w.Write(builtin2)
	if writeErr != nil {
		return writeErr
	}
	var var2 string = aliastemplate.HTMLEscaper(tplY)
	_, writeErr = io.WriteString(w, var2)
	if writeErr != nil {
		return writeErr
	}
	_, writeErr = w.Write(builtin2)
	if writeErr != nil {
		return writeErr
	}
	var var3 string = aliastemplate.HTMLEscaper(data.Some)
	_, writeErr = io.WriteString(w, var3)
	if writeErr != nil {
		return writeErr
	}
	_, writeErr = w.Write(builtin2)
	if writeErr != nil {
		return writeErr
	}
	var tplP string = data.Some
	_, writeErr = w.Write(builtin2)
	if writeErr != nil {
		return writeErr
	}
	var var4 string = aliastemplate.HTMLEscaper(data.Some)
	_, writeErr = io.WriteString(w, var4)
	if writeErr != nil {
		return writeErr
	}
	_, writeErr = w.Write(builtin2)
	if writeErr != nil {
		return writeErr
	}
	var var5 string = aliastemplate.HTMLEscaper(data.Some)
	_, writeErr = io.WriteString(w, var5)
	if writeErr != nil {
		return writeErr
	}
	_, writeErr = w.Write(builtin2)
	if writeErr != nil {
		return writeErr
	}
	var var6 string = aliastemplate.HTMLEscaper(data.Some)
	_, writeErr = io.WriteString(w, var6)
	if writeErr != nil {
		return writeErr
	}
	_, writeErr = w.Write(builtin2)
	if writeErr != nil {
		return writeErr
	}
	var var7 string = aliastemplate.HTMLEscaper(tplP)
	_, writeErr = io.WriteString(w, var7)
	if writeErr != nil {
		return writeErr
	}
	_, writeErr = w.Write(builtin2)
	if writeErr != nil {
		return writeErr
	}
	return nil
}

func fndTpl(t parse.Templater, w io.Writer, indata interface{}) error {
	var writeErr error
	_, writeErr = w.Write(builtin2)
	if writeErr != nil {
		return writeErr
	}
	writeErr = t.ExecuteTemplate(w, "tt", nil)
	if writeErr != nil {
		return writeErr
	}
	_, writeErr = w.Write(builtin3)
	if writeErr != nil {
		return writeErr
	}
	return nil
}

func fndTplTt(t parse.Templater, w io.Writer, indata interface{}) error {
	var writeErr error
	_, writeErr = w.Write(builtin4)
	if writeErr != nil {
		return writeErr
	}
	return nil
}

