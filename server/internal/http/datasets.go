package http

import (
	"encoding/json"
	"net/http"
)

func (s *Server) DatasetsHandler(w http.ResponseWriter, r *http.Request) {
    Cors(w)

    if r.Method == http.MethodOptions {
        return
    }

    w.Header().Set("Content-Type", "application/json")

    names := make([]string, 0)

    for name := range s.Datasets {
        names = append(names, name)
    }

    json.NewEncoder(w).Encode(names)
}