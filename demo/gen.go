package main

//golint:ignore

import (
	"io"
	"strconv"
	"text/template"

	"github.com/mh-cbon/template-compiler/compiled"
	dataalias "github.com/mh-cbon/template-compiler/demo/data"
	"github.com/mh-cbon/template-compiler/std/text/template/parse"
)

func init() {
	compiledTemplates = compiled.NewRegistry()
	compiledTemplates.Add("a.tpl", fnaTplaTpl0)
	compiledTemplates.Add("b.tpl", fnbTplbTpl0)
	compiledTemplates.Add("c.tpl", fncTplcTpl0)
	compiledTemplates.Add("d.tpl", fndTpldTpl0)
	compiledTemplates.Add("tt", fndTpltt1)
	tpl3X0 := compiledTemplates.MustGet("d.tpl")
	tpl3Y0 := compiledTemplates.MustGet("tt")
	tpl3X0, _ = tpl3X0.Compiled(tpl3Y0)
	compiledTemplates.Set("d.tpl", tpl3X0)
}

func fnaTplaTpl0(t parse.Templater, w io.Writer, indata interface {
}) error {
	var writeErr error
	_, writeErr = w.Write(builtin0)
	if writeErr != nil {
		return writeErr
	}
	return nil
}

func fnbTplbTpl0(t parse.Templater, w io.Writer, indata interface {
}) error {
	var writeErr error
	_, writeErr = w.Write(builtin1)
	if writeErr != nil {
		return writeErr
	}
	return nil
}

func fncTplcTpl0(t parse.Templater, w io.Writer, indata interface {
}) error {
	var data dataalias.MyTemplateData
	if d, ok := indata.(dataalias.MyTemplateData); ok {
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
	_, writeErr = io.WriteString(w, strconv.Itoa(tplZ))
	if writeErr != nil {
		return writeErr
	}
	_, writeErr = w.Write(builtin2)
	if writeErr != nil {
		return writeErr
	}
	_, writeErr = io.WriteString(w, tplA)
	if writeErr != nil {
		return writeErr
	}
	_, writeErr = w.Write(builtin2)
	if writeErr != nil {
		return writeErr
	}
	_, writeErr = io.WriteString(w, tplY)
	if writeErr != nil {
		return writeErr
	}
	_, writeErr = w.Write(builtin2)
	if writeErr != nil {
		return writeErr
	}
	var var0 string = data.Some
	_, writeErr = io.WriteString(w, var0)
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
	var var1 string = data.Some
	_, writeErr = io.WriteString(w, var1)
	if writeErr != nil {
		return writeErr
	}
	_, writeErr = w.Write(builtin2)
	if writeErr != nil {
		return writeErr
	}
	var var2 string = data.Some
	_, writeErr = io.WriteString(w, var2)
	if writeErr != nil {
		return writeErr
	}
	_, writeErr = w.Write(builtin2)
	if writeErr != nil {
		return writeErr
	}
	var var3 string = data.Some
	_, writeErr = io.WriteString(w, var3)
	if writeErr != nil {
		return writeErr
	}
	_, writeErr = w.Write(builtin2)
	if writeErr != nil {
		return writeErr
	}
	_, writeErr = io.WriteString(w, tplP)
	if writeErr != nil {
		return writeErr
	}
	_, writeErr = w.Write(builtin2)
	if writeErr != nil {
		return writeErr
	}
	return nil
}

func fndTpldTpl0(t parse.Templater, w io.Writer, indata interface {
}) error {
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

func fndTpltt1(t parse.Templater, w io.Writer, indata interface {
}) error {
	var writeErr error
	_, writeErr = w.Write(builtin4)
	if writeErr != nil {
		return writeErr
	}
	return nil
}

var builtin2 = []byte("\n")
var builtin3 = []byte(" World!\n")
var builtin4 = []byte("Hello")
var builtin0 = []byte("Hello from a!\n")
var builtin1 = []byte("Hello from b!\n")
