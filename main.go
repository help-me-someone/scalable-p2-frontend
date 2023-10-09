package main

import (
	"embed"
	"html/template"
	"log"
	"net/http"
)

func getTemplates() *template.Template {
	t := template.Must(template.ParseGlob("tmpl/*.html"))
	template.Must(t.ParseGlob("tmpl/base/*.html"))

	return t
}

// NOTE(Appy): Yes globals are ugly, but live with it for now.
var (
	templates *template.Template
)

func init() {
	templates = getTemplates()
}

//go:generate npm run build
//go:embed static
var static embed.FS

func main() {
	mux := http.NewServeMux()
	mux.Handle("/static/", http.FileServer(http.FS(static)))

	hello_handler := func(w http.ResponseWriter, req *http.Request) {
		err := templates.ExecuteTemplate(w, "baseHTML", nil)

		if err != nil {
			log.Print(err)
		}
	}

	mux.HandleFunc("/", hello_handler)
	log.Println("Listening for requests at http://localhost:8000")
	log.Fatal(http.ListenAndServe(":8000", mux))
}
