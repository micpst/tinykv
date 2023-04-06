package main

import (
	"flag"
	"log"
	"strings"

	"github.com/micpst/tinykv/api"
)

func main() {
	db := flag.String("db", "", "Path to leveldb")
	port := flag.Int("port", 3000, "Port for the server to listen on")
	volumes := flag.String("volumes", "", "Volumes to use for storage (comma separated)")
	flag.Parse()

	s, err := api.New(&api.Config{
		Db:      *db,
		Port:    *port,
		Volumes: strings.Split(*volumes, ","),
	})
	if err != nil {
		log.Println("Failed to create server", err)
	}
	s.Run()
}
