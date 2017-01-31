package compiler

import (
	"testing"
	"text/template"

	"github.com/mh-cbon/template-tree-simplifier/simplifier"
)

type TestData struct {
	// the template source string
	tplstr string
	// the template data
	dataValue interface{}
	// the expected function result
	expectCompiledFn string
	// the expected list of builtins text node variables
	expectBuiltins map[string]string
	// the func to compile the template
	funcs template.FuncMap
	// the list of func map with a public identifier
	funcsMapPublic []map[string]string
	// the expected imports of the resulting go compilation
	expectImports []string
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
			tplstr:    `Hello!`,
			dataValue: TemplateData{},
			expectCompiledFn: `func fn0(t parse.Templater, w io.Writer, indata interface {}) error {
  var writeErr error
  _, writeErr = w.Write(builtin0)
  if writeErr != nil {
    return writeErr
  }
  return nil
}`,
			expectBuiltins: map[string]string{"Hello!": "builtin0"},
			funcs:          map[string]interface{}{},
			expectImports: []string{
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
			expectBuiltins: map[string]string{},
			funcs:          map[string]interface{}{},
			expectImports: []string{
				"io",
				"github.com/mh-cbon/template-compiler/std/text/template/parse",
			},
		},
		TestData{
			tplstr:    `{{$y := "Hello!"}}{{$y}}`,
			dataValue: TemplateData{},
			expectCompiledFn: `func fn0(t parse.Templater, w io.Writer, indata interface {}) error {
  var writeErr error
  var tplY string = "Hello!"
  _, writeErr = io.WriteString(w, tplY)
  if writeErr != nil {
    return writeErr
  }
  return nil
}`,
			expectBuiltins: map[string]string{},
			funcs:          map[string]interface{}{},
			expectImports: []string{
				"io",
				"github.com/mh-cbon/template-compiler/std/text/template/parse",
			},
		},
		TestData{
			tplstr:    `{{if true}}true{{else}}false{{end}}`,
			dataValue: TemplateData{},
			expectCompiledFn: `func fn0(t parse.Templater, w io.Writer, indata interface {}) error {
  var writeErr error
  if true {
    _, writeErr = w.Write(builtin0)
    if writeErr != nil {
      return writeErr
    }
  } else {
    _, writeErr = w.Write(builtin1)
    if writeErr != nil {
      return writeErr
    }
  }
  return nil
}`,
			expectBuiltins: map[string]string{"true": "builtin0", "false": "builtin1"},
			funcs:          map[string]interface{}{},
			expectImports: []string{
				"io",
				"github.com/mh-cbon/template-compiler/std/text/template/parse",
			},
		},
		TestData{
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
  var writeErr error
  var var0 string = data.SomeString
  _, writeErr = io.WriteString(w, var0)
  if writeErr != nil {
    return writeErr
  }
  _, writeErr = w.Write(builtin0)
  if writeErr != nil {
    return writeErr
  }
  var var1 int = data.SomeInt
  _, writeErr = io.WriteString(w, strconv.Itoa(var1))
  if writeErr != nil {
    return writeErr
  }
  _, writeErr = w.Write(builtin0)
  if writeErr != nil {
    return writeErr
  }
  var var2 bool = data.SomeBool
  _, writeErr = io.WriteString(w, strconv.FormatBool(var2))
  if writeErr != nil {
    return writeErr
  }
  _, writeErr = w.Write(builtin0)
  if writeErr != nil {
    return writeErr
  }
  var var3 int8 = data.SomeInt8
  _, writeErr = io.WriteString(w, strconv.FormatInt(int64(var3), 10))
  if writeErr != nil {
    return writeErr
  }
  _, writeErr = w.Write(builtin0)
  if writeErr != nil {
    return writeErr
  }
  var var4 int16 = data.SomeInt16
  _, writeErr = io.WriteString(w, strconv.FormatInt(int64(var4), 10))
  if writeErr != nil {
    return writeErr
  }
  _, writeErr = w.Write(builtin0)
  if writeErr != nil {
    return writeErr
  }
  var var5 int32 = data.SomeInt32
  _, writeErr = io.WriteString(w, strconv.FormatInt(int64(var5), 10))
  if writeErr != nil {
    return writeErr
  }
  _, writeErr = w.Write(builtin0)
  if writeErr != nil {
    return writeErr
  }
  var var6 int64 = data.SomeInt64
  _, writeErr = io.WriteString(w, strconv.FormatInt(var6, 10))
  if writeErr != nil {
    return writeErr
  }
  _, writeErr = w.Write(builtin0)
  if writeErr != nil {
    return writeErr
  }
  var var7 uint = data.SomeUint
  _, writeErr = io.WriteString(w, strconv.FormatUint(var7, 10))
  if writeErr != nil {
    return writeErr
  }
  _, writeErr = w.Write(builtin0)
  if writeErr != nil {
    return writeErr
  }
  var var8 uint8 = data.SomeUint8
  _, writeErr = io.WriteString(w, strconv.FormatUint(var8, 10))
  if writeErr != nil {
    return writeErr
  }
  _, writeErr = w.Write(builtin0)
  if writeErr != nil {
    return writeErr
  }
  var var9 uint16 = data.SomeUint16
  _, writeErr = io.WriteString(w, strconv.FormatUint(var9, 10))
  if writeErr != nil {
    return writeErr
  }
  _, writeErr = w.Write(builtin0)
  if writeErr != nil {
    return writeErr
  }
  var var10 uint32 = data.SomeUint32
  _, writeErr = io.WriteString(w, strconv.FormatUint(var10, 10))
  if writeErr != nil {
    return writeErr
  }
  _, writeErr = w.Write(builtin0)
  if writeErr != nil {
    return writeErr
  }
  var var11 uint64 = data.SomeUint64
  _, writeErr = io.WriteString(w, strconv.FormatUint(var11, 10))
  if writeErr != nil {
    return writeErr
  }
  _, writeErr = w.Write(builtin0)
  if writeErr != nil {
    return writeErr
  }
  var var12 float32 = data.SomeFloat32
  _, writeErr = io.WriteString(w, strconv.FormatFloat(float64(var12), "f", -1, 32))
  if writeErr != nil {
    return writeErr
  }
  _, writeErr = w.Write(builtin0)
  if writeErr != nil {
    return writeErr
  }
  var var13 float64 = data.SomeFloat64
  _, writeErr = io.WriteString(w, strconv.FormatFloat(var13, "f", -1, 64))
  if writeErr != nil {
    return writeErr
  }
  _, writeErr = w.Write(builtin0)
  if writeErr != nil {
    return writeErr
  }
  var var14 int32 = data.SomeRune
  _, writeErr = io.WriteString(w, strconv.FormatInt(int64(var14), 10))
  if writeErr != nil {
    return writeErr
  }
  _, writeErr = w.Write(builtin0)
  if writeErr != nil {
    return writeErr
  }
  var var15 uint8 = data.SomeByte
  _, writeErr = io.WriteString(w, strconv.FormatUint(var15, 10))
  if writeErr != nil {
    return writeErr
  }
  _, writeErr = w.Write(builtin0)
  if writeErr != nil {
    return writeErr
  }
  var var16 []uint8 = data.SomeByteSlice
  _, writeErr = w.Write(var16)
  if writeErr != nil {
    return writeErr
  }
  _, writeErr = w.Write(builtin0)
  if writeErr != nil {
    return writeErr
  }
  var var17 []int32 = data.SomeRuneSlice
  _, writeErr = fmt.Fprintf(w, "%v", var17)
  if writeErr != nil {
    return writeErr
  }
  return nil
}`,
			expectBuiltins: map[string]string{"\n": "builtin0"},
			funcs:          map[string]interface{}{},
			expectImports: []string{
				"io",
				"github.com/mh-cbon/template-compiler/std/text/template/parse",
				"github.com/mh-cbon/template-compiler/compiler",
				"strconv",
				"fmt",
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
  var writeErr error
  var var0 []uint8 = data.SomeByteSlice
  for _, iterable := range var0 {
    _, writeErr = io.WriteString(w, strconv.FormatUint(iterable, 10))
    if writeErr != nil {
      return writeErr
    }
  }
  return nil
}`,
			expectBuiltins: map[string]string{},
			funcs:          map[string]interface{}{},
			expectImports: []string{
				"io",
				"github.com/mh-cbon/template-compiler/std/text/template/parse",
				"github.com/mh-cbon/template-compiler/compiler",
				"strconv",
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
  var writeErr error
  var var0 []uint8 = data.SomeByteSlice
  for tplI, tplV := range var0 {
    _, writeErr = io.WriteString(w, strconv.FormatUint(tplV, 10))
    if writeErr != nil {
      return writeErr
    }
  }
  return nil
}`,
			expectBuiltins: map[string]string{},
			funcs:          map[string]interface{}{},
			expectImports: []string{
				"io",
				"github.com/mh-cbon/template-compiler/std/text/template/parse",
				"github.com/mh-cbon/template-compiler/compiler",
				"strconv",
			},
		},
		TestData{
			tplstr: `{{range $i, $v := .SomeTemplateDataSlice}}
{{range $i, $v := $v.SomeTemplateDataSlice}}
{{end}}
{{end}}`,
			dataValue: TemplateData{},
			expectCompiledFn: `func fn0(t parse.Templater, w io.Writer, indata interface {}) error {
  var data compiler.TemplateData
  if d, ok := indata.(compiler.TemplateData); ok {
    data = d
  }
  var writeErr error
  var var0 []*compiler.TemplateData = data.SomeTemplateDataSlice
  for tplI, tplV := range var0 {
    _, writeErr = w.Write(builtin0)
    if writeErr != nil {
      return writeErr
    }
    var var1 []*compiler.TemplateData = tplV.SomeTemplateDataSlice
    for tplIShadow, tplVShadow := range var1 {
      _, writeErr = w.Write(builtin0)
      if writeErr != nil {
        return writeErr
      }
    }
    _, writeErr = w.Write(builtin0)
    if writeErr != nil {
      return writeErr
    }
  }
  return nil
}`,
			expectBuiltins: map[string]string{"\n": "builtin0"},
			funcs:          map[string]interface{}{},
			expectImports: []string{
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
			expectCompiledFn: `func fn0(t parse.Templater, w io.Writer, indata interface {}) error {
  var data compiler.TemplateData
  if d, ok := indata.(compiler.TemplateData); ok {
    data = d
  }
  var writeErr error
  var var0 []*compiler.TemplateData = data.SomeTemplateDataSlice
  for tplI, tplV := range var0 {
    _, writeErr = w.Write(builtin0)
    if writeErr != nil {
      return writeErr
    }
  }
  if len(var0) == 0 {
    _, writeErr = w.Write(builtin1)
    if writeErr != nil {
      return writeErr
    }
  }
  return nil
}`,
			expectBuiltins: map[string]string{
				"\nHello range branch!\n": "builtin0",
				"\nHello else branch!\n":  "builtin1",
			},
			funcs: map[string]interface{}{},
			expectImports: []string{
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
			expectCompiledFn: `func fn0(t parse.Templater, w io.Writer, indata interface {}) error {
  var data compiler.TemplateData
  if d, ok := indata.(compiler.TemplateData); ok {
    data = d
  }
  var writeErr error
  var var0 compiler.TemplateData = data
  {
    if true {
      _, writeErr = w.Write(builtin0)
      if writeErr != nil {
        return writeErr
      }
    } else {
      _, writeErr = w.Write(builtin1)
      if writeErr != nil {
        return writeErr
      }
    }
  }
  return nil
}`,
			expectBuiltins: map[string]string{
				"\nHello with branch!\n":    "builtin0",
				"\nHello without branch!\n": "builtin1",
			},
			funcs: map[string]interface{}{},
			expectImports: []string{
				"io",
				"github.com/mh-cbon/template-compiler/std/text/template/parse",
				"github.com/mh-cbon/template-compiler/compiler",
			},
		},
		TestData{
			tplstr:    `{{with .}}{{.SomeString}}{{else}}{{.SomeString}}{{end}}`,
			dataValue: TemplateData{},
			expectCompiledFn: `func fn0(t parse.Templater, w io.Writer, indata interface {}) error {
  var data compiler.TemplateData
  if d, ok := indata.(compiler.TemplateData); ok {
    data = d
  }
  var writeErr error
  var var0 compiler.TemplateData = data
  {
    if true {
      var var1 string = var0.SomeString
      _, writeErr = io.WriteString(w, var1)
      if writeErr != nil {
        return writeErr
      }
    } else {
      var var2 string = data.SomeString
      _, writeErr = io.WriteString(w, var2)
      if writeErr != nil {
        return writeErr
      }
    }
  }
  return nil
}`,
			expectBuiltins: map[string]string{},
			funcs:          map[string]interface{}{},
			expectImports: []string{
				"io",
				"github.com/mh-cbon/template-compiler/std/text/template/parse",
				"github.com/mh-cbon/template-compiler/compiler",
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
			expectBuiltins: map[string]string{},
			funcs:          map[string]interface{}{},
			expectImports: []string{
				"io",
				"github.com/mh-cbon/template-compiler/std/text/template/parse",
				"github.com/mh-cbon/template-compiler/compiler",
			},
		},
		TestData{
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
  var writeErr error
  var var0 string = data.SomeString
  if var0 != "" {
  }
  _, writeErr = w.Write(builtin0)
  if writeErr != nil {
    return writeErr
  }
  var var1 int = data.SomeInt
  if var1 != 0 {
  }
  _, writeErr = w.Write(builtin0)
  if writeErr != nil {
    return writeErr
  }
  var var2 bool = data.SomeBool
  if var2 {
  }
  _, writeErr = w.Write(builtin0)
  if writeErr != nil {
    return writeErr
  }
  var var3 int8 = data.SomeInt8
  if var3 != 0 {
  }
  _, writeErr = w.Write(builtin0)
  if writeErr != nil {
    return writeErr
  }
  var var4 int16 = data.SomeInt16
  if var4 != 0 {
  }
  _, writeErr = w.Write(builtin0)
  if writeErr != nil {
    return writeErr
  }
  var var5 int32 = data.SomeInt32
  if var5 != 0 {
  }
  _, writeErr = w.Write(builtin0)
  if writeErr != nil {
    return writeErr
  }
  var var6 int64 = data.SomeInt64
  if var6 != 0 {
  }
  _, writeErr = w.Write(builtin0)
  if writeErr != nil {
    return writeErr
  }
  var var7 uint = data.SomeUint
  if var7 != 0 {
  }
  _, writeErr = w.Write(builtin0)
  if writeErr != nil {
    return writeErr
  }
  var var8 uint8 = data.SomeUint8
  if var8 != 0 {
  }
  _, writeErr = w.Write(builtin0)
  if writeErr != nil {
    return writeErr
  }
  var var9 uint16 = data.SomeUint16
  if var9 != 0 {
  }
  _, writeErr = w.Write(builtin0)
  if writeErr != nil {
    return writeErr
  }
  var var10 uint32 = data.SomeUint32
  if var10 != 0 {
  }
  _, writeErr = w.Write(builtin0)
  if writeErr != nil {
    return writeErr
  }
  var var11 uint64 = data.SomeUint64
  if var11 != 0 {
  }
  _, writeErr = w.Write(builtin0)
  if writeErr != nil {
    return writeErr
  }
  var var12 float32 = data.SomeFloat32
  if var12 != 0 {
  }
  _, writeErr = w.Write(builtin0)
  if writeErr != nil {
    return writeErr
  }
  var var13 float64 = data.SomeFloat64
  if var13 != 0 {
  }
  _, writeErr = w.Write(builtin0)
  if writeErr != nil {
    return writeErr
  }
  var var14 int32 = data.SomeRune
  if var14 != 0 {
  }
  _, writeErr = w.Write(builtin0)
  if writeErr != nil {
    return writeErr
  }
  var var15 uint8 = data.SomeByte
  if var15 != 0 {
  }
  _, writeErr = w.Write(builtin0)
  if writeErr != nil {
    return writeErr
  }
  var var16 []uint8 = data.SomeByteSlice
  if len(var16) > 0 {
  }
  _, writeErr = w.Write(builtin0)
  if writeErr != nil {
    return writeErr
  }
  var var17 []int32 = data.SomeRuneSlice
  if len(var17) > 0 {
  }
  return nil
}`,
			expectBuiltins: map[string]string{"\n": "builtin0"},
			funcs:          map[string]interface{}{},
			expectImports: []string{
				"io",
				"github.com/mh-cbon/template-compiler/std/text/template/parse",
				"github.com/mh-cbon/template-compiler/compiler",
			},
		},
		TestData{
			tplstr: `{{$int := 4}}
{{$float := 4.0}}
{{$complex := 1i}}`,
			dataValue: TemplateData{SomeString: "Hello!"},
			expectCompiledFn: `func fn0(t parse.Templater, w io.Writer, indata interface {}) error {
  var writeErr error
  var tplInt int = 4
  _, writeErr = w.Write(builtin0)
  if writeErr != nil {
    return writeErr
  }
  var tplFloat float64 = 4.0
  _, writeErr = w.Write(builtin0)
  if writeErr != nil {
    return writeErr
  }
  var tplComplex complex128 = 1i
  return nil
}`,
			expectBuiltins: map[string]string{"\n": "builtin0"},
			funcs:          map[string]interface{}{},
			expectImports: []string{
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
  var writeErr error
  var var0 string = data.MethodHello()
  _, writeErr = io.WriteString(w, var0)
  if writeErr != nil {
    return writeErr
  }
  return nil
}`,
			expectBuiltins: map[string]string{},
			funcs:          map[string]interface{}{},
			expectImports: []string{
				"io",
				"github.com/mh-cbon/template-compiler/std/text/template/parse",
				"github.com/mh-cbon/template-compiler/compiler",
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
  var writeErr error
  var var0 string = data.MethodArgHello("me")
  _, writeErr = io.WriteString(w, var0)
  if writeErr != nil {
    return writeErr
  }
  return nil
}`,
			expectBuiltins: map[string]string{},
			funcs:          map[string]interface{}{},
			expectImports: []string{
				"io",
				"github.com/mh-cbon/template-compiler/std/text/template/parse",
				"github.com/mh-cbon/template-compiler/compiler",
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
  var writeErr error
  var var0 string = data.MethodArgHello2("me", "you")
  _, writeErr = io.WriteString(w, var0)
  if writeErr != nil {
    return writeErr
  }
  return nil
}`,
			expectBuiltins: map[string]string{},
			funcs:          map[string]interface{}{},
			expectImports: []string{
				"io",
				"github.com/mh-cbon/template-compiler/std/text/template/parse",
				"github.com/mh-cbon/template-compiler/compiler",
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
  var writeErr error
  var0, err := data.MethodArgHelloMultipleReturn("me", "you")
  if err != nil {
    return err
  }
  _, writeErr = io.WriteString(w, var0)
  if writeErr != nil {
    return writeErr
  }
  return nil
}`,
			expectBuiltins: map[string]string{},
			funcs:          map[string]interface{}{},
			expectImports: []string{
				"io",
				"github.com/mh-cbon/template-compiler/std/text/template/parse",
				"github.com/mh-cbon/template-compiler/compiler",
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
  var writeErr error
  var tplX compiler.TemplateData = data
  var var0 string = tplX.MethodHello()
  _, writeErr = io.WriteString(w, var0)
  if writeErr != nil {
    return writeErr
  }
  return nil
}`,
			expectBuiltins: map[string]string{},
			funcs:          map[string]interface{}{},
			expectImports: []string{
				"io",
				"github.com/mh-cbon/template-compiler/std/text/template/parse",
				"github.com/mh-cbon/template-compiler/compiler",
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
  var writeErr error
  var tplX compiler.TemplateData = data
  var var0 string = tplX.MethodArgHello("me")
  _, writeErr = io.WriteString(w, var0)
  if writeErr != nil {
    return writeErr
  }
  return nil
}`,
			expectBuiltins: map[string]string{},
			funcs:          map[string]interface{}{},
			expectImports: []string{
				"io",
				"github.com/mh-cbon/template-compiler/std/text/template/parse",
				"github.com/mh-cbon/template-compiler/compiler",
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
  var writeErr error
  var tplX compiler.TemplateData = data
  var var0 string = tplX.MethodArgHello2("me", "you")
  _, writeErr = io.WriteString(w, var0)
  if writeErr != nil {
    return writeErr
  }
  return nil
}`,
			expectBuiltins: map[string]string{},
			funcs:          map[string]interface{}{},
			expectImports: []string{
				"io",
				"github.com/mh-cbon/template-compiler/std/text/template/parse",
				"github.com/mh-cbon/template-compiler/compiler",
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
  var writeErr error
  var tplX compiler.TemplateData = data
  var0, err := tplX.MethodArgHelloMultipleReturn("me", "you")
  if err != nil {
    return err
  }
  _, writeErr = io.WriteString(w, var0)
  if writeErr != nil {
    return writeErr
  }
  return nil
}`,
			expectBuiltins: map[string]string{},
			funcs:          map[string]interface{}{},
			expectImports: []string{
				"io",
				"github.com/mh-cbon/template-compiler/std/text/template/parse",
				"github.com/mh-cbon/template-compiler/compiler",
			},
		},
		TestData{
			tplstr: `{{up "rr"}}`,
			expectCompiledFn: `func fn0(t parse.Templater, w io.Writer, indata interface {}) error {
  var writeErr error
  var var0 string = t.GetFuncs()["up"].(func(string) string)("rr")
  _, writeErr = io.WriteString(w, var0)
  if writeErr != nil {
    return writeErr
  }
  return nil
}`,
			expectBuiltins: map[string]string{},
			funcs: map[string]interface{}{
				"up": func(s string) string {
					return s
				},
			},
			expectImports: []string{
				"io",
				"github.com/mh-cbon/template-compiler/std/text/template/parse",
			},
		},
		TestData{
			tplstr: `{{split "rr" "r"}}`,
			expectCompiledFn: `func fn0(t parse.Templater, w io.Writer, indata interface {}) error {
  var writeErr error
  var var0 string = t.GetFuncs()["split"].(func(string, string) string)("rr", "r")
  _, writeErr = io.WriteString(w, var0)
  if writeErr != nil {
    return writeErr
  }
  return nil
}`,
			expectBuiltins: map[string]string{},
			funcs: map[string]interface{}{
				"split": func(s string, v string) string {
					return s
				},
			},
			expectImports: []string{
				"io",
				"github.com/mh-cbon/template-compiler/std/text/template/parse",
			},
		},
		TestData{
			tplstr: `{{fnerr "r"}}`,
			expectCompiledFn: `func fn0(t parse.Templater, w io.Writer, indata interface {}) error {
  var writeErr error
  var0, err := t.GetFuncs()["fnerr"].(func(string) (string, error))("r")
  if err != nil {
    return err
  }
  _, writeErr = io.WriteString(w, var0)
  if writeErr != nil {
    return writeErr
  }
  return nil
}`,
			expectBuiltins: map[string]string{},
			funcs: map[string]interface{}{
				"fnerr": func(s string) (string, error) {
					return s, nil
				},
			},
			expectImports: []string{
				"io",
				"github.com/mh-cbon/template-compiler/std/text/template/parse",
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
  var writeErr error
  var var0 interface {} = data.SomeInterface
  _, writeErr = fmt.Fprintf(w, "%v", var0)
  if writeErr != nil {
    return writeErr
  }
  return nil
}`,
			expectBuiltins: map[string]string{},
			funcs:          map[string]interface{}{},
			expectImports: []string{
				"io",
				"fmt",
				"github.com/mh-cbon/template-compiler/std/text/template/parse",
				"github.com/mh-cbon/template-compiler/compiler",
			},
		},
		TestData{
			tplstr:    `{{.SomeInterface.SomeInterface}}`,
			dataValue: TemplateData{SomeInterface: TemplateData{}},
			expectCompiledFn: `func fn0(t parse.Templater, w io.Writer, indata interface {}) error {
  var data compiler.TemplateData
  if d, ok := indata.(compiler.TemplateData); ok {
    data = d
  }
  var writeErr error
  var var0 interface{} = funcmap.BrowsePropertyPath(data, "SomeInterface.SomeInterface")
  _, writeErr = fmt.Fprintf(w, "%v", var0)
  if writeErr != nil {
    return writeErr
  }
  return nil
}`,
			expectBuiltins: map[string]string{},
			funcs: map[string]interface{}{
				"browsePropertyPath": func(x interface{}, p string, args ...interface{}) interface{} { return nil },
			},
			expectImports: []string{
				"io",
				"fmt",
				"github.com/mh-cbon/template-tree-simplifier/funcmap",
				"github.com/mh-cbon/template-compiler/std/text/template/parse",
				"github.com/mh-cbon/template-compiler/compiler",
			},
			funcsMapPublic: []map[string]string{
				map[string]string{
					"FuncName": "browsePropertyPath",
					"Sel":      "funcmap.BrowsePropertyPath",
					"Pkg":      "github.com/mh-cbon/template-tree-simplifier/funcmap",
				},
			},
		},
		TestData{
			tplstr:    `{{$x := .SomeInterface}}{{$x.SomeInterface}}`,
			dataValue: TemplateData{SomeInterface: TemplateData{}},
			expectCompiledFn: `func fn0(t parse.Templater, w io.Writer, indata interface {}) error {
  var data compiler.TemplateData
  if d, ok := indata.(compiler.TemplateData); ok {
    data = d
  }
  var writeErr error
  var tplX interface{} = data.SomeInterface
  var var0 interface{} = funcmap.BrowsePropertyPath(tplX, "SomeInterface")
  _, writeErr = fmt.Fprintf(w, "%v", var0)
  if writeErr != nil {
    return writeErr
  }
  return nil
}`,
			expectBuiltins: map[string]string{},
			funcs: map[string]interface{}{
				"browsePropertyPath": func(x interface{}, p string, args ...interface{}) interface{} { return nil },
			},
			expectImports: []string{
				"io",
				"fmt",
				"github.com/mh-cbon/template-tree-simplifier/funcmap",
				"github.com/mh-cbon/template-compiler/std/text/template/parse",
				"github.com/mh-cbon/template-compiler/compiler",
			},
			funcsMapPublic: []map[string]string{
				map[string]string{
					"FuncName": "browsePropertyPath",
					"Sel":      "funcmap.BrowsePropertyPath",
					"Pkg":      "github.com/mh-cbon/template-tree-simplifier/funcmap",
				},
			},
		},
		TestData{
			tplstr:    `{{$x := .SomeInterface}}{{$x.MethodHello}}`,
			dataValue: TemplateData{SomeInterface: TemplateData{}},
			expectCompiledFn: `func fn0(t parse.Templater, w io.Writer, indata interface {}) error {
  var data compiler.TemplateData
  if d, ok := indata.(compiler.TemplateData); ok {
    data = d
  }
  var writeErr error
  var tplX interface{} = data.SomeInterface
  var var0 interface{} = funcmap.BrowsePropertyPath(tplX, "MethodHello")
  _, writeErr = fmt.Fprintf(w, "%v", var0)
  if writeErr != nil {
    return writeErr
  }
  return nil
}`,
			expectBuiltins: map[string]string{},
			funcs: map[string]interface{}{
				"browsePropertyPath": func(x interface{}, p string, args ...interface{}) interface{} { return nil },
			},
			expectImports: []string{
				"io",
				"fmt",
				"github.com/mh-cbon/template-tree-simplifier/funcmap",
				"github.com/mh-cbon/template-compiler/std/text/template/parse",
				"github.com/mh-cbon/template-compiler/compiler",
			},
			funcsMapPublic: []map[string]string{
				map[string]string{
					"FuncName": "browsePropertyPath",
					"Sel":      "funcmap.BrowsePropertyPath",
					"Pkg":      "github.com/mh-cbon/template-tree-simplifier/funcmap",
				},
			},
		},
		TestData{
			tplstr:    `{{$x := .SomeInterface}}{{$x.MethodArgHello2 "me" "you"}}`,
			dataValue: TemplateData{SomeInterface: TemplateData{}},
			expectCompiledFn: `func fn0(t parse.Templater, w io.Writer, indata interface {}) error {
  var data compiler.TemplateData
  if d, ok := indata.(compiler.TemplateData); ok {
    data = d
  }
  var writeErr error
  var tplX interface{} = data.SomeInterface
  var var0 interface{} = funcmap.BrowsePropertyPath(tplX, "MethodArgHello2", "me", "you")
  _, writeErr = fmt.Fprintf(w, "%v", var0)
  if writeErr != nil {
    return writeErr
  }
  return nil
}`,
			expectBuiltins: map[string]string{},
			funcs: map[string]interface{}{
				"browsePropertyPath": func(x interface{}, p string, args ...interface{}) interface{} { return nil },
			},
			expectImports: []string{
				"io",
				"fmt",
				"github.com/mh-cbon/template-tree-simplifier/funcmap",
				"github.com/mh-cbon/template-compiler/std/text/template/parse",
				"github.com/mh-cbon/template-compiler/compiler",
			},
			funcsMapPublic: []map[string]string{
				map[string]string{
					"FuncName": "browsePropertyPath",
					"Sel":      "funcmap.BrowsePropertyPath",
					"Pkg":      "github.com/mh-cbon/template-tree-simplifier/funcmap",
				},
			},
		},
		TestData{
			tplstr:    `{{define "rr"}}what{{end}}ww{{template "rr" (up "rr")}}`,
			dataValue: TemplateData{SomeInterface: TemplateData{}},
			expectCompiledFn: `func fn0(t parse.Templater, w io.Writer, indata interface {}) error {
  var writeErr error
  _, writeErr = w.Write(builtin0)
  if writeErr != nil {
    return writeErr
  }
  var var0 string = t.GetFuncs()["up"].(func(string) string)("rr")
  writeErr = t.ExecuteTemplate(w, "rr", var0)
  if writeErr != nil {
    return writeErr
  }
  return nil
}`,
			expectBuiltins: map[string]string{"ww": "builtin0"},
			funcs: map[string]interface{}{
				"browsePropertyPath": func(x interface{}, p string, args ...interface{}) interface{} { return nil },
				"up":                 func(s string) string { return s },
			},
			expectImports: []string{
				"io",
				"github.com/mh-cbon/template-compiler/std/text/template/parse",
			},
		},
		TestData{
			tplstr:    `{{html "rr"}}`,
			dataValue: TemplateData{SomeInterface: TemplateData{}},
			expectCompiledFn: `func fn0(t parse.Templater, w io.Writer, indata interface {}) error {
  var writeErr error
  var var0 string = template.HTMLEscapeString("rr")
  _, writeErr = io.WriteString(w, var0)
  if writeErr != nil {
    return writeErr
  }
  return nil
}`,
			expectBuiltins: map[string]string{},
			funcs: map[string]interface{}{
				"browsePropertyPath": func(x interface{}, p string, args ...interface{}) interface{} { return nil },
				"html":               func(s string) string { return s },
			},
			expectImports: []string{
				"io",
				"github.com/mh-cbon/template-compiler/std/text/template/parse",
				"html/template",
			},
			funcsMapPublic: []map[string]string{
				map[string]string{
					"FuncName": "html",
					"Sel":      "template.HTMLEscaper",
					"Pkg":      "text/template",
				},
			},
		},
		TestData{
			tplstr:    `{{len .SomeTemplateDataSlice}}`,
			dataValue: TemplateData{},
			expectCompiledFn: `func fn0(t parse.Templater, w io.Writer, indata interface{}) error {
  var data compiler.TemplateData
  if d, ok := indata.(compiler.TemplateData); ok {
    data = d
  }
  var writeErr error
  var var0 []*compiler.TemplateData = data.SomeTemplateDataSlice
  var var1 int = len(var0)
  _, writeErr = io.WriteString(w, strconv.Itoa(var1))
  if writeErr != nil {
    return writeErr
  }
  return nil
}`,
			expectBuiltins: map[string]string{},
			funcs: map[string]interface{}{
				"browsePropertyPath": func(x interface{}, p string, args ...interface{}) interface{} { return nil },
				"len":                func(item interface{}) (int, error) { return 0, nil },
			},
			expectImports: []string{
				"io",
				"github.com/mh-cbon/template-compiler/std/text/template/parse",
				"github.com/mh-cbon/template-compiler/compiler",
				"strconv",
			},
			funcsMapPublic: []map[string]string{
				map[string]string{
					"FuncName": "html",
					"Sel":      "template.HTMLEscaper",
					"Pkg":      "text/template",
				},
			},
		},
		TestData{
			tplstr:    `{{if eq true true}}{{end}}`,
			dataValue: TemplateData{},
			expectCompiledFn: `func fn0(t parse.Templater, w io.Writer, indata interface{}) error {
  var var0 bool = true == true
  if var0 {
  }
  return nil
}`,
			expectBuiltins: map[string]string{},
			funcs: map[string]interface{}{
				"browsePropertyPath": func(x interface{}, p string, args ...interface{}) interface{} { return nil },
				"eq":                 func(arg1 interface{}, arg2 ...interface{}) (bool, error) { return false, nil },
			},
			expectImports: []string{
				"io",
				"github.com/mh-cbon/template-compiler/std/text/template/parse",
			},
			funcsMapPublic: []map[string]string{
				map[string]string{
					"FuncName": "html",
					"Sel":      "template.HTMLEscaper",
					"Pkg":      "text/template",
				},
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
			expectCompiledFn: `func fn0(t parse.Templater, w io.Writer, indata interface{}) error {
  var data compiler.TemplateData
  if d, ok := indata.(compiler.TemplateData); ok {
    data = d
  }
  var writeErr error
  var var0 []*compiler.TemplateData = data.SomeTemplateDataSlice
  for tplI, tplV := range var0 {
    _, writeErr = w.Write(builtin0)
    if writeErr != nil {
      return writeErr
    }
    var tplY bool = false
    _, writeErr = w.Write(builtin1)
    if writeErr != nil {
      return writeErr
    }
  }
  if len(var0) == 0 {
    _, writeErr = w.Write(builtin2)
    if writeErr != nil {
      return writeErr
    }
    var tplYShadow bool = false
    _, writeErr = w.Write(builtin1)
    if writeErr != nil {
      return writeErr
    }
  }
  _, writeErr = w.Write(builtin1)
  if writeErr != nil {
    return writeErr
  }
  var tplYShadow0 bool = true
  _, writeErr = w.Write(builtin1)
  if writeErr != nil {
    return writeErr
  }
  if tplYShadow0 {
    _, writeErr = w.Write(builtin3)
    if writeErr != nil {
      return writeErr
    }
  } else {
    _, writeErr = w.Write(builtin4)
    if writeErr != nil {
      return writeErr
    }
  }
  _, writeErr = w.Write(builtin1)
  if writeErr != nil {
    return writeErr
  }
  {
    if tplYShadow0 {
      _, writeErr = io.WriteString(w, strconv.FormatBool(tplYShadow0))
      if writeErr != nil {
        return writeErr
      }
    } else {
      _, writeErr = io.WriteString(w, strconv.FormatBool(data))
      if writeErr != nil {
        return writeErr
      }
    }
  }
  _, writeErr = w.Write(builtin1)
  if writeErr != nil {
    return writeErr
  }
  return nil
}`,
			expectBuiltins: map[string]string{
				"\nHello range branch!\n": "builtin0",
				"\n": "builtin1",
				"\nHello else branch!\n": "builtin2",
				" if branch ":            "builtin3",
				" else branch ":          "builtin4",
			},
			funcs: map[string]interface{}{},
			expectImports: []string{
				"io",
				"strconv",
				"github.com/mh-cbon/template-compiler/std/text/template/parse",
				"github.com/mh-cbon/template-compiler/compiler",
			},
		},
	}

	for i, testData := range allTestData {

		// parse and compile the template file
		tpl, err := template.New("").Funcs(testData.funcs).Parse(testData.tplstr)
		if err != nil {
			t.Errorf("Test(%v): Expected to compile the template, but got an error=%v", i, err)
			return
		}

		// convert it to go code
		compiledProgram := NewCompiledTemplatesProgram("ee")
		typeCheck := simplifier.TransformTree(tpl.Tree, testData.dataValue, testData.funcs)
		err = convertTplTree(
			"fn0",
			tpl.Tree,
			testData.funcs,
			testData.funcsMapPublic,
			makeDataConfiguration(testData.dataValue),
			typeCheck,
			compiledProgram,
		)
		if err != nil {
			t.Errorf("Test(%v): Expected to succeed, but got an error=%v", i, err)
			return
		}

		// ensure the compiled function matches
		astFunc := compiledProgram.funcs[0]

		compiledFn := astNodeToString(astFunc)
		compiledFn = formatGoCode(compiledFn)
		testData.expectCompiledFn = formatGoCode(testData.expectCompiledFn)

		if compiledFn != testData.expectCompiledFn {
			t.Errorf(
				"Unexpected compiled function. Expected=\n%v\n-----\nGot=\n%v\nTEMPLATE:\n%v\nSIMPLIFIED TEMPLATE:\n%v\n",
				testData.expectCompiledFn,
				compiledFn,
				testData.tplstr,
				tpl.Tree.Root.String(),
			)
			return
		}

		// ensure builtins text node are transformed into builtin variables
		builtinTexts := compiledProgram.builtinTexts
		for text, varname := range builtinTexts {
			if expectVarName, ok := testData.expectBuiltins[text]; ok == false {
				t.Errorf(
					"Unexpected builtin variable. Compilation produced an unexpected builtin variable %v for the text %q\nTEMPLATE:\n%v\n",
					varname,
					text,
					testData.tplstr,
				)
				return
			} else if expectVarName != varname {
				t.Errorf(
					"Incorrect variable name for builtin text. The text %q should be registered in the variable %v\nTEMPLATE:\n%v\n",
					text,
					varname,
					testData.tplstr,
				)
				return
			}
		}
		for text, varname := range testData.expectBuiltins {
			if _, ok := builtinTexts[text]; ok == false {
				t.Errorf(
					"Expected builtin variable was not found. Compilation did not produce the builtin variable %v with the text %q\nTEMPLATE:\n%v\n",
					varname,
					text,
					testData.tplstr,
				)
				return
			}
		}

		// ensure the import list matches
		gotImports := convertImportsSpecs(compiledProgram.imports)
		expectedImports := testData.expectImports
		if len(expectedImports) != len(gotImports) {
			t.Errorf(
				"Unexpected additionnal imports. Expected=\n%v\n-----\nGot=\n%v\nTEMPLATE:\n%v\nSIMPLIFIED TEMPLATE:\n%v\n",
				expectedImports,
				gotImports,
				testData.tplstr,
				tpl.Tree.Root.String(),
			)
			return
		}
		for _, i := range expectedImports {
			if strExists(i, gotImports) == false {
				t.Errorf(
					"Missing additionnal imports. Missing=%v\nTEMPLATE:\n%v\nSIMPLIFIED TEMPLATE:\n%v\n",
					i,
					testData.tplstr,
					tpl.Tree.Root.String(),
				)
				return
			}
		}
		for _, i := range gotImports {
			if strExists(i, expectedImports) == false {
				t.Errorf(
					"Unexpected additionnal imports. Unwanted=%v\nTEMPLATE:\n%v\nSIMPLIFIED TEMPLATE:\n%v\n",
					i,
					testData.tplstr,
					tpl.Tree.Root.String(),
				)
				return
			}
		}
	}
}

func strExists(s string, in []string) bool {
	for _, i := range in {
		if i == s {
			return true
		}
	}
	return false
}
