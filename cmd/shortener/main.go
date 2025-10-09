package main

import (
	"log"
	"net/http"

	"github.com/Guldana11/shortener/internal/handler"
)

func main() {
	h := handler.NewURLHandler()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == "/" {
			h.PostHandler(w, r)
			return
		}
		if r.Method == http.MethodGet {
			h.GetHandler(w, r)
			return
		}
		http.Error(w, "bad request", http.StatusBadRequest)
	})

	log.Println("Server is running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
