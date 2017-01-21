package compiler

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"reflect"
	"strings"
	"text/template/parse"

	"github.com/mh-cbon/template-tree-simplifier/simplifier"
)

type converter struct {
	tree              *parse.Tree
	funcs             map[string]interface{}
	funcsMapPublic    []map[string]string
	writerName        string
	builtinTexts      map[string]string
	fn                *ast.FuncDecl
	state             *state
	errvars           int
	additionalImports []string
}

func (c *converter) createErrVars() string {
	c.errvars++
	if c.errvars == 0 {
		return "err"
	}
	return fmt.Sprintf("err%v", c.errvars)
}

type state struct {
	current *scope
}
type scope struct {
	dotVars []string
	body    *ast.BlockStmt
	parent  *scope
}

func (s *state) addNode(n ast.Stmt) {
	s.current.body.List = append(s.current.body.List, n)
}

func (s *state) enter(n *ast.BlockStmt, currentDotVar string) {
	if n == nil {
		err := fmt.Errorf("state.enter: Impossible to enter a nil BlockStmt")
		panic(err)
	}
	s.current = &scope{[]string{currentDotVar}, n, s.current}
}

func (s *state) dotVar() string {
	return s.current.dotVars[len(s.current.dotVars)-1]
}

func (s *state) leave() {
	if s.current != nil {
		s.current = s.current.parent
	}
}

func convertTplTree(
	tree *parse.Tree,
	typeCheck *simplifier.State,
	builtinTexts map[string]string,
	dataQualifier string,
	funcs map[string]interface{},
	funcsMapPublic []map[string]string,
	fnName string,
) (*ast.FuncDecl, []string) {
	c := converter{tree, funcs, funcsMapPublic, "w", builtinTexts, nil, nil, -1, []string{}}
	c.fn = makeFunc(fnName)

	if simplifier.IsUsingDot(tree) {
		c.fn.Body.List = append(c.fn.Body.List, makePrelude(dataQualifier)...)
	}
	if simplifier.PrintsAnything(tree) {
		c.fn.Body.List = append(c.fn.Body.List, makeWriteErrorDecl()...)
	}

	c.state = &state{}
	c.state.enter(c.fn.Body, "data")
	typeCheck.Enter()
	c.convert(tree.Root, typeCheck)
	typeCheck.Leave()
	c.state.leave()
	injectReturnNil(c.fn)
	return c.fn, c.additionalImports
}

func makeFunc(fnName string) *ast.FuncDecl {
	fn := &ast.FuncDecl{Name: &ast.Ident{Name: fnName}}
	fn.Type = &ast.FuncType{}

	fn.Type.Params = &ast.FieldList{}
	fn.Type.Params.List = make([]*ast.Field, 0)

	names := []*ast.Ident{&ast.Ident{Name: "t"}}
	typee := &ast.SelectorExpr{X: &ast.Ident{Name: "parse"}, Sel: &ast.Ident{Name: "Templater"}}
	fn.Type.Params.List = append(fn.Type.Params.List, &ast.Field{Names: names, Type: typee})

	names = []*ast.Ident{&ast.Ident{Name: "w"}}
	typee = &ast.SelectorExpr{X: &ast.Ident{Name: "io"}, Sel: &ast.Ident{Name: "Writer"}}
	fn.Type.Params.List = append(fn.Type.Params.List, &ast.Field{Names: names, Type: typee})

	names = []*ast.Ident{&ast.Ident{Name: "indata"}}
	intf := &ast.InterfaceType{Methods: &ast.FieldList{}}
	fn.Type.Params.List = append(fn.Type.Params.List, &ast.Field{Names: names, Type: intf})

	fn.Type.Results = &ast.FieldList{}
	fn.Type.Results.List = make([]*ast.Field, 0)

	fn.Type.Results.List = append(fn.Type.Results.List, &ast.Field{Type: &ast.Ident{Name: "error"}})

	fn.Body = &ast.BlockStmt{}
	return fn
}

