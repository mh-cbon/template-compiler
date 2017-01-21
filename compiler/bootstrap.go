package compiler

import (
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"path/filepath"
	"strings"

	"github.com/mh-cbon/export-funcmap/export"
)

// this is for the tests only.
var emptyFunc map[string]interface{}

// GenerateProgramBootstrap generates the bootstrap program that handles
// the compilation of the templates.
func GenerateProgramBootstrap(
	tplsPath []string,
	outPath, varName string,
	isHTML bool,
	data string,
	funcExports ...string,
) (string, error) {

	dataPkg, err := pkgathFromDataSelector(data)
	if err != nil {
		return "", err
	}
	dataType, err := dataTypeFromDataSelector(data)
	if err != nil {
		return "", err
	}
	if isDataTypeExported(dataType) == false {
		return "", fmt.Errorf(
			"Unexpected private data type. The data type must be exported, got <%v>",
			dataType,
		)
	}
	dataValueInit, err := dataValueInitExpr(dataPkg, dataType)
	if err != nil {
		return "", err
	}

	targets := export.Targets{}
	if err = targets.Parse(funcExports); err != nil {
		return "", err
	}

	resFile, err := export.Export(targets, "gen.go", "main", "funcsMap")
	if err != nil {
		return "", err
	}

	funcsMapDecl := export.GetVarDecl(resFile, "funcsMap")
	publicIdentsDecl := export.GetVarDecl(resFile, "funcsMapPublic")
	importDecl := export.MustGetImportDecl(resFile)
	importDecl.Lparen = token.Pos(1)

	export.InjectImportPaths([]string{
		"io/ioutil",
		"os",
		"fmt",
		"github.com/mh-cbon/template-compiler/compiler",
		dataPkg, //todo: check uniqueness
	}, importDecl)

	programMain := fmt.Sprintf(`package main

%v

var tplsPath = %#v

var outPath = %#v

var varName = %#v

var dataPkg = %#v

var dataType = %#v

var dataValue = %v

var isHTML = %#v

%v

%v

func main () {
	pkgName, err := compiler.LookupPackageName(outPath)
	if err != nil {
    panic(fmt.Errorf("Failed to lookup for the package name: %%v", err))
	}
  tpls, err := compiler.PrepareTemplates(tplsPath, isHTML, funcsMap)
  if err != nil {
    panic(fmt.Errorf("Failed to prepare the templates: %%v", err))
  }
  program, err := compiler.Compile(tpls, pkgName, varName, dataPkg, dataType, dataValue, funcsMap, funcsMapPublic)
  if err != nil {
    panic(fmt.Errorf("Failed to compile the templates: %%v", err))
  }
  if err := ioutil.WriteFile(outPath, []byte(program), os.ModePerm); err != nil {
    panic(fmt.Errorf("Failed to write the compiled templates: %%v", err))
  }
}
`,
		astNodeToString(importDecl),
		tplsPath,
		outPath,
		varName,
		dataPkg,
		dataType,
		dataValueInit,
		isHTML,
		astNodeToString(funcsMapDecl),
		astNodeToString(publicIdentsDecl),
	)

	return formatGoCode(programMain), nil
}

func formatGoCode(s string) string {
	fmtExpected, err := format.Source([]byte(s))
	if err != nil {
		panic(err)
	}
	return string(fmtExpected)
}

func pkgathFromDataSelector(pkgTypeSelector string) (string, error) {
	dataArr := strings.Split(pkgTypeSelector, ":")
	if len(dataArr) != 2 {
		return "", fmt.Errorf(
			"unexpected data type format. Expected <pkg path:type name>, got <%v>",
			pkgTypeSelector,
		)
	}
	return dataArr[0], nil
}

func dataTypeFromDataSelector(pkgTypeSelector string) (string, error) {
	dataArr := strings.Split(pkgTypeSelector, ":")
	if len(dataArr) != 2 {
		return "", fmt.Errorf(
			"unexpected data type format. Expected <pkg path:type name>, got <%v>",
			pkgTypeSelector,
		)
	}
	return dataArr[1], nil
}

func isDataTypeExported(dataType string) bool {
	if len(dataType) > 0 {
		isPtr := dataType[0] == []byte("*")[0]
		isPtrOrStuct := dataType[0] == []byte("?")[0] // tbd later.
		if isPtr || isPtrOrStuct {
			dataType = dataType[1:]
		}
		return ast.IsExported(dataType)
	}
	return false
}

func dataValueInitExpr(dataPkgPath string, dataType string) (string, error) {
	if len(dataType) > 0 {
		dataPtr := ""
		if dataType[0] == []byte("*")[0] {
			dataPtr = "&"
		}
		dataPkgPath = filepath.Base(dataPkgPath)
		return fmt.Sprintf("%v%v.%v{}", dataPtr, dataPkgPath, dataType), nil
	}
	return "", fmt.Errorf(
		"unexpected empty data type format. got <%v>",
		dataType,
	)
}
