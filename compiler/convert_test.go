package compiler

import (
	html "html/template"
	"io/ioutil"
	"testing"
	"text/template"
	"text/template/parse"

	"github.com/mh-cbon/template-tree-simplifier/simplifier"
)

type TestData struct {
	// the template source string
	tplstr string
	// the template data
	dataValue interface{}
	// the expected function result
	expectCompiledFn string
	// the expected function result for an HTML template
	expectHTMLCompiledFn string
	// the expected list of builtins text node variables
	expectBuiltins map[string]string
	// the func to compile the template
	funcs template.FuncMap
	// the list of func map with a public identifier
	funcsMapPublic []map[string]string
	// the expected imports of the resulting go compilation
	expectImports []string
	// the expected imports of the resulting go compilation for an HTML template
	expectHTMLImports []string
}

type TemplateData struct {
	SomeString            string
	SomeBool              bool
	SomeInt               int
	SomeInt8              int8
	SomeInt16             int16
	SomeInt32             int32
	SomeInt64             int64
	SomeUint              uint
	SomeUint8             uint8
	SomeUint16            uint16
	SomeUint32            uint32
	SomeUint64            uint64
	SomeFloat32           float32
	SomeFloat64           float64
	SomeRune              rune
	SomeRuneSlice         []rune
	SomeByte              byte
	SomeByteSlice         []byte
	SomeInterface         interface{}
	SomeTemplateData      *TemplateData
	SomeTemplateDataSlice []*TemplateData
}

func (t TemplateData) MethodHello() string {
	return "hello!"
}

func (t TemplateData) MethodArgHello(name string) string {
	return "hello " + name + "!"
}

func (t TemplateData) MethodArgHello2(name string, name2 string) string {
	return "hello " + name + " " + name2 + "!"
}

func (t TemplateData) MethodArgHelloMultipleReturn(name string, name2 string) (string, error) {
	return "hello " + name + " " + name2 + "!", nil
}

