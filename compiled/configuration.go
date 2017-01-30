package compiled

import (
	"path/filepath"
	"reflect"
)

// New creates a new configuration instance
func New(outpath string, templates []TemplateConfiguration, funcsmap ...string) *Configuration {
	ret := &Configuration{
		OutPath:   outpath,
		Registry:  NewRegistry(),
		Templates: templates,
		FuncsMap:  funcsmap,
	}
	for i := range ret.Templates {
		t := &ret.Templates[i]
		t.TemplatesDataConfiguration = map[string]DataConfiguration{}
		for n, d := range t.TemplatesData {
			t.TemplatesDataConfiguration[n] = makeDataConfiguration(d)
		}
	}
	return ret
}

// SetPkg configures the output package of the produced compilation.
func (c *Configuration) SetPkg(s string) *Configuration {
	c.OutPkg = s
	return c
}

// Configuration holds all information to run the template compiler.
type Configuration struct {
	*Registry
	OutPath   string
	OutPkg    string
	Templates []TemplateConfiguration
	FuncsMap  []string
}

// TemplateConfiguration holds the configuration for a set of template files.
type TemplateConfiguration struct {
	HTML                       bool
	TemplatesPath              string
	TemplateName               string
	TemplateContent            string
	TemplatesData              map[string]interface{}
	TemplatesDataConfiguration map[string]DataConfiguration
	FuncsMap                   []string
	FuncsExport                map[string]interface{}
	PublicIdents               []map[string]string
}

//DataConfiguration holds information about the data type consumed by the template.
type DataConfiguration struct {
	IsPtr        bool
	DataTypeName string
	DataType     string
	PkgPath      string
}

// makeDataConfiguration transforms data into a DataConfiguration
func makeDataConfiguration(some interface{}) DataConfiguration {
	ret := DataConfiguration{}
	if some == nil {
		return ret
	}
	r := reflect.TypeOf(some)
	isPtr := r.Kind() == reflect.Ptr
	if isPtr {
		r = r.Elem()
	}
	ret.IsPtr = isPtr
	ret.DataTypeName = r.Name()
	ret.PkgPath = r.PkgPath()
	ret.DataType = filepath.Base(r.PkgPath()) + "." + r.Name()
	return ret
}
