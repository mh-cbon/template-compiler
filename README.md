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

# CLI

```sh
template-compiler - 0.0.0

  -help | -h   Show this help.
  -version     Show program version.
  -keep        Keep bootstrap program compiler.
  -print       Print bootstrap program compiler.
  -var         The variable name of the configuration in your program
               default: compiledTemplates
  -wdir        The working directory where the bootstrap program is written
               default: $GOPATH/src/template-compilerxx/

Examples
  template-compiler -h
  template-compiler -version
  template-compiler -keep -var theVarName
  template-compiler -keep -var theVarName -wdir /tmp
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
)

//go:generate template-compiler
var compiledTemplates = compiled.New(
  "gen.go",
  []compiled.TemplateConfiguration{
    compiled.TemplateConfiguration{
      HTML:          true,
      TemplatesPath: "tmpl/*.tpl",
      Data:          TplData{},
      FuncsMap:      []string{
        "somewhere/mypackage:tplFuncs",
      },
    },
  },
)

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
 "path/to/mypackage"
)

func init () {
  compiledTemplates = compiled.NewRegistry()
  compiledTemplates.Add("welcome.tpl", fnaTplaTpl0)
}

func fnaTplaTpl0(t parse.Templater, w io.Writer, indata interface {
}) error {
  var data mypackage.TplData
  if d, ok := indata.(mypackage.TplData); ok {
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

Given the templates available [here](https://github.com/mh-cbon/template-compiler/tree/master/demo/templates)

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

# Understanding

This paragraph will describe and explain the various steps from the `go:generate` command,
to the write of the compiled go code.

1. When `go:generate` is invoked, the go tool will parse and invoke your calls to
`template-compiler`. `template-compiler` is invoked in the directory containing the file
with the `go:generate` comment, `go generate` also declares an environment variable `GOFILE`.
With those hints `template-compiler` can locate and consume the variable declared with `-var` parameter.
[We are here](https://github.com/mh-cbon/template-compiler/blob/master/main.go#L17)
2. `template-compiler` will generate a bootstrap program.
[We are here](https://github.com/mh-cbon/template-compiler/blob/master/compiler/bootstrap.go#L21)
3. The generation of the bootstrap program is about parsing, browsing, and re exporting
an updated version of your configuration variable.
It specifically looks for each `compiled.TemplateConfiguration{}`:
- If the configuration is set to generate html content with the key `HTML:true`,
  it ensure that stdfunc are appropriately declared into the configuration.
- It read and evaluates the data field `Data: your.struct{}`,
  generates a `DataConfiguration{}` of it, and adds it to the template configuration.
- It checks for `FuncsMap` key, and export those variable targets
  (with the help of [this package](https://github.com/mh-cbon/export-funcmap))
  to `FuncsExport` and `PublicIdents` keys.
 [We are here](https://github.com/mh-cbon/template-compiler/blob/master/compiler/bootstrap.go#L117)
4. `template-compiler` writes and compiles a go program into
`$GOPATH/src/template-compilerxxx`.
This program is made to compile the templates with the updated configuration.
[We are here](https://github.com/mh-cbon/template-compiler/blob/master/compiler/bootstrap.go#L93)
5. `bootstrap-program` is now invoked.
[We are here](https://github.com/mh-cbon/template-compiler/blob/master/main.go#L72)
6. `bootstrap-program` browses the configuration value,
for each template path, it compiles it as `text/template` or `html/template`.
This steps creates the standard template AST Tree. Each template tree is then
transformed and simplified [with the help of this package](https://github.com/mh-cbon/template-tree-simplifier).
[We are here](https://github.com/mh-cbon/template-compiler/blob/master/compiler/compile.go#L103)
7. `template-tree-simplifier` takes in input the template tree and apply transformations:
  - It unshadows all variables declaration within the template.
  - It renames all template variables to prefix them with `tpl`
  - It simplifies structure such as `{{"son" | split "wat"}}` to `{{$var0 := split "wat" "son"}}{{$var0}}`
  - It produces a small type checker structure which registers variable and their type for each scope of the template.
[We are here](https://github.com/mh-cbon/template-compiler/blob/master/compiler/compile.go#L323)
8. `bootstrap-program` browses each simplified template tree, generates a go function corresponding to it.
[We are here](https://github.com/mh-cbon/template-compiler/blob/master/compiler/compile.go#L77)
9. `bootstrap-program` generates an `init` function with registers the new functions as their templates names to
your configuration variable.
[We are here](https://github.com/mh-cbon/template-compiler/blob/master/compiler/compile.go#L204)
10. `bootstrap-program` writes the fully generated program.

### Working with funcmap

`template-compiler` needs to be able to evaluate the `funcmap` consumed by the templates.

In that matter `template-compiler` can take in input a path to a variable declaring this functions.

`pkgPath:variableName` where `pkgPath` is the go pakage path such as `text/template`,
the variable name is the name of the variable declaring the funcmap such as `builtins`.
See [this](https://golang.org/src/text/template/funcs.go#L26).

It can read `map[string]interface{}` or `template.FuncMap`.

It can extract `exported` or `unexported` variables.

Functions declared into the funcmap can be `exported`, `unexported`, or inlined.

Note that `unexported` functions needs some runtime type checking.

__examples__

If you like [sprig](https://github.com/Masterminds/sprig), you d be able to consume those functions with the path,

`github.com/Masterminds/sprig:genericMap`

If you prefer [gtf](https://github.com/leekchan/gtf), you d be able to consume those functions with the path,

`github.com/leekchan/gtf:GtfFuncMap`


__beware__

it can t evaluate a function call! It must be a variable declaration into the top level context such as

```go
package yy

