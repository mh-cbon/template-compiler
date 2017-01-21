package compiler

import (
	"bytes"
	"fmt"
	"go/format"
	"go/token"
	html "html/template"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	text "text/template"
	"text/template/parse"

	"github.com/mh-cbon/template-tree-simplifier/simplifier"
	"github.com/serenize/snaker"
)

// Compile convert the templates tree into go code
// and generates the resulting go file.
func Compile(
	tpls []*TemplateToCompile,
	programPkg string,
	varName string,
	dataPkg string,
	dataType string,
	dataValue interface{},
	funcsMap map[string]interface{},
	funcsMapPublic []map[string]string,
) (string, error) {

	knownIdents := gatherIdents(varName, tpls)
	dataPkgIdent, dataPkgAlias := uncollidePkg(dataPkg, knownIdents)
	dataPkg = fmt.Sprintf("%v %q", dataPkgAlias, dataPkg) // must include the alias
	dataQualifier := fmt.Sprintf("%v.%v", dataPkgIdent, dataType)
	if dataType[0] == []byte("*")[0] {
		dataQualifier = fmt.Sprintf("*%v.%v", dataPkgIdent, dataType[1:])
	}

	builtinTexts := map[string]string{}
	fnContent := ""
	var allAdditionnalImports []string
	for _, tpl := range tpls {
		typeCheck := simplifier.TransformTree(tpl.tree, dataValue, funcsMap)

		astTree, additionalImports := convertTplTree(
			tpl.tree,
			typeCheck,
			builtinTexts,
			dataQualifier,
			funcsMap,
			funcsMapPublic,
			tpl.fnname,
		)

		allAdditionnalImports = append(allAdditionnalImports, additionalImports...)

		var b bytes.Buffer
		fset := token.NewFileSet()
		format.Node(&b, fset, astTree)

		fnContent += b.String() + "\n\n"
	}
	program := generateBaseProgram(programPkg, dataPkg, varName, tpls, allAdditionnalImports)
	program += fnContent

	// add the builtin text values
	for text, name := range builtinTexts {
		program += fmt.Sprintf("var %v = []byte(%q)\n", name, text)
	}

	// fmt.Println(program)

	return program, nil
}

func generateBaseProgram(
	pkgName, dataPkg, varName string,
	tpls []*TemplateToCompile,
	allAdditionnalImports []string,
) string {
	program := fmt.Sprintf("package %v\n\n", pkgName)
	program += fmt.Sprintf("//golint:ignore\n\n")
	program += fmt.Sprintf("import (\n")
	program += fmt.Sprintln(` "io"`)
	program += fmt.Sprintln(` "github.com/mh-cbon/template-compiler/compiled"`)
	program += fmt.Sprintln(` "github.com/mh-cbon/template-compiler/std/text/template/parse"`)
	for _, i := range allAdditionnalImports {
		if i != "io" &&
			i != "github.com/mh-cbon/template-compiler/compiled" &&
			i != "github.com/mh-cbon/template-compiler/std/text/template/parse" {
			program += fmt.Sprintf(" %q\n", i)
		}
	}
	program += fmt.Sprintln(` ` + dataPkg)
	program += fmt.Sprintf(")\n\n")
	program += fmt.Sprintf("func init () {\n")
	program += fmt.Sprintf("  %v = compiled.NewRegistry()\n", varName)
	for _, t := range tpls {
		program += fmt.Sprintf("  %v.Add(%#v, %v)\n", varName, t.name, t.fnname)
	}
	for i, t := range tpls {
		for e, a := range t.definedTemplateNames {
			varX := fmt.Sprintf("tpl%vX%v", i, e)
			varY := fmt.Sprintf("tpl%vY%v", i, e)
			program += fmt.Sprintf("  %v := %v.MustGet(%#v)\n", varX, varName, t.name)
			program += fmt.Sprintf("  %v := %v.MustGet(%#v)\n", varY, varName, a)
			program += fmt.Sprintf("  %v, _ = %v.Compiled(%v)\n", varX, varX, varY)
			program += fmt.Sprintf("  %v.Set(%#v, %v)\n", varName, t.name, varX)
		}
	}
	program += fmt.Sprintf("}\n\n")
	return program
}

func gatherIdents(varName string, tpls []*TemplateToCompile) []string {
	knownIdents := []string{"parse", "t", "w", "io", "data", "indata", varName}
	for _, tpl := range tpls {
		knownIdents = append(knownIdents, tpl.fnname)
	}
	return knownIdents
}

func uncollidePkg(pkgPath string, knownIdents []string) (string, string) {
	pkgIdent := filepath.Base(pkgPath)
	pkgAlias := ""
	for _, ident := range knownIdents {
		if ident == pkgIdent {
			pkgAlias = pkgIdent + "alias"
			break
		}
	}
	if pkgAlias != "" {
		pkgIdent = pkgAlias
	}
	return pkgIdent, pkgAlias
}

