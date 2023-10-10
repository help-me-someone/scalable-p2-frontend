package main

import (
	"log"
	"net/http"

	"github.com/unrolled/render"
	// "strings"
)

// This is a minimalistic frontend which will be able to serve requests
// by using the URL path as the name of the template which should be served.
// Eg. localhost:8000/register --> tmpl/register.html
//
// NOTE(Appy): You should NOT need to really change anything in here.
// HTMX, tailwind and alpinejs has been integrated into the HTML.

// Set this to true, for production.
const debug = true
const template_path = "templates/*.tmpl"

var executor TemplateExecutor

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
}

func main() {

	// Create templates.
	if debug {
		executor = DebugTemplateExecutor{template_path}
	} else {
		executor = ReleaseTemplateExecutor{
			r: render.New(render.Options{
				DisableHTTPErrorRendering: true,
				Directory:                 "templates",
				Layout:                    "baseof",
				FileSystem:                &render.EmbedFileSystem{FS: tmplFS},
				Extensions:                []string{".html", ".tmpl"}}),
		}
	}

	mux := http.NewServeMux()
	mux.Handle("/static/", http.FileServer(http.FS(static)))

	// Feel free to add this later.
	favicon_handler := func(w http.ResponseWriter, req *http.Request) {
		// Do nothing... unless...
	}

	// Template substitutor. (Check template.go for more info)
	template_handler := func(w http.ResponseWriter, req *http.Request) {
		enableCors(&w)
		path := req.URL.Path

		// Handling index.
		if len(path) == 0 {
			path = "index"
		}

		err := executor.ExecuteTemplate(w, path, nil)

		if err != nil {
			executor.ExecuteTemplateStatus(w, "404", nil, http.StatusNotFound)
		}
	}

	// Handlers.
	mux.HandleFunc("/favicon.ico", favicon_handler)
	mux.Handle("/", http.StripPrefix("/", http.HandlerFunc(template_handler)))

	// Serve.
	port := ":8000"
	log.Println("Listening for requests at http://localhost" + port)
	log.Fatal(http.ListenAndServe(port, mux))
}
