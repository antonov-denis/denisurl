package web

import (
	"net/http"

	"github.com/antonov-denis/denisurl/internal/limiter"
	"github.com/antonov-denis/denisurl/internal/store"
)

type Server struct {
	baseURL string
	store   *store.Store
	limiter *limiter.Limiter
	mux     *http.ServeMux
}

func (s *Server) routes() {
	s.mux = http.NewServeMux()
	s.mux.Handle("GET /static/", staticFiles())

	s.mux.HandleFunc("GET /health", s.handleHealth)

	s.mux.HandleFunc("GET /{$}", s.handleIndex)
	s.mux.HandleFunc("GET /{code}", s.handleRedirect)
	s.mux.Handle("POST /api/links", s.limit(http.HandlerFunc(s.handleCreate)))
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) { s.mux.ServeHTTP(w, r) }

func New(baseURL string, s *store.Store, l *limiter.Limiter) *Server {
	server := &Server{baseURL: baseURL, store: s, limiter: l}
	server.routes()

	return server
}