// PrepareTemplates parses and compiles template files.
func PrepareTemplates(
	tplsPath []string,
	isHTML bool,
	funcsMap map[string]interface{},
) ([]*TemplateToCompile, error) {

	var tpls []*TemplateToCompile
	for _, tplPath := range tplsPath {
		var treeNames map[string]*parse.Tree
		var err error
		if isHTML {
			treeNames, err = compileHTMLTemplate(tplPath, funcsMap)
		} else {
			treeNames, err = compileTextTemplate(tplPath, funcsMap)
		}
		if err != nil {
			return tpls, err
		}
		mainName := filepath.Base(tplPath)
		tpls = append(tpls, PrepareTemplate(mainName, treeNames)...)
	}
	return tpls, nil
}

// PrepareTemplate ...
func PrepareTemplate(
	name string,
	trees map[string]*parse.Tree,
) []*TemplateToCompile {
	var ret []*TemplateToCompile
	i := 0
	for treeName, tree := range trees {
		tplFnname := cleanTplName("fn" + name + treeName + "_" + strconv.Itoa(i))
		ret = append(ret, &TemplateToCompile{
			treeName,
			snakeToCamel(tplFnname),
			tree,
			[]string{},
		})
		i++
	}
	associateTemplates(name, ret)
	return ret
}

// TemplateToCompile is a struct which can hold
// test/template or html/template.
// it holds
// - fnname the resulting compiled function for a template
// - name, the template named
// - tree, the result of template parsing
// - definedTemplateNames, the list of defined templates
//    with the define instruction in a file template.
type TemplateToCompile struct {
	name                 string
	fnname               string
	tree                 *parse.Tree
	definedTemplateNames []string
}

// compileTextTemplate comiles a file template as a text/template.
func compileTextTemplate(tplPath string, funcsMap map[string]interface{}) (map[string]*parse.Tree, error) {
	ret := map[string]*parse.Tree{}

	t, err := text.New("").Funcs(funcsMap).ParseFiles(tplPath)
	if err != nil {
		return ret, err
	}
	t.Execute(ioutil.Discard, nil) // ignore err, it is just to force parse.

	for _, tpl := range t.Templates() {
		if tpl.Tree != nil {
			ret[tpl.Name()] = tpl.Tree
		}
	}

	return ret, nil
}

// compileHTMLTemplate comiles a file template as an html/template.
func compileHTMLTemplate(tplPath string, funcsMap map[string]interface{}) (map[string]*parse.Tree, error) {
	ret := map[string]*parse.Tree{}

	t, err := html.New("").Funcs(funcsMap).ParseFiles(tplPath)
	if err != nil {
		return ret, err
	}
	t.Execute(ioutil.Discard, nil) // ignore err, it is just to force parse.

	for _, tpl := range t.Templates() {
		if tpl.Tree != nil {
			ret[tpl.Name()] = tpl.Tree
		}
	}

	return ret, nil
}

func associateTemplates(mainTemplateName string, templatesToCompile []*TemplateToCompile) {
	var mainTemplate *TemplateToCompile
	for _, r := range templatesToCompile {
		if r.name == mainTemplateName {
			mainTemplate = r
			break
		}
	}
	for _, r := range templatesToCompile {
		if r.name != mainTemplateName {
			mainTemplate.definedTemplateNames = append(mainTemplate.definedTemplateNames, r.name)
		}
	}
}

// LookupPackageName search a directory for its delcaring package.
func LookupPackageName(someDir string) (string, error) {
	dir := filepath.Dir(someDir)
	// the dir must exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return "", err
	}
	files, err := filepath.Glob(filepath.Dir(someDir) + "/*.go")
	if err != nil {
		return "", err
	}
	if len(files) == 0 {
		// maybe it is pointing to an empty dir
		dir := filepath.Dir(someDir)
		// the package name will be basename of dir
		return filepath.Base(dir), nil
	}
	f := files[0]
	b, err := ioutil.ReadFile(f)
	if err != nil {
		return "", err
	}
	bs := string(b)
	// improve this. really q&d.
	bs = bs[strings.Index(bs, "package"):]
	bs = bs[0:strings.Index(bs, "\n")]
	return strings.Split(bs, "package ")[1], nil
}

func snakeToCamel(s string) string {
	s = snaker.SnakeToCamel(s)
	if len(s) > 0 {
		s = strings.ToLower(s[:1]) + s[1:]
	}
	return s
}

func cleanTplName(name string) string {
	return replaceAll(name, []string{"."}, "_")
}

func replaceAll(old string, removals []string, replacement string) string {
	for _, r := range removals {
		old = strings.Replace(old, r, replacement, -1)
	}
	return old
}