func (c *converter) convert(node interface{}, typeCheck *simplifier.State) {
	switch node := node.(type) {

	case *parse.TextNode:
		if len(node.Text) > 0 {
			for _, stmt := range c.handleTextNode(node) {
				c.state.addNode(stmt)
			}
		}

	case *parse.ListNode:
		for _, n := range node.Nodes {
			c.convert(n, typeCheck)
		}

	case *parse.ActionNode:
		for _, stmt := range c.handleActionNode(node, typeCheck) {
			c.state.addNode(stmt)
		}

	case *parse.IfNode:
		ifStmt := c.handleIfNode(node, typeCheck)
		c.state.addNode(ifStmt)
		c.state.enter(ifStmt.Body, c.state.dotVar())
		for _, n := range node.List.Nodes {
			c.convert(n, typeCheck)
		}
		c.state.leave()
		if ifStmt.Else != nil {
			elseStmt := ifStmt.Else.(*ast.BlockStmt)
			c.state.enter(elseStmt, c.state.dotVar())
			for _, n := range node.ElseList.Nodes {
				c.convert(n, typeCheck)
			}
			c.state.leave()
		}

	case *parse.RangeNode:
		rangeStmt, dotVarName := c.handleRangeNode(node, typeCheck)
		c.state.addNode(rangeStmt)
		c.state.enter(rangeStmt.Body, dotVarName)
		typeCheck.Enter()
		for _, n := range node.List.Nodes {
			c.convert(n, typeCheck)
		}
		c.state.leave()
		typeCheck.Leave()
		if node.ElseList != nil {
			typeCheck.Enter()
			elseStmt := c.handleRangeElseNode(node, typeCheck)
			c.state.addNode(elseStmt)
			c.state.enter(elseStmt.Body, c.state.dotVar())
			for _, n := range node.ElseList.Nodes {
				c.convert(n, typeCheck)
			}
			c.state.leave()
			typeCheck.Leave()
		}

	case *parse.WithNode:
		// pretty much the same as ifStmt,
		// a with node turns into a
		// if something is truelike{}else{}
		// note, it is embeded with a BlockStmt
		// to respect the with nature of the template node.
		ifStmt, dotVarName := c.handleWithNode(node, typeCheck)
		c.state.addNode(embedInBlockStmt(ifStmt))
		c.state.enter(ifStmt.Body, dotVarName)
		typeCheck.Enter()
		for _, n := range node.List.Nodes {
			c.convert(n, typeCheck)
		}
		typeCheck.Leave()
		c.state.leave()
		if ifStmt.Else != nil {
			typeCheck.Enter()
			elseStmt := ifStmt.Else.(*ast.BlockStmt)
			c.state.enter(elseStmt, c.state.dotVar())
			for _, n := range node.ElseList.Nodes {
				c.convert(n, typeCheck)
			}
			typeCheck.Leave()
			c.state.leave()
		}

	case *parse.TemplateNode:
		for _, stmt := range c.handleTemplateNode(node, typeCheck) {
			c.state.addNode(stmt)
		}

	default:
		err := fmt.Errorf("converter.convert: Node type unhandled\n%v\n%#v", node, node)
		panic(err)
	}
}

