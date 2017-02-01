package compiler

import (
	"go/ast"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/mh-cbon/template-compiler/compiled"
	"github.com/mh-cbon/template-compiler/demo/data"
)

type CompileTestData struct {
	varName string
	outPkg  string
	conf    []*TemplateToCompile
	// tpls        []*TemplateToCompile
	expectedErr error
	// the list of expected imports to find in the program, noted as alias:path, or just path
	expectedImports  []string
	expectedInitFunc string
	expectedTplsFunc map[string]string
	expectedBuiltins map[string]string
}

func TestCompile(t *testing.T) {

	allDataTest := []CompileTestData{
		CompileTestData{
			varName: "xx",
			outPkg:  "gen",
			conf: []*TemplateToCompile{
				makeConf(false,
					map[string]interface{}{
						"a.tpl": data.MyTemplateData{},
					},
					map[string]string{
						"a.tpl": ``,
					},
				),
			},
			expectedImports: []string{
				"io",
				"github.com/mh-cbon/template-compiler/std/text/template/parse",
			},
			expectedInitFunc: `func init() {
  xx.Add("a.tpl", fnaTpl)
}`,
			expectedTplsFunc: map[string]string{
				"fnaTpl": `func fnaTpl(t parse.Templater, w io.Writer, indata interface{}) error {
return nil
}`,
			},
			expectedBuiltins: map[string]string{},
		},
		CompileTestData{
			varName: "yy",
			outPkg:  "notgen",
			conf: []*TemplateToCompile{
				makeConf(false,
					map[string]interface{}{
						"b.tpl": data.MyTemplateData{},
					},
					map[string]string{
						"b.tpl": `{{$y := 4}}{{$y}}{{.}}`,
					}),
			},
			expectedImports: []string{
				"io",
				"strconv",
				"fmt",
				"github.com/mh-cbon/template-compiler/std/text/template/parse",
				"aliasdata:github.com/mh-cbon/template-compiler/demo/data",
			},
			expectedInitFunc: `func init() {
yy.Add("b.tpl", fnbTpl)
}`,
			expectedTplsFunc: map[string]string{
				"fnbTpl": `func fnbTpl(t parse.Templater, w io.Writer, indata interface{}) error {
  var data aliasdata.MyTemplateData
  if d, ok := indata.(aliasdata.MyTemplateData); ok {
    data = d
  }
  var tplY int = 4
  if _, werr := io.WriteString(w, strconv.Itoa(tplY)); werr != nil {
    return werr
  }
  if _, werr := fmt.Fprintf(w, "%v", data); werr != nil {
    return werr
  }
  return nil
}`,
			},
			expectedBuiltins: map[string]string{},
		},
		CompileTestData{
			varName: "yy",
			outPkg:  "notgen",
			conf: []*TemplateToCompile{
				makeConf(true,
					map[string]interface{}{
						"b.tpl": data.MyTemplateData{},
					}, map[string]string{
						"b.tpl": `samebuiltin{{$y := 4}}samebuiltin{{$y}}samebuiltin`,
					}),
			},
			expectedImports: []string{
				"io",
				"github.com/mh-cbon/template-compiler/std/text/template/parse",
				"github.com/mh-cbon/template-compiler/std/html/template",
			},
			expectedInitFunc: `func init() {
yy.Add("b.tpl", fnbTpl)
}`,
			expectedTplsFunc: map[string]string{
				"fnbTpl": `func fnbTpl(t parse.Templater, w io.Writer, indata interface{}) error {
  if _, werr := w.Write(builtin0); werr != nil {
    return werr
  }
  var tplY int = 4
  if _, werr := w.Write(builtin0); werr != nil {
    return werr
  }
  var var0 string = template.HTMLEscaper(tplY)
  if _, werr := io.WriteString(w, var0); werr != nil {
    return werr
  }
  if _, werr := w.Write(builtin0); werr != nil {
    return werr
  }
  return nil
}`,
			},
			expectedBuiltins: map[string]string{
				"builtin0": `[]byte("samebuiltin")`,
			},
		},
		CompileTestData{
			varName: "yy",
			outPkg:  "notgen",
			conf: []*TemplateToCompile{
				makeConf(true,
					map[string]interface{}{
						"b.tpl": &data.MyTemplateData{},
					},
					map[string]string{
						"b.tpl": `{{$y := true}}{{$y}}{{.}}`,
					}),
			},
			expectedImports: []string{
				"io",
				"github.com/mh-cbon/template-compiler/std/text/template/parse",
				"aliasdata:github.com/mh-cbon/template-compiler/demo/data",
				"github.com/mh-cbon/template-compiler/std/html/template",
			},
			expectedInitFunc: `func init() {
yy.Add("b.tpl", fnbTpl)
}`,
			expectedTplsFunc: map[string]string{
				"fnbTpl": `func fnbTpl(t parse.Templater, w io.Writer, indata interface{}) error {
  var data *aliasdata.MyTemplateData
  if d, ok := indata.(*aliasdata.MyTemplateData); ok {
    data = d
  }
  var tplY bool = true
  var var0 string = template.HTMLEscaper(tplY)
  if _, werr := io.WriteString(w, var0); werr != nil {
    return werr
  }
  var var1 string = template.HTMLEscaper(data)
  if _, werr := io.WriteString(w, var1); werr != nil {
    return werr
  }
  return nil
}`,
			},
			expectedBuiltins: map[string]string{},
		},
		CompileTestData{
			varName: "yy",
			outPkg:  "notgen",
			conf: []*TemplateToCompile{
				makeConf(false,
					map[string]interface{}{
						"b.tpl": &data.MyTemplateData{},
						"z":     &data.MyTemplateData{},
					}, map[string]string{
						"b.tpl": `{{define "z"}}z template{{end}}b template`,
					}),
			},
			expectedImports: []string{
				"io",
				"github.com/mh-cbon/template-compiler/std/text/template/parse",
			},
			expectedInitFunc: `func init() {
      yy.Add("b.tpl", fnbTpl)
      yy.Add("z", fnbTplZ)
      tpl0X0 := yy.MustGet("b.tpl")
      tpl0Y0 := yy.MustGet("z")
      tpl0X0, _ = tpl0X0.Compiled(tpl0Y0)
      yy.Set("b.tpl", tpl0X0)
}`,
			expectedTplsFunc: map[string]string{
				"fnbTpl": `func fnbTpl(t parse.Templater, w io.Writer, indata interface{}) error {
  if _, werr := w.Write(builtin0); werr != nil {
    return werr
  }
  return nil
}`,
				"fnbTplZ": `func fnbTplZ(t parse.Templater, w io.Writer, indata interface{}) error {
  if _, werr := w.Write(builtin1); werr != nil {
    return werr
  }
  return nil
}`,
			},
			expectedBuiltins: map[string]string{
				"builtin0": `[]byte("b template")`,
				"builtin1": `[]byte("z template")`,
			},
		},
		CompileTestData{
			varName: "yy",
			outPkg:  "notgen",
			conf: []*TemplateToCompile{
				makeConf(false,
					map[string]interface{}{
						"b.tpl": data.MyTemplateData{},
						"z":     data.MyTemplateData{},
					}, map[string]string{
						"b.tpl": `{{define "z"}}z template{{end}}b template`,
					}),
				makeConf(false,
					map[string]interface{}{
						"b.tpl": data.MyTemplateData{},
						"x":     data.MyTemplateData{},
					}, map[string]string{
						"b.tpl": `{{define "x"}}x template{{end}}b template 2`,
					}),
			},
			expectedImports: []string{
				"io",
				"github.com/mh-cbon/template-compiler/std/text/template/parse",
			},
			expectedInitFunc: `func init() {
        yy.Add("b.tpl", fnbTpl)
        yy.Add("z", fnbTplZ)
        yy.Add("b.tpl", fn0fnbTpl)
        yy.Add("x", fnbTplX)
        tpl0X0 := yy.MustGet("b.tpl")
        tpl0Y0 := yy.MustGet("z")
        tpl0X0, _ = tpl0X0.Compiled(tpl0Y0)
        yy.Set("b.tpl", tpl0X0)
        tpl0X0 := yy.MustGet("b.tpl")
        tpl0Y0 := yy.MustGet("x")
        tpl0X0, _ = tpl0X0.Compiled(tpl0Y0)
        yy.Set("b.tpl", tpl0X0)
}`,
			expectedTplsFunc: map[string]string{
				"fnbTpl": `func fnbTpl(t parse.Templater, w io.Writer, indata interface{}) error {
  if _, werr := w.Write(builtin0); werr != nil {
    return werr
  }
  return nil
}`,
				"fnbTplZ": `func fnbTplZ(t parse.Templater, w io.Writer, indata interface{}) error {
  if _, werr := w.Write(builtin1); werr != nil {
    return werr
  }
  return nil
}`,
				"fn0fnbTpl": `func fn0fnbTpl(t parse.Templater, w io.Writer, indata interface{}) error {
  if _, werr := w.Write(builtin2); werr != nil {
    return werr
  }
  return nil
}`,
				"fnbTplX": `func fnbTplX(t parse.Templater, w io.Writer, indata interface{}) error {
  if _, werr := w.Write(builtin3); werr != nil {
    return werr
  }
  return nil
}`,
			},
			expectedBuiltins: map[string]string{
				"builtin0": `[]byte("b template")`,
				"builtin1": `[]byte("z template")`,
				"builtin2": `[]byte("b template 2")`,
				"builtin3": `[]byte("x template")`,
			},
		},
	}

	for i, dataTest := range allDataTest {
		compiler := NewCompiledTemplatesProgram(dataTest.varName)
		program, err := compiler.compileTemplates(dataTest.outPkg, dataTest.conf)

		if dataTest.expectedErr != nil {
			if err == nil {
				t.Errorf("Test(%v): Expected to fail, but got err=%v", i, err)
				return
			} else if err.Error() != dataTest.expectedErr.Error() {
				t.Errorf("Test(%v): Expected to fail with the message=%v but got err=%v", i, dataTest.expectedErr, err)
				return
			}
		}

		extractedPkg := LookupPackageNameFromStr(program)
		if dataTest.outPkg != extractedPkg {
			t.Errorf("Test(%v): Unexpected output package expected=%v, got=%v", i, dataTest.outPkg, extractedPkg)
			return
		}

		f := stringToAst(program)

		importSpecs := extractImports(f)
		imports := convertImportsSpecs(importSpecs)
		if len(dataTest.expectedImports) != len(imports) {
			t.Errorf("Test(%v): Expected to get %v import statements, but found %v\n\n%v",
				i, len(dataTest.expectedImports), len(imports), program)
			return
		}

		for _, im := range imports {
			if containsStr(dataTest.expectedImports, im) == false {
				t.Errorf("Test(%v): Found unexpected import=%v\n\n%v", i, im, program)
				return
			}
		}

		for _, im := range dataTest.expectedImports {
			if containsStr(imports, im) == false {
				t.Errorf("Test(%v): Expected to find import=%v\n\n%v", i, im, program)
				return
			}
		}

		initfn := extractFunc(f, "init")
		if initfn == nil {
			t.Errorf("Test(%v): Expected to find init function\n\n%v", i, program)
			return
		}
		initFnString := astNodeToString(initfn)
		initFnString = formatGoCode(initFnString)
		expectedInit := dataTest.expectedInitFunc
		expectedInit = formatGoCode(expectedInit)
		if expectedInit != initFnString {
			t.Errorf("Test(%v): Unexpected content of init function\nexpected=\n%v\ngot\n%v\n\n%v", i, expectedInit, initFnString, program)
			return
		}

		for _, conf := range dataTest.conf {
			for _, tplfile := range conf.files {
				for _, fnname := range tplfile.tplsFunc {

					tfn := extractFunc(f, fnname)
					if tfn == nil {
						t.Errorf("Test(%v): Expected to find a function=%v\n\n%v", i, fnname, program)
						return
					}
					funcString := astNodeToString(tfn)
					funcString = formatGoCode(funcString)

					expectedFn, foundFn := dataTest.expectedTplsFunc[fnname]
					if foundFn == false {
						t.Errorf("Test(%v): Unexpected compiled func=%v\n\n%v", i, fnname, program)
						return
					}
					expectedFn = formatGoCode(expectedFn)
					if expectedFn != funcString {
						t.Errorf(
							"Test(%v): Unexpected content of template function %v\nexpected=\n%v\ngot\n%v\n\n%v",
							i, fnname, expectedFn, funcString, program,
						)
						return
					}

				}
			}
		}
		//-
		builtins := extractVarLike(f, "builtin")
		expectedBuiltins := dataTest.expectedBuiltins

		if len(builtins) != len(expectedBuiltins) {
			t.Errorf("Test(%v): Expected to get %v builtins variables, but found %v\n\n%v",
				i, len(expectedBuiltins), len(builtins), program)
			return
		}

		for name := range expectedBuiltins {
			found := false
			for _, b := range builtins {
				v := b.Specs[0].(*ast.ValueSpec)
				if name == v.Names[0].Name {
					found = true
					break
				}
			}
			if found == false {
				t.Errorf("Test(%v): Expected to find a builtin variable %v but it is missing\n\n%v",
					i, name, program)
				return
			}
		}

		for _, b := range builtins {
			v := b.Specs[0].(*ast.ValueSpec)
			name := v.Names[0].Name
			if _, ok := expectedBuiltins[name]; ok == false {
				t.Errorf("Test(%v): Unexpected builtin variable was found with name=%v\n\n%v",
					i, name, program)
				return
			}
			expectedContent := expectedBuiltins[name]
			content := astNodeToString(v.Values[0])
			if content != expectedContent {
				t.Errorf("Test(%v): Unexpected content found for builtin variable %v\nexpected:%v\ngot:%v\n\n%v",
					i, name, expectedContent, content, program)
				return
			}
		}
		//-
	}
	//-
}

