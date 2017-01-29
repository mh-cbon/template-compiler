package compiled

// New creates a new configuration instance
func New(outpath string, templates []TemplateConfiguration, funcsmap ...string) *Configuration {
	return &Configuration{
		OutPath:   outpath,
		Registry:  NewRegistry(),
		Templates: templates,
		FuncsMap:  funcsmap,
	}
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
	HTML              bool
	TemplatesPath     string
	Data              interface{}
	DataConfiguration DataConfiguration
	FuncsMap          []string
	FuncsExport       map[string]interface{}
	PublicIdents      []map[string]string
}

//DataConfiguration holds information about the data type consumed by the template.
type DataConfiguration struct {
	IsPtr        bool
	DataTypeName string
	DataType     string
	PkgPath      string
}
