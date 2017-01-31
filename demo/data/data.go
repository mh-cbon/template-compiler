package data

// MyTemplateData ...
type MyTemplateData struct {
	Some  string
	Items []string
}

// MethodItems ...
func (m MyTemplateData) MethodItems() []string {
	return []string{"", "", "", ""}
}