func TestConvert(t *testing.T) {

	allTestData := []TestData{
		TestData{
			tplstr:         `Hello!`,
			dataValue:      TemplateData{},
			expectBuiltins: map[string]string{"Hello!": "builtin0"},
			expectCompiledFn: `func fn0(t parse.Templater, w io.Writer, indata interface {}) error {
  if _, werr := w.Write(builtin0); werr != nil {
    return werr
  }
  return nil
}`,
			expectImports: []string{
				"io",
				"github.com/mh-cbon/template-compiler/std/text/template/parse",
			},
			expectHTMLCompiledFn: `func fn0(t parse.Templater, w io.Writer, indata interface{}) error {
if _, werr := w.Write(builtin0); werr != nil {
return werr
}
return nil
}`,
			expectHTMLImports: []string{
				"io",
				"github.com/mh-cbon/template-compiler/std/text/template/parse",
			},
		},
		TestData{
			tplstr:    `{{$y := "Hello!"}}`,
			dataValue: TemplateData{},
			expectCompiledFn: `func fn0(t parse.Templater, w io.Writer, indata interface {}) error {
  var tplY string = "Hello!"
  return nil
}`,
			expectImports: []string{
				"io",
				"github.com/mh-cbon/template-compiler/std/text/template/parse",
			},
			expectHTMLCompiledFn: `func fn0(t parse.Templater, w io.Writer, indata interface{}) error {
  var tplY string = "Hello!"
  return nil
}`,
			expectHTMLImports: []string{
				"io",
				"github.com/mh-cbon/template-compiler/std/text/template/parse",
			},
		},
		TestData{
			tplstr:    `{{$y := "Hello!"}}{{$y}}`,
			dataValue: TemplateData{},
			expectCompiledFn: `func fn0(t parse.Templater, w io.Writer, indata interface {}) error {
  var tplY string = "Hello!"
  if _, werr := io.WriteString(w, tplY); werr != nil {
    return werr
  }
  return nil
}`,
			expectImports: []string{
				"io",
				"github.com/mh-cbon/template-compiler/std/text/template/parse",
			},
			expectHTMLCompiledFn: `func fn0(t parse.Templater, w io.Writer, indata interface{}) error {
  var bw bytes.Buffer
  var tplY string = "Hello!"
  bw.WriteString(tplY)
  template.HTMLEscape(w, bw.Bytes())
  bw.Reset()
  return nil
}`,
			expectHTMLImports: []string{
				"io",
				"github.com/mh-cbon/template-compiler/std/text/template/parse",
				"text/template",
				"bytes",
			},
		},
		TestData{
			tplstr:         `{{if true}}true{{else}}false{{end}}`,
			dataValue:      TemplateData{},
			expectBuiltins: map[string]string{"true": "builtin0", "false": "builtin1"},
			expectCompiledFn: `func fn0(t parse.Templater, w io.Writer, indata interface {}) error {
  if true {
    if _, werr := w.Write(builtin0); werr != nil {
      return werr
    }
  } else {
    if _, werr := w.Write(builtin1); werr != nil {
      return werr
    }
  }
  return nil
}`,
			expectImports: []string{
				"io",
				"github.com/mh-cbon/template-compiler/std/text/template/parse",
			},
			expectHTMLCompiledFn: `func fn0(t parse.Templater, w io.Writer, indata interface {}) error {
  if true {
    if _, werr := w.Write(builtin0); werr != nil {
      return werr
    }
  } else {
    if _, werr := w.Write(builtin1); werr != nil {
      return werr
    }
  }
  return nil
}`,
			expectHTMLImports: []string{
				"io",
				"github.com/mh-cbon/template-compiler/std/text/template/parse",
			},
		},
		TestData{
			expectBuiltins: map[string]string{"\n": "builtin0"},
			tplstr: `{{.SomeString}}
{{.SomeInt}}
{{.SomeBool}}
{{.SomeInt8}}
{{.SomeInt16}}
{{.SomeInt32}}
{{.SomeInt64}}
{{.SomeUint}}
{{.SomeUint8}}
{{.SomeUint16}}
{{.SomeUint32}}
{{.SomeUint64}}
{{.SomeFloat32}}
{{.SomeFloat64}}
{{.SomeRune}}
{{.SomeByte}}
{{.SomeByteSlice}}
{{.SomeRuneSlice}}`,
			dataValue: TemplateData{SomeString: "Hello!"},
			expectCompiledFn: `func fn0(t parse.Templater, w io.Writer, indata interface {}) error {
  var data compiler.TemplateData
  if d, ok := indata.(compiler.TemplateData); ok {
    data = d
  }
  var var0 string = data.SomeString
  if _, werr := io.WriteString(w, var0); werr != nil {
    return werr
  }
  if _, werr := w.Write(builtin0); werr != nil {
    return werr
  }
  var var1 int = data.SomeInt
  if _, werr := io.WriteString(w, strconv.Itoa(var1)); werr != nil {
    return werr
  }
  if _, werr := w.Write(builtin0); werr != nil {
    return werr
  }
  var var2 bool = data.SomeBool
  if _, werr := io.WriteString(w, strconv.FormatBool(var2)); werr != nil {
    return werr
  }
  if _, werr := w.Write(builtin0); werr != nil {
    return werr
  }
  var var3 int8 = data.SomeInt8
  if _, werr := io.WriteString(w, strconv.FormatInt(int64(var3), 10)); werr != nil {
    return werr
  }
  if _, werr := w.Write(builtin0); werr != nil {
    return werr
  }
  var var4 int16 = data.SomeInt16
  if _, werr := io.WriteString(w, strconv.FormatInt(int64(var4), 10)); werr != nil {
    return werr
  }
  if _, werr := w.Write(builtin0); werr != nil {
    return werr
  }
  var var5 int32 = data.SomeInt32
  if _, werr := io.WriteString(w, strconv.FormatInt(int64(var5), 10)); werr != nil {
    return werr
  }
  if _, werr := w.Write(builtin0); werr != nil {
    return werr
  }
  var var6 int64 = data.SomeInt64
  if _, werr := io.WriteString(w, strconv.FormatInt(var6, 10)); werr != nil {
    return werr
  }
  if _, werr := w.Write(builtin0); werr != nil {
    return werr
  }
  var var7 uint = data.SomeUint
  if _, werr := io.WriteString(w, strconv.FormatUint(var7, 10)); werr != nil {
    return werr
  }
  if _, werr := w.Write(builtin0); werr != nil {
    return werr
  }
  var var8 uint8 = data.SomeUint8
  if _, werr := io.WriteString(w, strconv.FormatUint(var8, 10)); werr != nil {
    return werr
  }
  if _, werr := w.Write(builtin0); werr != nil {
    return werr
  }
  var var9 uint16 = data.SomeUint16
  if _, werr := io.WriteString(w, strconv.FormatUint(var9, 10)); werr != nil {
    return werr
  }
  if _, werr := w.Write(builtin0); werr != nil {
    return werr
  }
  var var10 uint32 = data.SomeUint32
  if _, werr := io.WriteString(w, strconv.FormatUint(var10, 10)); werr != nil {
    return werr
  }
  if _, werr := w.Write(builtin0); werr != nil {
    return werr
  }
  var var11 uint64 = data.SomeUint64
  if _, werr := io.WriteString(w, strconv.FormatUint(var11, 10)); werr != nil {
    return werr
  }
  if _, werr := w.Write(builtin0); werr != nil {
    return werr
  }
  var var12 float32 = data.SomeFloat32
  if _, werr := io.WriteString(w, strconv.FormatFloat(float64(var12), "f", -1, 32)); werr != nil {
    return werr
  }
  if _, werr := w.Write(builtin0); werr != nil {
    return werr
  }
  var var13 float64 = data.SomeFloat64
  if _, werr := io.WriteString(w, strconv.FormatFloat(var13, "f", -1, 64)); werr != nil {
    return werr
  }
  if _, werr := w.Write(builtin0); werr != nil {
    return werr
  }
  var var14 int32 = data.SomeRune
  if _, werr := io.WriteString(w, strconv.FormatInt(int64(var14), 10)); werr != nil {
    return werr
  }
  if _, werr := w.Write(builtin0); werr != nil {
    return werr
  }
  var var15 uint8 = data.SomeByte
  if _, werr := io.WriteString(w, strconv.FormatUint(var15, 10)); werr != nil {
    return werr
  }
  if _, werr := w.Write(builtin0); werr != nil {
    return werr
  }
  var var16 []uint8 = data.SomeByteSlice
  if _, werr := w.Write(var16); werr != nil {
    return werr
  }
  if _, werr := w.Write(builtin0); werr != nil {
    return werr
  }
  var var17 []int32 = data.SomeRuneSlice
  if _, werr := fmt.Fprintf(w, "%v", var17); werr != nil {
    return werr
  }
  return nil
}`,
			expectImports: []string{
				"io",
				"github.com/mh-cbon/template-compiler/std/text/template/parse",
				"github.com/mh-cbon/template-compiler/compiler",
				"strconv",
				"fmt",
			},
			expectHTMLCompiledFn: `func fn0(t parse.Templater, w io.Writer, indata interface {}) error {
  var bw bytes.Buffer
  var data compiler.TemplateData
  if d, ok := indata.(compiler.TemplateData); ok {
    data = d
  }
  var var1 string = data.SomeString
  bw.WriteString(var1)
  template.HTMLEscape(w, bw.Bytes())
  bw.Reset()
  if _, werr := w.Write(builtin0); werr != nil {
    return werr
  }
  var var3 int = data.SomeInt
  var var2 string = aliastemplate.HTMLEscaper(var3)
  if _, werr := io.WriteString(w, var2); werr != nil {
    return werr
  }
  if _, werr := w.Write(builtin0); werr != nil {
    return werr
  }
  var var5 bool = data.SomeBool
  var var4 string = aliastemplate.HTMLEscaper(var5)
  if _, werr := io.WriteString(w, var4); werr != nil {
    return werr
  }
  if _, werr := w.Write(builtin0); werr != nil {
    return werr
  }
  var var7 int8 = data.SomeInt8
  var var6 string = aliastemplate.HTMLEscaper(var7)
  if _, werr := io.WriteString(w, var6); werr != nil {
    return werr
  }
  if _, werr := w.Write(builtin0); werr != nil {
    return werr
  }
  var var9 int16 = data.SomeInt16
  var var8 string = aliastemplate.HTMLEscaper(var9)
  if _, werr := io.WriteString(w, var8); werr != nil {
    return werr
  }
  if _, werr := w.Write(builtin0); werr != nil {
    return werr
  }
  var var11 int32 = data.SomeInt32
  var var10 string = aliastemplate.HTMLEscaper(var11)
  if _, werr := io.WriteString(w, var10); werr != nil {
    return werr
  }
  if _, werr := w.Write(builtin0); werr != nil {
    return werr
  }
  var var13 int64 = data.SomeInt64
  var var12 string = aliastemplate.HTMLEscaper(var13)
  if _, werr := io.WriteString(w, var12); werr != nil {
    return werr
  }
  if _, werr := w.Write(builtin0); werr != nil {
    return werr
  }
  var var15 uint = data.SomeUint
  var var14 string = aliastemplate.HTMLEscaper(var15)
  if _, werr := io.WriteString(w, var14); werr != nil {
    return werr
  }
  if _, werr := w.Write(builtin0); werr != nil {
    return werr
  }
  var var17 uint8 = data.SomeUint8
  var var16 string = aliastemplate.HTMLEscaper(var17)
  if _, werr := io.WriteString(w, var16); werr != nil {
    return werr
  }
  if _, werr := w.Write(builtin0); werr != nil {
    return werr
  }
  var var19 uint16 = data.SomeUint16
  var var18 string = aliastemplate.HTMLEscaper(var19)
  if _, werr := io.WriteString(w, var18); werr != nil {
    return werr
  }
  if _, werr := w.Write(builtin0); werr != nil {
    return werr
  }
  var var21 uint32 = data.SomeUint32
  var var20 string = aliastemplate.HTMLEscaper(var21)
  if _, werr := io.WriteString(w, var20); werr != nil {
    return werr
  }
  if _, werr := w.Write(builtin0); werr != nil {
    return werr
  }
  var var23 uint64 = data.SomeUint64
  var var22 string = aliastemplate.HTMLEscaper(var23)
  if _, werr := io.WriteString(w, var22); werr != nil {
    return werr
  }
  if _, werr := w.Write(builtin0); werr != nil {
    return werr
  }
  var var25 float32 = data.SomeFloat32
  var var24 string = aliastemplate.HTMLEscaper(var25)
  if _, werr := io.WriteString(w, var24); werr != nil {
    return werr
  }
  if _, werr := w.Write(builtin0); werr != nil {
    return werr
  }
  var var27 float64 = data.SomeFloat64
  var var26 string = aliastemplate.HTMLEscaper(var27)
  if _, werr := io.WriteString(w, var26); werr != nil {
    return werr
  }
  if _, werr := w.Write(builtin0); werr != nil {
    return werr
  }
  var var29 int32 = data.SomeRune
  var var28 string = aliastemplate.HTMLEscaper(var29)
  if _, werr := io.WriteString(w, var28); werr != nil {
    return werr
  }
  if _, werr := w.Write(builtin0); werr != nil {
    return werr
  }
  var var31 uint8 = data.SomeByte
  var var30 string = aliastemplate.HTMLEscaper(var31)
  if _, werr := io.WriteString(w, var30); werr != nil {
    return werr
  }
  if _, werr := w.Write(builtin0); werr != nil {
    return werr
  }
  var var33 []uint8 = data.SomeByteSlice
  var var32 string = aliastemplate.HTMLEscaper(var33)
  if _, werr := io.WriteString(w, var32); werr != nil {
    return werr
  }
  if _, werr := w.Write(builtin0); werr != nil {
    return werr
  }
  var var35 []int32 = data.SomeRuneSlice
  var var34 string = aliastemplate.HTMLEscaper(var35)
  if _, werr := io.WriteString(w, var34); werr != nil {
    return werr
  }
  return nil
}`,
			expectHTMLImports: []string{
				"io",
				"bytes",
				"github.com/mh-cbon/template-compiler/std/text/template/parse",
				"github.com/mh-cbon/template-compiler/compiler",
				"text/template",
				"aliastemplate:github.com/mh-cbon/template-compiler/std/html/template",
			},
		},
		TestData{
			tplstr:    `{{range .SomeByteSlice}}{{.}}{{end}}`,
			dataValue: TemplateData{SomeString: "Hello!"},
			expectCompiledFn: `func fn0(t parse.Templater, w io.Writer, indata interface{}) error {
  var data compiler.TemplateData
  if d, ok := indata.(compiler.TemplateData); ok {
    data = d
  }
  var var0 []uint8 = data.SomeByteSlice
  for _, iterable := range var0 {
    if _, werr := io.WriteString(w, strconv.FormatUint(iterable, 10)); werr != nil {
      return werr
    }
  }
  return nil
}`,
			expectImports: []string{
				"io",
				"github.com/mh-cbon/template-compiler/std/text/template/parse",
				"github.com/mh-cbon/template-compiler/compiler",
				"strconv",
			},
			expectHTMLCompiledFn: `func fn0(t parse.Templater, w io.Writer, indata interface{}) error {
  var data compiler.TemplateData
  if d, ok := indata.(compiler.TemplateData); ok {
    data = d
  }
  var var0 []uint8 = data.SomeByteSlice
  for _, iterable := range var0 {
    var var1 string = template.HTMLEscaper(iterable)
    if _, werr := io.WriteString(w, var1); werr != nil {
      return werr
    }
  }
  return nil
}`,
			expectHTMLImports: []string{
				"io",
				"github.com/mh-cbon/template-compiler/std/text/template/parse",
				"github.com/mh-cbon/template-compiler/compiler",
				"github.com/mh-cbon/template-compiler/std/html/template",
			},
		},
		TestData{
			tplstr:    `{{range $i, $v := .SomeByteSlice}}{{.}}{{end}}`,
			dataValue: TemplateData{SomeString: "Hello!"},
			expectCompiledFn: `func fn0(t parse.Templater, w io.Writer, indata interface {}) error {
  var data compiler.TemplateData
  if d, ok := indata.(compiler.TemplateData); ok {
    data = d
  }
  var var0 []uint8 = data.SomeByteSlice
  for tplI, tplV := range var0 {
    if _, werr := io.WriteString(w, strconv.FormatUint(tplV, 10)); werr != nil {
      return werr
    }
  }
  return nil
}`,
			expectImports: []string{
				"io",
				"github.com/mh-cbon/template-compiler/std/text/template/parse",
				"github.com/mh-cbon/template-compiler/compiler",
				"strconv",
			},
			expectHTMLCompiledFn: `func fn0(t parse.Templater, w io.Writer, indata interface{}) error {
  var data compiler.TemplateData
  if d, ok := indata.(compiler.TemplateData); ok {
    data = d
  }
  var var0 []uint8 = data.SomeByteSlice
  for tplI, tplV := range var0 {
    var var1 string = template.HTMLEscaper(tplV)
    if _, werr := io.WriteString(w, var1); werr != nil {
      return werr
    }
  }
  return nil
}`,
			expectHTMLImports: []string{
				"io",
				"github.com/mh-cbon/template-compiler/std/text/template/parse",
				"github.com/mh-cbon/template-compiler/compiler",
				"github.com/mh-cbon/template-compiler/std/html/template",
			},
		},
		TestData{
			tplstr: `{{range $i, $v := .SomeTemplateDataSlice}}
{{range $i, $v := $v.SomeTemplateDataSlice}}
{{end}}
{{end}}`,
			dataValue:      TemplateData{},
			expectBuiltins: map[string]string{"\n": "builtin0"},
			expectCompiledFn: `func fn0(t parse.Templater, w io.Writer, indata interface {}) error {
  var data compiler.TemplateData
  if d, ok := indata.(compiler.TemplateData); ok {
    data = d
  }
  var var0 []*compiler.TemplateData = data.SomeTemplateDataSlice
  for tplI, tplV := range var0 {
    if _, werr := w.Write(builtin0); werr != nil {
      return werr
    }
    var var1 []*compiler.TemplateData = tplV.SomeTemplateDataSlice
    for tplIShadow, tplVShadow := range var1 {
      if _, werr := w.Write(builtin0); werr != nil {
        return werr
      }
    }
    if _, werr := w.Write(builtin0); werr != nil {
      return werr
    }
  }
  return nil
}`,
			expectImports: []string{
				"io",
				"github.com/mh-cbon/template-compiler/std/text/template/parse",
				"github.com/mh-cbon/template-compiler/compiler",
			},
			expectHTMLCompiledFn: `func fn0(t parse.Templater, w io.Writer, indata interface{}) error {
  var data compiler.TemplateData
  if d, ok := indata.(compiler.TemplateData); ok {
    data = d
  }
  var var0 []*compiler.TemplateData = data.SomeTemplateDataSlice
  for tplI, tplV := range var0 {
    if _, werr := w.Write(builtin0); werr != nil {
      return werr
    }
    var var1 []*compiler.TemplateData = tplV.SomeTemplateDataSlice
    for tplIShadow, tplVShadow := range var1 {
    if _, werr := w.Write(builtin0); werr != nil {
      return werr
    }
    }
    if _, werr := w.Write(builtin0); werr != nil {
      return werr
    }
  }
  return nil
}`,
			expectHTMLImports: []string{
				"io",
				"github.com/mh-cbon/template-compiler/std/text/template/parse",
				"github.com/mh-cbon/template-compiler/compiler",
			},
		},
		TestData{
			tplstr: `{{range $i, $v := .SomeTemplateDataSlice}}
Hello range branch!
{{else}}
Hello else branch!
{{end}}`,
			dataValue: TemplateData{},
			expectBuiltins: map[string]string{
				"\nHello range branch!\n": "builtin0",
				"\nHello else branch!\n":  "builtin1",
			},
			expectCompiledFn: `func fn0(t parse.Templater, w io.Writer, indata interface {}) error {
  var data compiler.TemplateData
  if d, ok := indata.(compiler.TemplateData); ok {
    data = d
  }
  var var0 []*compiler.TemplateData = data.SomeTemplateDataSlice
  for tplI, tplV := range var0 {
    if _, werr := w.Write(builtin0); werr != nil {
      return werr
    }
  }
  if len(var0) == 0 {
    if _, werr := w.Write(builtin1); werr != nil {
      return werr
    }
  }
  return nil
}`,
			expectImports: []string{
				"io",
				"github.com/mh-cbon/template-compiler/std/text/template/parse",
				"github.com/mh-cbon/template-compiler/compiler",
			},
			expectHTMLCompiledFn: `func fn0(t parse.Templater, w io.Writer, indata interface{}) error {
  var data compiler.TemplateData
  if d, ok := indata.(compiler.TemplateData); ok {
    data = d
  }
  var var0 []*compiler.TemplateData = data.SomeTemplateDataSlice
  for tplI, tplV := range var0 {
    if _, werr := w.Write(builtin0); werr != nil {
      return werr
    }
  }
  if len(var0) == 0 {
    if _, werr := w.Write(builtin1); werr != nil {
      return werr
    }
  }
  return nil
}`,
			expectHTMLImports: []string{
				"io",
				"github.com/mh-cbon/template-compiler/std/text/template/parse",
				"github.com/mh-cbon/template-compiler/compiler",
			},
		},
		TestData{
			tplstr: `{{with .}}
Hello with branch!
{{else}}
Hello without branch!
{{end}}`,
			dataValue: TemplateData{},
			expectBuiltins: map[string]string{
				"\nHello with branch!\n":    "builtin0",
				"\nHello without branch!\n": "builtin1",
			},
			expectCompiledFn: `func fn0(t parse.Templater, w io.Writer, indata interface {}) error {
  var data compiler.TemplateData
  if d, ok := indata.(compiler.TemplateData); ok {
    data = d
  }
  var var0 compiler.TemplateData = data
  {
    if true {
      if _, werr := w.Write(builtin0); werr != nil {
        return werr
      }
    } else {
      if _, werr := w.Write(builtin1); werr != nil {
        return werr
      }
    }
  }
  return nil
}`,
			expectImports: []string{
				"io",
				"github.com/mh-cbon/template-compiler/std/text/template/parse",
				"github.com/mh-cbon/template-compiler/compiler",
			},
			expectHTMLCompiledFn: `func fn0(t parse.Templater, w io.Writer, indata interface{}) error {
  var data compiler.TemplateData
  if d, ok := indata.(compiler.TemplateData); ok {
    data = d
  }
  var var0 compiler.TemplateData = data
  {
    if true {
      if _, werr := w.Write(builtin0); werr != nil {
        return werr
      }
    } else {
      if _, werr := w.Write(builtin1); werr != nil {
        return werr
      }
    }
  }
  return nil
}`,
			expectHTMLImports: []string{
				"io",
				"github.com/mh-cbon/template-compiler/std/text/template/parse",
				"github.com/mh-cbon/template-compiler/compiler",
			},
		},
		TestData{
			tplstr: `{{with .}}{{.SomeString}}{{else}}{{.SomeString}}{{end}}`,
			// remember, an (if struct) is always true.
			dataValue: TemplateData{},
			expectCompiledFn: `func fn0(t parse.Templater, w io.Writer, indata interface {}) error {
  var data compiler.TemplateData
  if d, ok := indata.(compiler.TemplateData); ok {
    data = d
  }
  var var0 compiler.TemplateData = data
  {
    if true {
      var var1 string = var0.SomeString
      if _, werr := io.WriteString(w, var1); werr != nil {
        return werr
      }
    } else {
      var var2 string = data.SomeString
      if _, werr := io.WriteString(w, var2); werr != nil {
        return werr
      }
    }
  }
  return nil
}`,
			expectImports: []string{
				"io",
				"github.com/mh-cbon/template-compiler/std/text/template/parse",
				"github.com/mh-cbon/template-compiler/compiler",
			},
			expectHTMLCompiledFn: `func fn0(t parse.Templater, w io.Writer, indata interface{}) error {
  var bw bytes.Buffer
  var data compiler.TemplateData
  if d, ok := indata.(compiler.TemplateData); ok {
    data = d
  }
  var var0 compiler.TemplateData = data
  {
    if true {
      var var2 string = var0.SomeString
      bw.WriteString(var2)
      template.HTMLEscape(w, bw.Bytes())
      bw.Reset()
    } else {
      var var4 string = data.SomeString
      bw.WriteString(var4)
      template.HTMLEscape(w, bw.Bytes())
      bw.Reset()
    }
  }
  return nil
}`,
			expectHTMLImports: []string{
				"io",
				"bytes",
				"github.com/mh-cbon/template-compiler/std/text/template/parse",
				"github.com/mh-cbon/template-compiler/compiler",
				"text/template",
			},
		},
		TestData{
			tplstr:    `{{if .SomeString}}{{end}}{{if .SomeString}}{{end}}`,
			dataValue: TemplateData{SomeString: "Hello!"},
			expectCompiledFn: `func fn0(t parse.Templater, w io.Writer, indata interface {}) error {
  var data compiler.TemplateData
  if d, ok := indata.(compiler.TemplateData); ok {
    data = d
  }
  var var0 string = data.SomeString
  if var0 != "" {
  }
  var var1 string = data.SomeString
  if var1 != "" {
  }
  return nil
}`,
			expectImports: []string{
				"io",
				"github.com/mh-cbon/template-compiler/std/text/template/parse",
				"github.com/mh-cbon/template-compiler/compiler",
			},
			expectHTMLCompiledFn: `func fn0(t parse.Templater, w io.Writer, indata interface{}) error {
  var data compiler.TemplateData
  if d, ok := indata.(compiler.TemplateData); ok {
    data = d
  }
  var var0 string = data.SomeString
  if var0 != "" {
  }
  var var1 string = data.SomeString
  if var1 != "" {
  }
  return nil
}`,
			expectHTMLImports: []string{
				"io",
				"github.com/mh-cbon/template-compiler/std/text/template/parse",
				"github.com/mh-cbon/template-compiler/compiler",
			},
		},
		TestData{
			expectBuiltins: map[string]string{"\n": "builtin0"},
			tplstr: `{{if .SomeString}}{{end}}
{{if .SomeInt}}{{end}}
{{if .SomeBool}}{{end}}
{{if .SomeInt8}}{{end}}
{{if .SomeInt16}}{{end}}
{{if .SomeInt32}}{{end}}
{{if .SomeInt64}}{{end}}
{{if .SomeUint}}{{end}}
{{if .SomeUint8}}{{end}}
{{if .SomeUint16}}{{end}}
{{if .SomeUint32}}{{end}}
{{if .SomeUint64}}{{end}}
{{if .SomeFloat32}}{{end}}
{{if .SomeFloat64}}{{end}}
{{if .SomeRune}}{{end}}
{{if .SomeByte}}{{end}}
{{if .SomeByteSlice}}{{end}}
{{if .SomeRuneSlice}}{{end}}`,
			dataValue: TemplateData{SomeString: "Hello!"},
			expectCompiledFn: `func fn0(t parse.Templater, w io.Writer, indata interface {}) error {
  var data compiler.TemplateData
  if d, ok := indata.(compiler.TemplateData); ok {
    data = d
  }
  var var0 string = data.SomeString
  if var0 != "" {
  }
  if _, werr := w.Write(builtin0); werr != nil {
    return werr
  }
  var var1 int = data.SomeInt
  if var1 != 0 {
  }
  if _, werr := w.Write(builtin0); werr != nil {
    return werr
  }
  var var2 bool = data.SomeBool
  if var2 {
  }
  if _, werr := w.Write(builtin0); werr != nil {
    return werr
  }
  var var3 int8 = data.SomeInt8
  if var3 != 0 {
  }
  if _, werr := w.Write(builtin0); werr != nil {
    return werr
  }
  var var4 int16 = data.SomeInt16
  if var4 != 0 {
  }
  if _, werr := w.Write(builtin0); werr != nil {
    return werr
  }
  var var5 int32 = data.SomeInt32
  if var5 != 0 {
  }
  if _, werr := w.Write(builtin0); werr != nil {
    return werr
  }
  var var6 int64 = data.SomeInt64
  if var6 != 0 {
  }
  if _, werr := w.Write(builtin0); werr != nil {
    return werr
  }
  var var7 uint = data.SomeUint
  if var7 != 0 {
  }
  if _, werr := w.Write(builtin0); werr != nil {
    return werr
  }
  var var8 uint8 = data.SomeUint8
  if var8 != 0 {
  }
  if _, werr := w.Write(builtin0); werr != nil {
    return werr
  }
  var var9 uint16 = data.SomeUint16
  if var9 != 0 {
  }
  if _, werr := w.Write(builtin0); werr != nil {
    return werr
  }
  var var10 uint32 = data.SomeUint32
  if var10 != 0 {
  }
  if _, werr := w.Write(builtin0); werr != nil {
    return werr
  }
  var var11 uint64 = data.SomeUint64
  if var11 != 0 {
  }
  if _, werr := w.Write(builtin0); werr != nil {
    return werr
  }
  var var12 float32 = data.SomeFloat32
  if var12 != 0 {
  }
  if _, werr := w.Write(builtin0); werr != nil {
    return werr
  }
  var var13 float64 = data.SomeFloat64
  if var13 != 0 {
  }
  if _, werr := w.Write(builtin0); werr != nil {
    return werr
  }
  var var14 int32 = data.SomeRune
  if var14 != 0 {
  }
  if _, werr := w.Write(builtin0); werr != nil {
    return werr
  }
  var var15 uint8 = data.SomeByte
  if var15 != 0 {
  }
  if _, werr := w.Write(builtin0); werr != nil {
    return werr
  }
  var var16 []uint8 = data.SomeByteSlice
  if len(var16) > 0 {
  }
  if _, werr := w.Write(builtin0); werr != nil {
    return werr
  }
  var var17 []int32 = data.SomeRuneSlice
  if len(var17) > 0 {
  }
  return nil
}`,
			expectImports: []string{
				"io",
				"github.com/mh-cbon/template-compiler/std/text/template/parse",
				"github.com/mh-cbon/template-compiler/compiler",
			},
			expectHTMLCompiledFn: `func fn0(t parse.Templater, w io.Writer, indata interface{}) error {
  var data compiler.TemplateData
  if d, ok := indata.(compiler.TemplateData); ok {
    data = d
  }
  var var0 string = data.SomeString
  if var0 != "" {
  }
  if _, werr := w.Write(builtin0); werr != nil {
    return werr
  }
  var var1 int = data.SomeInt
  if var1 != 0 {
  }
  if _, werr := w.Write(builtin0); werr != nil {
    return werr
  }
  var var2 bool = data.SomeBool
  if var2 {
  }
  if _, werr := w.Write(builtin0); werr != nil {
    return werr
  }
  var var3 int8 = data.SomeInt8
  if var3 != 0 {
  }
  if _, werr := w.Write(builtin0); werr != nil {
    return werr
  }
  var var4 int16 = data.SomeInt16
  if var4 != 0 {
  }
  if _, werr := w.Write(builtin0); werr != nil {
    return werr
  }
  var var5 int32 = data.SomeInt32
  if var5 != 0 {
  }
  if _, werr := w.Write(builtin0); werr != nil {
    return werr
  }
  var var6 int64 = data.SomeInt64
  if var6 != 0 {
  }
  if _, werr := w.Write(builtin0); werr != nil {
    return werr
  }
  var var7 uint = data.SomeUint
  if var7 != 0 {
  }
  if _, werr := w.Write(builtin0); werr != nil {
    return werr
  }
  var var8 uint8 = data.SomeUint8
  if var8 != 0 {
  }
  if _, werr := w.Write(builtin0); werr != nil {
    return werr
  }
  var var9 uint16 = data.SomeUint16
  if var9 != 0 {
  }
  if _, werr := w.Write(builtin0); werr != nil {
    return werr
  }
  var var10 uint32 = data.SomeUint32
  if var10 != 0 {
  }
  if _, werr := w.Write(builtin0); werr != nil {
    return werr
  }
  var var11 uint64 = data.SomeUint64
  if var11 != 0 {
  }
  if _, werr := w.Write(builtin0); werr != nil {
    return werr
  }
  var var12 float32 = data.SomeFloat32
  if var12 != 0 {
  }
  if _, werr := w.Write(builtin0); werr != nil {
    return werr
  }
  var var13 float64 = data.SomeFloat64
  if var13 != 0 {
  }
  if _, werr := w.Write(builtin0); werr != nil {
    return werr
  }
  var var14 int32 = data.SomeRune
  if var14 != 0 {
  }
  if _, werr := w.Write(builtin0); werr != nil {
    return werr
  }
  var var15 uint8 = data.SomeByte
  if var15 != 0 {
  }
  if _, werr := w.Write(builtin0); werr != nil {
    return werr
  }
  var var16 []uint8 = data.SomeByteSlice
  if len(var16) > 0 {
  }
  if _, werr := w.Write(builtin0); werr != nil {
    return werr
  }
  var var17 []int32 = data.SomeRuneSlice
  if len(var17) > 0 {
  }
  return nil
}`,
			expectHTMLImports: []string{
				"io",
				"github.com/mh-cbon/template-compiler/std/text/template/parse",
				"github.com/mh-cbon/template-compiler/compiler",
			},
		},
		TestData{
			tplstr: `{{$int := 4}}
{{$float := 4.0}}
{{$complex := 1i}}`,
			dataValue:      TemplateData{SomeString: "Hello!"},
			expectBuiltins: map[string]string{"\n": "builtin0"},
			expectCompiledFn: `func fn0(t parse.Templater, w io.Writer, indata interface {}) error {
  var tplInt int = 4
  if _, werr := w.Write(builtin0); werr != nil {
    return werr
  }
  var tplFloat float64 = 4.0
  if _, werr := w.Write(builtin0); werr != nil {
    return werr
  }
  var tplComplex complex128 = 1i
  return nil
}`,
			expectImports: []string{
				"io",
				"github.com/mh-cbon/template-compiler/std/text/template/parse",
			},
			expectHTMLCompiledFn: `func fn0(t parse.Templater, w io.Writer, indata interface{}) error {
  var tplInt int = 4
  if _, werr := w.Write(builtin0); werr != nil {
    return werr
  }
  var tplFloat float64 = 4.0
  if _, werr := w.Write(builtin0); werr != nil {
    return werr
  }
  var tplComplex complex128 = 1i
  return nil
}`,
			expectHTMLImports: []string{
				"io",
				"github.com/mh-cbon/template-compiler/std/text/template/parse",
			},
		},
		TestData{
			tplstr:    `{{.MethodHello}}`,
			dataValue: TemplateData{SomeString: "Hello!"},
			expectCompiledFn: `func fn0(t parse.Templater, w io.Writer, indata interface {}) error {
  var data compiler.TemplateData
  if d, ok := indata.(compiler.TemplateData); ok {
    data = d
  }
  var var0 string = data.MethodHello()
  if _, werr := io.WriteString(w, var0); werr != nil {
    return werr
  }
  return nil
}`,
			expectImports: []string{
				"io",
				"github.com/mh-cbon/template-compiler/std/text/template/parse",
				"github.com/mh-cbon/template-compiler/compiler",
			},
			expectHTMLCompiledFn: `func fn0(t parse.Templater, w io.Writer, indata interface{}) error {
  var bw bytes.Buffer
  var data compiler.TemplateData
  if d, ok := indata.(compiler.TemplateData); ok {
    data = d
  }
  var var1 string = data.MethodHello()
  bw.WriteString(var1)
  template.HTMLEscape(w, bw.Bytes())
  bw.Reset()
  return nil
}`,
			expectHTMLImports: []string{
				"io",
				"bytes",
				"github.com/mh-cbon/template-compiler/std/text/template/parse",
				"github.com/mh-cbon/template-compiler/compiler",
				"text/template",
			},
		},
		TestData{
			tplstr:    `{{.MethodArgHello "me"}}`,
			dataValue: TemplateData{SomeString: "Hello!"},
			expectCompiledFn: `func fn0(t parse.Templater, w io.Writer, indata interface {}) error {
  var data compiler.TemplateData
  if d, ok := indata.(compiler.TemplateData); ok {
    data = d
  }
  var var0 string = data.MethodArgHello("me")
  if _, werr := io.WriteString(w, var0); werr != nil {
    return werr
  }
  return nil
}`,
			expectImports: []string{
				"io",
				"github.com/mh-cbon/template-compiler/std/text/template/parse",
				"github.com/mh-cbon/template-compiler/compiler",
			},
			expectHTMLCompiledFn: `func fn0(t parse.Templater, w io.Writer, indata interface{}) error {
  var bw bytes.Buffer
  var data compiler.TemplateData
  if d, ok := indata.(compiler.TemplateData); ok {
    data = d
  }
  var var0 string = data.MethodArgHello("me")
  bw.WriteString(var0)
  template.HTMLEscape(w, bw.Bytes())
  bw.Reset()
  return nil
}`,
			expectHTMLImports: []string{
				"io",
				"bytes",
				"github.com/mh-cbon/template-compiler/std/text/template/parse",
				"github.com/mh-cbon/template-compiler/compiler",
				"text/template",
			},
		},
		TestData{
			tplstr:    `{{.MethodArgHello2 "me" "you"}}`,
			dataValue: TemplateData{SomeString: "Hello!"},
			expectCompiledFn: `func fn0(t parse.Templater, w io.Writer, indata interface {}) error {
  var data compiler.TemplateData
  if d, ok := indata.(compiler.TemplateData); ok {
    data = d
  }
  var var0 string = data.MethodArgHello2("me", "you")
  if _, werr := io.WriteString(w, var0); werr != nil {
    return werr
  }
  return nil
}`,
			expectImports: []string{
				"io",
				"github.com/mh-cbon/template-compiler/std/text/template/parse",
				"github.com/mh-cbon/template-compiler/compiler",
			},
			expectHTMLCompiledFn: `func fn0(t parse.Templater, w io.Writer, indata interface{}) error {
  var bw bytes.Buffer
  var data compiler.TemplateData
  if d, ok := indata.(compiler.TemplateData); ok {
    data = d
  }
  var var0 string = data.MethodArgHello2("me", "you")
  bw.WriteString(var0)
  template.HTMLEscape(w, bw.Bytes())
  bw.Reset()
  return nil
}`,
			expectHTMLImports: []string{
				"io",
				"bytes",
				"github.com/mh-cbon/template-compiler/std/text/template/parse",
				"github.com/mh-cbon/template-compiler/compiler",
				"text/template",
			},
		},
		TestData{
			tplstr:    `{{.MethodArgHelloMultipleReturn "me" "you"}}`,
			dataValue: TemplateData{SomeString: "Hello!"},
			expectCompiledFn: `func fn0(t parse.Templater, w io.Writer, indata interface {}) error {
  var data compiler.TemplateData
  if d, ok := indata.(compiler.TemplateData); ok {
    data = d
  }
  var0, err := data.MethodArgHelloMultipleReturn("me", "you")
  if err != nil {
    return err
  }
  if _, werr := io.WriteString(w, var0); werr != nil {
    return werr
  }
  return nil
}`,
			expectImports: []string{
				"io",
				"github.com/mh-cbon/template-compiler/std/text/template/parse",
				"github.com/mh-cbon/template-compiler/compiler",
			},
			expectHTMLCompiledFn: `func fn0(t parse.Templater, w io.Writer, indata interface{}) error {
  var bw bytes.Buffer
  var data compiler.TemplateData
  if d, ok := indata.(compiler.TemplateData); ok {
    data = d
  }
  var0, err := data.MethodArgHelloMultipleReturn("me", "you")
  if err != nil {
    return err
  }
  bw.WriteString(var0)
  template.HTMLEscape(w, bw.Bytes())
  bw.Reset()
  return nil
}`,
			expectHTMLImports: []string{
				"io",
				"bytes",
				"github.com/mh-cbon/template-compiler/std/text/template/parse",
				"github.com/mh-cbon/template-compiler/compiler",
				"text/template",
			},
		},
		TestData{
			tplstr:    `{{$x := .}}{{$x.MethodHello}}`,
			dataValue: TemplateData{SomeString: "Hello!"},
			expectCompiledFn: `func fn0(t parse.Templater, w io.Writer, indata interface {}) error {
  var data compiler.TemplateData
  if d, ok := indata.(compiler.TemplateData); ok {
    data = d
  }
  var tplX compiler.TemplateData = data
  var var0 string = tplX.MethodHello()
  if _, werr := io.WriteString(w, var0); werr != nil {
    return werr
  }
  return nil
}`,
			expectImports: []string{
				"io",
				"github.com/mh-cbon/template-compiler/std/text/template/parse",
				"github.com/mh-cbon/template-compiler/compiler",
			},
			expectHTMLCompiledFn: `func fn0(t parse.Templater, w io.Writer, indata interface{}) error {
  var bw bytes.Buffer
  var data compiler.TemplateData
  if d, ok := indata.(compiler.TemplateData); ok {
    data = d
  }
  var tplX compiler.TemplateData = data
  var var1 string = tplX.MethodHello()
  bw.WriteString(var1)
  template.HTMLEscape(w, bw.Bytes())
  bw.Reset()
  return nil
}`,
			expectHTMLImports: []string{
				"io",
				"bytes",
				"github.com/mh-cbon/template-compiler/std/text/template/parse",
				"github.com/mh-cbon/template-compiler/compiler",
				"text/template",
			},
		},
		TestData{
			tplstr:    `{{$x := .}}{{$x.MethodArgHello "me"}}`,
			dataValue: TemplateData{SomeString: "Hello!"},
			expectCompiledFn: `func fn0(t parse.Templater, w io.Writer, indata interface {}) error {
  var data compiler.TemplateData
  if d, ok := indata.(compiler.TemplateData); ok {
    data = d
  }
  var tplX compiler.TemplateData = data
  var var0 string = tplX.MethodArgHello("me")
  if _, werr := io.WriteString(w, var0); werr != nil {
    return werr
  }
  return nil
}`,
			expectImports: []string{
				"io",
				"github.com/mh-cbon/template-compiler/std/text/template/parse",
				"github.com/mh-cbon/template-compiler/compiler",
			},
			expectHTMLCompiledFn: `func fn0(t parse.Templater, w io.Writer, indata interface{}) error {
  var bw bytes.Buffer
  var data compiler.TemplateData
  if d, ok := indata.(compiler.TemplateData); ok {
    data = d
  }
  var tplX compiler.TemplateData = data
  var var0 string = tplX.MethodArgHello("me")
  bw.WriteString(var0)
  template.HTMLEscape(w, bw.Bytes())
  bw.Reset()
  return nil
}`,
			expectHTMLImports: []string{
				"io",
				"bytes",
				"github.com/mh-cbon/template-compiler/std/text/template/parse",
				"github.com/mh-cbon/template-compiler/compiler",
				"text/template",
			},
		},
		TestData{
			tplstr:    `{{$x := .}}{{$x.MethodArgHello2 "me" "you"}}`,
			dataValue: TemplateData{SomeString: "Hello!"},
			expectCompiledFn: `func fn0(t parse.Templater, w io.Writer, indata interface {}) error {
  var data compiler.TemplateData
  if d, ok := indata.(compiler.TemplateData); ok {
    data = d
  }
  var tplX compiler.TemplateData = data
  var var0 string = tplX.MethodArgHello2("me", "you")
  if _, werr := io.WriteString(w, var0); werr != nil {
    return werr
  }
  return nil
}`,
			expectImports: []string{
				"io",
				"github.com/mh-cbon/template-compiler/std/text/template/parse",
				"github.com/mh-cbon/template-compiler/compiler",
			},
			expectHTMLCompiledFn: `func fn0(t parse.Templater, w io.Writer, indata interface{}) error {
  var bw bytes.Buffer
  var data compiler.TemplateData
  if d, ok := indata.(compiler.TemplateData); ok {
    data = d
  }
  var tplX compiler.TemplateData = data
  var var0 string = tplX.MethodArgHello2("me", "you")
  bw.WriteString(var0)
  template.HTMLEscape(w, bw.Bytes())
  bw.Reset()
  return nil
}`,
			expectHTMLImports: []string{
				"io",
				"bytes",
				"github.com/mh-cbon/template-compiler/std/text/template/parse",
				"github.com/mh-cbon/template-compiler/compiler",
				"text/template",
			},
		},
		TestData{
			tplstr:    `{{$x := .}}{{$x.MethodArgHelloMultipleReturn "me" "you"}}`,
			dataValue: TemplateData{SomeString: "Hello!"},
			expectCompiledFn: `func fn0(t parse.Templater, w io.Writer, indata interface {}) error {
  var data compiler.TemplateData
  if d, ok := indata.(compiler.TemplateData); ok {
    data = d
  }
  var tplX compiler.TemplateData = data
  var0, err := tplX.MethodArgHelloMultipleReturn("me", "you")
  if err != nil {
    return err
  }
  if _, werr := io.WriteString(w, var0); werr != nil {
    return werr
  }
  return nil
}`,
			expectImports: []string{
				"io",
				"github.com/mh-cbon/template-compiler/std/text/template/parse",
				"github.com/mh-cbon/template-compiler/compiler",
			},
			expectHTMLCompiledFn: `func fn0(t parse.Templater, w io.Writer, indata interface{}) error {
  var bw bytes.Buffer
  var data compiler.TemplateData
  if d, ok := indata.(compiler.TemplateData); ok {
    data = d
  }
  var tplX compiler.TemplateData = data
  var0, err := tplX.MethodArgHelloMultipleReturn("me", "you")
  if err != nil {
    return err
  }
  bw.WriteString(var0)
  template.HTMLEscape(w, bw.Bytes())
  bw.Reset()
  return nil
}`,
			expectHTMLImports: []string{
				"io",
				"bytes",
				"github.com/mh-cbon/template-compiler/std/text/template/parse",
				"github.com/mh-cbon/template-compiler/compiler",
				"text/template",
			},
		},
		TestData{
			tplstr: `{{up "rr"}}`,
			funcs: map[string]interface{}{
				"up": func(s string) string {
					return s
				},
			},
			expectCompiledFn: `func fn0(t parse.Templater, w io.Writer, indata interface {}) error {
  var var0 string = t.GetFuncs()["up"].(func(string) string)("rr")
  if _, werr := io.WriteString(w, var0); werr != nil {
    return werr
  }
  return nil
}`,
			expectImports: []string{
				"io",
				"github.com/mh-cbon/template-compiler/std/text/template/parse",
			},
			expectHTMLCompiledFn: `func fn0(t parse.Templater, w io.Writer, indata interface{}) error {
  var bw bytes.Buffer
  var var1 string = t.GetFuncs()["up"].(func(string) string)("rr")
  bw.WriteString(var1)
  template.HTMLEscape(w, bw.Bytes())
  bw.Reset()
  return nil
}`,
			expectHTMLImports: []string{
				"io",
				"bytes",
				"github.com/mh-cbon/template-compiler/std/text/template/parse",
				"text/template",
			},
		},
		TestData{
			tplstr: `{{split "rr" "r"}}`,
			funcs: map[string]interface{}{
				"split": func(s string, v string) string {
					return s
				},
			},
			expectCompiledFn: `func fn0(t parse.Templater, w io.Writer, indata interface {}) error {
  var var0 string = t.GetFuncs()["split"].(func(string, string) string)("rr", "r")
  if _, werr := io.WriteString(w, var0); werr != nil {
    return werr
  }
  return nil
}`,
			expectImports: []string{
				"io",
				"github.com/mh-cbon/template-compiler/std/text/template/parse",
			},
			expectHTMLCompiledFn: `func fn0(t parse.Templater, w io.Writer, indata interface{}) error {
  var bw bytes.Buffer
  var var1 string = t.GetFuncs()["split"].(func(string, string) string)("rr", "r")
  bw.WriteString(var1)
  template.HTMLEscape(w, bw.Bytes())
  bw.Reset()
  return nil
}`,
			expectHTMLImports: []string{
				"io",
				"bytes",
				"github.com/mh-cbon/template-compiler/std/text/template/parse",
				"text/template",
			},
		},
		TestData{
			tplstr: `{{fnerr "r"}}`,
			funcs: map[string]interface{}{
				"fnerr": func(s string) (string, error) {
					return s, nil
				},
			},
			expectCompiledFn: `func fn0(t parse.Templater, w io.Writer, indata interface {}) error {
  var0, err := t.GetFuncs()["fnerr"].(func(string) (string, error))("r")
  if err != nil {
    return err
  }
  if _, werr := io.WriteString(w, var0); werr != nil {
    return werr
  }
  return nil
}`,
			expectImports: []string{
				"io",
				"github.com/mh-cbon/template-compiler/std/text/template/parse",
			},
			expectHTMLCompiledFn: `func fn0(t parse.Templater, w io.Writer, indata interface{}) error {
  var bw bytes.Buffer
  var1, err := t.GetFuncs()["fnerr"].(func(string) (string, error))("r")
  if err != nil {
    return err
  }
  bw.WriteString(var1)
  template.HTMLEscape(w, bw.Bytes())
  bw.Reset()
  return nil
}`,
			expectHTMLImports: []string{
				"io",
				"bytes",
				"github.com/mh-cbon/template-compiler/std/text/template/parse",
				"text/template",
			},
		},
		TestData{
			tplstr:    `{{.SomeInterface}}`,
			dataValue: TemplateData{SomeInterface: TemplateData{}},
			expectCompiledFn: `func fn0(t parse.Templater, w io.Writer, indata interface {}) error {
  var data compiler.TemplateData
  if d, ok := indata.(compiler.TemplateData); ok {
    data = d
  }
  var var0 interface {} = data.SomeInterface
  if _, werr := fmt.Fprintf(w, "%v", var0); werr != nil {
    return werr
  }
  return nil
}`,
			expectImports: []string{
				"io",
				"fmt",
				"github.com/mh-cbon/template-compiler/std/text/template/parse",
				"github.com/mh-cbon/template-compiler/compiler",
			},
			expectHTMLCompiledFn: `func fn0(t parse.Templater, w io.Writer, indata interface{}) error {
  var data compiler.TemplateData
  if d, ok := indata.(compiler.TemplateData); ok {
    data = d
  }
  var var1 interface{} = data.SomeInterface
  var var0 string = template.HTMLEscaper(var1)
  if _, werr := io.WriteString(w, var0); werr != nil {
    return werr
  }
  return nil
}`,
			expectHTMLImports: []string{
				"io",
				"github.com/mh-cbon/template-compiler/std/text/template/parse",
				"github.com/mh-cbon/template-compiler/compiler",
				"github.com/mh-cbon/template-compiler/std/html/template",
			},
		},
		TestData{
			tplstr:    `{{.SomeInterface.SomeInterface}}`,
			dataValue: TemplateData{SomeInterface: TemplateData{}},
			funcs: map[string]interface{}{
				"browsePropertyPath": func(x interface{}, p string, args ...interface{}) interface{} { return nil },
			},
			funcsMapPublic: []map[string]string{
				map[string]string{
					"FuncName": "browsePropertyPath",
					"Sel":      "funcmap.BrowsePropertyPath",
					"Pkg":      "github.com/mh-cbon/template-tree-simplifier/funcmap",
				},
			},
			expectCompiledFn: `func fn0(t parse.Templater, w io.Writer, indata interface {}) error {
  var data compiler.TemplateData
  if d, ok := indata.(compiler.TemplateData); ok {
    data = d
  }
  var var0 interface{} = funcmap.BrowsePropertyPath(data, "SomeInterface.SomeInterface")
  if _, werr := fmt.Fprintf(w, "%v", var0); werr != nil {
    return werr
  }
  return nil
}`,
			expectImports: []string{
				"io",
				"fmt",
				"github.com/mh-cbon/template-tree-simplifier/funcmap",
				"github.com/mh-cbon/template-compiler/std/text/template/parse",
				"github.com/mh-cbon/template-compiler/compiler",
			},
			expectHTMLCompiledFn: `func fn0(t parse.Templater, w io.Writer, indata interface{}) error {
  var data compiler.TemplateData
  if d, ok := indata.(compiler.TemplateData); ok {
    data = d
  }
  var var1 interface{} = funcmap.BrowsePropertyPath(data, "SomeInterface.SomeInterface")
  var var0 string = template.HTMLEscaper(var1)
  if _, werr := io.WriteString(w, var0); werr != nil {
    return werr
  }
  return nil
}`,
			expectHTMLImports: []string{
				"io",
				"github.com/mh-cbon/template-compiler/std/text/template/parse",
				"github.com/mh-cbon/template-compiler/compiler",
				"github.com/mh-cbon/template-compiler/std/html/template",
				"github.com/mh-cbon/template-tree-simplifier/funcmap",
			},
		},
		TestData{
			tplstr:    `{{$x := .SomeInterface}}{{$x.SomeInterface}}`,
			dataValue: TemplateData{SomeInterface: TemplateData{}},
			funcs: map[string]interface{}{
				"browsePropertyPath": func(x interface{}, p string, args ...interface{}) interface{} { return nil },
			},
			funcsMapPublic: []map[string]string{
				map[string]string{
					"FuncName": "browsePropertyPath",
					"Sel":      "funcmap.BrowsePropertyPath",
					"Pkg":      "github.com/mh-cbon/template-tree-simplifier/funcmap",
				},
			},
			expectCompiledFn: `func fn0(t parse.Templater, w io.Writer, indata interface {}) error {
  var data compiler.TemplateData
  if d, ok := indata.(compiler.TemplateData); ok {
    data = d
  }
  var tplX interface{} = data.SomeInterface
  var var0 interface{} = funcmap.BrowsePropertyPath(tplX, "SomeInterface")
  if _, werr := fmt.Fprintf(w, "%v", var0); werr != nil {
    return werr
  }
  return nil
}`,
			expectImports: []string{
				"io",
				"fmt",
				"github.com/mh-cbon/template-tree-simplifier/funcmap",
				"github.com/mh-cbon/template-compiler/std/text/template/parse",
				"github.com/mh-cbon/template-compiler/compiler",
			},
			expectHTMLCompiledFn: `func fn0(t parse.Templater, w io.Writer, indata interface{}) error {
  var data compiler.TemplateData
  if d, ok := indata.(compiler.TemplateData); ok {
    data = d
  }
  var tplX interface{} = data.SomeInterface
  var var1 interface{} = funcmap.BrowsePropertyPath(tplX, "SomeInterface")
  var var0 string = template.HTMLEscaper(var1)
  if _, werr := io.WriteString(w, var0); werr != nil {
    return werr
  }
  return nil
}`,
			expectHTMLImports: []string{
				"io",
				"github.com/mh-cbon/template-compiler/std/text/template/parse",
				"github.com/mh-cbon/template-compiler/compiler",
				"github.com/mh-cbon/template-compiler/std/html/template",
				"github.com/mh-cbon/template-tree-simplifier/funcmap",
			},
		},
		TestData{
			tplstr:    `{{$x := .SomeInterface}}{{$x.MethodHello}}`,
			dataValue: TemplateData{SomeInterface: TemplateData{}},
			funcs: map[string]interface{}{
				"browsePropertyPath": func(x interface{}, p string, args ...interface{}) interface{} { return nil },
			},
			funcsMapPublic: []map[string]string{
				map[string]string{
					"FuncName": "browsePropertyPath",
					"Sel":      "funcmap.BrowsePropertyPath",
					"Pkg":      "github.com/mh-cbon/template-tree-simplifier/funcmap",
				},
			},
			expectCompiledFn: `func fn0(t parse.Templater, w io.Writer, indata interface {}) error {
  var data compiler.TemplateData
  if d, ok := indata.(compiler.TemplateData); ok {
    data = d
  }
  var tplX interface{} = data.SomeInterface
  var var0 interface{} = funcmap.BrowsePropertyPath(tplX, "MethodHello")
  if _, werr := fmt.Fprintf(w, "%v", var0); werr != nil {
    return werr
  }
  return nil
}`,
			expectImports: []string{
				"io",
				"fmt",
				"github.com/mh-cbon/template-tree-simplifier/funcmap",
				"github.com/mh-cbon/template-compiler/std/text/template/parse",
				"github.com/mh-cbon/template-compiler/compiler",
			},
			expectHTMLCompiledFn: `func fn0(t parse.Templater, w io.Writer, indata interface{}) error {
  var data compiler.TemplateData
  if d, ok := indata.(compiler.TemplateData); ok {
    data = d
  }
  var tplX interface{} = data.SomeInterface
  var var1 interface{} = funcmap.BrowsePropertyPath(tplX, "MethodHello")
  var var0 string = template.HTMLEscaper(var1)
  if _, werr := io.WriteString(w, var0); werr != nil {
    return werr
  }
  return nil
}`,
			expectHTMLImports: []string{
				"io",
				"github.com/mh-cbon/template-compiler/std/text/template/parse",
				"github.com/mh-cbon/template-compiler/compiler",
				"github.com/mh-cbon/template-tree-simplifier/funcmap",
				"github.com/mh-cbon/template-compiler/std/html/template",
			},
		},
		TestData{
			tplstr:         `{{$x := .SomeInterface}}{{$x.MethodArgHello2 "me" "you"}}`,
			dataValue:      TemplateData{SomeInterface: TemplateData{}},
			expectBuiltins: map[string]string{},
			funcs: map[string]interface{}{
				"browsePropertyPath": func(x interface{}, p string, args ...interface{}) interface{} { return nil },
			},
			funcsMapPublic: []map[string]string{
				map[string]string{
					"FuncName": "browsePropertyPath",
					"Sel":      "funcmap.BrowsePropertyPath",
					"Pkg":      "github.com/mh-cbon/template-tree-simplifier/funcmap",
				},
			},
			expectCompiledFn: `func fn0(t parse.Templater, w io.Writer, indata interface {}) error {
  var data compiler.TemplateData
  if d, ok := indata.(compiler.TemplateData); ok {
    data = d
  }
  var tplX interface{} = data.SomeInterface
  var var0 interface{} = funcmap.BrowsePropertyPath(tplX, "MethodArgHello2", "me", "you")
  if _, werr := fmt.Fprintf(w, "%v", var0); werr != nil {
    return werr
  }
  return nil
}`,
			expectImports: []string{
				"io",
				"fmt",
				"github.com/mh-cbon/template-tree-simplifier/funcmap",
				"github.com/mh-cbon/template-compiler/std/text/template/parse",
				"github.com/mh-cbon/template-compiler/compiler",
			},
			expectHTMLCompiledFn: `func fn0(t parse.Templater, w io.Writer, indata interface{}) error {
  var data compiler.TemplateData
  if d, ok := indata.(compiler.TemplateData); ok {
    data = d
  }
  var tplX interface{} = data.SomeInterface
  var var0 interface{} = funcmap.BrowsePropertyPath(tplX, "MethodArgHello2", "me", "you")
  var var1 string = template.HTMLEscaper(var0)
  if _, werr := io.WriteString(w, var1); werr != nil {
    return werr
  }
  return nil
}`,
			expectHTMLImports: []string{
				"io",
				"github.com/mh-cbon/template-compiler/std/text/template/parse",
				"github.com/mh-cbon/template-compiler/compiler",
				"github.com/mh-cbon/template-tree-simplifier/funcmap",
				"github.com/mh-cbon/template-compiler/std/html/template",
			},
		},
		TestData{
			tplstr:         `{{define "rr"}}what{{end}}ww{{template "rr" (up "rr")}}`,
			dataValue:      TemplateData{SomeInterface: TemplateData{}},
			expectBuiltins: map[string]string{"ww": "builtin0"},
			funcs: map[string]interface{}{
				"browsePropertyPath": func(x interface{}, p string, args ...interface{}) interface{} { return nil },
				"up":                 func(s string) string { return s },
			},
			expectCompiledFn: `func fn0(t parse.Templater, w io.Writer, indata interface {}) error {
  if _, werr := w.Write(builtin0); werr != nil {
    return werr
  }
  var var0 string = t.GetFuncs()["up"].(func(string) string)("rr")
  if werr := t.ExecuteTemplate(w, "rr", var0); werr != nil {
    return werr
  }
  return nil
}`,
			expectImports: []string{
				"io",
				"github.com/mh-cbon/template-compiler/std/text/template/parse",
			},
			expectHTMLCompiledFn: `func fn0(t parse.Templater, w io.Writer, indata interface{}) error {
  if _, werr := w.Write(builtin0); werr != nil {
    return werr
  }
  var var0 string = t.GetFuncs()["up"].(func(string) string)("rr")
  if werr := t.ExecuteTemplate(w, "rr", var0); werr != nil {
    return werr
  }
  return nil
}`,
			expectHTMLImports: []string{
				"io",
				"github.com/mh-cbon/template-compiler/std/text/template/parse",
			},
		},
		TestData{
			tplstr:    `{{html "rr"}}`,
			dataValue: TemplateData{SomeInterface: TemplateData{}},
			funcs: map[string]interface{}{
				"browsePropertyPath": func(x interface{}, p string, args ...interface{}) interface{} { return nil },
			},
			expectCompiledFn: `func fn0(t parse.Templater, w io.Writer, indata interface {}) error {
  var bw bytes.Buffer
  bw.WriteString("rr")
  template.HTMLEscape(w, bw.Bytes())
  bw.Reset()
  return nil
}`,
			expectImports: []string{
				"io",
				"bytes",
				"github.com/mh-cbon/template-compiler/std/text/template/parse",
				"text/template",
			},
			expectHTMLCompiledFn: `func fn0(t parse.Templater, w io.Writer, indata interface{}) error {
  var bw bytes.Buffer
  bw.WriteString("rr")
  template.HTMLEscape(w, bw.Bytes())
  bw.Reset()
  return nil
}`,
			expectHTMLImports: []string{
				"io",
				"bytes",
				"github.com/mh-cbon/template-compiler/std/text/template/parse",
				"text/template",
			},
		},
		TestData{
			tplstr:    `{{len .SomeTemplateDataSlice}}`,
			dataValue: TemplateData{},
			funcs: map[string]interface{}{
				"browsePropertyPath": func(x interface{}, p string, args ...interface{}) interface{} { return nil },
			},
			expectCompiledFn: `func fn0(t parse.Templater, w io.Writer, indata interface{}) error {
  var data compiler.TemplateData
  if d, ok := indata.(compiler.TemplateData); ok {
    data = d
  }
  var var0 []*compiler.TemplateData = data.SomeTemplateDataSlice
  var var1 int = len(var0)
  if _, werr := io.WriteString(w, strconv.Itoa(var1)); werr != nil {
    return werr
  }
  return nil
}`,
			expectImports: []string{
				"io",
				"github.com/mh-cbon/template-compiler/std/text/template/parse",
				"github.com/mh-cbon/template-compiler/compiler",
				"strconv",
			},
			expectHTMLCompiledFn: `func fn0(t parse.Templater, w io.Writer, indata interface{}) error {
  var data compiler.TemplateData
  if d, ok := indata.(compiler.TemplateData); ok {
    data = d
  }
  var var0 []*compiler.TemplateData = data.SomeTemplateDataSlice
  var var2 int = len(var0)
  var var1 string = template.HTMLEscaper(var2)
  if _, werr := io.WriteString(w, var1); werr != nil {
    return werr
  }
  return nil
}`,
			expectHTMLImports: []string{
				"io",
				"github.com/mh-cbon/template-compiler/std/text/template/parse",
				"github.com/mh-cbon/template-compiler/compiler",
				"github.com/mh-cbon/template-compiler/std/html/template",
			},
		},
		TestData{
			tplstr:    `{{if eq true true}}{{end}}`,
			dataValue: TemplateData{},
			funcs: map[string]interface{}{
				"browsePropertyPath": func(x interface{}, p string, args ...interface{}) interface{} { return nil },
			},
			expectCompiledFn: `func fn0(t parse.Templater, w io.Writer, indata interface{}) error {
  var var0 bool = true == true
  if var0 {
  }
  return nil
}`,
			expectImports: []string{
				"io",
				"github.com/mh-cbon/template-compiler/std/text/template/parse",
			},
			expectHTMLCompiledFn: `func fn0(t parse.Templater, w io.Writer, indata interface{}) error {
  var var0 bool = true == true
  if var0 {
  }
  return nil
}`,
			expectHTMLImports: []string{
				"io",
				"github.com/mh-cbon/template-compiler/std/text/template/parse",
			},
		},
		TestData{
			tplstr: `{{range $i, $v := .SomeTemplateDataSlice}}
Hello range branch!
{{$y := false}}
{{else}}
Hello else branch!
{{$y := false}}
{{end}}
{{$y := true}}
{{if $y}} if branch {{else}} else branch {{end}}
{{with $y}}{{.}}{{else}}{{.}}{{end}}
`,
			dataValue: TemplateData{},
			expectBuiltins: map[string]string{
				"\nHello range branch!\n": "builtin0",
				"\n": "builtin1",
				"\nHello else branch!\n": "builtin2",
				" if branch ":            "builtin3",
				" else branch ":          "builtin4",
			},
			expectCompiledFn: `func fn0(t parse.Templater, w io.Writer, indata interface{}) error {
  var data compiler.TemplateData
  if d, ok := indata.(compiler.TemplateData); ok {
    data = d
  }
  var var0 []*compiler.TemplateData = data.SomeTemplateDataSlice
  for tplI, tplV := range var0 {
    if _, werr := w.Write(builtin0); werr != nil {
      return werr
    }
    var tplY bool = false
    if _, werr := w.Write(builtin1); werr != nil {
      return werr
    }
  }
  if len(var0) == 0 {
    if _, werr := w.Write(builtin2); werr != nil {
      return werr
    }
    var tplYShadow bool = false
    if _, werr := w.Write(builtin1); werr != nil {
      return werr
    }
  }
  if _, werr := w.Write(builtin1); werr != nil {
    return werr
  }
  var tplYShadow0 bool = true
  if _, werr := w.Write(builtin1); werr != nil {
    return werr
  }
  if tplYShadow0 {
    if _, werr := w.Write(builtin3); werr != nil {
      return werr
    }
  } else {
    if _, werr := w.Write(builtin4); werr != nil {
      return werr
    }
  }
  if _, werr := w.Write(builtin1); werr != nil {
    return werr
  }
  {
    if tplYShadow0 {
      if _, werr := io.WriteString(w, strconv.FormatBool(tplYShadow0)); werr != nil {
        return werr
      }
    } else {
      if _, werr := io.WriteString(w, strconv.FormatBool(data)); werr != nil {
        return werr
      }
    }
  }
  if _, werr := w.Write(builtin1); werr != nil {
    return werr
  }
  return nil
}`,
			expectImports: []string{
				"io",
				"strconv",
				"github.com/mh-cbon/template-compiler/std/text/template/parse",
				"github.com/mh-cbon/template-compiler/compiler",
			},
			expectHTMLCompiledFn: `func fn0(t parse.Templater, w io.Writer, indata interface{}) error {
  var data compiler.TemplateData
  if d, ok := indata.(compiler.TemplateData); ok {
    data = d
  }
  var var0 []*compiler.TemplateData = data.SomeTemplateDataSlice
  for tplI, tplV := range var0 {
  if _, werr := w.Write(builtin0); werr != nil {
    return werr
  }
  var tplY bool = false
  if _, werr := w.Write(builtin1); werr != nil {
    return werr
  }
  }
  if len(var0) == 0 {
    if _, werr := w.Write(builtin2); werr != nil {
      return werr
    }
    var tplYShadow bool = false
    if _, werr := w.Write(builtin1); werr != nil {
      return werr
    }
  }
  if _, werr := w.Write(builtin1); werr != nil {
    return werr
  }
  var tplYShadow0 bool = true
  if _, werr := w.Write(builtin1); werr != nil {
    return werr
  }
  if tplYShadow0 {
    if _, werr := w.Write(builtin3); werr != nil {
      return werr
    }
  } else {
    if _, werr := w.Write(builtin4); werr != nil {
      return werr
    }
  }
  if _, werr := w.Write(builtin1); werr != nil {
    return werr
  }
  {
    if tplYShadow0 {
      var var1 string = template.HTMLEscaper(tplYShadow0)
      if _, werr := io.WriteString(w, var1); werr != nil {
        return werr
      }
    } else {
      var var2 string = template.HTMLEscaper(data)
      if _, werr := io.WriteString(w, var2); werr != nil {
        return werr
      }
    }
  }
  if _, werr := w.Write(builtin1); werr != nil {
    return werr
  }
  return nil
}`,
			expectHTMLImports: []string{
				"io",
				"github.com/mh-cbon/template-compiler/std/text/template/parse",
				"github.com/mh-cbon/template-compiler/compiler",
				"github.com/mh-cbon/template-compiler/std/html/template",
			},
		},
	}

	for i, testData := range allTestData {

		failedText := checkTextTemplate(t, i, testData)
		if failedText {
			return
		}

		failedHTML := checkHTMLTemplate(t, i, testData)
		if failedHTML {
			return
		}

	}
}

