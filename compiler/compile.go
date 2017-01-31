package compiler

import (
	"fmt"
	"go/ast"
	html "html/template"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	text "text/template"
	"text/template/parse"

	"github.com/mh-cbon/template-compiler/compiled"
	"github.com/mh-cbon/template-tree-simplifier/simplifier"
	"github.com/serenize/snaker"
)

// CompiledTemplatesProgram ...
type CompiledTemplatesProgram struct {
	varName      string
	imports      []*ast.ImportSpec
	funcs        []*ast.FuncDecl
	idents       []string
	builtinTexts map[string]string
}

// NewCompiledTemplatesProgram prepare a new instance.
// it automatically adds io and template/parse packages,
// and declares static idents.
func NewCompiledTemplatesProgram(varName string) *CompiledTemplatesProgram {
	ret := &CompiledTemplatesProgram{
		varName: varName,
		idents: []string{
			"t", "w", "data", "indata", varName,
		},
		builtinTexts: map[string]string{},
	}
	ret.addImport("io")
	ret.addImport("github.com/mh-cbon/template-compiler/std/text/template/parse")
	return ret
}

// CompileAndWrite the configuration and write the resulting program to config.OutPath.
func (c *CompiledTemplatesProgram) CompileAndWrite(config *compiled.Configuration) error {
	program, err := c.Compile(config)
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(config.OutPath, []byte(program), os.ModePerm); err != nil {
		return fmt.Errorf("Failed to write the compiled templates: %v", err)
	}
	return nil
}

//Compile the configuration, it returns a string of the output program.
func (c *CompiledTemplatesProgram) Compile(config *compiled.Configuration) (string, error) {
	if err := updateOutPkg(config); err != nil {
		return "", err
	}

	templatesToCompile, err := c.getTemplatesToCompile(config)
	if err != nil {
		return "", err
	}

	return c.compileTemplates(config.OutPkg, templatesToCompile)
}

//compileTemplates generates the output program for the given templates to compile.
func (c *CompiledTemplatesProgram) compileTemplates(outpkg string, templatesToCompile []*TemplateToCompile) (string, error) {
	if err := c.convertTemplates(templatesToCompile); err != nil {
		return "", err
	}
	return c.generateProgram(outpkg, templatesToCompile), nil
}

