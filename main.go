package main

import (
	"log"
	"net"
	"os"
	"sync"
	"time"
)

func test() {
	conn, err := net.Dial("udp", "localhost:12345")
	if err != nil {
		log.Fatalf("error: unable to dial in %s\n", err.Error())
	}
	defer conn.Close()
	conn.Write([]byte("PUT key value"))
	var buf [1024]byte
	_, err = conn.Read(buf[:])
	if err != nil {
		log.Fatalf("error: unable to read: %s\n", err.Error())
	}
	log.Println(string(buf[:]))
}

func main() {
	m := NewMap()
	var wg sync.WaitGroup
	wg.Add(1)
	us, err := NewUDPMapServer("localhost", 12345, &wg, m, true)
	if err != nil {
		log.Printf("error: unable to start the UDP server: %s\n", err.Error())
		os.Exit(1)
	}
	go us.Serve()
	ts, err := NewTCPMapServer("localhost", 12346, &wg, m, true)
	if err != nil {
		log.Printf("error: unable to start the TCP server: %s\n", err.Error())
		os.Exit(1)
	}
	go ts.Serve()
	hs, err := NewHTTPMapServer("localhost", 8080, &wg, m, true)
	if err != nil {
		log.Printf("error: unable to start the HTTP server: %s\n", err.Error())
		os.Exit(1)
	}
	go hs.Serve()
	time.Sleep(1 * time.Second)
	go test()
	wg.Wait()
}