func checkTextTemplate(
	t *testing.T,
	testIndex int,
	testData TestData,
) bool {

	funcs, publicIdents, tree, prepfailed := prepTextTemplate(t, testIndex, testData)
	if prepfailed {
		return true
	}

	failed := checkTree(
		t,
		testIndex,
		testData,
		tree,
		funcs,
		publicIdents,
		false,
	)
	if failed {
		return true
	}
	return false
}

func checkHTMLTemplate(
	t *testing.T,
	testIndex int,
	testData TestData,
) bool {

	funcs, publicIdents, tree, prepfailed := prepHTMLTemplate(t, testIndex, testData)
	if prepfailed {
		return true
	}

	failed := checkTree(
		t,
		testIndex,
		testData,
		tree,
		funcs,
		publicIdents,
		true,
	)
	if failed {
		return true
	}
	return false
}

func prepTextTemplate(
	t *testing.T,
	testIndex int,
	testData TestData,
) (map[string]interface{}, []map[string]string, *parse.Tree, bool) {

	// parse and compile the template to text
	funcs := map[string]interface{}{}
	for k, v := range textTemplateFuncExports {
		funcs[k] = v
	}
	for k, v := range testData.funcs {
		funcs[k] = v
	}
	publicIdents := []map[string]string{}
	publicIdents = append(publicIdents, testData.funcsMapPublic...)
	publicIdents = append(publicIdents, textTemplatePublicIdents...)

	tpl, err := template.New("").Funcs(funcs).Parse(testData.tplstr)
	if err != nil {
		t.Errorf(
			"Test(%v): Expected to compile the template, but got an error=%v",
			testIndex, err,
		)
		return nil, nil, nil, true
	}

	return funcs, publicIdents, tpl.Tree, false
}

