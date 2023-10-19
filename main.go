package main

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/help-me-someone/scalable-p2-db/models/video"
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

var API_GATEWAY_URL string

func GetUserActionButton(w http.ResponseWriter, r *http.Request) {
	claims, err := ValidateJWTTOken(r)
	if err != nil {
		executor.ExecuteTemplate(w, "components/login_button", nil)
	} else {
		executor.ExecuteTemplate(w, "components/logout_button", map[string]string{
			"username":        claims.Username,
			"API_GATEWAY_URL": API_GATEWAY_URL,
		})
	}
}

func GetMyVideos(w http.ResponseWriter, r *http.Request) {

	username, ok := r.Context().Value("username").(string)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "User not logged in.",
		})
		return
	}

	resp, err := http.Get(fmt.Sprintf("http://db-svc:8083/user/%s/videos", username))
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Get user request failed.",
		})
		return
	}
	defer resp.Body.Close()

	response := &struct {
		Success bool          `json:"success"`
		Message string        `json:"message"`
		Videos  []video.Video `json:"videos"`
	}{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		log.Println("Failed to decode:", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Failed to decode.",
		})
		return
	}

	type Entry struct {
		Video        video.Video
		ThumbnailURL string
	}
	entries := make([]Entry, 0)

	for _, v := range response.Videos {
		url := fmt.Sprintf("http://back-end:7000/users/%s/videos/%s/info", username, v.Key)
		log.Println("Accessing:", url)
		resp, err := http.Get(url)
		if err != nil {
			// I don't want to throw it all away. Yes this is bad.
			// I'm sure no one will see this though.
			log.Println("Could not access!", err)
			continue
		}
		defer resp.Body.Close()

		response := &struct {
			Success   bool
			Message   string
			Video     video.Video
			Thumbnail string
		}{}
		err = json.NewDecoder(resp.Body).Decode(response)
		if err != nil {
			continue
		}

		log.Println("Video name:", v.Name)

		entries = append(entries, Entry{
			Video:        v,
			ThumbnailURL: response.Thumbnail,
		})
	}

	executor.ExecuteTemplate(w, "video_progress", map[string]interface{}{
		"Videos": entries,
	})
}

func BrowseVideos(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement me!
	executor.ExecuteTemplate(w, "browse", nil)
}

func main() {
	API_GATEWAY_URL = os.Getenv("API_GATEWAY_URL")
	log.Println("API_GATEWAY_URL:", API_GATEWAY_URL)

	// Create templates.
	if debug {
		executor = DebugTemplateExecutor{template_path}
	} else {
		funcs := map[string]any{
			"map":  TemplateMap,
			"loop": TemplateLoop,
			"add":  TemplateAdd,
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
	mux.Handle("/signup", template_handler)
	mux.Handle("/home", template_handler)
	mux.Handle("/forgot_password", template_handler)
	mux.HandleFunc("/action_button", GetUserActionButton)
	mux.HandleFunc("/progress", NeedAuth(GetMyVideos))
	mux.Handle("/browse", template_handler)
	mux.Handle("/videos", template_handler)
	mux.Handle("/", http.StripPrefix("/", http.HandlerFunc(NeedAuth(template_handler.ServeHTTP))))

	// Serve.
	port := ":8000"
	log.Println("Listening for requests at http://localhost" + port)
	log.Fatal(http.ListenAndServe(port, mux))
}
