package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/0990/httpproxy"
	"log"
	"net/http"
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

	stats := httpproxy.NewStats()
	stats.Run()

	if cfg.StatsAddr != "" {
		go func() {
			http.HandleFunc("/stats", func(writer http.ResponseWriter, request *http.Request) {
				log := stats.WaitStatsLog()
				writer.Write([]byte(log))
			})
			if err := http.ListenAndServe(cfg.StatsAddr, nil); err != nil {
				log.Panic(err)
			}
		}()
	}

	s := httpproxy.NewServer(cfg)
	s.AddCallBack(func(r *http.Request) {
		stats.LogReq(r)
	}, func(r *http.Request) {
		stats.LogFail(r)
	}, func(r *http.Request) {
		stats.LogRelay(r)
	})
	if err := s.ListenAndServe(); err != nil {
		fmt.Println(err)
	}
}
