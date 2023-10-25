package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
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
func HandleFeed(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	log.Println("HANDLING FEED!")
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
	amount := p.ByName("amount")
	page := p.ByName("page")

	if len(page) == 0 || len(amount) == 0 {
		// Do nothing... I guess.. something went wrong...
		log.Println("Something went wrong while handing feed.")
		return
	}
	// Retrieve the videos.
	url := fmt.Sprintf("http://%s/api/video/feed/%s/%s", API_GATEWAY_URL, amount, page)
	log.Println("Fetching from:", url)
	resp, err := http.Get(url)
	if err != nil {
		log.Println(err)
		log.Println("Failed to request from the backend.")
		return
	}
	defer resp.Body.Close()

	// Decode the information.
	type Entry struct {
		Video        video.VideoWithUserEntry `json:"video"`
		ThumbnailURL string                   `json:"thumbnail_url"`
	}
	type ResponsePayload struct {
		Success bool    `json:"success"`
		Message string  `json:"message"`
		Entries []Entry `json:"entries"`
	}
	response := ResponsePayload{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		log.Println("Failed to decode!!")
		return
	}

	pageNumber, _ := strconv.Atoi(page)

	executor.ExecuteTemplate(w, "videos", map[string]interface{}{
		"Entries":         response.Entries,
		"Page":            pageNumber,
		"API_GATEWAY_URL": API_GATEWAY_URL,
	})
}

func GetVideoFromRank(rank int) (*video.VideoWithUserEntry, error) {
	type Entry struct {
		Video        video.VideoWithUserEntry `json:"video"`
		ThumbnailURL string                   `json:"thumbnail_url"`
	}

	type ResponsePayload struct {
		Success bool
		Message string
		Entry   Entry
	}

	res := &ResponsePayload{}

	// Perform the request.
	url := fmt.Sprintf("http://%s/api/video/rank/%d", API_GATEWAY_URL, rank-1)
	resp, err := http.Get(url)
	if err != nil || resp.StatusCode != http.StatusOK {
		log.Println("Failed to request from the backend.")
		return nil, err
	}
	defer resp.Body.Close()

	// Decode the struct
	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		log.Println("Could not decode response body for getting info by rank number.")
		return nil, err
	}

	return &res.Entry.Video, nil
}

func GetNextAndPreviousVideo(rank int) (map[string]*video.VideoWithUserEntry, error) {
	res := make(map[string]*video.VideoWithUserEntry)

	// Get previous
	if rank > 0 {
		previous, err := GetVideoFromRank(rank - 1)
		if err != nil {
			log.Println("Failed to get previous entry.")
			return nil, err
		}
		res["previous"] = previous
	} else {
		res["previous"] = nil
	}

	// Get next
	// This failing is not too bad.
	next, err := GetVideoFromRank(rank + 1)
	if err != nil {
		log.Println("Failed to get next entry.")
		res["next"] = nil
		// TODO: Watch out for this.
		return res, nil
	}
	res["next"] = next

	return res, nil
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

	url := fmt.Sprintf("http://%s/api/users/%s/videos", API_GATEWAY_URL, username)
	resp, err := http.Get(url)
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

	log.Println("Request send to backend")

	response := &struct {
		Success bool           `json:"success"`
		Message string         `json:"message"`
		Videos  []*video.Video `json:"videos"`
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

	log.Printf("Response: %+v\n", response)

	type Entry struct {
		Video        *video.Video
		ThumbnailURL string
	}
	entries := make([]Entry, 0)

	for _, v := range response.Videos {
		url := fmt.Sprintf("http://%s/api/users/%s/videos/%s/info", API_GATEWAY_URL, username, v.Key)
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
		"Videos":   entries,
		"Username": username,
	})
}

// HandleWatchPage - This serves the page for watching a video.
func HandleWatchPage(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	// Retrieve the url parameters
	username := p.ByName("username")
	videoKey := p.ByName("video") // This is the video key.
	rankStr := p.ByName("rank")
	if len(username) == 0 || len(videoKey) == 0 || len(rankStr) == 0 {
		log.Println("Something went terribly wrong.")
		return
	}

	// Request for the video information.
	url := fmt.Sprintf("http://%s/api/watch/%s/%s/info", API_GATEWAY_URL, username, videoKey)
	resp, err := http.Get(url)
	if err != nil {
		log.Println("Could not access!", err)
		return
	}
	defer resp.Body.Close()

	// Request for next and previous videos name.
	rank, _ := strconv.Atoi(rankStr)
	// TODO: Do something about this...
	nextPrev, err := GetNextAndPreviousVideo(rank + 1)
	if err != nil {
		log.Println("Get next and previous failed.")
		return
	}

	// Decode the response.
	response := &struct {
		Success   bool
		Message   string
		Video     video.Video
		Thumbnail string
	}{}
	err = json.NewDecoder(resp.Body).Decode(response)
	if err != nil {
		log.Println("Failed to decode response.")
		return
	}

	if v, ok := nextPrev["previous"]; ok && v != nil {
		log.Println("Previous Video: ", v.Name)
	}
	if v, ok := nextPrev["next"]; ok && v != nil {
		log.Println("Next Video: ", v.Name)
	}

	executor.ExecuteTemplate(w, "watch", map[string]interface{}{
		"Thumbnail":       response.Thumbnail,
		"API_GATEWAY_URL": API_GATEWAY_URL,
		"Username":        username,
		"Video":           response.Video,
		"VideoKey":        videoKey,
		"VideoName":       response.Video.Name,
		"PreviousVideo":   nextPrev["previous"],
		"NextVideo":       nextPrev["next"],
		"RankNumber":      rank,
	})
}

func HandleEditPage(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	// Retrieve the url parameters
	username := p.ByName("username")
	videoKey := p.ByName("video") // This is the video key.
	if len(username) == 0 || len(videoKey) == 0 {
		log.Println("Something went terribly wrong.")
		return
	}

	// Request for the video information.
	url := fmt.Sprintf("http://%s/api/users/%s/videos/%s/info", API_GATEWAY_URL, username, videoKey)
	resp, err := http.Get(url)
	if err != nil {
		log.Println("Could not access!", err)
		return
	}
	defer resp.Body.Close()

	// Decode the response.
	response := &struct {
		Success   bool
		Message   string
		Video     video.Video
		Thumbnail string
	}{}
	err = json.NewDecoder(resp.Body).Decode(response)
	if err != nil {
		log.Println("Failed to decode response.")
		return
	}

	executor.ExecuteTemplate(w, "watch", map[string]interface{}{
		"Thumbnail":       response.Thumbnail,
		"API_GATEWAY_URL": API_GATEWAY_URL,
		"Username":        username,
		"Video":           response.Video,
		"VideoKey":        videoKey,
		"VideoName":       response.Video.Name,
	})
}
