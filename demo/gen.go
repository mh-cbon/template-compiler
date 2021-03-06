package main

//golint:ignore

import (
	"io"
	"github.com/mh-cbon/template-compiler/std/text/template/parse"
	aliasdata "github.com/mh-cbon/template-compiler/demo/data"
	"text/template"
	aliastemplate "github.com/mh-cbon/template-compiler/std/html/template"
	"bytes"
	"fmt"
)

var builtin6 = []byte("\n  <ul>\n  ")
var builtin10 = []byte("\nNo items!\n")
var builtin2 = []byte("\n")
var builtin4 = []byte("Hello")
var builtin3 = []byte(" World!\n")
var builtin5 = []byte("This is a template!\n\n")
var builtin7 = []byte("\n    <li>")
var builtin8 = []byte("</li>\n  ")
var builtin9 = []byte("\n  </ul>\n")
var builtin11 = []byte("hello!")
var builtin0 = []byte("Hello from a!\n")
var builtin1 = []byte("Hello from b!\n")


func init () {
  compiledTemplates.Add("a.tpl", fnaTpl)
  compiledTemplates.Add("b.tpl", fnbTpl)
  compiledTemplates.Add("c.tpl", fncTpl)
  compiledTemplates.Add("d.tpl", fndTpl)
  compiledTemplates.Add("tt", fndTplTt)
  compiledTemplates.Add("e.tpl", fneTpl)
  compiledTemplates.Add("f.tpl", fnfTpl)
  compiledTemplates.Add("embed", fnnotafileEmbed)
  compiledTemplates.Add("notafile", fnnotafile)
  tpl3X0 := compiledTemplates.MustGet("d.tpl")
  tpl3Y0 := compiledTemplates.MustGet("tt")
  tpl3X0, _ = tpl3X0.Compiled(tpl3Y0)
  compiledTemplates.Set("d.tpl", tpl3X0)
  tpl0X0 := compiledTemplates.MustGet("notafile")
  tpl0Y0 := compiledTemplates.MustGet("embed")
  tpl0X0, _ = tpl0X0.Compiled(tpl0Y0)
  compiledTemplates.Set("notafile", tpl0X0)
}

func fnaTpl(t parse.Templater, w io.Writer, indata interface{}) error {
	if _, werr := w.Write(builtin0); werr != nil {
		return werr
	}
	return nil
}

func fnbTpl(t parse.Templater, w io.Writer, indata interface{}) error {
	if _, werr := w.Write(builtin1); werr != nil {
		return werr
	}
	return nil
}

func fncTpl(t parse.Templater, w io.Writer, indata interface{}) error {
	var bw bytes.Buffer
	var data aliasdata.MyTemplateData
	if d, ok := indata.(aliasdata.MyTemplateData); ok {
		data = d
	}
	var tplZ int = 4
	if _, werr := w.Write(builtin2); werr != nil {
		return werr
	}
	var tplA string = template.HTMLEscaper(tplZ)
	if _, werr := w.Write(builtin2); werr != nil {
		return werr
	}
	var tplY string = data.Some
	if _, werr := w.Write(builtin2); werr != nil {
		return werr
	}
	var var0 string = aliastemplate.HTMLEscaper(tplZ)
	if _, werr := io.WriteString(w, var0); werr != nil {
		return werr
	}
	if _, werr := w.Write(builtin2); werr != nil {
		return werr
	}
	bw.WriteString(tplA)
	template.HTMLEscape(w, bw.Bytes())
	bw.Reset()
	if _, werr := w.Write(builtin2); werr != nil {
		return werr
	}
	bw.WriteString(tplY)
	template.HTMLEscape(w, bw.Bytes())
	bw.Reset()
	if _, werr := w.Write(builtin2); werr != nil {
		return werr
	}
	var var4 string = data.Some
	bw.WriteString(var4)
	template.HTMLEscape(w, bw.Bytes())
	bw.Reset()
	if _, werr := w.Write(builtin2); werr != nil {
		return werr
	}
	var tplP string = data.Some
	if _, werr := w.Write(builtin2); werr != nil {
		return werr
	}
	var var6 string = data.Some
	bw.WriteString(var6)
	template.HTMLEscape(w, bw.Bytes())
	bw.Reset()
	if _, werr := w.Write(builtin2); werr != nil {
		return werr
	}
	var var8 string = data.Some
	bw.WriteString(var8)
	template.HTMLEscape(w, bw.Bytes())
	bw.Reset()
	if _, werr := w.Write(builtin2); werr != nil {
		return werr
	}
	var var10 string = data.Some
	bw.WriteString(var10)
	template.HTMLEscape(w, bw.Bytes())
	bw.Reset()
	if _, werr := w.Write(builtin2); werr != nil {
		return werr
	}
	bw.WriteString(tplP)
	template.HTMLEscape(w, bw.Bytes())
	bw.Reset()
	if _, werr := w.Write(builtin2); werr != nil {
		return werr
	}
	return nil
}

