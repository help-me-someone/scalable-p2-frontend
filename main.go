package main

import (
	"log"
	"net/http"
	"strings"
)

// This is a minimalistic frontend which will be able to serve requests
// by using the URL path as the name of the template which should be served.
// Eg. localhost:8000/register --> tmpl/register.html
//
// NOTE(Appy): You should NOT need to really change anything in here.
// HTMX, tailwind and alpinejs has been integrated into the HTML.

func main() {
	// Create templates.
	templates := NewTemplates()

	mux := http.NewServeMux()
	mux.Handle("/static/", http.FileServer(http.FS(static)))

	// Feel free to add this later.
	favicon_handler := func(w http.ResponseWriter, req *http.Request) {
		// Do nothing... unless...
	}

	error404_handler := func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		templates.Render(w, "404.html", nil)
	}

	// Template substitutor. (Check template.go for more info)
	template_handler := func(w http.ResponseWriter, req *http.Request) {
		path := req.URL.Path

		// Handling index.
		if len(path) == 0 {
			path = "index.html"
		}

		// Append html extension if missing.
		if !strings.Contains(path, ".html") {
			path = path + ".html"
		}
		err := templates.Render(w, path, nil)

		if err != nil {
			error404_handler(w, req)
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
