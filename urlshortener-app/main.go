package main

import (
	"log"
	"net/http"
	"urlshortener/handlers"
)

func main() {
	store := handlers.NewURLStore()

	http.HandleFunc("/shorten", store.ShortenURL)
	http.HandleFunc("/count/", store.GetCount)
	http.HandleFunc("/valid/", store.CheckValidity)
	http.HandleFunc("/", store.Redirect)

	log.Println("start listening on port 8080")

	http.ListenAndServe(":8080", nil)
}