func prepHTMLTemplate(
	t *testing.T,
	testIndex int,
	testData TestData,
) (map[string]interface{}, []map[string]string, *parse.Tree, bool) {

	// parse and compile the template to text
	funcs := map[string]interface{}{}
	for k, v := range htmlFuncsExport {
		funcs[k] = v
	}
	for k, v := range testData.funcs {
		funcs[k] = v
	}
	publicIdents := []map[string]string{}
	publicIdents = append(publicIdents, testData.funcsMapPublic...)
	publicIdents = append(publicIdents, htmlPublicIdents...)

	tpl, err := html.New("").Funcs(funcs).Parse(testData.tplstr)
	if err != nil {
		t.Errorf(
			"Test(%v): Expected to compile the template, but got an error=%v",
			testIndex, err,
		)
		return nil, nil, nil, true
	}
	// DO NOT FORGET TO EXECUTE THE TEMPLATE. otherwise it is not escaped...
	tpl.Execute(ioutil.Discard, nil) // ignore errors.
	return funcs, publicIdents, tpl.Tree, false
}

func checkTree(
	t *testing.T,
	testIndex int,
	testData TestData,
	tree *parse.Tree,
	funcs map[string]interface{},
	publicIdents []map[string]string,
	isHTML bool,
) bool {

	compiledProgram, failed := compile(
		t,
		testIndex,
		tree,
		testData.dataValue,
		funcs,
		publicIdents,
		isHTML,
	)
	if failed {
		return true
	}

	// ensure the compiled function matches
	expectedCompiledFn := testData.expectCompiledFn
	if isHTML {
		expectedCompiledFn = testData.expectHTMLCompiledFn
	}
	failed = checkCompiledFunc(
		t,
		testIndex,
		expectedCompiledFn,
		astNodeToString(compiledProgram.funcs[0]),
		testData.tplstr,
		tree,
		isHTML,
	)
	if failed {
		return true
	}

	// ensure builtins text node are transformed into builtin variables
	gotBuiltins := compiledProgram.builtinTexts
	expectedBuiltins := testData.expectBuiltins
	failed = checkBuiltins(
		t,
		testIndex,
		testData.tplstr,
		expectedBuiltins,
		gotBuiltins,
		isHTML,
	)
	if failed {
		return true
	}

	// ensure the import list matches
	expectedImports := testData.expectImports
	if isHTML {
		expectedImports = testData.expectHTMLImports
	}
	gotImports := convertImportsSpecs(compiledProgram.imports)
	failed = checkImports(
		t,
		testIndex,
		tree,
		testData.tplstr,
		expectedImports,
		gotImports,
		isHTML,
	)
	if failed {
		return true
	}
	return false
}

