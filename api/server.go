package api

import (
	"fmt"
	"log"
	"net/http"

	"github.com/micpst/tinykv/pkg/hash"
	"github.com/micpst/tinykv/pkg/rpc"
	"github.com/syndtr/goleveldb/leveldb"
)

type Config struct {
	Db      string
	Port    int
	Volumes []string
}

type Server struct {
	db      *leveldb.DB
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
		port:    cfg.Port,
		volumes: cfg.Volumes,
	}, nil
}

func (s *Server) Run() {
	log.Println("Staring master server on port", s.port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", s.port), s))
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	key := []byte(r.URL.Path)

	switch r.Method {
	case http.MethodGet, http.MethodHead:
		data, err := s.db.Get(key, nil)
		if err == leveldb.ErrNotFound {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		remote := fmt.Sprintf("http://%s%s", string(data), hash.KeyToPath(key))
		w.Header().Set("Location", remote)
		w.WriteHeader(http.StatusFound)

	case http.MethodPut:
		// no empty values
		if r.ContentLength == 0 {
			w.WriteHeader(http.StatusLengthRequired)
			return
		}

		// check if we already have the key
		if _, err := s.db.Get(key, nil); err != leveldb.ErrNotFound {
			w.WriteHeader(http.StatusConflict)
			return
		}

		// we don't, compute the remote URL
		volume := hash.KeyToVolume(key, s.volumes)
		remote := fmt.Sprintf("http://%s%s", volume, hash.KeyToPath(key))

		if !rpc.Put(remote, r.ContentLength, r.Body) {
			rpc.Delete(remote)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// note, this currently is a race
		// TODO: put get and put in the same transaction
		if _, err := s.db.Get(key, nil); err != leveldb.ErrNotFound {
			rpc.Delete(remote)
			w.WriteHeader(http.StatusConflict)
			return
		}

		// push to leveldb
		err := s.db.Put(key, []byte(volume), nil)
		if err != nil {
			return
		}
		w.WriteHeader(http.StatusCreated)

	case http.MethodDelete:
		// delete the key
		data, err := s.db.Get(key, nil)
		if err == leveldb.ErrNotFound {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		_ = s.db.Delete(key, nil)

		remote := fmt.Sprintf("http://%s%s", string(data), hash.KeyToPath(key))
		if !rpc.Delete(remote) {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
