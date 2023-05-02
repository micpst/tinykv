package api

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/micpst/tinykv/pkg/hash"
	"github.com/micpst/tinykv/pkg/rpc"
)

type RebalanceRequest struct {
	Key  []byte
	From string
	To   string
}

type RebuildRequest struct {
	Volume string
	Url    string
}

type File struct {
	Name  string
	Type  string
	Mtime string
}

func (s *Server) rebalance(r *RebalanceRequest) bool {
	if r.From == r.To {
		return true
	}

	path := hash.KeyToPath(r.Key)
	remoteFrom := fmt.Sprintf("http://%s%s", r.From, path)
	remoteTo := fmt.Sprintf("http://%s%s", r.To, path)

	data, err := rpc.Get(remoteFrom)
	if err != nil {
		return false
	}

	if err := rpc.Put(remoteTo, int64(len(data)), strings.NewReader(data)); err != nil {
		return false
	}

	if err := s.db.Put(r.Key, []byte(r.To), nil); err != nil {
		return false
	}

	if err = rpc.Delete(remoteFrom); err != nil {
		return false
	}

	return true
}

func (s *Server) rebuild(r *RebuildRequest) bool {
	data, err := rpc.Get(r.Url)
	if err != nil {
		return false
	}

	var files []File
	if err := json.Unmarshal([]byte(data), &files); err != nil {
		return false
	}

	for _, file := range files {
		key, err := base64.StdEncoding.DecodeString(file.Name)
		if err != nil {
			return false
		}

		if err := s.db.Put(key, []byte(r.Volume), nil); err != nil {
			return false
		}

		fmt.Println(string(key), r.Volume)
	}

	return true
}