func makeConf(
	isHTML bool,
	data map[string]interface{},
	tpls map[string]string,
) *TemplateToCompile {
	funcsmap := textTemplateFuncExports
	publicIdents := textTemplatePublicIdents
	if isHTML {
		funcsmap = htmlFuncsExport
		publicIdents = htmlPublicIdents
	}
	ret := makeTemplateToCompile(
		compiled.TemplateConfiguration{
			HTML:                       isHTML,
			TemplatesData:              data,
			TemplatesDataConfiguration: makeMapDataConfiguration(data),
			FuncsExport:                funcsmap,
			PublicIdents:               publicIdents,
		},
	)
	for name, content := range tpls {
		t, err := makeTemplateFileToCompileFromStr(
			name, content, ret,
		)
		if err != nil {
			panic(err)
		}
		ret.files = append(ret.files, t)
	}
	return ret
}

// given a go file ast.Node, extracts a (top-level) variable declaration with name starting with given name.
func extractVarLike(file *ast.File, varName string) []*ast.GenDecl {
	var found []*ast.GenDecl
	ast.Inspect(file, func(n ast.Node) bool {
		if d, ok := n.(*ast.GenDecl); ok {
			if v, okk := d.Specs[0].(*ast.ValueSpec); okk {
				if len(v.Names) > 0 && len(v.Names[0].Name) >= len(varName) && v.Names[0].Name[0:len(varName)] == varName {
					found = append(found, d)
				}
			}
		}
		return true
	})

	return found
}

