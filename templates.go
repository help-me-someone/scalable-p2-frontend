package main

// References.
// Setup: https://stackoverflow.com/questions/36617949/how-to-use-base-template-file-for-golang-html-template
// Hot-reload: https://stackoverflow.com/questions/36951855/is-it-possible-to-reload-html-templates-after-app-is-started

import (
	"embed"
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/http"

	"github.com/unrolled/render"
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

// Template utility functions
func TemplateMap(pairs ...any) (map[string]any, error) {
	if len(pairs)%2 != 0 {
		return nil, errors.New("misaligned map")
	}

	m := make(map[string]any, len(pairs)/2)

	for i := 0; i < len(pairs); i += 2 {
		key, ok := pairs[i].(string)

		if !ok {
			return nil, fmt.Errorf("cannot use type %T as map key", pairs[i])
		}
		m[key] = pairs[i+1]
	}
	return m, nil
}

func TemplateLoop(from, to int) <-chan int {
	ch := make(chan int)
	go func() {
		for i := from; i <= to; i++ {
			ch <- i
		}
		close(ch)
	}()
	return ch
}

func TemplateAdd(a, b int) int {
	return a + b
}

func TemplateSub(a, b int) int {
	return a - b
}

func (e DebugTemplateExecutor) ExecuteTemplateStatus(w io.Writer, name string, data interface{}, status int) error {

	funcs := map[string]any{
		"map":  TemplateMap,
		"loop": TemplateLoop,
		"add":  TemplateAdd,
		"sub":  TemplateSub,
	}
	r := render.New(render.Options{
		DisableHTTPErrorRendering: true,
		Directory:                 "templates",
		Layout:                    "baseof",
		Extensions:                []string{".html", ".tmpl"},
		Funcs:                     []template.FuncMap{funcs},
	})

	return r.HTML(w, status, name, data)
}

func (e DebugTemplateExecutor) ExecuteTemplate(w io.Writer, name string, data interface{}) error {
	return e.ExecuteTemplateStatus(w, name, data, http.StatusOK)
}

type ReleaseTemplateExecutor struct {
	r *render.Render
}

func (e ReleaseTemplateExecutor) ExecuteTemplateStatus(w io.Writer, name string, data interface{}, status int) error {
	return e.r.HTML(w, status, name, data)
}

func (e ReleaseTemplateExecutor) ExecuteTemplate(w io.Writer, name string, data interface{}) error {
	return e.ExecuteTemplateStatus(w, name, data, http.StatusOK)
}
