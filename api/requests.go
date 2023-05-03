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
	From []string
	To   []string
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

func commonVolumes(oldVolumes []string, newVolumes []string) map[string]struct{} {
	common := make(map[string]struct{})
	set := make(map[string]struct{}, len(oldVolumes))

	for _, v := range oldVolumes {
		set[v] = struct{}{}
	}

	for _, v := range newVolumes {
		if _, ok := set[v]; ok {
			common[v] = struct{}{}
		}
	}

	return common
}

func (s *Server) rebalance(r *RebalanceRequest) bool {
	if len(r.From) == 0 {
		return false
	}

	common := commonVolumes(r.From, r.To)
	if len(common) == len(r.To) {
		return true
	}

	path := hash.KeyToPath(r.Key)

	for _, v := range r.From {
		remote := fmt.Sprintf("http://%s%s", v, path)
		if err := rpc.Head(remote); err != nil {
			return false
		}
	}

	remote := fmt.Sprintf("http://%s%s", r.From[0], path)
	data, err := rpc.Get(remote)
	if err != nil {
		return false
	}

	for _, v := range r.To {
		if _, ok := common[v]; !ok {
			remote := fmt.Sprintf("http://%s%s", v, path)
			if err := rpc.Put(remote, int64(len(data)), strings.NewReader(data)); err != nil {
				return false
			}
		}
	}

	if err := s.db.Put(r.Key, []byte(strings.Join(r.To, ",")), nil); err != nil {
		return false
	}

	for _, v := range r.From {
		if _, ok := common[v]; !ok {
			remote := fmt.Sprintf("http://%s%s", v, path)
			_ = rpc.Delete(remote)
		}
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
