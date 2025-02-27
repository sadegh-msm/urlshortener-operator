package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func generateShortURL() string {
	rand.Seed(time.Now().UnixNano())
	b := make([]rune, 4)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

type URLRecord struct {
	LongURL  string
	ExpireAt *time.Time
}

type URLStore struct {
	mu    sync.Mutex
	store map[string]URLRecord
	count map[string]int
}

func NewURLStore() *URLStore {
	return &URLStore{
		store: make(map[string]URLRecord),
		count: make(map[string]int),
	}
}

func (u *URLStore) ShortenURL(w http.ResponseWriter, r *http.Request) {
	var req struct {
		LongURL  string `json:"long_url"`
		ExpireAt string `json:"expire_at,omitempty"` // "2025-03-01T15:04:05Z"
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	var expireAt *time.Time

	if req.ExpireAt != "" {
		loc := time.FixedZone("Local", 3*3600+30*60)
		expireStr := req.ExpireAt
		if strings.HasSuffix(expireStr, "Z") {
			expireStr = strings.TrimSuffix(expireStr, "Z")
		}
		parsedTime, err := time.ParseInLocation("2006-01-02T15:04:05", expireStr, loc)
		if err != nil {
			http.Error(w, "Invalid expiration date format. Use format: YYYY-MM-DDTHH:MM:SS", http.StatusBadRequest)
			return
		}
		expireAt = &parsedTime
	}

	u.mu.Lock()
	shortURL := generateShortURL()
	u.store[shortURL] = URLRecord{
		LongURL:  req.LongURL,
		ExpireAt: expireAt,
	}
	u.mu.Unlock()

	response := map[string]string{"short_url": shortURL}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (u *URLStore) Redirect(w http.ResponseWriter, r *http.Request) {
	shortURL := r.URL.Path[1:]

	u.mu.Lock()
	record, exists := u.store[shortURL]
	if exists {
		if record.ExpireAt != nil && time.Until(*record.ExpireAt) < 0 {
			u.mu.Unlock()
			http.Error(w, "URL expired", http.StatusGone)
			return
		}
		u.count[shortURL]++
	}
	u.mu.Unlock()

	if !exists {
		http.NotFound(w, r)
		return
	}

	http.Redirect(w, r, record.LongURL, http.StatusFound)
}

func (u *URLStore) GetCount(w http.ResponseWriter, r *http.Request) {
	shortURL := r.URL.Path[len("/count/"):]

	u.mu.Lock()
	count, exists := u.count[shortURL]
	u.mu.Unlock()

	if !exists {
		http.NotFound(w, r)
		return
	}

	response := map[string]int{"click_count": count}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (u *URLStore) CheckValidity(w http.ResponseWriter, r *http.Request) {
	shortURL := r.URL.Path[len("/valid/"):]

	u.mu.Lock()
	record, exists := u.store[shortURL]
	u.mu.Unlock()

	if !exists {
		http.NotFound(w, r)
		return
	}

	if record.ExpireAt != nil {
		log.Println("Time until expiration:", time.Until(*record.ExpireAt))
	}

	isValid := true
	if record.ExpireAt != nil && time.Until(*record.ExpireAt) < 0 {
		isValid = false
	}

	response := map[string]bool{"is_valid": isValid}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func main() {
	store := NewURLStore()

	http.HandleFunc("/shorten", store.ShortenURL)
	http.HandleFunc("/count/", store.GetCount)
	http.HandleFunc("/valid/", store.CheckValidity)
	http.HandleFunc("/", store.Redirect)

	log.Println("start listening on port 8080")

	http.ListenAndServe(":8080", nil)
}
