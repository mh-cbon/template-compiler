

Manual testing,

```sh
go run main.go -keep -html -tpl "demo/templates/*.tpl" -out "demo/gen.go" -var compiledTemplates -data github.com/mh-cbon/template-compiler/demo/data:MyTemplateData text/template:builtins

go run main.go  -html -tpl "demo/templates/*.tpl" -out "demo/gen.go" -var compiledTemplates -data github.com/mh-cbon/template-compiler/demo/data:MyTemplateData

```
