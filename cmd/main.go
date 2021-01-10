package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/0990/httpproxy"
	"log"
	"os"
)

var confFile = flag.String("c", "config.json", "config file")

func main() {
	file, err := os.Open(*confFile)
	if err != nil {
		log.Fatalln(err)
	}

	var cfg httpproxy.Config
	err = json.NewDecoder(file).Decode(&cfg)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println("config:", cfg)

	s := httpproxy.NewServer(cfg)
	if err := s.ListenAndServe(); err != nil {
		fmt.Println(err)
	}
}