func fndTpl(t parse.Templater, w io.Writer, indata interface{}) error {
	if _, werr := w.Write(builtin2); werr != nil {
		return werr
	}
	if werr := t.ExecuteTemplate(w, "tt", nil); werr != nil {
		return werr
	}
	if _, werr := w.Write(builtin3); werr != nil {
		return werr
	}
	return nil
}

func fndTplTt(t parse.Templater, w io.Writer, indata interface{}) error {
	if _, werr := w.Write(builtin4); werr != nil {
		return werr
	}
	return nil
}

func fneTpl(t parse.Templater, w io.Writer, indata interface{}) error {
	var bw bytes.Buffer
	var data aliasdata.MyTemplateData
	if d, ok := indata.(aliasdata.MyTemplateData); ok {
		data = d
	}
	if _, werr := w.Write(builtin5); werr != nil {
		return werr
	}
	var var2 []string = data.Items
	var var1 int = len(var2)
	var var0 bool = 0 != var1
	if var0 {
		if _, werr := w.Write(builtin6); werr != nil {
			return werr
		}
		var var3 []string = data.Items
		for _, iterable := range var3 {
			if _, werr := w.Write(builtin7); werr != nil {
				return werr
			}
			bw.WriteString(iterable)
			template.HTMLEscape(w, bw.Bytes())
			bw.Reset()
			if _, werr := w.Write(builtin8); werr != nil {
				return werr
			}
		}
		if _, werr := w.Write(builtin9); werr != nil {
			return werr
		}
	} else {
		if _, werr := w.Write(builtin10); werr != nil {
			return werr
		}
	}
	if _, werr := w.Write(builtin2); werr != nil {
		return werr
	}
	return nil
}

func fnfTpl(t parse.Templater, w io.Writer, indata interface{}) error {
	var bw bytes.Buffer
	var data aliasdata.MyTemplateData
	if d, ok := indata.(aliasdata.MyTemplateData); ok {
		data = d
	}
	if _, werr := w.Write(builtin5); werr != nil {
		return werr
	}
	var var2 []string = data.MethodItems()
	var var1 int = len(var2)
	var var0 bool = 0 != var1
	if var0 {
		if _, werr := w.Write(builtin6); werr != nil {
			return werr
		}
		var var3 []string = data.MethodItems()
		for _, iterable := range var3 {
			if _, werr := w.Write(builtin7); werr != nil {
				return werr
			}
			bw.WriteString(iterable)
			template.HTMLEscape(w, bw.Bytes())
			bw.Reset()
			if _, werr := w.Write(builtin8); werr != nil {
				return werr
			}
		}
		if _, werr := w.Write(builtin9); werr != nil {
			return werr
		}
	} else {
		if _, werr := w.Write(builtin10); werr != nil {
			return werr
		}
	}
	if _, werr := w.Write(builtin2); werr != nil {
		return werr
	}
	return nil
}

func fnnotafileEmbed(t parse.Templater, w io.Writer, indata interface{}) error {
	var data aliasdata.MyTemplateData
	if d, ok := indata.(aliasdata.MyTemplateData); ok {
		data = d
	}
	if _, werr := fmt.Fprintf(w, "%v", data); werr != nil {
		return werr
	}
	return nil
}

func fnnotafile(t parse.Templater, w io.Writer, indata interface{}) error {
	if _, werr := w.Write(builtin11); werr != nil {
		return werr
	}
	return nil
}

