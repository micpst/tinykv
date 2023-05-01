package api

import (
	"fmt"
	"strings"

	"github.com/micpst/tinykv/pkg/hash"
	"github.com/micpst/tinykv/pkg/rpc"
)

type rebalanceRequest struct {
	key  []byte
	from string
	to   string
}

func (s *Server) rebalance(r *rebalanceRequest) bool {
	if r.from == r.to {
		return true
	}

	path := hash.KeyToPath(r.key)
	remoteFrom := fmt.Sprintf("http://%s%s", r.from, path)
	remoteTo := fmt.Sprintf("http://%s%s", r.to, path)

	data, err := rpc.Get(remoteFrom)
	if err != nil {
		return false
	}

	if err := rpc.Put(remoteTo, int64(len(data)), strings.NewReader(data)); err != nil {
		return false
	}

	if err := s.db.Put(r.key, []byte(r.to), nil); err != nil {
		return false
	}

	if err = rpc.Delete(remoteFrom); err != nil {
		return false
	}

	return true
}