// given a go file ast.Node, extracts a (top-level) func declaration.
func extractFunc(file *ast.File, funcName string) *ast.FuncDecl {
	var found *ast.FuncDecl
	ast.Inspect(file, func(n ast.Node) bool {
		if d, ok := n.(*ast.FuncDecl); ok {
			if d.Name.Name == funcName {
				found = d
			}
		}
		return found == nil
	})

	return found
}

func makeMapDataConfiguration(some map[string]interface{}) map[string]compiled.DataConfiguration {
	ret := map[string]compiled.DataConfiguration{}
	for name, s := range some {
		ret[name] = makeDataConfiguration(s)
	}
	return ret
}

func makeDataConfiguration(some interface{}) compiled.DataConfiguration {
	ret := compiled.DataConfiguration{}
	if some != nil {
		r := reflect.TypeOf(some)
		isPtr := r.Kind() == reflect.Ptr
		if isPtr {
			r = r.Elem()
		}
		ret.IsPtr = isPtr
		ret.DataTypeName = r.Name()
		ret.PkgPath = r.PkgPath()
		ret.DataType = filepath.Base(r.PkgPath()) + "." + r.Name()
	}
	return ret
}

func convertImportsSpecs(importSpecs []*ast.ImportSpec) []string {
	imports := make([]string, 0)
	for _, im := range importSpecs {
		p := im.Path.Value
		p = p[1 : len(p)-1] // remove quotes
		if im.Name != nil {
			p = im.Name.Name + ":" + p
		}
		imports = append(imports, p)
	}
	return imports
}