func compile(
	t *testing.T,
	testIndex int,
	tree *parse.Tree,
	data interface{},
	funcsMap map[string]interface{},
	publicIdents []map[string]string,
	isHTML bool,
) (*CompiledTemplatesProgram, bool) {
	compiledProgram := NewCompiledTemplatesProgram("ee")
	typeCheck := simplifier.TransformTree(tree, data, funcsMap)
	err := convertTplTree(
		"fn0",
		tree,
		funcsMap,
		publicIdents,
		makeDataConfiguration(data),
		typeCheck,
		compiledProgram,
	)
	if err != nil {
		t.Errorf(
			"Test(%v) html(%v): Expected to succeed, but got an error=%v",
			testIndex,
			isHTML,
			err,
		)
		return nil, true
	}
	return compiledProgram, false
}

func checkCompiledFunc(
	t *testing.T,
	testIndex int,
	expectedFunc string,
	gotFunc string,
	tplstr string,
	tree *parse.Tree,
	isHTML bool,
) bool {
	expectedFunc = formatGoCode(expectedFunc)
	gotFunc = formatGoCode(gotFunc)
	if expectedFunc != gotFunc {
		t.Errorf(
			"Test(%v) html(%v): Unexpected compiled function. Expected=\n%v\n-----\nGot=\n%v\nTEMPLATE:\n%v\nSIMPLIFIED TEMPLATE:\n%v\n",
			testIndex,
			isHTML,
			expectedFunc,
			gotFunc,
			tplstr,
			tree.Root.String(),
		)
		return true
	}
	return false
}

