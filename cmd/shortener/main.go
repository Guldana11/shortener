package main

import (
	"log"
	"net/http"

	"github.com/Guldana11/shortener/internal/handler"
	"github.com/go-chi/chi/v5"
)

func main() {
	h := handler.NewURLHandler()
	r := chi.NewRouter()

	r.Post("/", h.PostHandler)

	r.Get("/{id}", h.GetHandler)

	log.Println("Server is running on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
