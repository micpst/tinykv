package api

import (
	"net/http"
	"strings"
)

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if len(r.URL.RawQuery) > 0 && r.Method == http.MethodGet {
		s.dispatchQuery(w, r)
	} else {
		s.dispatchMethod(w, r)
	}
}

func (s *Server) dispatchMethod(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPut, http.MethodDelete:
		if ok := s.locks.Add(r.URL.Path); !ok {
			w.WriteHeader(http.StatusConflict)
			return
		}
		defer s.locks.Remove(r.URL.Path)
	}

	switch r.Method {
	case http.MethodGet, http.MethodHead:
		s.fetchData(w, r)
	case http.MethodPut:
		s.putData(w, r)
	case http.MethodDelete:
		s.deleteData(w, r)
	}
}

func (s *Server) dispatchQuery(w http.ResponseWriter, r *http.Request) {
	operation := strings.Split(r.URL.RawQuery, "&")[0]
	switch operation {
	case "list":
		s.listKeys(w, r)
	default:
		w.WriteHeader(http.StatusForbidden)
	}
}
