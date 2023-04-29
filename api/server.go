package api

import (
	"fmt"
	"log"
	"net/http"

	"github.com/micpst/tinykv/pkg/syncset"
	"github.com/syndtr/goleveldb/leveldb"
)

type Config struct {
	Db      string
	Port    int
	Volumes []string
}

type Server struct {
	db      *leveldb.DB
	locks   *syncset.SyncSet
	port    int
	volumes []string
}

func New(cfg *Config) (*Server, error) {
	db, err := leveldb.OpenFile(cfg.Db, nil)
	if err != nil {
		return nil, fmt.Errorf("LevelDB open failed %s", err)
	}
	return &Server{
		db:      db,
		locks:   syncset.New(),
		port:    cfg.Port,
		volumes: cfg.Volumes,
	}, nil
}

func (s *Server) Run() {
	log.Println("Staring master server on port", s.port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", s.port), s))
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPut || r.Method == http.MethodDelete {
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
