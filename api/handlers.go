package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/micpst/tinykv/pkg/hash"
	"github.com/micpst/tinykv/pkg/rpc"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
)

func (s *Server) fetchData(w http.ResponseWriter, r *http.Request) {
	key := []byte(r.URL.Path)

	data, err := s.db.Get(key, nil)
	if err == leveldb.ErrNotFound {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	remote := fmt.Sprintf("http://%s%s", string(data), hash.KeyToPath(key))

	w.Header().Set("Location", remote)
	w.WriteHeader(http.StatusFound)
}

func (s *Server) putData(w http.ResponseWriter, r *http.Request) {
	key := []byte(r.URL.Path)

	// no empty values
	if r.ContentLength == 0 {
		w.WriteHeader(http.StatusLengthRequired)
		return
	}

	// check if we already have the key
	if _, err := s.db.Get(key, nil); err != leveldb.ErrNotFound {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// we don't, compute the remote URL
	volume := hash.KeyToVolume(key, s.volumes)
	remote := fmt.Sprintf("http://%s%s", volume, hash.KeyToPath(key))

	if err := rpc.Put(remote, r.ContentLength, r.Body); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// push to leveldb
	// note that the key is locked, so nobody wrote to the leveldb
	_ = s.db.Put(key, []byte(volume), nil)

	w.WriteHeader(http.StatusCreated)
}

func (s *Server) deleteData(w http.ResponseWriter, r *http.Request) {
	key := []byte(r.URL.Path)

	// delete the key
	data, err := s.db.Get(key, nil)
	if err == leveldb.ErrNotFound {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	_ = s.db.Delete(key, nil)

	remote := fmt.Sprintf("http://%s%s", string(data), hash.KeyToPath(key))
	if err := rpc.Delete(remote); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) listKeys(w http.ResponseWriter, r *http.Request) {
	key := []byte(r.URL.Path)

	iter := s.db.NewIterator(util.BytesPrefix(key), nil)
	defer iter.Release()

	keys := make([]string, 0)
	for iter.Next() {
		keys = append(keys, string(iter.Key()))
		if len(keys) > 1000000 {
			w.WriteHeader(http.StatusRequestEntityTooLarge)
			return
		}
	}

	str, err := json.Marshal(keys)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(str)
}
