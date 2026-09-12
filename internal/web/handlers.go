package web

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/antonov-denis/denisurl/internal/store"
)

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.Write([]byte("Ok!"))
}

type indexView struct {
	BaseURL string
}

func (s *Server) handleIndex(w http.ResponseWriter, _ *http.Request) {
	s.render(w, http.StatusOK, "index.html", indexView{BaseURL: s.baseURL})
}

func (s *Server) handleRedirect(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")

	target, err := s.store.GetTarget(r.Context(), code)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			http.Error(w, "Not Found", http.StatusNotFound)
		} else {
			http.Error(w, "Something Went Wrong", http.StatusInternalServerError)
		}
		return
	}

	http.Redirect(w, r, target, 301)
}

type createBody struct {
	TargetURL string `json:"target"`
}

type createResponse struct {
	ShortURL string `json:"short_url"`
}

func (s *Server) handleCreate(w http.ResponseWriter, r *http.Request) {
	body := r.Body
	defer body.Close()

	var parsedBody createBody
	err := json.NewDecoder(body).Decode(&parsedBody)
	if err != nil {
		slog.Error(err.Error())
		http.Error(w, "couldn't parse request", http.StatusBadRequest)
		return
	}

	target, err := normalizeURL(parsedBody.TargetURL)
	if err != nil {
		http.Error(w, "target must be a http or https URL", http.StatusBadRequest)
		return
	}

	code := rand.Text()[:7]
	if err := s.store.CreateTarget(r.Context(), code, target); err != nil {
		http.Error(w, "Something Went Wrong", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createResponse{ShortURL: s.baseURL + "/" + code})
}
