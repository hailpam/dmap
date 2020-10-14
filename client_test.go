package dmap

import (
	"sync"
	"testing"
	"time"
)

func TestUDPMapClient(t *testing.T) {
	m := NewMap()
	var wg sync.WaitGroup
	wg.Add(1)
	us, err := NewUDPMapServer("localhost", 12346, &wg, m, true)
	if err != nil {
		t.Logf("error: unable to start the UDP server: %s\n", err.Error())
		t.Fail()
	}
	go us.Serve()
	time.Sleep(1 * time.Second)
	uc := NewUDPMapClient("localhost", 12346)
	err = uc.Dial()
	if err != nil {
		t.Logf("error: unable to dial in: %s\n", err.Error())
		t.Fail()
	}
	err = uc.Put("key12345", []byte("value12345"))
	if err != nil {
		t.Logf("error: unable to store: %s\n", err.Error())
		t.Fatal()
	}
	b, err := uc.Get("key12345")
	if err != nil {
		t.Logf("error: unable to retrieve: %s\n", err.Error())
		t.Fail()
	}
	if string(b) != "value12345" {
		t.Logf("error: unexpected value: %s\n", string(b))
		t.Fail()
	}
	s, err := uc.Size()
	if err != nil {
		t.Logf("error: unable to retrieve the size: %s\n", err.Error())
		t.Fail()
	}
	err = uc.Clear()
	if err != nil {
		t.Logf("error: unable to retrieve the clear: %s\n", err.Error())
		t.Fail()
	}
	s, err = uc.Size()
	if err != nil {
		t.Logf("error: unable to retrieve the size: %s\n", err.Error())
		t.Fail()
	}
	if s != 0 {
		t.Logf("error: enexpected size: %d\n", s)
		t.Fail()
	}
	uc.Close()
}

func TestTCPMapClient(t *testing.T) {
	m := NewMap()
	var wg sync.WaitGroup
	wg.Add(1)
	us, err := NewTCPMapServer("localhost", 12345, &wg, m, true)
	if err != nil {
		t.Logf("error: unable to start the TCP server: %s\n", err.Error())
		t.Fail()
	}
	go us.Serve()
	time.Sleep(1 * time.Second)
	uc := NewTCPMapClient("localhost", 12345)
	err = uc.Dial()
	if err != nil {
		t.Logf("error: unable to dial in: %s\n", err.Error())
		t.Fail()
	}
	err = uc.Put("key12345", []byte("value12345"))
	if err != nil {
	 	t.Logf("error: unable to store: %s\n", err.Error())
	 	t.Fatal()
	 }
	b, err := uc.Get("key12345")
	if err != nil {
		t.Logf("error: unable to retrieve: %s\n", err.Error())
		t.Fail()
	}
	if string(b) != "value12345" {
		t.Logf("error: unexpected value: %s\n", string(b))
		t.Fail()
	}
	s, err := uc.Size()
	if err != nil {
		t.Logf("error: unable to retrieve the size: %s\n", err.Error())
		t.Fail()
	}
	err = uc.Clear()
	if err != nil {
		t.Logf("error: unable to retrieve the clear: %s\n", err.Error())
		t.Fail()
	}
	s, err = uc.Size()
	if err != nil {
		t.Logf("error: unable to retrieve the size: %s\n", err.Error())
		t.Fail()
	}
	if s != 0 {
		t.Logf("error: enexpected size: %d\n", s)
		t.Fail()
	}
	uc.Close()
	time.Sleep(1 * time.Second)
}

func TestHTTPMapClient(t *testing.T) {
	m := NewMap()
	var wg sync.WaitGroup
	wg.Add(1)
	us, err := NewHTTPMapServer("localhost", 8080, &wg, m, true)
	if err != nil {
		t.Logf("error: unable to start the UDP server: %s\n", err.Error())
		t.Fail()
	}
	go us.Serve()
	time.Sleep(1 * time.Second)
	uc := NewHTTPMapClient("localhost", 8080)
	uc.Dial()
	if err != nil {
		t.Logf("error: unable to dial in: %s\n", err.Error())
		t.Fail()
	}
	err = uc.Put("key12345", []byte("value12345"))
	if err != nil {
		t.Logf("error: unable to store: %s\n", err.Error())
		t.Fatal()
	}
	b, err := uc.Get("key12345")
	if err != nil {
		t.Logf("error: unable to retrieve: %s\n", err.Error())
		t.Fail()
	}
	if string(b) != "value12345" {
		t.Logf("error: unexpected value: %s\n", string(b))
		t.Fail()
	}
	s, err := uc.Size()
	if err != nil {
		t.Logf("error: unable to retrieve the size: %s\n", err.Error())
		t.Fail()
	}
	if s != 1 {
		t.Logf("error: expected a map of size 1")
		t.Fail()
	}
	err = uc.Clear()
	if err != nil {
		t.Logf("error: unable to retrieve the clear: %s\n", err.Error())
		t.Fail()
	}
	s, err = uc.Size()
	if err != nil {
		t.Logf("error: unable to retrieve the size: %s\n", err.Error())
		t.Fail()
	}
	if s != 0 {
		t.Logf("error: enexpected size: %d\n", s)
		t.Fail()
	}
	uc.Close()
	time.Sleep(1 * time.Second)
}
