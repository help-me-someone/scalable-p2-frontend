package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/help-me-someone/scalable-p2-db/models/video"
	"github.com/julienschmidt/httprouter"
)

type URLContext struct {
	API_GATEWAY_URL string
}

func (u *URLContext) ServeHTTP(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
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

type BackwardCompatibleHandler struct {
	next http.HandlerFunc
}

func HttpNeedAuth(handler httprouter.Handle) http.HandlerFunc {
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
		handler(wr, re.WithContext(ctx), nil)
	}
}

func HttpRouterNeedAuth(handler httprouter.Handle) httprouter.Handle {
	return func(wr http.ResponseWriter, re *http.Request, p httprouter.Params) {

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
		handler(wr, re.WithContext(ctx), p)
	}
}

// We perform a feed call to the backend and pass the data into our template.
func HandleFeed(w http.ResponseWriter, r *http.Request) {
	type EntryType struct {
		Video        video.VideoWithUserEntry `json:"video"`
		ThumbnailURL string                   `json:"thumbnail_url"`
	}
	type PayloadType struct {
		Success bool        `json:"success"`
		Message string      `json:"message"`
		Entries []EntryType `json:"entries"`
	}

	// Retrieve params.

	// Retrieve the information.
	// resp, err := http.Get()
}

func GetUserActionButton(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
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

// TODO: Move this to the backend!!
func GetMyVideos(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

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
