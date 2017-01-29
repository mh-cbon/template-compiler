package compiler

import (
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"path/filepath"
	"strconv"

	"github.com/mh-cbon/export-funcmap/export"
)

// this is only for the tests.
var emptyFunc map[string]interface{}

// GenerateProgramBootstrap generates the bootstrap program that handles
// the compilation of the templates from the source file imports
// and the configuration variable declaration ast.
func GenerateProgramBootstrap(
	importsContext []*ast.ImportSpec,
	confNode *ast.GenDecl,
	varName string,
) (string, error) {

	programImport := export.NewImportDecl()
	programImport.Lparen = token.Pos(1)
	export.InjectImportPaths([]string{
		"github.com/mh-cbon/template-compiler/compiler",
		"fmt",
	}, programImport)

	newImports, err := prepareConfiguration(importsContext, confNode)
	if err != nil {
		return "", err
	}
	for _, i := range newImports {
		programImport.Specs = append(programImport.Specs, i)
	}

	return makeProgram(programImport, confNode, varName), nil
}

// GenerateProgramBootstrapFromAstFile generates the bootstrap program that handles
// the compilation of the templates from a parsed go source.
func GenerateProgramBootstrapFromAstFile(
	parsedFile *ast.File,
	varName string,
) (string, error) {

	var configurationVar *ast.GenDecl
	configurationVar = extractVar(parsedFile, varName)
	if configurationVar == nil {
		return "", fmt.Errorf("Configuration variable %v not found in %v", varName, parsedFile.Name.String())
	}

	importsContext := extractImports(parsedFile)

	return GenerateProgramBootstrap(importsContext, configurationVar, varName)
}

// GenerateProgramBootstrapFromFile generates the bootstrap program that handles
// the compilation of the templates from a file path.
func GenerateProgramBootstrapFromFile(
	file string,
	varName string,
) (string, error) {

	parsedFile, err := parseGoFilePath(file)
	if err != nil {
		return "", err
	}

	return GenerateProgramBootstrapFromAstFile(parsedFile, varName)
}

// GenerateProgramBootstrapFromString generates the bootstrap program that handles
// the compilation of the templates from a go code string.
func GenerateProgramBootstrapFromString(
	content string,
	varName string,
) (string, error) {

	parsedFile, err := parseGoString(content)
	if err != nil {
		return "", err
	}

	return GenerateProgramBootstrapFromAstFile(parsedFile, varName)
}

func makeProgram(imports *ast.GenDecl, confNode *ast.GenDecl, varName string) string {
	programMain := fmt.Sprintf(`package main

%v

%v

func main () {
  compiler := compiler.NewCompiledTemplatesProgram(%q)
  if err := compiler.CompileAndWrite(%v); err != nil {
    panic(fmt.Errorf("Failed to compile the templates: %%v", err))
  }
}
`,
		astNodeToString(imports),
		astNodeToString(confNode),
		varName,
		varName,
	)

	return formatGoCode(programMain)
}

