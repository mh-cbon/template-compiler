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

// VERSION is the current program version.
var VERSION = "0.0.0"

func main() {

	var err error
	var versionPtr = flag.Bool("version", false, "Show version")
	var help = flag.Bool("help", false, "Show help")
	var shelp = flag.Bool("h", false, "Show help")
	var keep = flag.Bool("keep", false, "Keep program generator")
	var print = flag.Bool("print", false, "Keep program generator")
	var varNamePtr = flag.String("var", "", "Name of the compiled.Registry variable to use")
	var wdirPtr = flag.String("wdir", "", "Working directory")

	flag.Parse()

	if *versionPtr {
		showVersion()
		return
	}

	if *help || *shelp {
		showHelp()
		return
	}

	wdir := *wdirPtr
	varName := *varNamePtr

	if varName == "" {
		varName = "compiledTemplates"
	}

	if wdir == "" {
		wdir, err = eludeWorkingDirectory(wdir)
		panicOnErr(err)
	}

	w, _ := os.Getwd()
	file := filepath.Join(w, os.Getenv("GOFILE"))

	prog, err := compiler.GenerateProgramBootstrapFromFile(
		file,
		varName,
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

func showVersion() {
	fmt.Println(`template-compiler - ` + VERSION)
}

func showHelp() {
	showVersion()
	fmt.Println(`
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

func invokeProgram(wdir string) error {
	c := exec.Command("go", []string{"run", wdir + "/main.go"}...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}
