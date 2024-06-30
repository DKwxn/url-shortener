package main

import (
    "encoding/json"
    "fmt"
    "log"
    "math/rand"
    "net/http"
    "time"

    "github.com/dkwxn/url-shortener/internal/storage"
)

var (
    urlStore *storage.Storage
    letters  = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
)

func init() {
    time.Now().UnixNano()
}

func generateShortURL(n int) string {
    b := make([]rune, n)
    for i := range b {
        b[i] = letters[rand.Intn(len(letters))]
    }
    return string(b)
}

func shortenURLHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
        return
    }

    var req struct {
        URL string `json:"url"`
    }

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    shortURL := generateShortURL(6)
    if err := urlStore.Store(shortURL, req.URL); err != nil {
        http.Error(w, "Could not save URL", http.StatusInternalServerError)
        return
    }

    resp := map[string]string{
        "short_url": fmt.Sprintf("http://localhost:8080/%s", shortURL),
    }
    json.NewEncoder(w).Encode(resp)
}

func redirectHandler(w http.ResponseWriter, r *http.Request) {
    shortURL := r.URL.Path[1:]
    if originalURL, ok := urlStore.Load(shortURL); ok {
        http.Redirect(w, r, originalURL, http.StatusFound)
    } else {
        http.Error(w, "URL not found", http.StatusNotFound)
    }
}

func main() {
    var err error
    urlStore, err = storage.NewStorage("urls.json")
    if err != nil {
        log.Fatalf("Could not create storage: %s\n", err)
    }

    http.HandleFunc("/shorten", shortenURLHandler)
    http.HandleFunc("/", redirectHandler)

    log.Println("Starting server on :8080")
    if err := http.ListenAndServe(":8080", nil); err != nil {
        log.Fatalf("Could not start server: %s\n", err)
    }
}
