package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strconv"
	"strings"

	"github.com/micpst/tinykv/pkg/hash"
	"github.com/micpst/tinykv/pkg/rpc"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
)

type ListResponse struct {
	Next string   `json:"next"`
	Keys []string `json:"keys"`
}

func (s *Server) fetchData(w http.ResponseWriter, r *http.Request) {
	key := []byte(r.URL.Path)

	data, err := s.db.Get(key, nil)
	if err == leveldb.ErrNotFound {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	volumes := strings.Split(string(data), ",")
	for _, v := range rand.Perm(len(volumes)) {
		remote := fmt.Sprintf("http://%s%s", volumes[v], hash.KeyToPath(key))

		if err := rpc.Head(remote); err == nil {
			w.Header().Set("Location", remote)
			w.WriteHeader(http.StatusFound)
			return
		}
	}

	w.WriteHeader(http.StatusNotFound)
}

func (s *Server) putData(w http.ResponseWriter, r *http.Request) {
	key := []byte(r.URL.Path)

	if r.ContentLength == 0 {
		w.WriteHeader(http.StatusLengthRequired)
		return
	}

	if _, err := s.db.Get(key, nil); err != leveldb.ErrNotFound {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	var buf bytes.Buffer
	body := io.TeeReader(r.Body, &buf)
	volumes := hash.KeyToVolumes(key, s.volumes, s.replicas)

	for i, volume := range volumes {
		if i != 0 {
			body = bytes.NewReader(buf.Bytes())
		}

		remote := fmt.Sprintf("http://%s%s", volume, hash.KeyToPath(key))
		if err := rpc.Put(remote, r.ContentLength, body); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	_ = s.db.Put(key, []byte(strings.Join(volumes, ",")), nil)

	w.WriteHeader(http.StatusCreated)
}

func (s *Server) deleteData(w http.ResponseWriter, r *http.Request) {
	key := []byte(r.URL.Path)

	data, err := s.db.Get(key, nil)
	if err == leveldb.ErrNotFound {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	_ = s.db.Delete(key, nil)

	deleteError := false
	volumes := strings.Split(string(data), ",")

	for _, volume := range volumes {
		remote := fmt.Sprintf("http://%s%s", volume, hash.KeyToPath(key))
		if err := rpc.Delete(remote); err != nil {
			deleteError = true
		}
	}

	if deleteError {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) listKeys(w http.ResponseWriter, r *http.Request) {
	key := []byte(r.URL.Path)
	slice := util.BytesPrefix(key)
	if start := r.URL.Query().Get("start"); start != "" {
		slice.Start = []byte(start)
	}

	limit := 10
	if qLimit := r.URL.Query().Get("limit"); qLimit != "" {
		parsed, err := strconv.Atoi(qLimit)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		limit = parsed
	}

	data := ListResponse{
		Next: "",
		Keys: make([]string, 0),
	}

	iter := s.db.NewIterator(slice, nil)
	defer iter.Release()

	for iter.Next() {
		if len(data.Keys) > 1000000 {
			w.WriteHeader(http.StatusRequestEntityTooLarge)
			return
		}
		if len(data.Keys) == limit {
			data.Next = string(iter.Key())
			break
		}
		data.Keys = append(data.Keys, string(iter.Key()))
	}

	response, err := json.Marshal(data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(response)
}
