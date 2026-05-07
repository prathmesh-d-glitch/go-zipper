package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prathmesh-d-glitch/go-zipper/api"
)

// refractored main inspired by github.com/gmhafiz/go8

const defaultPort = "8080"
const defaultHost = ""
const defaultVersion = "0.0.1"

type Server struct {
	Version string
	host    string
	port    string

	router     http.Handler
	httpServer *http.Server
}

type Options func(*Server) error

func New(opts ...Options) *Server {
	s := defaultServer()
	for _, opt := range opts {
		if err := opt(s); err != nil {
			log.Fatalf("server option error: %v", err)
		}
	}
	return s
}

func defaultServer() *Server {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}
	return &Server{
		Version: defaultVersion,
		host:    defaultHost,
		port:    port,
		router:  api.NewRouter(),
	}
}

func WithVersion(version string) Options {
	return func(s *Server) error {
		log.Printf("Starting go-zipper version: %s\n", version)
		s.Version = version
		return nil
	}
}

func WithPort(port string) Options {
	return func(s *Server) error {
		s.port = port
		return nil
	}
}

func WithHost(host string) Options {
	return func(s *Server) error {
		s.host = host
		return nil
	}
}

func (s *Server) Init() {
	s.httpServer = &http.Server{
		Addr:              fmt.Sprintf("%s:%s", s.host, s.port),
		Handler:           s.router,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
}

func (s *Server) Run() {
	go start(s)
	gracefulShutdown(context.Background(), s)
}

func start(s *Server) {
	log.Printf("go-zipper listening on %s:%s\n", s.host, s.port)
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}

func gracefulShutdown(ctx context.Context, s *Server) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		log.Fatalf("forced shutdown: %v", err)
	}

	log.Println("Server stopped.")
}

func main() {
	s := New(
		WithVersion("1.0.0"),
		// WithPort("9090"),   // override port if needed
		// WithHost("0.0.0.0"),
	)
	s.Init()
	s.Run()
}