func injectReturnNil(fn *ast.FuncDecl) {
	strFunc := `func ww ()error{return nil}`
	n := getReturnFuncAst(strFunc)
	fn.Body.List = append(fn.Body.List, n)
}
func getReturnFuncAst(strFunc string) ast.Stmt {
	gocode := `package aa
` + strFunc
	f := stringToAst(gocode)
	return f.Decls[0].(*ast.FuncDecl).Body.List[0]
}
func embedInBlockStmt(s ast.Stmt) *ast.BlockStmt {
	return &ast.BlockStmt{List: []ast.Stmt{s}}
}
func (c *converter) addImport(i string) {
	for _, im := range c.additionalImports {
		if im == i {
			return
		}
	}
	c.additionalImports = append(c.additionalImports, i)
}
func (c *converter) handleTextNode(node *parse.TextNode) []ast.Stmt {
	var builtinName string
	if b, ok := c.builtinTexts[string(node.Text)]; ok {
		builtinName = b
	} else {
		builtinName = fmt.Sprintf("%v%v", "builtin", len(c.builtinTexts))
		c.builtinTexts[string(node.Text)] = builtinName
	}
	return c.makeIoWrite(builtinName, reflect.TypeOf([]byte{}))
}
func (c *converter) handleActionNode(node *parse.ActionNode, typeCheck *simplifier.State) []ast.Stmt {
	ret := []ast.Stmt{}
	if len(node.Pipe.Decl) == 0 { // likely a print
		t, _ := c.getTypesOfCommandNode(node.Pipe.Cmds[0], typeCheck)

		expr := c.handleCommandNode(node.Pipe.Cmds[0], typeCheck)
		ret = c.makeIoWrite(astNodeToString(expr), t[0])

	} else if len(node.Pipe.Cmds) == 1 { // likely a simple assignment $z := 4
		// this case could go into the next one, it would produce an assignement (:=)
		// but this case is designed spcifically to produce var declaration with their type.
		expr := c.handleCommandNode(node.Pipe.Cmds[0], typeCheck)
		exprType, outTypes := c.getTypesOfCommandNode(node.Pipe.Cmds[0], typeCheck)
		if len(outTypes) > 0 {
			// the method return more than 1 parameters,
			// the declaration switch to an assignment
			// x, err := call(...)
			// It is assumed that the second return parameter
			// is an err of type error.
			assign := &ast.AssignStmt{}
			assign.Lhs = make([]ast.Expr, 0)
			for _, n := range node.Pipe.Decl {
				assign.Lhs = append(assign.Lhs, c.convertVariableNode(n, typeCheck))
			}
			errVar := c.createErrVars()
			assign.Lhs = append(assign.Lhs, &ast.Ident{Name: errVar})
			assign.Tok = token.DEFINE
			assign.Rhs = []ast.Expr{expr}
			ret = append(ret, assign)

			// Add the error check
			gocode := `package aa
        func zz (indata interface{}) {
          if ` + errVar + ` != nil {
            return ` + errVar + `
          }
        }`
			f := stringToAst(gocode)
			// locate the if...
			ifErr := f.Decls[0].(*ast.FuncDecl).Body.List[0]
			ret = append(ret, ifErr)

		} else {
			// this is a variable declaration,
			// var x string = ""
			vspec := &ast.ValueSpec{
				Names:  []*ast.Ident{c.convertVariableNode(node.Pipe.Decl[0], typeCheck).(*ast.Ident)},
				Type:   &ast.Ident{Name: exprType[0].String()},
				Values: []ast.Expr{expr},
			}
			decl := &ast.GenDecl{Tok: token.VAR, Specs: []ast.Spec{vspec}}
			ret = append(ret, &ast.DeclStmt{Decl: decl})
		}

	} else { // likely a complex assignment
		expr := c.handleCommandNode(node.Pipe.Cmds[0], typeCheck)
		assign := &ast.AssignStmt{}
		assign.Lhs = make([]ast.Expr, 0)
		assign.Tok = token.DEFINE
		assign.Rhs = make([]ast.Expr, 0)
		assign.Rhs = append(assign.Rhs, expr)
		for _, n := range node.Pipe.Decl {
			assign.Lhs = append(assign.Lhs, c.convertVariableNode(n, typeCheck))
		}
		ret = append(ret, assign)
	}
	return ret
}
func (c *converter) handleIfNode(node *parse.IfNode, typeCheck *simplifier.State) *ast.IfStmt {
	if len(node.Pipe.Decl) > 0 {
		err := fmt.Errorf(
			"converter.handleIfNode: Unhandled If node with declarations\n%v\n%#v",
			node, node)
		panic(err)
	}
	if len(node.Pipe.Cmds) > 1 {
		err := fmt.Errorf(
			"converter.handleIfNode: Unhandled If node with more than 1 Cmd\n%v\n%#v",
			node, node)
		panic(err)
	}
	ifStmt := &ast.IfStmt{
		Body: &ast.BlockStmt{},
	}
	exprToTest := c.handleCommandNode(node.Pipe.Cmds[0], typeCheck)
	typeToTest, _ := c.getTypesOfCommandNode(node.Pipe.Cmds[0], typeCheck)
	ifStmt.Cond = c.makeBinaryTest(exprToTest, typeToTest[0])
	if node.ElseList != nil && len(node.ElseList.Nodes) > 0 {
		ifStmt.Else = &ast.BlockStmt{}
	}
	return ifStmt
}
func (c *converter) handleTemplateNode(node *parse.TemplateNode, typeCheck *simplifier.State) []ast.Stmt {
	if node.Pipe != nil {
		if len(node.Pipe.Decl) > 0 {
			err := fmt.Errorf(
				"converter.handleTemplateNode: Unhandled Template node with declarations\n%v\n%#v",
				node, node)
			panic(err)
		}
		if len(node.Pipe.Cmds) > 1 {
			err := fmt.Errorf(
				"converter.handleTemplateNode: Unhandled Template node with more than 1 Cmd\n%v\n%#v",
				node, node)
			panic(err)
		}
	}

	expr := ", nil"
	if node.Pipe != nil {
		exprStmt := c.handleCommandNode(node.Pipe.Cmds[0], typeCheck)
		expr = astNodeToString(exprStmt)
		expr = ", " + expr
	}
	gocode := `package aa
	func x() {
    writeErr = t.ExecuteTemplate(` + c.writerName + `, "` + node.Name + `"` + expr + `)
    if writeErr != nil {
      return writeErr
    }
  }`
	f := stringToAst(gocode)
	return f.Decls[0].(*ast.FuncDecl).Body.List
}
func (c *converter) handleRangeNode(node *parse.RangeNode, typeCheck *simplifier.State) (*ast.RangeStmt, string) {
	var dotVarName string
	decl := node.BranchNode.Pipe.Decl

	ret := &ast.RangeStmt{Tok: token.DEFINE, Body: &ast.BlockStmt{}}
	if len(decl) == 1 {
		ret.Value = c.convertNode(decl[0], typeCheck)
		dotVarName = decl[0].Ident[0][1:]

	} else if len(decl) == 2 {
		ret.Key = c.convertNode(decl[0], typeCheck)
		ret.Value = c.convertNode(decl[1], typeCheck)
		dotVarName = decl[1].Ident[0][1:]

	} else {
		dotVarName = c.state.dotVar()
	}
	ret.X = c.convertNode(node.BranchNode.Pipe.Cmds[0].Args[0], typeCheck)

	return ret, dotVarName
}
func (c *converter) handleRangeElseNode(node *parse.RangeNode, typeCheck *simplifier.State) *ast.IfStmt {

	gocode := `package aa
  func zz (indata interface{}) {
    if len(x)==0 {}
  }`
	f := stringToAst(gocode)
	//c.convertNode(node.BranchNode.Pipe.Cmds[0].Args[0])

	// locate the if
	ifStmt := f.Decls[0].(*ast.FuncDecl).Body.List[0].(*ast.IfStmt)
	// locate the call to len()
	lenCall := ifStmt.Cond.(*ast.BinaryExpr).X.(*ast.CallExpr)
	// replace variable x with the correct ident
	lenCall.Args[0] = c.convertNode(node.BranchNode.Pipe.Cmds[0].Args[0], typeCheck)

	return ifStmt
}
func (c *converter) handleWithNode(node *parse.WithNode, typeCheck *simplifier.State) (*ast.IfStmt, string) {
	var dotVarName string
	if len(node.Pipe.Cmds) > 1 {
		err := fmt.Errorf(
			"converter.handleIfNode: Unhandled With node with more than 1 Cmd\n%v\n%#v",
			node, node)
		panic(err)
	}
	ifStmt := &ast.IfStmt{
		Body: &ast.BlockStmt{},
	}
	if len(node.Pipe.Decl) > 0 {
		expr := c.handleCommandNode(node.Pipe.Cmds[0], typeCheck)
		assign := &ast.AssignStmt{}
		assign.Tok = token.DEFINE
		assign.Lhs = make([]ast.Expr, 0)
		assign.Rhs = make([]ast.Expr, 0)
		assign.Rhs = append(assign.Rhs, expr)
		for _, n := range node.Pipe.Decl {
			y := c.convertVariableNode(n, typeCheck)
			assign.Lhs = append(assign.Lhs, y)
		}
		ifStmt.Init = assign
		varToTest := c.convertNode(node.Pipe.Decl[0], typeCheck)
		typeToTest := typeCheck.GetVar(node.Pipe.Decl[0].Ident[0])
		ifStmt.Cond = c.makeBinaryTest(varToTest, typeToTest)
		dotVarName = node.Pipe.Decl[0].Ident[0]

	} else {
		dotVarName = node.Pipe.Cmds[0].Args[0].(*parse.VariableNode).Ident[0][1:] // must be a var.
		expr := c.handleCommandNode(node.Pipe.Cmds[0], typeCheck)
		typeToTest, _ := c.getTypesOfCommandNode(node.Pipe.Cmds[0], typeCheck)
		ifStmt.Cond = c.makeBinaryTest(expr, typeToTest[0])

	}
	if node.ElseList != nil && len(node.ElseList.Nodes) > 0 {
		ifStmt.Else = &ast.BlockStmt{}
	}
	return ifStmt, dotVarName
}
func (c *converter) handleCommandNode(node *parse.CommandNode, typeCheck *simplifier.State) ast.Expr {
	var ret ast.Expr
	if len(node.Args) == 1 {
		e := c.convertNode(node.Args[0], typeCheck)
		if e == nil {
			err := fmt.Errorf(
				"converter.handleCommandNode: Node.Cmd.Arg[0] conversion failed\n%v\n%#v",
				node, node)
			panic(err)
		}
		ret = e
	} else {
		var fnCall *ast.CallExpr
		switch x := node.Args[0].(type) {
		case *parse.IdentifierNode:
			fnCall = c.convertIdentifierNode(x).(*ast.ExprStmt).X.(*ast.CallExpr)
		case *parse.FieldNode:
			fnCall = c.convertFieldNodeMethod(x, typeCheck).(*ast.CallExpr)
		case *parse.VariableNode:
			fnCall = c.convertVariableNode(x, typeCheck).(*ast.CallExpr)
		default:
			err := fmt.Errorf(
				"converter.handleCommandNode: Unhandled node type\n%v\n%#v",
				node, node)
			panic(err)
		}
		for i, a := range node.Args[1:] {
			e := c.convertNode(a, typeCheck)
			if e == nil {
				err := fmt.Errorf(
					"converter.handleCommandNode: Node.Cmd.Arg[%v] conversion failed\n%v\n%#v",
					i, node, node)
				panic(err)
			}
			fnCall.Args = append(fnCall.Args, e)
		}
		ret = fnCall
	}
	return ret
}
func (c *converter) getTypesOfCommandNode(node *parse.CommandNode, typeCheck *simplifier.State) ([]reflect.Type, []reflect.Type) {
	ret := []reflect.Type{}
	out := []reflect.Type{}
	switch x := node.Args[0].(type) {
	case *parse.FieldNode:
		y := typeCheck.Dot()

		if typeCheck.IsMethodPath(x.Ident, y) {
			methType := typeCheck.ReflectPath(x.Ident, y)
			for i := 0; i < methType.NumOut(); i++ {
				if i == 0 {
					ret = append(ret, methType.Out(i))
				} else {
					out = append(out, methType.Out(i))
				}
			}
		} else {
			rType := typeCheck.BrowsePathType(x.Ident, y)
			ret = append(ret, rType)
		}

	case *parse.VariableNode:
		y := typeCheck.GetVar(x.Ident[0])

		if typeCheck.IsMethodPath(x.Ident[1:], y) {
			methType := typeCheck.ReflectPath(x.Ident[1:], y)
			for i := 0; i < methType.NumOut(); i++ {
				if i == 0 {
					ret = append(ret, methType.Out(i))
				} else {
					out = append(out, methType.Out(i))
				}
			}
		} else {
			if len(x.Ident) > 1 {
				y = typeCheck.BrowsePathType(x.Ident[1:], y)
			}
			ret = append(ret, y)
		}

	case *parse.NumberNode:
		if x.IsFloat && !isHexConstant(x.Text) && strings.ContainsAny(x.Text, ".eE") {
			ret = append(ret, reflect.TypeOf(1.0))
		} else if x.IsComplex {
			ret = append(ret, reflect.TypeOf(1i))
		} else {
			ret = append(ret, reflect.TypeOf(1))
		}

	case *parse.StringNode:
		ret = append(ret, reflect.TypeOf(""))

	case *parse.DotNode:
		ret = append(ret, typeCheck.Dot())

	case *parse.BoolNode:
		ret = append(ret, reflect.TypeOf(x.True))

	case *parse.IdentifierNode:
		fn := c.funcs[x.Ident]
		fnType := reflect.TypeOf(fn)
		for i := 0; i < fnType.NumOut(); i++ {
			if i == 0 {
				ret = append(ret, fnType.Out(i))
			} else {
				out = append(out, fnType.Out(i))
			}
		}

	default:
		err := fmt.Errorf(
			"converter.getTypesOfCommandNode: Unhandled node type\n%v\n%#v",
			node, node)
		panic(err)
	}
	return ret, out
}

