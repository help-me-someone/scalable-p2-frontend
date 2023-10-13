package main

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
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

type Handler = func(http.ResponseWriter, *http.Request)

func NeedAuth(handler Handler) Handler {
	return func(wr http.ResponseWriter, re *http.Request) {

		// Retrieve claims, if DNE then just redirect.
		claims, err := ValidateJWTTOken(re)
		if err != nil {
			log.Println("NO TOKEN!!!")
			http.Redirect(wr, re, "/login", http.StatusFound)
			return
		}

		// Debug
		log.Printf("Sending request with user: %s\n", claims.Username)

		ctx := context.WithValue(re.Context(), "username", claims.Username)

		// Okay. We are authenticated.
		handler(wr, re.WithContext(ctx))
	}
}

type URLContext struct {
	API_GATEWAY_URL string
}

func (u *URLContext) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/")

	log.Println("Handling:", path)

	// Handling index.
	if len(path) == 0 {
		path = "index"
	}

	// Cursed.... I hope no one ever sees this.
	err := executor.ExecuteTemplate(w, path, map[string]string{
		"API_GATEWAY_URL": u.API_GATEWAY_URL,
	})

	if err != nil {
		executor.ExecuteTemplateStatus(w, "404", nil, http.StatusNotFound)
	}
}

func main() {
	API_GATEWAY_URL := os.Getenv("API_GATEWAY_URL")
	log.Println("API_GATEWAY_URL:", API_GATEWAY_URL)

	// Create templates.
	if debug {
		executor = DebugTemplateExecutor{template_path}
	} else {
		funcs := map[string]any{
			"map": func(pairs ...any) (map[string]any, error) {
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
			},
		}
		executor = ReleaseTemplateExecutor{
			r: render.New(render.Options{
				DisableHTTPErrorRendering: true,
				Directory:                 "templates",
				Layout:                    "baseof",
				FileSystem:                &render.EmbedFileSystem{FS: tmplFS},
				Extensions:                []string{".html", ".tmpl"},
				Funcs:                     []template.FuncMap{funcs},
			}),
		}
	}

	mux := http.NewServeMux()
	mux.Handle("/static/", http.FileServer(http.FS(static)))

	// Feel free to add this later.
	favicon_handler := func(w http.ResponseWriter, req *http.Request) {
		// Do nothing... unless...
	}

	template_handler := &URLContext{
		API_GATEWAY_URL: API_GATEWAY_URL,
	}

	// Handlers.
	mux.HandleFunc("/favicon.ico", favicon_handler)

	log.Println("API_GATEWAY_URL gateway url:", API_GATEWAY_URL)
	mux.Handle("/login", template_handler)
	mux.Handle("/forgot_password", template_handler)
	mux.Handle("/", http.StripPrefix("/", http.HandlerFunc(NeedAuth(template_handler.ServeHTTP))))

	// Serve.
	port := ":8000"
	log.Println("Listening for requests at http://localhost" + port)
	log.Fatal(http.ListenAndServe(port, mux))
}
