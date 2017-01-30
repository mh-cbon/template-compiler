package compiler

import (
	"go/ast"
	"strings"
	"testing"
)

type BootstrapTestData struct {
	// the content of the source program
	srcProgram string
	// the name of the configuration variable in the source prgoram
	srcVar string
	// if it is expected to fail to compile, this error should contain the same message
	expectErr error
	// the list of expected imports to find in the program, noted as alias:path, or just path
	expectedImports []string
	// is the new var ast.node expect to be an Ident or a CallExpr
	expectedNewVarAsCall bool
	// the expected alias of compiled package, if any
	expectedCompiledAlias string
	// the expected value of the first argument
	expectedOutPath string
	// the expected length of template configuration values
	expectedLenOfTemplateConfig int
	// the expected length of shared funcmap
	expectedLenOfAllTemplatesFuncsMap int
	// self explanatory
	expectedToBeHTMLTemplates []bool
}

func TestBootstrap(t *testing.T) {
	allTestData := []BootstrapTestData{
		BootstrapTestData{
			srcProgram: `package yy

import (
	"github.com/mh-cbon/template-compiler/compiled"
	"github.com/mh-cbon/template-compiler/demo/data"
)

var compiled = compiled.New(
	"gen.go",
	[]compiled.TemplateConfiguration{
		compiled.TemplateConfiguration{
			TemplatesPath: "templates/*.tpl",
			TemplatesData: map[string]interface{}{
				"*": data.MyTemplateData{},
			},
		},
	},
)`,
			srcVar: `compiled`,
			expectedImports: []string{
				"fmt",
				"github.com/mh-cbon/template-compiler/compiled",
				"github.com/mh-cbon/template-compiler/compiler",
				"github.com/mh-cbon/template-compiler/demo/data",
			},
			expectedOutPath:             "gen.go",
			expectedLenOfTemplateConfig: 1,
			expectedToBeHTMLTemplates:   []bool{false},
		},
		BootstrapTestData{
			srcProgram: `package yy

import (
	"github.com/mh-cbon/template-compiler/compiled"
	aliasdata "github.com/mh-cbon/template-compiler/demo/data"
)

var compiled = compiled.New(
	"gen.go",
	[]compiled.TemplateConfiguration{
		compiled.TemplateConfiguration{
			HTML:          true,
			TemplatesPath: "templates/*.tpl",
			TemplatesData: map[string]interface{}{
				"*": aliasdata.MyTemplateData{},
			},
			FuncsMap:      []string{"github.com/mh-cbon/template-compiler/compiler:emptyFunc"},
		},
	},
).SetPkg("main")`,
			srcVar: `compiled`,
			expectedImports: []string{
				"fmt",
				"github.com/mh-cbon/template-compiler/compiled",
				"github.com/mh-cbon/template-compiler/compiler",
				"aliasdata:github.com/mh-cbon/template-compiler/demo/data",
			},
			expectedNewVarAsCall:        true,
			expectedOutPath:             "gen.go",
			expectedLenOfTemplateConfig: 1,
			expectedToBeHTMLTemplates:   []bool{true},
		},
		BootstrapTestData{
			srcProgram: `package yy

import (
	tomate "github.com/mh-cbon/template-compiler/compiled"
	"github.com/mh-cbon/template-compiler/demo/data"
	aliasdata "github.com/mh-cbon/template-compiler/demo/data"
)

var compiled = tomate.New(
	"somethingelse.go",
	[]compiled.TemplateConfiguration{
		compiled.TemplateConfiguration{
			HTML:          true,
			TemplatesPath: "templates/*.tpl",
			TemplatesData: map[string]interface{}{
				"*": data.MyTemplateData{},
				"*": data.MyTemplateData{},
			},
			FuncsMap:      []string{},
		},
		compiled.TemplateConfiguration{
			TemplatesPath: "templates/*.tpl",
			TemplatesData: map[string]interface{}{
				"*": aliasdata.MyTemplateData{},
				"*": aliasdata.MyTemplateData{},
			},
			FuncsMap:      []string{},
		},
	},
  "github.com/mh-cbon/template-compiler/compiler:emptyFunc",
)`,
			srcVar: `compiled`,
			expectedImports: []string{
				"fmt",
				"tomate:github.com/mh-cbon/template-compiler/compiled",
				"github.com/mh-cbon/template-compiler/compiler",
				"github.com/mh-cbon/template-compiler/demo/data",
				"aliasdata:github.com/mh-cbon/template-compiler/demo/data",
			},
			expectedCompiledAlias:       "tomate",
			expectedOutPath:             "somethingelse.go",
			expectedLenOfTemplateConfig: 2,
			expectedToBeHTMLTemplates:   []bool{true, false},
		},
	}

	for i, testData := range allTestData {
		program, err := GenerateProgramBootstrapFromString(testData.srcProgram, testData.srcVar)

		if testData.expectErr != nil {
			if err == nil {
				t.Errorf("Test(%v): Expected to fail, but got err=%v", i, err)
				return
			} else if err.Error() != testData.expectErr.Error() {
				t.Errorf("Test(%v): Expected to fail with the message=%v but got err=%v", i, testData.expectErr, err)
				return
			}
		}

		if err != nil {
			t.Errorf("Test(%v): Expected to succeed, but got an error=%v", i, err)
			return
		}

		if program == "" {
			t.Errorf("Test(%v): Unexpected empty program", i)
			return
		}

		parsedProgram, err := parseGoString(program)
		if err != nil {
			t.Errorf("Test(%v): Failed to parse the program=%v", i, program)
			return
		}

		if parsedProgram.Name.Name != "main" {
			t.Errorf("Test(%v): Invalid package name=%v", i, parsedProgram.Name.Name)
			return
		}

		importSpecs := extractImports(parsedProgram)
		if len(testData.expectedImports) != len(importSpecs) {
			t.Errorf("Test(%v): Expected to get %v import statements, but found %v\n\n%v",
				i, len(testData.expectedImports), len(importSpecs), program)
			return
		}

		imports := convertImportsSpecs(importSpecs)
		for _, im := range imports {
			if containsStr(testData.expectedImports, im) == false {
				t.Errorf("Test(%v): Found unexpected import=%v\n\n%v", i, im, program)
				return
			}
		}

		for _, im := range testData.expectedImports {
			if containsStr(imports, im) == false {
				t.Errorf("Test(%v): Expected to find import=%v\n\n%v", i, im, program)
				return
			}
		}

		newConfVar := extractVar(parsedProgram, testData.srcVar)
		if newConfVar == nil {
			t.Errorf("Test(%v): Expected to find the configuration var=%v\n\n%v", i, testData.srcVar, program)
			return
		}

		var compileConf *ast.CallExpr
		rightHand := newConfVar.Specs[0].(*ast.ValueSpec).Values[0].(*ast.CallExpr).Fun.(*ast.SelectorExpr)
		if testData.expectedNewVarAsCall == false {
			if _, ok := rightHand.X.(*ast.Ident); ok == false {
				t.Errorf("Test(%v): Expected to find an Ident, but found=%T\n\n%v", i, rightHand.X, program)
				return
			}
			compileConf = newConfVar.Specs[0].(*ast.ValueSpec).Values[0].(*ast.CallExpr)
		} else {
			if _, ok := rightHand.X.(*ast.CallExpr); ok == false {
				t.Errorf("Test(%v): Expected to find a CallExpr, but found=%T\n\n%v", i, rightHand.X, program)
				return
			}
			compileConf = rightHand.X.(*ast.CallExpr)
		}

		compiledAlias := compileConf.Fun.(*ast.SelectorExpr).X.(*ast.Ident).Name
		if testData.expectedCompiledAlias == "" && compiledAlias != "compiled" {
			t.Errorf("Test(%v): Expected to find compiled alias as 'compiled', but found=%v\n\n%v", i, compiledAlias, program)
			return
		} else if testData.expectedCompiledAlias != "" && testData.expectedCompiledAlias != compiledAlias {
			t.Errorf("Test(%v): Expected to find compiled alias as '%v', but found=%v\n\n%v",
				i, testData.expectedCompiledAlias, compiledAlias, program)
			return
		}

		firstArg := compileConf.Args[0].(*ast.BasicLit).Value
		firstArg = firstArg[1 : len(firstArg)-1]
		if testData.expectedOutPath != firstArg {
			t.Errorf("Test(%v): Expected to find the first arg value=%v, but got=%v\n\n%v",
				i, testData.expectedOutPath, firstArg, program)
			return
		}

		templatesConf := compileConf.Args[1].(*ast.CompositeLit)
		if testData.expectedLenOfTemplateConfig != len(templatesConf.Elts) {
			t.Errorf("Test(%v): Expected a length of template configuration=%v, but got=%v\n\n%v",
				i, testData.expectedLenOfTemplateConfig, len(templatesConf.Elts), program)
			return
		}

		for e, y := range templatesConf.Elts {
			templateConf := y.(*ast.CompositeLit)
			templateConfStr := astNodeToString(templateConf)
			//-
			expectedHTML := testData.expectedToBeHTMLTemplates[e]
			gotHTML := isAnHTMLTemplateConf(templateConf)
			if expectedHTML != gotHTML {
				t.Errorf("Test(%v): Expected template configuration(%v) to be HTML=%v, but got=%v\n\n%v",
					i, e, expectedHTML, gotHTML, templateConfStr)
				return
			}
			//-
			if PublicIdents := getKeyValue(templateConf, "PublicIdents"); PublicIdents == nil {
				t.Errorf("Test(%v): Expected template configuration(%v) to contain %v key, but got=nil\n\n%v",
					i, e, "PublicIdents", templateConfStr)
				break
			} else {
				PublicIdentsStr := astNodeToString(PublicIdents)
				// lets do static check
				if strings.Index(PublicIdentsStr, `"Sel": "template.HTMLEscaper"`) == -1 {
					t.Errorf("Test(%v): Expected template data configuration(%v) to contain a Public ident for html func, but it was not found\n\n%v",
						i, e, PublicIdentsStr)
					return
				}
				if strings.Index(PublicIdentsStr, `"Sel": "funcmap.BrowsePropertyPath"`) == -1 {
					t.Errorf("Test(%v): Expected template data configuration(%v) to contain a Public ident for browsePropertyPath func, but it was not found\n\n%v",
						i, e, PublicIdentsStr)
					return
				}
				if expectedHTML {
					if strings.Index(PublicIdentsStr, `"Sel": "template.RcdataEscaper"`) == -1 {
						t.Errorf("Test(%v): Expected template data configuration(%v) to contain a Public ident for _html_template_rcdataescaper func, but it was not found\n\n%v",
							i, e, PublicIdentsStr)
						return
					}
				} else {
					if strings.Index(PublicIdentsStr, `"Sel": "template.RcdataEscaper"`) > -1 {
						t.Errorf("Test(%v): Expected template data configuration(%v) to NOT contain a Public ident for _html_template_rcdataescaper func, but it WAS found\n\n%v",
							i, e, PublicIdentsStr)
						return
					}
				}
			}
			//-
			if FuncsExport := getKeyValue(templateConf, "FuncsExport"); FuncsExport == nil {
				t.Errorf("Test(%v): Expected template configuration(%v) to contain %v key, but got=nil\n\n%v",
					i, e, "FuncsExport", templateConfStr)
				break
			} else {
				FuncsExportStr := astNodeToString(FuncsExport)
				// lets do static check
				if strings.Index(FuncsExportStr, `"html": func(args ...interface{}) string`) == -1 {
					t.Errorf("Test(%v): Expected template data configuration(%v) to contain a FuncExport for html func, but it was not found\n\n%v",
						i, e, FuncsExportStr)
					return
				}
				if strings.Index(FuncsExportStr, `"browsePropertyPath": func(some interface{}, propertypath string, args ...interface{}) interface{}`) == -1 {
					t.Errorf("Test(%v): Expected template data configuration(%v) to contain a FuncExport for browsePropertyPath func, but it was not found\n\n%v",
						i, e, FuncsExportStr)
					return
				}
				if expectedHTML {
					if strings.Index(FuncsExportStr, `"_html_template_urlnormalizer": func(args ...interface{}) string`) == -1 {
						t.Errorf("Test(%v): Expected template data configuration(%v) to contain a FuncExport for _html_template_urlnormalizer func, but it was not found\n\n%v",
							i, e, FuncsExportStr)
						return
					}
				} else {
					if strings.Index(FuncsExportStr, `"_html_template_urlnormalizer": func(args ...interface{}) string`) > -1 {
						t.Errorf("Test(%v): Expected template data configuration(%v) to NOT contain a FuncExport for _html_template_urlnormalizer func, but it WAS found\n\n%v",
							i, e, FuncsExportStr)
						return
					}
				}
			}
		}
	}
}
