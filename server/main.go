package main

import (
	"dmap"
	"log"
	"os"
	"sync"
	"time"
)

func main() {
	m := dmap.NewMap()
	var wg sync.WaitGroup
	wg.Add(3)
	us, err := dmap.NewUDPMapServer("localhost", 12345, &wg, m, true)
	if err != nil {
		log.Printf("error: unable to start the UDP server: %s\n", err.Error())
		os.Exit(1)
	}
	go us.Serve()
	ts, err := dmap.NewTCPMapServer("localhost", 12346, &wg, m, true)
	if err != nil {
		log.Printf("error: unable to start the TCP server: %s\n", err.Error())
		os.Exit(1)
	}
	go ts.Serve()
	hs, err := dmap.NewHTTPMapServer("localhost", 8080, &wg, m, true)
	if err != nil {
		log.Printf("error: unable to start the HTTP server: %s\n", err.Error())
		os.Exit(1)
	}
	go hs.Serve()
	time.Sleep(1 * time.Second)
	wg.Wait()
}
