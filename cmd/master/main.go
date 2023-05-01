package main

import (
	"flag"
	"fmt"
	"strings"

	"github.com/micpst/tinykv/api"
)

const (
	RebalanceCmd = "rebalance"
	RunCmd       = "run"
)

func main() {
	cmd := flag.String("cmd", RunCmd, "Master command to execute: \"run\" or \"rebalance\"")
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
		panic(err)
	}

	switch *cmd {
	case RunCmd:
		s.Run()
	case RebalanceCmd:
		s.Rebalance()
	default:
		fmt.Println("Unknown command", *cmd)
		flag.PrintDefaults()
	}
}