// copied from template/exec.go?#L478
func isHexConstant(s string) bool {
	return len(s) > 2 && s[0] == '0' && (s[1] == 'x' || s[1] == 'X')
}

// copied from template/exec.go?#L478
func (c *converter) makeBinaryTest(expr ast.Expr, exprType reflect.Type) ast.Expr {
	// returns a binary expr, except for a struct which is always true (turns into an ast.Ident)
	ret := &ast.BinaryExpr{X: expr}
	switch exprType.Kind() {
	case reflect.String:
		ret.Op = token.NEQ
		ret.Y = &ast.BasicLit{Kind: token.STRING, Value: `""`}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		ret.Op = token.NEQ
		ret.Y = &ast.BasicLit{Kind: token.INT, Value: `0`}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		ret.Op = token.NEQ
		ret.Y = &ast.BasicLit{Kind: token.INT, Value: `0`}
	case reflect.Float32, reflect.Float64:
		ret.Op = token.NEQ
		ret.Y = &ast.BasicLit{Kind: token.INT, Value: `0`}
	case reflect.Bool:
		// a bool expr, return it as is
		return expr
	case reflect.Struct:
		// a struct is always true
		// https://golang.org/src/text/template/exec.go#L299
		return &ast.Ident{Name: "true"}
	case reflect.Array, reflect.Map, reflect.Slice /*, reflect.String*/ :
		// truth = val.Len() > 0
		gocode := `package aa
      func zz (indata interface{}) {
        len(x)
      }`
		f := stringToAst(gocode)
		// locate the len()
		lenCall := f.Decls[0].(*ast.FuncDecl).Body.List[0].(*ast.ExprStmt).X.(*ast.CallExpr)
		// replace variable x with the correct ident
		lenCall.Args[0] = expr
		ret.X = lenCall
		ret.Op = token.GTR
		ret.Y = &ast.BasicLit{Kind: token.INT, Value: `0`}
	default:
		err := fmt.Errorf(
			"converter.makeBinaryTest: Unhandled expression relfect.type\n%v\n%#v",
			exprType, exprType.Kind())
		panic(err)
	}
	return ret
}
func (c *converter) convertNode(node parse.Node, typeCheck *simplifier.State) ast.Expr {
	var ret ast.Expr
	switch x := node.(type) {
	case *parse.FieldNode:
		ret = c.convertFieldNode(x, typeCheck)
	case *parse.VariableNode:
		ret = c.convertVariableNode(x, typeCheck)
	case *parse.NumberNode:
		ret = c.convertNumberNode(x)
	case *parse.StringNode:
		ret = c.convertStringNode(x)
	case *parse.BoolNode:
		ret = c.convertBoolNode(x)
	case *parse.DotNode:
		ret = c.convertVariableNode(&parse.VariableNode{Ident: []string{"$" + c.state.dotVar()}}, typeCheck)
	default:
		err := fmt.Errorf(
			"converter.convertNode: Unhandled node type\n%v\n%#v",
			node, node)
		panic(err)
	}
	return ret
}

