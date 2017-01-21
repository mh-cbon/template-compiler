package compiled

import (
	"reflect"

	"github.com/mh-cbon/template-compiler/std/text/template"
	"github.com/mh-cbon/template-compiler/std/text/template/parse"
)

// NewRegistry create new instance of Registry
func NewRegistry() *Registry {
	return &Registry{
		templates: make(map[string]*template.Compiled),
	}
}

// Registry is a regisrty to map template path to their compiled funcs.
type Registry struct {
	templates map[string]*template.Compiled
}

// Add registers a compiled template func.
func (t Registry) Add(name string, fn parse.CompiledTemplateFunc) {
	t.templates[name] = template.NewCompiled(name, fn)
}

// Get provides a compiled func for a template path
func (t Registry) Get(name string) *template.Compiled {
	return t.templates[name]
}

// Set a compiled template.
func (t Registry) Set(name string, tpl *template.Compiled) {
	t.templates[name] = tpl
}

// MustGet provides a compiled func for a template path,
// it panics if the template cannot be found.
func (t Registry) MustGet(name string) *template.Compiled {
	if t, ok := t.templates[name]; ok {
		return t
	}
	panic("template not found")
}

// EvaluateFuncCall ...
func EvaluateFuncCall(fn reflect.Value, args ...interface{}) interface{} {
	return nil
}