// some static vars :x
var textTemplateFuncExports = map[string]interface{}{
	"and":                func(arg0 interface{}, args ...interface{}) interface{} { return nil },
	"call":               func(fn interface{}, args ...interface{}) (interface{}, error) { return nil, nil },
	"html":               func(args ...interface{}) string { return "" },
	"index":              func(item interface{}, indices ...interface{}) (interface{}, error) { return nil, nil },
	"js":                 func(args ...interface{}) string { return "" },
	"len":                func(item interface{}) (int, error) { return 0, nil },
	"not":                func(arg interface{}) bool { return false },
	"or":                 func(arg0 interface{}, args ...interface{}) interface{} { return nil },
	"print":              func(a ...interface{}) string { return "" },
	"printf":             func(format string, a ...interface{}) string { return "" },
	"println":            func(a ...interface{}) string { return "" },
	"urlquery":           func(args ...interface{}) string { return "" },
	"eq":                 func(arg1 interface{}, arg2 ...interface{}) (bool, error) { return false, nil },
	"ge":                 func(arg1 interface{}, arg2 interface{}) (bool, error) { return false, nil },
	"gt":                 func(arg1 interface{}, arg2 interface{}) (bool, error) { return false, nil },
	"le":                 func(arg1 interface{}, arg2 interface{}) (bool, error) { return false, nil },
	"lt":                 func(arg1 interface{}, arg2 interface{}) (bool, error) { return false, nil },
	"ne":                 func(arg1 interface{}, arg2 interface{}) (bool, error) { return false, nil },
	"browsePropertyPath": func(some interface{}, propertypath string, args ...interface{}) interface{} { return nil },
}

