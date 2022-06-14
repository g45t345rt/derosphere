package cli

var AppTemplate = `{{range .VisibleCategories}}{{if .Name}}
{{.Name}}:{{range .VisibleCommands}}
{{join .Names ", "}}{{"\t"}}{{.Usage}}{{end}}{{else}}{{range .VisibleCommands}}
{{join .Names ", "}}{{"\t"}}{{.Usage}}{{end}}{{end}}{{end}}
`