// returns the selector such template JSEscaper
// and populates c.additionalImports
func (c *converter) identifierToPublicCall(name string) string {
	for _, i := range c.funcsMapPublic {
		if i["FuncName"] == name {
			c.addImport(i["Pkg"])
			return i["Sel"]
		}
	}
	return ""
}
func (c *converter) convertIdentifierNode(node *parse.IdentifierNode) ast.Stmt {
	// maybe this func can be called directly as pkg.func
	p := c.identifierToPublicCall(node.Ident)
	if len(p) > 0 {
		gocode := `package aa
  func x() {` + p + `()}`
		f := stringToAst(gocode)
		return f.Decls[0].(*ast.FuncDecl).Body.List[0]
	}
	// It s func to consume from the runtime funcmap
	fn, ok := c.funcs[node.Ident]
	if ok == false {
		panic(fmt.Errorf("convert.convertIdentifierNode: func map not found %v", node.Ident))
	} else {

	}
	// two cases now,
	// This func can be inlined into,
	// template.GetFuncs()[ident].(func (...)...)(...)
	// or it can t,
	// let s panic for now.
	rFn := reflect.TypeOf(fn)
	in := ""
	for i := 0; i < rFn.NumIn(); i++ {
		t := rFn.In(i)
		switch t.Kind() {
		case reflect.Struct:
			if ast.IsExported(t.String()) == false {
				panic(fmt.Errorf("convert.convertIdentifierNode: Impossible to use non exported in parameter of type %v in funcmap %v",
					t.String(),
					node.Ident,
				))
			}
		case reflect.Ptr:
			if ast.IsExported(t.Elem().String()) == false {
				panic(fmt.Errorf("convert.convertIdentifierNode: Impossible to use non exported in parameter of type %v in funcmap %v",
					t.String(),
					node.Ident,
				))
			}
		}
		in += t.String() + ","
	}
	out := ""
	for i := 0; i < rFn.NumOut(); i++ {
		t := rFn.Out(i)
		switch t.Kind() {
		case reflect.Struct:
			if ast.IsExported(t.String()) == false {
				panic(fmt.Errorf("convert.convertIdentifierNode: Impossible to use non exported in parameter of type %v in funcmap %v",
					t.String(),
					node.Ident,
				))
			}
		case reflect.Ptr:
			if ast.IsExported(t.Elem().String()) == false {
				panic(fmt.Errorf("convert.convertIdentifierNode: Impossible to use non exported in parameter of type %v in funcmap %v",
					t.String(),
					node.Ident,
				))
			}
		}
		out += t.String() + ","
	}

	if len(in) > 0 {
		in = in[0 : len(in)-1]
	}
	if len(out) > 0 {
		out = out[0 : len(out)-1]
	}

	gocode := `package aa
func x() {t.GetFuncs()["` + node.Ident + `"].(func (` + in + `) (` + out + `))()}`
	f := stringToAst(gocode)
	return f.Decls[0].(*ast.FuncDecl).Body.List[0]
}
func (c *converter) convertFieldNodeMethod(node *parse.FieldNode, typeCheck *simplifier.State) ast.Expr {
	return c.convertFieldNode(node, typeCheck)
}
func (c *converter) convertStringNode(node *parse.StringNode) *ast.BasicLit {
	ret := &ast.BasicLit{Kind: token.STRING, Value: node.Quoted}
	return ret
}
func (c *converter) convertBoolNode(node *parse.BoolNode) *ast.Ident {
	ret := &ast.Ident{Name: node.String()}
	return ret
}
func (c *converter) convertNumberNode(node *parse.NumberNode) *ast.BasicLit {
	ret := &ast.BasicLit{Kind: token.INT, Value: node.Text}
	if node.IsComplex {
		ret.Kind = token.IMAG
	} else if node.IsFloat {
		ret.Kind = token.FLOAT
	}
	return ret
}
func (c *converter) convertFieldNode(node *parse.FieldNode, typeCheck *simplifier.State) ast.Expr {
	var ret ast.Expr
	for i := 0; i < len(node.Ident); i += 2 {
		if ret == nil {
			ret = &ast.SelectorExpr{
				X:   &ast.Ident{Name: c.state.dotVar()},
				Sel: &ast.Ident{Name: node.Ident[i]},
			}
		} else {
			ret = &ast.SelectorExpr{
				X:   ret,
				Sel: &ast.Ident{Name: node.Ident[i]},
			}
		}
	}
	ismethod := typeCheck.IsMethodPath(node.Ident, typeCheck.Dot())
	if ismethod {
		// need to embed the last ast.SelectorExpr with a CallExpr
		ret = &ast.CallExpr{Fun: ret}
	}
	return ret
}
func (c *converter) convertVariableNode(node *parse.VariableNode, typeCheck *simplifier.State) ast.Expr {
	var ret ast.Expr
	if len(node.Ident) == 1 {
		ret = &ast.Ident{Name: node.Ident[0][1:]}
	} else {
		for i := 0; i < len(node.Ident); i += 2 {
			if ret == nil {
				ret = &ast.SelectorExpr{
					X:   &ast.Ident{Name: node.Ident[i][1:]},
					Sel: &ast.Ident{Name: node.Ident[i+1]},
				}
			} else {
				ret = &ast.SelectorExpr{
					X:   ret,
					Sel: &ast.Ident{Name: node.Ident[i]},
				}
			}
		}
	}
	ismethod := typeCheck.IsMethodPath(node.Ident[1:], typeCheck.GetVar(node.Ident[0]))
	if ismethod {
		// need to embed the last ast.SelectorExpr with a CallExpr
		ret = &ast.CallExpr{Fun: ret}
	}
	return ret
}

