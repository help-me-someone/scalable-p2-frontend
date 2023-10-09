package main

// References.
// Setup: https://stackoverflow.com/questions/36617949/how-to-use-base-template-file-for-golang-html-template
// Hot-reload: https://stackoverflow.com/questions/36951855/is-it-possible-to-reload-html-templates-after-app-is-started

import (
	"embed"
	"github.com/unrolled/render"
	"io"
	"net/http"
)

// DO NOT TOUCH.

//go:generate npm run build

//go:embed static
var static embed.FS

//go:embed templates/*.tmpl
var tmplFS embed.FS

// Supporting hot reload
type TemplateExecutor interface {
	ExecuteTemplate(w io.Writer, name string, data interface{}) error
	ExecuteTemplateStatus(w io.Writer, name string, data interface{}, status int) error
}

type DebugTemplateExecutor struct {
	root string
}

func (e DebugTemplateExecutor) ExecuteTemplateStatus(w io.Writer, name string, data interface{}, status int) error {
	r := render.New(render.Options{
		DisableHTTPErrorRendering: true,
		Directory:                 "templates",
		Layout:                    "baseof",
		Extensions:                []string{".html", ".tmpl"},
	})

	return r.HTML(w, status, name, nil)
}

func (e DebugTemplateExecutor) ExecuteTemplate(w io.Writer, name string, data interface{}) error {
	return e.ExecuteTemplateStatus(w, name, data, http.StatusOK)
}

type ReleaseTemplateExecutor struct {
	r *render.Render
}

func (e ReleaseTemplateExecutor) ExecuteTemplateStatus(w io.Writer, name string, data interface{}, status int) error {
	return e.r.HTML(w, status, name, nil)
}

func (e ReleaseTemplateExecutor) ExecuteTemplate(w io.Writer, name string, data interface{}) error {
	return e.ExecuteTemplateStatus(w, name, data, http.StatusOK)
}
