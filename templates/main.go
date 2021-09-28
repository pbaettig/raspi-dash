package templates

import (
	"embed"
	"html/template"
)

type IndexPageData struct {
	LoadAvg1  string
	LoadAvg5  string
	LoadAvg15 string
	CPUTemp   string
	Plots     map[string]string
}

//go:embed *.tmpl
var fs embed.FS
var IndexPage = template.Must(template.ParseFS(fs, "index.html.tmpl"))

func init() {
	// IndexPage = template.Must(template.New("index").Funcs(template.FuncMap{
	// 	"nilDefault": func(v *interface{}) interface{} {
	// 		if v == nil {
	// 			return "-"
	// 		}
	// 		return v
	// 	},
	// }).ParseFS(fs, "index.html.tmpl"))

}
