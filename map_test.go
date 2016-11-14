package main

import (
	"fmt"
	"sync"
	"testing"
)

func TestMap(t *testing.T) {
	m := NewMap()
	m.Put("test", []byte("Hello, World!"))
	r := string(m.Get("test"))
	if r != "Hello, World!" {
		t.Logf("retrieved %s\n", r)
		t.Fail()
	}
	m.Put("test1", []byte("Hello, World! 1"))
	r = string(m.Get("test1"))
	if r != "Hello, World! 1" {
		t.Logf("retrieved %s\n", r)
		t.Fail()
	}
	s := m.Size()
	if s != 2 {
		t.Logf("expected size 2: %d\n", s)
		t.Fail()
	}
}

func TestMapConcurrent(t *testing.T) {
	m := NewMap()
	var wg sync.WaitGroup
	ids := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go tester(ids[i], m, t, &wg)
	}
	wg.Wait()
	t.Logf("size: %d\n", m.Size())
	m.Clear()
	size := m.Size()
	if size != 0 {
		t.Logf("not cleaned up: %d\n", size)
		t.Fail()
	}
}

func tester(id string, m *Map, t *testing.T, wg *sync.WaitGroup) {
	defer wg.Done()
	for i := 0; i < 100000; i++ {
		key := fmt.Sprintf("%s%d", id, i)
		m.Put(key, []byte(key))
		value := string(m.Get(key))
		if value != key {
			t.Logf("expected %s: got %s\n", key, value)
			t.Fail()
		}
	}
}
