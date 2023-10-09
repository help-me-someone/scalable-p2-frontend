package main

// Reference: https://stackoverflow.com/questions/36617949/how-to-use-base-template-file-for-golang-html-template

import (
	"embed"
	"html/template"
	"io"
)

// DO NOT TOUCH.

//go:generate npm run build

//go:embed static
var static embed.FS

//go:embed tmpl/*.html
var tmplFS embed.FS

type Template struct {
	templates *template.Template
}

func NewTemplates() *Template {
	funcMap := template.FuncMap{}

	templates := template.Must(template.New("").Funcs(funcMap).ParseFS(tmplFS, "tmpl/*.html"))
	return &Template{
		templates: templates,
	}
}

func (t *Template) Render(w io.Writer, name string, data interface{}) error {
	tmpl := template.Must(t.templates.Clone())
	target, err := tmpl.ParseFS(tmplFS, "tmpl/"+name)

	if err != nil {
		return err
	}

	tmpl = template.Must(target, err)
	return tmpl.ExecuteTemplate(w, name, data)
}
