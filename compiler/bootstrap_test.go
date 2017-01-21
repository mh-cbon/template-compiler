package compiler

import (
	"fmt"
	"testing"
)

type BootstrapTestData struct {
	// the templates file path to compile
	tplsPath []string
	// the path to the output path containing the compiled template
	outPath string
	// the variable name of the compiled.Registry
	varName string
	// the type of template compilation
	isHTML bool
	// the data type selector such pkgPath:type
	data string
	// the funcmap selectors to export
	// such text/template:builtins
	funcExports []string
	// the expected program output
	expectedProgram string
	// will it fail
	expectErr bool
	// the err message
	expectedErr string
	// notes, it is not possible to check for
	// funcExports, because it is exported as a map,
	// a map does not guarantee order of keys,
	// so the same result can produce different order...
}

func TestBootstrap(t *testing.T) {

	allTestData := []BootstrapTestData{
		BootstrapTestData{
			tplsPath:    []string{"a.tpl", "b.tpl"},
			outPath:     "program.go",
			varName:     "registryVarName",
			isHTML:      true,
			data:        "pkg/path:type",
			funcExports: []string{"text/template:builtins"},
			expectErr:   true,
			expectedErr: "Unexpected private data type. The data type must be exported, got <type>",
		},
		BootstrapTestData{
			tplsPath:    []string{"a.tpl", "b.tpl"},
			outPath:     "program.go",
			varName:     "registryVarName",
			isHTML:      true,
			data:        "pkg/path:Type",
			funcExports: []string{"github.com/mh-cbon/template-compiler/compiler:emptyFunc"},
			expectedProgram: `package main

import (
  "fmt"
  "github.com/mh-cbon/template-compiler/compiler"
  "io/ioutil"
  "os"
  "pkg/path"
)

var tplsPath = []string{"a.tpl", "b.tpl"}

var outPath = "program.go"

var varName = "registryVarName"

var dataPkg = "pkg/path"

var dataType = "Type"

var dataValue = path.Type{}

var isHTML = true

var funcsMap = map[string]interface {
}{}

var funcsMapPublic []map[string]string = []map[string]string(nil)

func main() {
  pkgName, err := compiler.LookupPackageName(outPath)
  if err != nil {
    panic(fmt.Errorf("Failed to lookup for the package name: %v", err))
  }
  tpls, err := compiler.PrepareTemplates(tplsPath, isHTML, funcsMap)
  if err != nil {
    panic(fmt.Errorf("Failed to prepare the templates: %v", err))
  }
  program, err := compiler.Compile(tpls, pkgName, varName, dataPkg, dataType, dataValue, funcsMap, funcsMapPublic)
  if err != nil {
    panic(fmt.Errorf("Failed to compile the templates: %v", err))
  }
  if err := ioutil.WriteFile(outPath, []byte(program), os.ModePerm); err != nil {
    panic(fmt.Errorf("Failed to write the compiled templates: %v", err))
  }
}`,
		},
	}

	for i, testData := range allTestData {
		program, err := GenerateProgramBootstrap(
			testData.tplsPath,
			testData.outPath,
			testData.varName,
			testData.isHTML,
			testData.data,
			testData.funcExports...,
		)
		if err != nil && !testData.expectErr {
			t.Errorf("Test(%v): got err=%v", i, err)
			break
		} else if testData.expectErr {
			if fmt.Sprint(err) != testData.expectedErr {
				t.Errorf(
					"Test(%v): Invalid error output. Expected\n%q\n\nGot\n%q\n",
					i,
					testData.expectedErr,
					err,
				)
			}
		} else {
			program = formatGoCode(program)
			expectedProgram := formatGoCode(testData.expectedProgram)
			if err := compare(program, expectedProgram); err != nil {
				fmt.Println(program)
				t.Errorf("Test(%v): got err=%v", i, err)
			}
			// if program != expectedProgram {
			// 	t.Errorf(
			// 		"Test(%v): Invalid program output. Expected\n%v\n\nGot\n%v\n",
			// 		i,
			// 		expectedProgram,
			// 		program,
			// 	)
			// }
		}
	}
}

func compare(s1, s2 string) error {
	line := 0
	leftContent := ""
	rightContent := ""
	for i, s := range s1 {
		leftContent += string(s)
		if s == rune('\n') {
			line++
		}
		if i >= len(s2) {
			return fmt.Errorf("content too small at line %v, line=%q", line, leftContent)
		}
		rightContent += string(s2[i])
		if rune(s2[i]) != s {
			return fmt.Errorf("invalid content at pos=%v line %v\nleft=%v\nright=%v", i, line, leftContent, rightContent)
		}
		if s == rune('\n') {
			leftContent = ""
		}
		if s2[i] == '\n' {
			rightContent = ""
		}
	}
	return nil
}
