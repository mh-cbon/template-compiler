# template-compiler

Compile your `text/template` / `html/template` to regular go code.

# Install

You will need both library and binary.

```sh
go get github.com/mh-cbon/template-compiler
cd $GOPATH/src/github.com/mh-cbon/template-compiler
glide install
go install
```

# Usage

Let s take this example package

```go
package mypackage

import(
  "net/http"
  "html/template"
)

var tplFuncs = map[string]interface{}{
  "up": strings.ToUpper,
}

type TplData struct {
  Email string
  Name string
}

func handler(w http.ResponseWriter, r *http.Request) {
    t := template.New("").Funcs(tplFuncs).ParseFiles("tmpl/welcome.html")
    t.Execute(w, TplData{})
}
```

With this template

`{{.Email}} {{.Name}}`

To generate compiled version of your template, change it to

```go
package mypackage

import (
  "net/http"

	"github.com/mh-cbon/template-compiler/compiled"
	"github.com/mh-cbon/template-compiler/std/text/template"
	"github.com/mh-cbon/template-compiler/std/text/template/parse"
)

//go:generate template-compiler -html -tpl "tmpl/*.tpl" -out "gen.go" -var compiledTemplates -data path/to/mypackage:TplData path/to/mypackage:tplFuncs
var compiledTemplates *compiled.Registry

var tplFuncs = map[string]interface{}{
  "up": strings.ToUpper,
}

type TplData struct {
  Email string
  Name string
}

func handler(w http.ResponseWriter, r *http.Request) {
    compiledTemplates.MustGet("welcome.tpl").Execute(w, TplData{})
}
```

Then run,

```sh
go generate
```

It will produce a file `gen.go` containing the code to declare and run the compiled templates,

```go
package main

//golint:ignore

import (
 "io"
 "github.com/mh-cbon/template-compiler/compiled"
 "github.com/mh-cbon/template-compiler/std/text/template/parse"
 "text/template"
 "strconv"
 dataalias "github.com/mh-cbon/template-compiler/demo/data"
)

func init () {
  compiledTemplates = compiled.NewRegistry()
  compiledTemplates.Add("welcome.tpl", fnaTplaTpl0)
}

func fnaTplaTpl0(t parse.Templater, w io.Writer, indata interface {
}) error {
	var data dataalias.MyTemplateData
	if d, ok := indata.(dataalias.MyTemplateData); ok {
		data = d
	}
	var writeErr error
  var var0 string = data.Email
	_, writeErr = w.Write(var0)
	if writeErr != nil {
		return writeErr
	}
	_, writeErr = w.Write(builtin0)
	if writeErr != nil {
		return writeErr
	}
  var var1 string = data.Name
	_, writeErr = w.Write(var0)
	if writeErr != nil {
		return writeErr
	}
	return nil
}
var builtin0 = []byte(" ")

```

### What would be the performance improvements ?

Given the templates available [here](...)

```sh
$ go test -bench=.

a.tpl
BenchmarkRenderWithCompiledTemplate-4    	100000000	        10.5 ns/op
BenchmarkRenderWithJitTemplate-4         	 5000000	       384 ns/op
                                          x20 times faster.

c.tpl
BenchmarkRenderWithCompiledTemplateC-4   	 2000000	       641 ns/op
BenchmarkRenderWithJitTemplateC-4        	  100000	     18064 ns/op
                                          x20 times faster.

d.tpl
BenchmarkRenderWithCompiledTemplateD-4   	50000000	        34.5 ns/op
BenchmarkRenderWithJitTemplateD-4        	 3000000	       478 ns/op
                                          x16 times faster.

PASS
ok  	github.com/mh-cbon/template-compiler/demo	10.993s
```

Generally speaking the more complex the template is,
the more output you ll get.


### More coming soon.
