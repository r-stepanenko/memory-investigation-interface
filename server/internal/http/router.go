package http

import (
	"net/http"
	"strings"
)

func (s *Server) SearchRouter(w http.ResponseWriter, r *http.Request) {
	if strings.Contains(r.URL.Path, "/candidates/") {
		s.ExplainHandler(w, r)
		return
	}

	s.SearchResultHandler(w, r)
}