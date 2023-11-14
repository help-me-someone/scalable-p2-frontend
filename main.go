package main

import (
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"

	"github.com/julienschmidt/httprouter"
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

var (
	API_GATEWAY_URL string
	BACKEND_URL     string
)

func loadEnvs() {
	API_GATEWAY_URL = os.Getenv("API_GATEWAY_URL")
	BACKEND_URL = os.Getenv("BACKEND_URL")
}

func main() {
	// Load environment variables.
	loadEnvs()

	log.Println("API_GATEWAY_URL:", API_GATEWAY_URL)

	// Create templates.
	if debug {
		executor = DebugTemplateExecutor{template_path}
		API_GATEWAY_URL = "toktik.localhost"
		BACKEND_URL = "localhost:7000"
	} else {
		funcs := map[string]any{
			"map":  TemplateMap,
			"loop": TemplateLoop,
			"add":  TemplateAdd,
			"sub":  TemplateSub,
			"mul":  TemplateMult,
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

	f := fs.FS(static)
	v, _ := fs.Sub(f, "static")
	mux := httprouter.New()
	mux.ServeFiles("/static/*filepath", http.FS(v))

	// Feel free to add this later.
	favicon_handler := func(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
		// Do nothing... unless...
	}

	template_handler := &URLContext{
		API_GATEWAY_URL: API_GATEWAY_URL,
	}

	// Handlers.
	mux.GET("/favicon.ico", favicon_handler)

	log.Println("API_GATEWAY_URL gateway url:", API_GATEWAY_URL)
	mux.GET("/login", template_handler.ServeHTTP)
	mux.GET("/signup", template_handler.ServeHTTP)
	mux.GET("/home", template_handler.ServeHTTP)
	mux.GET("/socket", template_handler.ServeHTTP)
	mux.GET("/forgot_password", template_handler.ServeHTTP)
	mux.GET("/action_button", GetUserActionButton)
	mux.GET("/progress", HttpRouterNeedAuth(GetMyVideos))
	mux.GET("/videos", template_handler.ServeHTTP)
	mux.GET("/watch/:username/:video/:rank", HttpRouterNeedAuth(HandleWatchPage))
	mux.GET("/edit/:username/:video", HandleEditPage)
	mux.GET("/feed/:amount/:page", HandleFeed)
	mux.GET("/comment/:video/:amount/:page", HandleFeed)
	mux.NotFound = http.StripPrefix("/", HttpNeedAuth(template_handler.ServeHTTP))

	// Serve.
	port := ":8000"
	log.Println("Listening for requests at http://localhost" + port)
	log.Fatal(http.ListenAndServe(port, mux))
}