var funcs = map[string]interface{}{
  "funcname": func(){},
  "funcname2": pkg.Func,
}
```

### Working with template data

The data consumed by your template must follow few rules:
- It must be an exported type.
- It must not be declared into a `main` package.

### Others warnings

As the resulting compilation is pure go code, the type system must be respected,
thus `unexported` types may not work.

# The ugly stuff

Unfortunately this package contains some ugly copy pastes :x :x :x

It duplicates both `text/template` and `html/template`.

It would be great to backport those changes into core go code to get ride of those duplications.

1. Added a new `text/template.Compiled` type. Much like a `text/template` or an `html/template`,
`Compiled` has a `*parse.Tree`. This tree is a bultin tree to hold only one node to execute the compiled function.
Doing so allow to mix compiled and non-compiled templates.
[see here](https://github.com/mh-cbon/template-compiler/blob/master/std/text/template/compiled.go#L31)
2. Added a new method `text/template.GetFuncs()` to get the funcs related to the template.
This is usefull to the compiled template functions to get access to those unexported functions.
[see here](https://github.com/mh-cbon/template-compiler/blob/master/std/text/template/compiled.go#L26)
3. Added `text/template.Compiled()` to attach a compiled template to a regular `text/template` instance.
[see here](https://github.com/mh-cbon/template-compiler/blob/master/std/text/template/compiled.go#L26)
4. Added a new tree node `text/template/parse.CompiledNode`, which knows the function to execute
for a compiled template.
[see here](https://github.com/mh-cbon/template-compiler/blob/master/std/text/template/parse/compiled.go#L10)
5. Added a new interface `text/template/parse.Templater`,
to use in the compiled function to receive the current template executed. This instance can be one of
`text/template.Template`, `html/template.Template` or `text/template.Compiled`.
[see here](https://github.com/mh-cbon/template-compiler/blob/master/std/text/template/parse/compiled.go#L18)
6. Added a new type `CompiledTemplateFunc` for the signature of a compiled template function.
[see here](https://github.com/mh-cbon/template-compiler/blob/master/std/text/template/parse/compiled.go#L24)
7. Added a new funcmap variable `html/template.publicFuncMap` to map all html template idents to a function.
It also delcares all escapers to a public function to improve performance of compiled templates.
[see here](https://github.com/mh-cbon/template-compiler/blob/master/std/html/template/compiled.go#L6)
8. Added support of CompiledNode to the state walker
[see here](https://github.com/mh-cbon/template-compiler/blob/master/std/text/template/exec.go#L247)

# TBD

Here are some optimizations/todos to implement later:

- When compiling templates, funcs like `_html_template_htmlescaper` will translate to `template.HTMLEscaper`.
It worth to note that many cases are probably `template.HTMLEscaper(string)`, but `template.HTMLEscaper` is doing
some extra job to type check this `string` value.
An optimization is to detect those calls `template.HTMLEscaper(string)` and transformedform them to `template.HTMLEscapeString(string)`
- Same as previous for most escaper functions of `html/template`
- Detect template calls such `eq(bool, bool)`, or `neq(int, int)` and transform them to an
appropriate go binary test `bool == bool`, ect.
- Detect templates calls such `len(some)` and transforms it to the builtin `len` function.
- Detect prints of `struct` or `*struct`, check if they implements `Stringer`,
or something like `Byter`, and make use of that to get ride of some `fmt.Sprintf` calls.
- review the install procedure, i suspect it is not yet correct. Make use of glide.
- consolidate additions to std `text/template`/`html/template` packages.
- version releases.
- implement cache for functions export.
