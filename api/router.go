package api

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter() chi.Router {
	r := chi.NewRouter()

	//standard middleware stack
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	//roytes
	r.Get("/health", healthHandler)
	r.Post("/compress", compressHandler)
	r.Post("/decompress", decompressHandler)
	r.Get("/decompress/{sessionID}/{filename}", decompressFileHandler)

	return r
}
