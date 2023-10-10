package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/unrolled/render"
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

type Handler = func(http.ResponseWriter, *http.Request)

func NeedAuth(handler Handler) Handler {
	return func(wr http.ResponseWriter, re *http.Request) {

		// Retrieve cookie, if DNE then just redirect.
		cookie, err := re.Cookie("token")
		if err != nil {
			http.Redirect(wr, re, "/login", http.StatusFound)
			return
		}
		_ = cookie

		// New client to perform out request.
		client := &http.Client{}

		// Set up a new request.
		req, err := http.NewRequest("GET", "http://localhost:7887/auth", nil)
		if err != nil {
			http.Redirect(wr, re, "/login", http.StatusFound)
			return
		}

		// Carry out the auth request with the cookie.
		req.AddCookie(cookie)

		// Carry out the request.
		resp, err := client.Do(req)
		if err != nil {
			// Just redirect by default.
			http.Redirect(wr, re, "/login", http.StatusFound)
			return
		}

		var response Response
		err = json.NewDecoder(resp.Body).Decode(&response)
		if err != nil {
			// Failed to deocde, let's redirect anyways.
			http.Redirect(wr, re, "/login", http.StatusFound)
			return
		}

		if !response.Authenticated {
			// Not authenticated, we redirect as well.
			http.Redirect(wr, re, "/login", http.StatusFound)
			return
		}

		// Okay. We are authenticated.
		handler(wr, re)
	}
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
		path := strings.TrimPrefix(req.URL.Path, "/")

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
	mux.HandleFunc("/login", template_handler)
	mux.Handle("/", http.StripPrefix("/", http.HandlerFunc(NeedAuth(template_handler))))

	// Serve.
	port := ":8000"
	log.Println("Listening for requests at http://localhost" + port)
	log.Fatal(http.ListenAndServe(port, mux))
}