var textTemplatePublicIdents = []map[string]string{
	map[string]string{"FuncName": "html", "Sel": "template.HTMLEscaper", "Pkg": "text/template"},
	map[string]string{"FuncName": "js", "Sel": "template.JSEscaper", "Pkg": "text/template"},
	map[string]string{"FuncName": "print", "Sel": "fmt.Sprint", "Pkg": "fmt"},
	map[string]string{"FuncName": "printf", "Sel": "fmt.Sprintf", "Pkg": "fmt"},
	map[string]string{"FuncName": "println", "Sel": "fmt.Sprintln", "Pkg": "fmt"},
	map[string]string{"FuncName": "urlquery", "Sel": "template.URLQueryEscaper", "Pkg": "text/template"},
	map[string]string{"FuncName": "browsePropertyPath", "Sel": "funcmap.BrowsePropertyPath", "Pkg": "github.com/mh-cbon/template-tree-simplifier/funcmap"},
}

var htmlFuncsExport = map[string]interface{}{
	"and":                            func(arg0 interface{}, args ...interface{}) interface{} { return nil },
	"call":                           func(fn interface{}, args ...interface{}) (interface{}, error) { return nil, nil },
	"html":                           func(args ...interface{}) string { return "" },
	"index":                          func(item interface{}, indices ...interface{}) (interface{}, error) { return nil, nil },
	"js":                             func(args ...interface{}) string { return "" },
	"len":                            func(item interface{}) (int, error) { return 0, nil },
	"not":                            func(arg interface{}) bool { return false },
	"or":                             func(arg0 interface{}, args ...interface{}) interface{} { return nil },
	"print":                          func(a ...interface{}) string { return "" },
	"printf":                         func(format string, a ...interface{}) string { return "" },
	"println":                        func(a ...interface{}) string { return "" },
	"urlquery":                       func(args ...interface{}) string { return "" },
	"eq":                             func(arg1 interface{}, arg2 ...interface{}) (bool, error) { return false, nil },
	"ge":                             func(arg1 interface{}, arg2 interface{}) (bool, error) { return false, nil },
	"gt":                             func(arg1 interface{}, arg2 interface{}) (bool, error) { return false, nil },
	"le":                             func(arg1 interface{}, arg2 interface{}) (bool, error) { return false, nil },
	"lt":                             func(arg1 interface{}, arg2 interface{}) (bool, error) { return false, nil },
	"ne":                             func(arg1 interface{}, arg2 interface{}) (bool, error) { return false, nil },
	"browsePropertyPath":             func(some interface{}, propertypath string, args ...interface{}) interface{} { return nil },
	"_html_template_attrescaper":     func(args ...interface{}) string { return "" },
	"_html_template_commentescaper":  func(args ...interface{}) string { return "" },
	"_html_template_cssescaper":      func(args ...interface{}) string { return "" },
	"_html_template_cssvaluefilter":  func(args ...interface{}) string { return "" },
	"_html_template_htmlnamefilter":  func(args ...interface{}) string { return "" },
	"_html_template_htmlescaper":     func(args ...interface{}) string { return "" },
	"_html_template_jsregexpescaper": func(args ...interface{}) string { return "" },
	"_html_template_jsstrescaper":    func(args ...interface{}) string { return "" },
	"_html_template_jsvalescaper":    func(args ...interface{}) string { return "" },
	"_html_template_nospaceescaper":  func(args ...interface{}) string { return "" },
	"_html_template_rcdataescaper":   func(args ...interface{}) string { return "" },
	"_html_template_urlescaper":      func(args ...interface{}) string { return "" },
	"_html_template_urlfilter":       func(args ...interface{}) string { return "" },
	"_html_template_urlnormalizer":   func(args ...interface{}) string { return "" },
}