func makeWriteErrorDecl() []ast.Stmt {
	gocode := `package aa
func zz (indata interface{}) {
  var writeErr error
}`
	f := stringToAst(gocode)
	return f.Decls[0].(*ast.FuncDecl).Body.List
}

func makePrelude(dataQualifier string) []ast.Stmt {
	gocode := `package aa
func zz (indata interface{}) {
  var data ` + dataQualifier + `
  if d, ok := indata.(` + dataQualifier + `); ok {
    data = d
  }
}`
	f := stringToAst(gocode)
	return f.Decls[0].(*ast.FuncDecl).Body.List
}

func (c *converter) makeIoWrite(expr string, exprType reflect.Type) []ast.Stmt {
	writeCall := ""
	switch exprType.Kind() {
	case reflect.String:
		writeCall = "io.WriteString(w, " + expr + ")"
	case reflect.Int:
		writeCall = "io.WriteString(w, strconv.Itoa(" + expr + "))"
		c.addImport("strconv")
	case reflect.Int8, reflect.Int16, reflect.Int32:
		writeCall = "io.WriteString(w, strconv.FormatInt(int64(" + expr + "), 10))"
		c.addImport("strconv")
	case reflect.Int64:
		writeCall = "io.WriteString(w, strconv.FormatInt(" + expr + ", 10))"
		c.addImport("strconv")
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		writeCall = "io.WriteString(w, strconv.FormatUint(" + expr + ", 10))"
		c.addImport("strconv")
	case reflect.Float32:
		writeCall = "io.WriteString(w, strconv.FormatFloat(float64(" + expr + "), \"f\", -1, 32))"
		c.addImport("strconv")
	case reflect.Float64:
		writeCall = "io.WriteString(w, strconv.FormatFloat(" + expr + ", \"f\", -1, 64))"
		c.addImport("strconv")
	case reflect.Bool:
		writeCall = "io.WriteString(w, strconv.FormatBool(" + expr + "))"
		c.addImport("strconv")
	case reflect.Slice:
		switch exprType.Elem().Kind() {
		case reflect.Uint8:
			writeCall = "w.Write(" + expr + ")"
		default:
			writeCall = "fmt.Fprintf(w, \"%v\", " + expr + ")"
			c.addImport("fmt")
		}
	case reflect.Struct, reflect.Interface:
		writeCall = "fmt.Fprintf(w, \"%v\", " + expr + ")"
		c.addImport("fmt")
	default:
		err := fmt.Errorf(
			"makeIoWrite: Unhandled expression relfect.type\n%v\n%#v",
			exprType, exprType.Kind())
		panic(err)
	}
	gocode := `package aa
func zz (indata interface{}) {
  _, writeErr = ` + writeCall + `
  if writeErr!=nil{
    return writeErr
  }
}`
	f := stringToAst(gocode)
	return f.Decls[0].(*ast.FuncDecl).Body.List
}

func stringToAst(gocode string) *ast.File {
	f, err := parser.ParseFile(token.NewFileSet(), "", gocode, 0)
	if err != nil {
		err := fmt.Errorf(
			"stringToAst: Failed to convert string to ast\n%v",
			gocode)
		panic(err)
	}
	return f
}

func astNodeToString(n ast.Node) string {
	var b bytes.Buffer
	err := format.Node(&b, token.NewFileSet(), n)
	if err != nil {
		err := fmt.Errorf(
			"astNodeToString: Failed to convert ast node to string\n%v\n%#v",
			n, n)
		panic(err)
	}
	return b.String()
}