// convertTemplates convert each TemplateToCompile into functions.
func (c *CompiledTemplatesProgram) convertTemplates(templatesToCompile []*TemplateToCompile) error {
	for _, t := range templatesToCompile {
		for _, f := range t.files {
			for _, name := range f.names() {
				f.tplsFunc[name] = c.makeFuncName(f.tplsFunc[name])
				f.tplsFunc[name] = snakeToCamel(f.tplsFunc[name])

				dataConfig, err := t.getDataConfiguration(name)
				if err != nil {
					return err
				}

				err = convertTplTree(
					f.tplsFunc[name],
					f.tplsTree[name],
					t.FuncsExport,
					t.PublicIdents,
					dataConfig,
					f.tplsTypeCheck[name],
					c,
				)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// getTemplatesToCompile prepares the templates for the given configuration.
func (c *CompiledTemplatesProgram) getTemplatesToCompile(conf *compiled.Configuration) ([]*TemplateToCompile, error) {
	templatesToCompile := convertConfigToTemplatesToCompile(conf)
	for _, t := range templatesToCompile {
		if err := t.prepare(); err != nil {
			return templatesToCompile, err
		}
	}
	return templatesToCompile, nil
}

// updatedOutPkg ensure the configuration OutPkg is set.
// if OutPkg is empty, it tries to detect the package automatically
// by reading the files existing in the OutPath directory and extracting the package declaration.
func updateOutPkg(conf *compiled.Configuration) error {
	if conf.OutPkg == "" {
		pkgName, err := LookupPackageName(conf.OutPath)
		if err != nil {
			return fmt.Errorf("Failed to lookup for the package name: %v", err)
		}
		conf.OutPkg = pkgName
	}
	return nil
}

// getDataQualifier returns the contextualized data qualifer for the program imports.
func (c *CompiledTemplatesProgram) getDataQualifier(dataConf compiled.DataConfiguration) string {
	dataAlias := c.addImport(dataConf.PkgPath)
	dataQualifier := fmt.Sprintf("%v.%v", dataAlias, dataConf.DataTypeName)
	if dataConf.IsPtr {
		dataQualifier = fmt.Sprintf("*%v.%v", dataAlias, dataConf.DataTypeName)
	}
	return dataQualifier
}

// addImport adds a new import spec to the output program.
// if the pkgpath is already imported, it is imported once only.
// if the pkgpath collides with another exisiting ident, it is renamed appropriately.
// it returns the alias of the pkgpath.
func (c *CompiledTemplatesProgram) addImport(pkgpath string) string {
	qpath := fmt.Sprintf("%q", pkgpath)
	bpath := filepath.Base(pkgpath)
	// if already imported, return the current alias
	for _, i := range c.imports {
		if i.Path.Value == qpath {
			if i.Name == nil {
				return bpath
			}
			return i.Name.Name
		}
	}
	newImport := &ast.ImportSpec{
		Path: &ast.BasicLit{Value: qpath},
	}
	c.imports = append(c.imports, newImport)
	duplicated := false
	i := 0
	okAlias := bpath
	for c.isCollidingIdent(okAlias) {
		if i == 0 {
			okAlias = "alias" + bpath
		} else {
			okAlias = fmt.Sprintf("alias%v%v", bpath, i)
		}
		i++
		duplicated = true
	}
	if duplicated {
		newImport.Name = &ast.Ident{Name: okAlias}
	}
	c.idents = append(c.idents, okAlias)
	return okAlias
}

// isCollidingIdent tells if given ident will collide exisiting idents.
func (c *CompiledTemplatesProgram) isCollidingIdent(ident string) bool {
	// check for imports
	for _, i := range c.imports {
		if i.Name != nil && i.Name.Name == ident {
			return true
		}
	}
	// check for static idents
	for _, i := range c.idents {
		if i == ident {
			return true
		}
	}
	return false
}

// makeFuncName produces a new unique func name.
func (c *CompiledTemplatesProgram) makeFuncName(baseName string) string {
	x := baseName
	i := 0
	for c.isCollidingIdent(x) {
		x = fmt.Sprintf("%v%v%v", "fn", i, baseName)
		i++
	}
	c.idents = append(c.idents, x)
	return x
}

// createFunc creates the ast code of a compiled template function with given name.
func (c *CompiledTemplatesProgram) createFunc(name string) *ast.FuncDecl {
	gocode := fmt.Sprintf(
		`package aa
func %v(t parse.Templater, w io.Writer, indata interface{}) error {}`,
		name,
	)
	f := stringToAst(gocode)
	fn := f.Decls[0].(*ast.FuncDecl)
	c.funcs = append(c.funcs, fn)
	return fn
}

// addBuiltintText registers a static builtin text to the program.
func (c *CompiledTemplatesProgram) addBuiltintText(text string) string {
	if x, ok := c.builtinTexts[text]; ok {
		return x
	}
	c.builtinTexts[text] = fmt.Sprintf("%v%v", "builtin", len(c.builtinTexts))
	return c.builtinTexts[text]
}

// generateInitFunc generates the init func body.
// the init func contains the code to declare the compiled templates
// and associate their defined templates.
func (c *CompiledTemplatesProgram) generateInitFunc(tpls []*TemplateToCompile) string {
	initfunc := ""
	initfunc += fmt.Sprintf("func init () {\n")
	for _, t := range tpls {
		for _, f := range t.files {
			for _, name := range f.names() {
				funcname := f.tplsFunc[name]
				initfunc += fmt.Sprintf("  %v.Add(%#v, %v)\n", c.varName, name, funcname)
			}
		}
	}

	for _, t := range tpls {
		for i, f := range t.files {
			for e, name := range f.definedTemplates {
				varX := fmt.Sprintf("tpl%vX%v", i, e)
				varY := fmt.Sprintf("tpl%vY%v", i, e)
				initfunc += fmt.Sprintf("  %v := %v.MustGet(%#v)\n", varX, c.varName, f.name)
				initfunc += fmt.Sprintf("  %v := %v.MustGet(%#v)\n", varY, c.varName, name)
				initfunc += fmt.Sprintf("  %v, _ = %v.Compiled(%v)\n", varX, varX, varY)
				initfunc += fmt.Sprintf("  %v.Set(%#v, %v)\n", c.varName, f.name, varX)
			}
		}
	}

	initfunc += fmt.Sprintf("}")
	return initfunc
}

// generateProgram generates the output program.
func (c *CompiledTemplatesProgram) generateProgram(outpkg string, tpls []*TemplateToCompile) string {
	program := fmt.Sprintf("package %v\n\n", outpkg)
	program += fmt.Sprintf("//golint:ignore\n\n")
	program += fmt.Sprintf("%v\n\n", c.generateImportStmt())
	program += fmt.Sprintf("%v\n\n", c.generateBuiltins())
	program += fmt.Sprintf("%v\n\n", c.generateInitFunc(tpls))
	for _, f := range c.funcs {
		program += fmt.Sprintf("%v\n\n", astNodeToString(f))
	}
	return program
}

// generateImportStmt generates all import statements.
func (c *CompiledTemplatesProgram) generateImportStmt() string {
	importStmt := ""
	importStmt += fmt.Sprintf("import (\n")
	for _, i := range c.imports {
		importStmt += fmt.Sprintf("\t")
		if i.Name != nil {
			importStmt += fmt.Sprintf("%v ", i.Name.Name)
		}
		importStmt += fmt.Sprintf("%v\n", i.Path.Value)
	}
	importStmt += fmt.Sprintf(")")
	return importStmt
}

// generateBuiltins generates the builtins text variable declarations.
func (c *CompiledTemplatesProgram) generateBuiltins() string {
	builtins := ""
	for text, name := range c.builtinTexts {
		builtins += fmt.Sprintf("var %v = []byte(%q)\n", name, text)
	}
	return builtins
}

// convertConfigToTemplatesToCompile convert the confguration into instances of TemplateToCompile
func convertConfigToTemplatesToCompile(conf *compiled.Configuration) []*TemplateToCompile {
	ret := []*TemplateToCompile{}
	for _, t := range conf.Templates {
		ret = append(ret, makeTemplateToCompile(t))
	}
	return ret
}

// TemplateToCompile links a configuration and all the template files it matches.
type TemplateToCompile struct {
	*compiled.TemplateConfiguration
	files []TemplateFileToCompile
}

// TemplateFileToCompile links a template file with all the templates defined in it.
type TemplateFileToCompile struct {
	name             string
	tplsTree         map[string]*parse.Tree
	tplsFunc         map[string]string
	tplsTypeCheck    map[string]*simplifier.State
	definedTemplates []string
}

// names returns all template names sorted asc.
func (t TemplateFileToCompile) names() []string {
	strs := []string{}
	for name := range t.tplsTree {
		strs = append(strs, name)
	}
	sort.Strings(strs)
	return strs
}

// getDataConfiguration returns the data configuration for the given template name.
func (t TemplateToCompile) getDataConfiguration(name string) (compiled.DataConfiguration, error) {
	if ret, ok := t.TemplatesDataConfiguration[name]; ok {
		return ret, nil
	}
	if ret, ok := t.TemplatesDataConfiguration["*"]; ok {
		return ret, nil
	}
	fmt.Printf("%#v\n", t.TemplatesData)
	fmt.Printf("%#v\n", t.TemplatesDataConfiguration)
	return compiled.DataConfiguration{}, fmt.Errorf("Template data configuration not found for %v", name)
}

// getData returns the data value for the given template name.
func (t TemplateToCompile) getData(name string) (interface{}, error) {
	if ret, ok := t.TemplatesData[name]; ok {
		return ret, nil
	}
	if ret, ok := t.TemplatesData["*"]; ok {
		return ret, nil
	}
	return nil, fmt.Errorf("Template data configuration not found for %v", name)
}

// makeTemplateToCompile creates a new instance of TemplateToCompile for the given TemplateConfiguration.
func makeTemplateToCompile(templateConf compiled.TemplateConfiguration) *TemplateToCompile {
	ret := &TemplateToCompile{
		TemplateConfiguration: &templateConf,
		files: []TemplateFileToCompile{},
	}
	return ret
}

// prepare evalutes the files of the TemplateConfiguration and prepares the resulting templates.
func (t *TemplateToCompile) prepare() error {
	if t.TemplatesPath != "" {
		tplsPath, err := filepath.Glob(t.TemplatesPath)
		if err != nil {
			return fmt.Errorf("Failed to glob the templates: %v %v", t.TemplatesPath, err)
		}
		for _, tplPath := range tplsPath {
			fileTpl, err := makeTemplateFileToCompileFromFile(tplPath, t)
			if err != nil {
				return err
			}
			t.files = append(t.files, fileTpl)
		}
	} else {
		fileTpl, err := makeTemplateFileToCompileFromStr(t.TemplateName, t.TemplateContent, t)
		if err != nil {
			return err
		}
		t.files = append(t.files, fileTpl)
	}
	return nil
}

//makeTemplateFileToCompileFromFile creates a new TemplateFileToCompile instance for the given template file.
func makeTemplateFileToCompileFromFile(tplPath string, tplToCompile *TemplateToCompile) (TemplateFileToCompile, error) {

	fileTpl := TemplateFileToCompile{
		name:             filepath.Base(tplPath),
		tplsTree:         map[string]*parse.Tree{},
		tplsFunc:         map[string]string{},
		tplsTypeCheck:    map[string]*simplifier.State{},
		definedTemplates: []string{},
	}

	content, err := ioutil.ReadFile(tplPath)
	if err != nil {
		return fileTpl, err
	}
	mainName := fileTpl.name
	funcs := tplToCompile.FuncsExport

	var treeNames map[string]*parse.Tree
	if tplToCompile.HTML {
		treeNames, err = compileHTMLTemplate(mainName, string(content), funcs)
	} else {
		treeNames, err = compileTextTemplate(mainName, string(content), funcs)
	}
	if err != nil {
		return fileTpl, err
	}
	fileTpl.tplsTree = treeNames
	for treeName, tree := range fileTpl.tplsTree {
		data, err := tplToCompile.getData(treeName)
		if err != nil {
			return fileTpl, err
		}
		fileTpl.tplsTypeCheck[treeName] = simplifier.TransformTree(tree, data, funcs)
		if treeName != mainName {
			fileTpl.tplsFunc[treeName] = cleanTplName("fn" + mainName + "_" + treeName)
			fileTpl.definedTemplates = append(fileTpl.definedTemplates, treeName)
		} else {
			fileTpl.tplsFunc[treeName] = cleanTplName("fn" + mainName)
		}
	}
	return fileTpl, nil
}

//makeTemplateFileToCompileFromStr creates a new TemplateFileToCompile instance for the given template content.
func makeTemplateFileToCompileFromStr(name, tplContent string, tplToCompile *TemplateToCompile) (TemplateFileToCompile, error) {

	fileTpl := TemplateFileToCompile{
		name:             name,
		tplsTree:         map[string]*parse.Tree{},
		tplsFunc:         map[string]string{},
		tplsTypeCheck:    map[string]*simplifier.State{},
		definedTemplates: []string{},
	}
	funcs := tplToCompile.FuncsExport

	var err error
	var treeNames map[string]*parse.Tree
	if tplToCompile.HTML {
		treeNames, err = compileHTMLTemplate(name, tplContent, funcs)
	} else {
		treeNames, err = compileTextTemplate(name, tplContent, funcs)
	}
	if err != nil {
		return fileTpl, err
	}
	fileTpl.tplsTree = treeNames
	mainName := fileTpl.name
	for treeName, tree := range fileTpl.tplsTree {
		data, err := tplToCompile.getData(treeName)
		if err != nil {
			return fileTpl, err
		}
		fileTpl.tplsTypeCheck[treeName] = simplifier.TransformTree(tree, data, funcs)
		if treeName != mainName {
			fileTpl.tplsFunc[treeName] = cleanTplName("fn" + mainName + "_" + treeName)
			fileTpl.definedTemplates = append(fileTpl.definedTemplates, treeName)
		} else {
			fileTpl.tplsFunc[treeName] = cleanTplName("fn" + mainName)
		}
	}
	return fileTpl, nil
}

// compileTextTemplate compiles a file template as a text/template, it returns a map of trees by their name.
func compileTextTemplate(name string, content string, funcsMap map[string]interface{}) (map[string]*parse.Tree, error) {
	ret := map[string]*parse.Tree{}

	t, err := text.New(name).Funcs(funcsMap).Parse(content)
	if err != nil {
		return ret, err
	}

	for _, tpl := range t.Templates() {
		tpl.Execute(ioutil.Discard, nil) // ignore err, it is just to force parse.
		if tpl.Tree != nil {
			ret[tpl.Name()] = tpl.Tree
		}
	}

	return ret, nil
}

// compileHTMLTemplate compiles a file template as an html/template, it returns a map of trees by their name.
func compileHTMLTemplate(name string, content string, funcsMap map[string]interface{}) (map[string]*parse.Tree, error) {
	ret := map[string]*parse.Tree{}

	t, err := html.New(name).Funcs(funcsMap).Parse(content)
	if err != nil {
		return ret, err
	}

	for _, tpl := range t.Templates() {
		tpl.Execute(ioutil.Discard, nil) // ignore err, it is just to force parse.
		if tpl.Tree != nil {
			ret[tpl.Name()] = tpl.Tree
		}
	}

	return ret, nil
}

// LookupPackageName search a directory for its declaring package.
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
	return LookupPackageNameFromStr(string(b)), nil
}

// LookupPackageNameFromStr extract the declaring package from given go source string.
func LookupPackageNameFromStr(gocode string) string {
	// improve this. really q&d.
	gocode = gocode[strings.Index(gocode, "package"):]
	gocode = gocode[0:strings.Index(gocode, "\n")]
	return strings.Split(gocode, "package ")[1]
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
