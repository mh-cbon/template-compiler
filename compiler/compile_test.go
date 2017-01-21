package compiler

import (
	"fmt"
	html "html/template"
	"io/ioutil"
	"testing"
	text "text/template"
	"text/template/parse"
)

type CompileTestData struct {
	tplName string
	tplStr  string
	isHTML  bool
	// tpls           []*TemplateToCompile
	programPkg      string
	varName         string
	dataPkg         string
	dataType        string
	dataValue       interface{}
	funcsMap        map[string]interface{}
	funcsMapPublic  []map[string]string
	expectedProgram string
}

// more to do here
func TestCompile(t *testing.T) {

	allTestData := []CompileTestData{
		CompileTestData{
			tplName:    "hello",
			tplStr:     ``,
			programPkg: "gen",
			varName:    "registryVarName",
			dataPkg:    "gen",
			dataType:   "gen",
			dataValue:  "gen",
			funcsMap: map[string]interface{}{
				"up": func(s string) string { return s },
			},
			funcsMapPublic: []map[string]string{},
			expectedProgram: `package gen

//golint:ignore

import (
  "gen"
  "github.com/mh-cbon/template-compiler/compiled"
  "github.com/mh-cbon/template-compiler/std/text/template/parse"
  "io"
)

func init() {
  registryVarName = compiled.NewRegistry()
  registryVarName.Add("hello", fnhellohello0)
}

func fnhellohello0(t parse.Templater, w io.Writer, indata interface {
}) error {
  return nil
}`,
		},
	}

	for i, testData := range allTestData {
		var tplsTree map[string]*parse.Tree
		if testData.isHTML {
			tplsTree = mustCompileHTMLTemplate(testData.tplName, testData.tplStr, testData.funcsMap)
		} else {
			tplsTree = mustCompileTextTemplate(testData.tplName, testData.tplStr, testData.funcsMap)
		}
		preparedTemplates := PrepareTemplate(
			testData.tplName,
			tplsTree,
		)
		program, err := Compile(
			preparedTemplates,
			testData.programPkg,
			testData.varName,
			testData.dataPkg,
			testData.dataType,
			testData.dataValue,
			testData.funcsMap,
			testData.funcsMapPublic,
		)
		if err != nil {
			t.Errorf("Test(%v): got err=%v", i, err)
			break
		}
		program = formatGoCode(program)
		expectedProgram := formatGoCode(testData.expectedProgram)
		if err := compare(program, expectedProgram); err != nil {
			fmt.Println(program)
			t.Errorf("Test(%v): got err=%v", i, err)
			break
		}
	}
}

func mustCompileTextTemplate(name string, tplStr string, funcsMap map[string]interface{}) map[string]*parse.Tree {
	ret := map[string]*parse.Tree{}

	t, err := text.New(name).Funcs(funcsMap).Parse(tplStr)
	if err != nil {
		panic(err)
	}
	t.Execute(ioutil.Discard, nil) // ignore err, it is just to force parse.

	for _, tpl := range t.Templates() {
		if tpl.Tree != nil {
			ret[tpl.Name()] = tpl.Tree
		}
	}

	return ret
}

func mustCompileHTMLTemplate(name string, tplStr string, funcsMap map[string]interface{}) map[string]*parse.Tree {
	ret := map[string]*parse.Tree{}

	t, err := html.New(name).Funcs(funcsMap).Parse(tplStr)
	if err != nil {
		panic(err)
	}
	t.Execute(ioutil.Discard, nil) // ignore err, it is just to force parse.

	for _, tpl := range t.Templates() {
		if tpl.Tree != nil {
			ret[tpl.Name()] = tpl.Tree
		}
	}

	return ret
}
