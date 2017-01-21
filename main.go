package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/mh-cbon/template-compiler/compiler"
)

func main() {

	var err error
	var help = flag.Bool("help", false, "Show help")
	var shelp = flag.Bool("h", false, "Show help")
	var keep = flag.Bool("keep", false, "Keep program generator")
	var print = flag.Bool("print", false, "Keep program generator")
	var isHTML = flag.Bool("html", false, "Compile template as HTML")
	var tplGlobPtr = flag.String("tpl", "", "Glob to the templates")
	var outPathPtr = flag.String("out", "", "Path to the output go file")
	var varNamePtr = flag.String("var", "", "Name of the compiled.Registry variable to use")
	var wdirPtr = flag.String("wdir", "", "Working directory")
	var dataPtr = flag.String("data", "", "Package Path and type name of the template data")

	flag.Parse()

	if *help || *shelp {
		showHelp()
		return
	}

	tplGlob := *tplGlobPtr
	outPath := *outPathPtr
	wdir := *wdirPtr
	varName := *varNamePtr
	data := *dataPtr

	if varName == "" {
		panic("You must provide a variabe name to use")
	}

	if tplGlob == "" {
		panic("You must provide a glob to the templates to compile")
	}

	if outPath == "" {
		panic("You must provide an output path for the compiled templates")
	}

	if wdir == "" {
		wdir, err = eludeWorkingDirectory(wdir)
		panicOnErr(err)
	}

	outPath, err = eludeOutPath(outPath)
	panicOnErr(err)

	outPath, err = filepath.Abs(outPath)
	panicOnErr(err)

	tplsPath, err := filepath.Glob(tplGlob)
	panicOnErr(err)

	funcMapExport := consolidateFuncMapToExport(flag.Args())
	prog, err := compiler.GenerateProgramBootstrap(
		tplsPath,
		outPath,
		varName,
		*isHTML,
		data,
		funcMapExport...,
	)
	panicOnErr(err)

	if *print {
		fmt.Println(prog)
	}

	if *keep {
		fmt.Printf("Program written at %v\n", wdir+"/main.go")
	}

	err = ioutil.WriteFile(wdir+"/main.go", []byte(prog), os.ModePerm)
	panicOnErr(err)

	err2 := invokeProgram(wdir)
	if *keep == false {
		os.RemoveAll(wdir)
	}
	if err2 != nil {
		os.Exit(1)
	}
}

func showHelp() {
	fmt.Println(`template-compiler - 0.0.0
`)
}

func panicOnErr(err error) {
	if err != nil {
		panic(err)
	}
}

// eludeWorkingDirectory returns a working directory
// within GOPATH
func eludeWorkingDirectory(wdir string) (string, error) {
	GoPath := os.Getenv("GOPATH") + "/src/"
	return ioutil.TempDir(GoPath, "template-compiler")
}

// eludeOutPath take input outPath and refines it
// to be a file path.
func eludeOutPath(outPath string) (string, error) {
	if filepath.Ext(outPath) != ".go" {
		// it must be a dir
		if s, err := os.Stat(outPath); err != nil {
			return outPath, err

		} else if s.IsDir() == false {
			return outPath, fmt.Errorf("Wrong output filepath: %v", outPath)

		}
		// set the filename
		outPath += "/gen.go"
	}
	return outPath, nil
}

// consolidateFuncMapToExport ensures that the funcmap to export
// contains test and template-tree-simplifier builtin functions.
func consolidateFuncMapToExport(ex []string) []string {
	templateFuncs := "text/template:builtins"
	simplifierFuncs := "github.com/mh-cbon/template-tree-simplifier/funcmap:tplFunc"
	if strIndex(ex, templateFuncs) == false {
		ex = append(ex, templateFuncs)
	}
	if strIndex(ex, simplifierFuncs) == false {
		ex = append(ex, simplifierFuncs)
	}
	return ex
}

func invokeProgram(wdir string) error {
	c := exec.Command("go", []string{"run", wdir + "/main.go"}...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}

func strIndex(list []string, search string) bool {
	for _, l := range list {
		if l == search {
			return true
		}
	}
	return false
}
