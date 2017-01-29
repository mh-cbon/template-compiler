package template

import "github.com/mh-cbon/template-compiler/std/text/template"

// publicFuncMap is the map of functions with public funcs
var publicFuncMap = template.FuncMap{
	"_html_template_attrescaper":     AttrEscaper,
	"_html_template_commentescaper":  CommentEscaper,
	"_html_template_cssescaper":      CSSEscaper,
	"_html_template_cssvaluefilter":  CSSValueFilter,
	"_html_template_htmlnamefilter":  HTMLNameFilter,
	"_html_template_htmlescaper":     HTMLEscaper,
	"_html_template_jsregexpescaper": JSRegexpEscaper,
	"_html_template_jsstrescaper":    JSStrEscaper,
	"_html_template_jsvalescaper":    JSValEscaper,
	"_html_template_nospaceescaper":  HTMLNospaceEscaper,
	"_html_template_rcdataescaper":   RcdataEscaper,
	"_html_template_urlescaper":      URLEscaper,
	"_html_template_urlfilter":       URLFilter,
	"_html_template_urlnormalizer":   urlNormalizer,
}

// URLNormalizer ...
func URLNormalizer(args ...interface{}) string {
	return urlNormalizer(args...)
}

// URLFilter ...
func URLFilter(args ...interface{}) string {
	return urlFilter(args...)
}

// URLEscaper ...
func URLEscaper(args ...interface{}) string {
	return urlEscaper(args...)
}

// JSValEscaper ...
func JSValEscaper(args ...interface{}) string {
	return jsValEscaper(args...)
}

// JSStrEscaper ...
func JSStrEscaper(args ...interface{}) string {
	return jsStrEscaper(args...)
}

// JSRegexpEscaper ...
func JSRegexpEscaper(args ...interface{}) string {
	return jsRegexpEscaper(args...)
}

// HTMLNameFilter ...
func HTMLNameFilter(args ...interface{}) string {
	return htmlNameFilter(args...)
}

// CSSValueFilter ...
func CSSValueFilter(args ...interface{}) string {
	return cssValueFilter(args...)
}

// CSSEscaper ...
func CSSEscaper(args ...interface{}) string {
	return cssEscaper(args...)
}

// HTMLNospaceEscaper escapes for inclusion in unquoted attribute values.
func HTMLNospaceEscaper(args ...interface{}) string {
	return htmlNospaceEscaper(args...)
}

// AttrEscaper escapes for inclusion in quoted attribute values.
func AttrEscaper(args ...interface{}) string {
	return attrEscaper(args...)
}

// RcdataEscaper escapes for inclusion in an RCDATA element body.
func RcdataEscaper(args ...interface{}) string {
	return rcdataEscaper(args...)
}

// CommentEscaper ...
func CommentEscaper(args ...interface{}) string {
	return commentEscaper(args...)
}
