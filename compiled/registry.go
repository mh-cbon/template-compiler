package compiled

import (
	"fmt"

	"github.com/mh-cbon/template-compiler/std/text/template"
	"github.com/mh-cbon/template-compiler/std/text/template/parse"
)

// NewRegistry create new instance of Registry
func NewRegistry() *Registry {
	return &Registry{
		templates: make(map[string]*template.Compiled),
	}
}

// Registry registers compiled templates by their name.
type Registry struct {
	templates map[string]*template.Compiled
}

// Add registers a func as a compiled template with given name.
func (t Registry) Add(name string, fn parse.CompiledTemplateFunc) {
	t.templates[name] = template.NewCompiled(name, fn)
}

// Get provides a compiled template matching given name.
func (t Registry) Get(name string) *template.Compiled {
	return t.templates[name]
}

// Set a compiled template as given name.
func (t Registry) Set(name string, tpl *template.Compiled) {
	t.templates[name] = tpl
}

// MustGet provides the compiled template matching given name,
// it panics if the template is not found.
func (t Registry) MustGet(name string) *template.Compiled {
	if t, ok := t.templates[name]; ok {
		return t
	}
	panic(
		fmt.Errorf("template not found: %v", name),
	)
}