var htmlPublicIdents = []map[string]string{
	map[string]string{"FuncName": "html", "Sel": "template.HTMLEscaper", "Pkg": "text/template"},
	map[string]string{"FuncName": "js", "Sel": "template.JSEscaper", "Pkg": "text/template"},
	map[string]string{"Pkg": "fmt", "FuncName": "print", "Sel": "fmt.Sprint"},
	map[string]string{"FuncName": "printf", "Sel": "fmt.Sprintf", "Pkg": "fmt"},
	map[string]string{"FuncName": "println", "Sel": "fmt.Sprintln", "Pkg": "fmt"},
	map[string]string{"FuncName": "urlquery", "Sel": "template.URLQueryEscaper", "Pkg": "text/template"},
	map[string]string{"FuncName": "browsePropertyPath", "Sel": "funcmap.BrowsePropertyPath", "Pkg": "github.com/mh-cbon/template-tree-simplifier/funcmap"},
	map[string]string{"FuncName": "_html_template_attrescaper", "Sel": "template.AttrEscaper", "Pkg": "github.com/mh-cbon/template-compiler/std/html/template"},
	map[string]string{"Sel": "template.CommentEscaper", "Pkg": "github.com/mh-cbon/template-compiler/std/html/template", "FuncName": "_html_template_commentescaper"},
	map[string]string{"FuncName": "_html_template_cssescaper", "Sel": "template.CSSEscaper", "Pkg": "github.com/mh-cbon/template-compiler/std/html/template"},
	map[string]string{"Pkg": "github.com/mh-cbon/template-compiler/std/html/template", "FuncName": "_html_template_cssvaluefilter", "Sel": "template.CSSValueFilter"},
	map[string]string{"FuncName": "_html_template_htmlnamefilter", "Sel": "template.HTMLNameFilter", "Pkg": "github.com/mh-cbon/template-compiler/std/html/template"},
	map[string]string{"FuncName": "_html_template_htmlescaper", "Sel": "template.HTMLEscaper", "Pkg": "github.com/mh-cbon/template-compiler/std/html/template"},
	map[string]string{"FuncName": "_html_template_jsregexpescaper", "Sel": "template.JSRegexpEscaper", "Pkg": "github.com/mh-cbon/template-compiler/std/html/template"},
	map[string]string{"FuncName": "_html_template_jsstrescaper", "Sel": "template.JSStrEscaper", "Pkg": "github.com/mh-cbon/template-compiler/std/html/template"},
	map[string]string{"FuncName": "_html_template_jsvalescaper", "Sel": "template.JSValEscaper", "Pkg": "github.com/mh-cbon/template-compiler/std/html/template"},
	map[string]string{"FuncName": "_html_template_nospaceescaper", "Sel": "template.HTMLNospaceEscaper", "Pkg": "github.com/mh-cbon/template-compiler/std/html/template"},
	map[string]string{"Sel": "template.RcdataEscaper", "Pkg": "github.com/mh-cbon/template-compiler/std/html/template", "FuncName": "_html_template_rcdataescaper"},
	map[string]string{"Pkg": "github.com/mh-cbon/template-compiler/std/html/template", "FuncName": "_html_template_urlescaper", "Sel": "template.URLEscaper"},
	map[string]string{"FuncName": "_html_template_urlfilter", "Sel": "template.URLFilter", "Pkg": "github.com/mh-cbon/template-compiler/std/html/template"},
}