// prepareConfiguration takes the configuration ast.Node and completes it
func prepareConfiguration(importsContext []*ast.ImportSpec, confNode *ast.GenDecl) ([]*ast.ImportSpec, error) {
	newImports := []*ast.ImportSpec{}

	compiledNew := confNode.Specs[0].(*ast.ValueSpec).Values[0].(*ast.CallExpr)
	// need to check the Fun Call is compiled.New,
	// it might not be if the callexpr compiled.New is followed
	// byt a SetPkg(...) call,
	// in such case callExpr.Fun.X is not an ast.ident.
	if _, ok := compiledNew.Fun.(*ast.SelectorExpr).X.(*ast.Ident); ok == false {
		// not compiled.New
		compiledNew = compiledNew.Fun.(*ast.SelectorExpr).X.(*ast.CallExpr)
	}

	// in the original program, we know compiled package is imported,
	// but it may use an alias, lets grab this now.
	compiledAlias := compiledNew.Fun.(*ast.SelectorExpr).X.(*ast.Ident).Name
	if compiledAlias == "compiled" {
		compiledAlias = ""
	}
	s := export.NewImportSpec("github.com/mh-cbon/template-compiler/compiled", compiledAlias)
	newImports = append(newImports, s)

	// in New(outpath string, templates []TemplateConfiguration, funcsmap ...string) *Configuration
	// find funcsmap arguments and add them to allTemplatesFuncs
	allTemplatesFuncs := []string{}
	if len(compiledNew.Args) > 2 {
		for _, a := range compiledNew.Args[2:] {
			arg := a.(*ast.BasicLit)
			value := arg.Value[1 : len(arg.Value)-1] // remove quotes
			if value != "" {
				allTemplatesFuncs = append(allTemplatesFuncs, value)
			}
		}
	}

	// ensure the template configuration contains text template std funcs
	if containsStr(allTemplatesFuncs, "text/template:builtins") == false {
		allTemplatesFuncs = append(allTemplatesFuncs, "text/template:builtins")
	}

	// ensure the template configuration contains template-tree-simplifier funcs
	if containsStr(allTemplatesFuncs, "github.com/mh-cbon/template-tree-simplifier/funcmap:tplFunc") == false {
		allTemplatesFuncs = append(allTemplatesFuncs, "github.com/mh-cbon/template-tree-simplifier/funcmap:tplFunc")
	}

	// in New(outpath string, templates []TemplateConfiguration, funcsmap ...string) *Configuration
	// find templates argument, then look for each compiled.TemplateConfiguration{},
	// then browse each key / value expression,
	// - for an HTML key, check if it says true/false
	// - for a Data key, creates and a DataConfiguration{}
	// - for a FuncsMap key, exports them to their symbolic version, and their public idents,
	//   add those new data to the configuration of the template.
	templatesConf := compiledNew.Args[1].(*ast.CompositeLit)

	for _, t := range templatesConf.Elts {
		templateConf := t.(*ast.CompositeLit)

		// manage HTML key
		isHTML := isAnHTMLTemplateConf(templateConf)

		// manage Data key
		if dataExpr, isPtr := getDataKeySelExpr(templateConf); dataExpr != nil {
			pkgPath, newKv := makeDataConfigurationKeyValue(importsContext, dataExpr, isPtr)
			templateConf.Elts = append(templateConf.Elts, newKv)
			newImports = append(newImports, pkgPath)
		} else {
			panic(
				fmt.Errorf("Expected template configuration to contain a Data key"),
			)
		}

		// manage FuncsMap key
		// todo, handle the nil case.
		var varToExport []string
		if funcsMapKey := getKeyValue(templateConf, "FuncsMap"); funcsMapKey != nil {
			varToExport = getFuncsMapKeyValues(funcsMapKey.Value.(*ast.CompositeLit))
		}
		varToExport = append(varToExport, allTemplatesFuncs...)

		// ensure the template configuration contains html template std funcs
		if isHTML {
			if containsStr(varToExport, "github.com/mh-cbon/template-compiler/std/html/template:publicFuncMap") == false {
				varToExport = append(varToExport, "github.com/mh-cbon/template-compiler/std/html/template:publicFuncMap")
			}
		}

		if len(varToExport) > 0 {
			funcsMapValue, publicIdentValue, imports, err := exportFuncsMap(varToExport)
			if err != nil {
				return newImports, err
			}
			kvFuncsExport := &ast.KeyValueExpr{
				Key:   &ast.Ident{Name: "FuncsExport"},
				Value: funcsMapValue,
			}
			templateConf.Elts = append(templateConf.Elts, kvFuncsExport)
			kvPublicIdents := &ast.KeyValueExpr{
				Key:   &ast.Ident{Name: "PublicIdents"},
				Value: publicIdentValue,
			}
			templateConf.Elts = append(templateConf.Elts, kvPublicIdents)
			newImports = append(newImports, imports...)
		}

	}

	return newImports, nil
}

// isAnHTMLTemplateConf tells if a TemplateConfiguration contains an HTML key and its value is true
func isAnHTMLTemplateConf(templateConf *ast.CompositeLit) bool {
	isHTML := false
	if isHTMLKey := getKeyValue(templateConf, "HTML"); isHTMLKey != nil {
		if x, ok := isHTMLKey.Value.(*ast.Ident); ok == false {
			panic(fmt.Errorf("Unexpected node\n%v\n%#v", isHTMLKey.Value, isHTMLKey.Value))
		} else {
			isHTML = x.Name == "true"
		}
	}
	return isHTML
}

// from a ast.CompositeLit such
// TypeOfData{Key: value}, extract the key/value
// matching search
func getKeyValue(templateConf *ast.CompositeLit, search string) *ast.KeyValueExpr {
	for _, keyValue := range templateConf.Elts {
		keyValueExpr := keyValue.(*ast.KeyValueExpr)
		if ident, ok := keyValueExpr.Key.(*ast.Ident); ok {
			if ident.Name == search {
				return keyValueExpr
			}
		}
	}
	return nil
}

// transforms the ast.Node of a value
// []string{"",""...} into a slice of string values.
func getFuncsMapKeyValues(value *ast.CompositeLit) []string {
	var varToExport []string
	for _, v := range value.Elts {
		value := v.(*ast.BasicLit).Value
		value = value[1 : len(value)-1]
		if value != "" {
			varToExport = append(varToExport, value)
		}
	}
	return varToExport
}

