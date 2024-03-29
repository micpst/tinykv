package main

import (
	"flag"
	"fmt"
	"strings"

	"github.com/micpst/tinykv/api"
)

const (
	RebalanceCmd = "rebalance"
	RebuildCmd   = "rebuild"
	RunCmd       = "run"
)

func main() {
	cmd := flag.String("cmd", RunCmd, "Master command to execute: \"run\", \"rebalance\" or \"rebuild\"")
	db := flag.String("db", "", "Path to leveldb")
	port := flag.Int("p", 3000, "Port for the server to listen on")
	replicas := flag.Int("replicas", 3, "Amount of replicas to make of the data")
	volumes := flag.String("volumes", "", "Volumes to use for storage (comma separated)")
	flag.Parse()

	s, err := api.New(&api.Config{
		Db:       *db,
		Port:     *port,
		Replicas: *replicas,
		Volumes:  strings.Split(*volumes, ","),
	})
	if err != nil {
		panic(err)
	}

	switch *cmd {
	case RebalanceCmd:
		s.Rebalance()
	case RebuildCmd:
		s.Rebuild()
	case RunCmd:
		s.Run()
	default:
		fmt.Println("Unknown command", *cmd)
		flag.PrintDefaults()
	}
}
