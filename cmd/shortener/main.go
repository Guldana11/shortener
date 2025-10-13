package main

import (
	"log"
	"net/http"

	"github.com/Guldana11/shortener/internal/config"
	"github.com/Guldana11/shortener/internal/handler"
	"github.com/go-chi/chi/v5"
)

func main() {
	cfg := config.Init()

	h := handler.NewURLHandler(cfg.BaseURL)
	r := chi.NewRouter()

	r.Post("/", h.PostHandler)
	r.Get("/{id}", h.GetHandler)

	log.Printf("Server is running on %s\n", cfg.Address)
	log.Fatal(http.ListenAndServe(cfg.Address, r))
}
