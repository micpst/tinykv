package api

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/micpst/tinykv/pkg/hash"
	"github.com/micpst/tinykv/pkg/rpc"
	"github.com/syndtr/goleveldb/leveldb"
)

type RebalanceRequest struct {
	Key  []byte
	From []string
	To   []string
}

type RebuildRequest struct {
	Key    []byte
	Volume string
}

type File struct {
	Name  string
	Type  string
	Mtime string
}

func commonVolumes(v1 []string, v2 []string) map[string]struct{} {
	common := make(map[string]struct{})
	set := make(map[string]struct{}, len(v1))

	for _, v := range v1 {
		set[v] = struct{}{}
	}

	for _, v := range v2 {
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

func fetchFiles(url string) []File {
	var files []File

	data, err := rpc.Get(url)
	if err != nil {
		return files
	}

	if err := json.Unmarshal([]byte(data), &files); err != nil {
		return nil
	}

	return files
}

func (s *Server) rebuild(r *RebuildRequest) bool {
	if ok := s.locks.Add(string(r.Key)); !ok {
		return false
	}
	defer s.locks.Remove(string(r.Key))

	volumes := []string{r.Volume}

	data, err := s.db.Get(r.Key, nil)
	if err != leveldb.ErrNotFound {
		volumes = append(volumes, strings.Split(string(data), ",")...)
	}

	if err := s.db.Put(r.Key, []byte(strings.Join(volumes, ",")), nil); err != nil {
		return false
	}

	fmt.Println(string(r.Key), r.Volume)

	return true
}
