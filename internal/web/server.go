package web

import (
	"net/http"

	"github.com/antonov-denis/denisurl/internal/store"
)

type Server struct {
	baseURL string
	store *store.Store
	mux   *http.ServeMux
}

func (s *Server) routes() {
	s.mux = http.NewServeMux()
	s.mux.HandleFunc("GET /health", s.handleHealth)
	s.mux.HandleFunc("GET /{code}", s.handleRedirect)
	s.mux.HandleFunc("POST /api/links", s.handleCreate)
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) { s.mux.ServeHTTP(w, r) }

func New(baseURL string, s *store.Store) *Server {
	server := &Server{baseURL: baseURL, store: s}
	server.routes()

	return server
}
