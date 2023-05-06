package api

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/micpst/tinykv/pkg/hash"
	"github.com/micpst/tinykv/pkg/syncset"
	"github.com/syndtr/goleveldb/leveldb"
)

type Config struct {
	Db       string
	Port     int
	Replicas int
	Volumes  []string
}

type Server struct {
	db       *leveldb.DB
	locks    *syncset.SyncSet
	port     int
	replicas int
	volumes  []string
}

func New(cfg *Config) (*Server, error) {
	db, err := leveldb.OpenFile(cfg.Db, nil)
	if err != nil {
		return nil, fmt.Errorf("LevelDB open failed %s", err)
	}
	return &Server{
		db:       db,
		locks:    syncset.New(),
		port:     cfg.Port,
		replicas: cfg.Replicas,
		volumes:  cfg.Volumes,
	}, nil
}

func (s *Server) Run() {
	log.Println("Staring master server on port", s.port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", s.port), s))
}

func (s *Server) Rebalance() {
	log.Println("Rebalancing to", s.volumes)

	var wg sync.WaitGroup
	requests := make(chan *RebalanceRequest, 20000)

	for i := 0; i < 16; i++ {
		go func() {
			for r := range requests {
				s.rebalance(r)
				wg.Done()
			}
		}()
	}

	iter := s.db.NewIterator(nil, nil)
	defer iter.Release()

	for iter.Next() {
		wg.Add(1)

		key := make([]byte, len(iter.Key()))
		copy(key, iter.Key())

		oldVolumes := strings.Split(string(iter.Value()), ",")
		newVolumes := hash.KeyToVolumes(key, s.volumes, s.replicas)

		requests <- &RebalanceRequest{
			Key:  key,
			From: oldVolumes,
			To:   newVolumes,
		}
	}

	close(requests)
	wg.Wait()
}

func (s *Server) Rebuild() {
	log.Println("Rebuilding on", s.volumes)

	var wg sync.WaitGroup
	requests := make(chan *RebuildRequest, 20000)

	for i := 0; i < 128; i++ {
		go func() {
			for r := range requests {
				s.rebuild(r)
				wg.Done()
			}
		}()
	}

	iter := s.db.NewIterator(nil, nil)
	defer iter.Release()

	for iter.Next() {
		_ = s.db.Delete(iter.Key(), nil)
	}

	for _, volume := range s.volumes {
		stack := []string{""}

		for len(stack) > 0 {
			dir := stack[len(stack)-1]
			stack = stack[:len(stack)-1]

			files := fetchFiles(fmt.Sprintf("http://%s/%s", volume, dir))
			for _, file := range files {
				if file.Type == "directory" {
					stack = append(stack, fmt.Sprintf("%s/%s", dir, file.Name))
				} else {
					requests <- &RebuildRequest{
						Key:    []byte(file.Name),
						Volume: volume,
					}
				}
			}
		}
	}

	close(requests)
	wg.Wait()
}