func exportFuncsMap(funcExports []string) (*ast.CompositeLit, *ast.CompositeLit, []*ast.ImportSpec, error) {

	imports := []*ast.ImportSpec{}
	targets := export.Targets{}
	if err := targets.Parse(funcExports); err != nil {
		return nil, nil, nil, err
	}

	resFile, err := export.Export(targets, "gen.go", "main", "funcsMap")
	if err != nil {
		return nil, nil, nil, err
	}

	funcsMapDecl := export.GetVarDecl(resFile, "funcsMap")
	publicIdentsDecl := export.GetVarDecl(resFile, "funcsMapPublic")
	importDecl := export.MustGetImportDecl(resFile)

	funcsMapValue := funcsMapDecl.Specs[0].(*ast.ValueSpec).Values[0].(*ast.CompositeLit)
	publicIdentValue := publicIdentsDecl.Specs[0].(*ast.ValueSpec).Values[0].(*ast.CompositeLit)

	for _, i := range importDecl.Specs {
		if ii, ok := i.(*ast.ImportSpec); ok {
			imports = append(imports, ii)
		}
	}

	return funcsMapValue, publicIdentValue, imports, nil
}

func getDataKeySelExpr(templateConf *ast.CompositeLit) (*ast.SelectorExpr, bool) {
	var ret *ast.SelectorExpr
	isPtr := false
	if dataKey := getKeyValue(templateConf, "Data"); dataKey != nil {
		switch valueType := dataKey.Value.(type) {
		case *ast.CompositeLit:
			ret = valueType.Type.(*ast.SelectorExpr)
		case *ast.UnaryExpr:
			ret = valueType.X.(*ast.CompositeLit).Type.(*ast.SelectorExpr)
			isPtr = true
		}
	}
	return ret, isPtr
}

// creates a DataConfiguration instance of a data instanciation ast.Node (yourpkg.DataType{})
func makeDataConfigurationKeyValue(importsContext []*ast.ImportSpec, expr *ast.SelectorExpr, isPtr bool) (*ast.ImportSpec, *ast.KeyValueExpr) {
	dataImportSpec := getPkgPath(importsContext, expr.X.(*ast.Ident).Name)
	pkgPath := dataImportSpec.Path.Value
	pkgPath = pkgPath[1 : len(pkgPath)-1]
	pkg := filepath.Base(pkgPath)
	gocode := `package xx
  var x = compiled.DataConfiguration{
      IsPtr: ` + strconv.FormatBool(isPtr) + `,
      DataTypeName: "` + expr.Sel.Name + `",
      DataType: "` + pkg + `.` + expr.Sel.Name + `",
      PkgPath: "` + pkgPath + `",
  }`
	newNode := stringToAst(gocode)
	return dataImportSpec, &ast.KeyValueExpr{
		Key:   &ast.Ident{Name: "DataConfiguration"},
		Value: newNode.Decls[0].(*ast.GenDecl).Specs[0].(*ast.ValueSpec).Values[0],
	}
}

func getPkgPath(importsContext []*ast.ImportSpec, name string) *ast.ImportSpec {
	for _, i := range importsContext {
		if i.Name != nil && i.Name.Name == name {
			return i
		}
	}
	for _, i := range importsContext {
		pkgPath := i.Path.Value[1 : len(i.Path.Value)-1] // remove quotes
		if filepath.Base(pkgPath) == name {
			return i
		}
	}
	return nil
}

func strIndex(list []string, search string) int {
	for i, l := range list {
		if l == search {
			return i
		}
	}
	return -1
}
func containsStr(list []string, search string) bool {
	return strIndex(list, search) > -1
}

func parseGoFilePath(file string) (*ast.File, error) {
	fset := token.NewFileSet()
	return parser.ParseFile(fset, file, nil, 0)
}

func parseGoString(content string) (*ast.File, error) {
	fset := token.NewFileSet()
	return parser.ParseFile(fset, "whatever.go", content, 0)
}

// given a go file ast.Node, extracts a (top-level) variable declaration.
func extractVar(file *ast.File, varName string) *ast.GenDecl {
	var found *ast.GenDecl
	ast.Inspect(file, func(n ast.Node) bool {
		if d, ok := n.(*ast.GenDecl); ok {
			if v, okk := d.Specs[0].(*ast.ValueSpec); okk {
				if len(v.Names) > 0 && v.Names[0].Name == varName {
					found = d
				}
			}
		}
		return found == nil
	})

	return found
}

// given a go file ast.Node, extract its import statements.
func extractImports(file *ast.File) []*ast.ImportSpec {
	var found []*ast.ImportSpec
	ast.Inspect(file, func(n ast.Node) bool {
		if d, ok := n.(*ast.GenDecl); ok {
			for _, s := range d.Specs {
				if v, okk := s.(*ast.ImportSpec); okk {
					found = append(found, v)
				}
			}
		}
		return true
	})

	return found
}

func formatGoCode(s string) string {
	fmtExpected, err := format.Source([]byte(s))
	if err != nil {
		panic(err)
	}
	return string(fmtExpected)
}