func checkBuiltins(
	t *testing.T,
	testIndex int,
	tplstr string,
	expectedBuiltins map[string]string,
	gotBuiltins map[string]string,
	isHTML bool,
) bool {
	for text, varname := range gotBuiltins {
		if expectVarName, ok := expectedBuiltins[text]; ok == false {
			t.Errorf(
				"Test(%v) html(%v): Found unexpected builtin variable %q for the text %q\nTEMPLATE:\n%v\n",
				testIndex,
				isHTML,
				varname,
				text,
				tplstr,
			)
			return true
		} else if expectVarName != varname {
			t.Errorf(
				"Test(%v) html(%v): Incorrect variable name %q for the builtin text %q\nTEMPLATE:\n%v\n",
				testIndex,
				isHTML,
				varname,
				text,
				tplstr,
			)
			return true
		}
	}
	for text, varname := range expectedBuiltins {
		if _, ok := gotBuiltins[text]; ok == false {
			t.Errorf(
				"Test(%v) html(%v): Expected builtin variable was not found %q with the text %q\nTEMPLATE:\n%v\n",
				testIndex,
				isHTML,
				varname,
				text,
				tplstr,
			)
			return true
		}
	}
	return false
}

func checkImports(
	t *testing.T,
	testIndex int,
	tree *parse.Tree,
	tplstr string,
	expectedImports []string,
	gotImports []string,
	isHTML bool,
) bool {
	for _, im := range expectedImports {
		if strExists(im, gotImports) == false {
			t.Errorf(
				"Test(%v) html(%v): Missing additionnal imports %q.\nTEMPLATE:\n%v\nSIMPLIFIED TEMPLATE:\n%v\n",
				testIndex,
				isHTML,
				im,
				tplstr,
				tree.Root.String(),
			)
			return true
		}
	}
	for _, im := range gotImports {
		if strExists(im, expectedImports) == false {
			t.Errorf(
				"Test(%v) html(%v): Unexpected additionnal imports %q. Unwanted=%v\nTEMPLATE:\n%v\nSIMPLIFIED TEMPLATE:\n%v\n",
				testIndex,
				isHTML,
				im,
				im,
				tplstr,
				tree.Root.String(),
			)
			return true
		}
	}
	if len(expectedImports) != len(gotImports) {
		t.Errorf(
			"Test(%v) html(%v): Unexpected additionnal imports. Expected=\n%v\n-----\nGot=\n%v\nTEMPLATE:\n%v\nSIMPLIFIED TEMPLATE:\n%v\n",
			testIndex,
			isHTML,
			expectedImports,
			gotImports,
			tplstr,
			tree.Root.String(),
		)
		return true
	}
	return false
}

func strExists(s string, in []string) bool {
	for _, i := range in {
		if i == s {
			return true
		}
	}
	return false
}
