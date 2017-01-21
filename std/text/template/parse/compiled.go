package parse

import (
	"fmt"
	"io"
)

// CompiledNode is a Node to represent
// a compiled template in the template nodes flow.
type CompiledNode struct {
	NodeType
	Pos
	tr      *Tree
	Execute CompiledTemplateFunc
}

// Templater is the interface for text/html templates.
type Templater interface {
	GetFuncs() map[string]interface{}
	ExecuteTemplate(io.Writer, string, interface{}) error
}

// CompiledTemplateFunc is the signature of the func responsible to render a compiled template.
type CompiledTemplateFunc func(t Templater, w io.Writer, data interface{}) error

// NewCompiledNode makes a new instance of CompiledNode
func (t *Tree) NewCompiledNode(Execute CompiledTemplateFunc) *CompiledNode {
	return &CompiledNode{tr: t, NodeType: NodeCompiledNode, Execute: Execute}
}

func (a *CompiledNode) String() string {
	return fmt.Sprintf("{{compiled}}")

}

func (a *CompiledNode) tree() *Tree {
	return a.tr
}

// Copy this Node to a new instance.
func (a *CompiledNode) Copy() Node {
	return a.tr.NewCompiledNode(a.Execute)
}
