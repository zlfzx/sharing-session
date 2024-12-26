package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	gonanoid "github.com/matoous/go-nanoid"
)

//go:embed views/index.html
var indexHTML []byte

type Progress struct {
	Percent int    `json:"percent"`
	Status  string `json:"status"`
	URL     string `json:"url"`
}

type Task struct {
	Progress *Progress
	Channel  chan *Progress
}

var (
	progressMap = make(map[int]*Progress)
	mu          sync.Mutex
)

func main() {

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(indexHTML)
	})
	r.Get("/download", downloadHandler)
	r.Get("/progress-download/{id}", progressHandler)

	fmt.Println("Server running on :8080")
	http.ListenAndServe(":8080", r)
	// http.ListenAndServeTLS(":8080", "localhost.pem", "localhost-key.pem", r)
}

func downloadHandler(w http.ResponseWriter, r *http.Request) {
	// Start download process
	id, err := gonanoid.Generate("1234567890", 8)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	taskId, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Buat task baru dengan channel
	progress := &Progress{
		Percent: 0,
		Status:  "in_progress",
	}

	mu.Lock()
	progressMap[taskId] = progress
	mu.Unlock()

	// Jalankan proses pembuatan file
	go startFileGeneration(taskId)

	response := map[string]string{"task_id": id}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func startFileGeneration(taskID int) {
	for percent := 0; percent <= 100; percent += 5 {
		time.Sleep(time.Second)

		// simulate delay
		switch percent {
		case 25:
			time.Sleep(3 * time.Second)
		case 50:
			time.Sleep(3 * time.Second)
		case 75:
			time.Sleep(4 * time.Second)
		}

		mu.Lock()
		progressMap[taskID].Percent = percent

		if percent == 100 {
			progressMap[taskID].Status = "done"
			progressMap[taskID].URL = "http://example.com/file.zip"
		}

		mu.Unlock()
	}
}

func progressHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	taskId, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Setup SSE response
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Transfer-Encoding", "chunked")

	var currentPercent int
	var clientGone = r.Context().Done()

	for {
		select {
		case <-clientGone:
			return
		default:
			// time.Sleep(1 * time.Second)

			mu.Lock()
			progress, exists := progressMap[taskId]
			mu.Unlock()

			if !exists {
				fmt.Fprintf(w, "event: error\ndata: {\"error\":\"task_id not found\"}\n\n")
				w.(http.Flusher).Flush()
				return
			}

			if progress.Percent == currentPercent {
				continue
			}

			currentPercent = progress.Percent

			// Kirim data progress ke client
			data, _ := json.Marshal(progress)
			fmt.Fprintf(w, "data: %s\n\n", data)
			w.(http.Flusher).Flush()

			// Hentikan jika proses selesai
			if progress.Status == "done" {
				return
			}
		}
	}
}
