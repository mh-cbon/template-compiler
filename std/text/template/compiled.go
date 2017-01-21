package template

import (
	"io"

	"github.com/mh-cbon/template-compiler/std/text/template/parse"
)

// additions to template.Template

// Compiled registers a compiled template.
func (t *Template) Compiled(c *Compiled) (*Template, error) {
	t.init()
	// Add the newly parsed trees, including the one for t, into our common structure.
	for name, tmpl := range c.compiledTmpl {
		t.tmpl[name] = tmpl.Template
	}
	// Add the newly parsed trees, including the one for t, into our common structure.
	for name, tmpl := range c.tmpl {
		t.tmpl[name] = tmpl
	}
	return t, nil
}

// GetFuncs returns the func map of the template.
func (t *Template) GetFuncs() map[string]interface{} {
	return t.parseFuncs
}

// Compiled ...
type Compiled struct {
	*Template
	compiledTmpl map[string]*Compiled
	executeFn    parse.CompiledTemplateFunc
}

// NewCompiled is the template type of a compiled template.
func (r *Compiled) init() {
	r.Template.init()
	r.compiledTmpl = make(map[string]*Compiled)
	r.Tree = &parse.Tree{}
	r.Tree.Root = &parse.ListNode{}
	r.Root = r.Tree.Root
	r.Tree.Root.Nodes = append(
		r.Tree.Root.Nodes,
		r.Tree.NewCompiledNode(r.executeFn))
}

// NewCompiled makes a new Compiled template instance of a func implementation.
func NewCompiled(name string, fn parse.CompiledTemplateFunc) *Compiled {
	r := &Compiled{
		Template:  &Template{},
		executeFn: fn,
	}
	r.init()
	r.compiledTmpl[name] = r
	return r
}

// Execute invokes the compiled template function.
func (r *Compiled) Execute(wr io.Writer, data interface{}) error {
	// its important to bypass Template.Execute method.
	return r.executeFn(r, wr, data)
}

type templateExecute interface {
	executeFn(io.Writer, string, interface{}) error
}

// ExecuteTemplate invokes the compiled template function.
func (r *Compiled) ExecuteTemplate(wr io.Writer, name string, data interface{}) error {
	// its important to bypass Template.ExecuteTemplate method.
	if t, ok := r.compiledTmpl[name]; ok {
		return t.executeFn(r, wr, data)
	}
	return r.Template.ExecuteTemplate(wr, name, data)
}

// Compiled registers a compiled template.
func (r *Compiled) Compiled(c *Compiled) (*Compiled, error) {
	for name, tmpl := range c.compiledTmpl {
		r.compiledTmpl[name] = tmpl
	}
	// Add the newly parsed trees, including the one for t, into our common structure.
	for name, tmpl := range c.tmpl {
		r.tmpl[name] = tmpl
	}
	return r, nil
}
