package main

import (
	"sync"
)

type Map struct {
	e []entry
	s uint
}

type entry struct {
	m map[string][]byte
	l sync.RWMutex
}

func NewMap() *Map {
	m := new(Map)
	m.e = make([]entry, 256)
	for i := 0; i < 256; i++ {
		m.e[i].m = make(map[string][]byte)
	}
	return m
}

func (m *Map) Put(key string, value []byte) {
	idx := index(key)
	m.e[idx].l.Lock()
	defer m.e[idx].l.Unlock()
	m.e[idx].m[key] = value
}

func (m *Map) Get(key string) []byte {
	idx := index(key)
	m.e[idx].l.RLock()
	defer m.e[idx].l.RUnlock()
	return m.e[idx].m[key]
}

func (m *Map) Clear() {
	for i := 0; i < 256; i++ {
		m.e[i].l.Lock()
		for k := range m.e[i].m {
			delete(m.e[i].m, k)
		}
		m.e[i].l.Unlock()
	}
}

func (m *Map) Delete(key string) {
	idx := index(key)
	m.e[idx].l.Lock()
	defer m.e[idx].l.Unlock()
	delete(m.e[idx].m, key)
}

func (m *Map) Size() int {
	var size int
	for i := 0; i < 256; i++ {
		m.e[i].l.RLock()
		size += len(m.e[i].m)
		m.e[i].l.RUnlock()
	}
	return size
}

func index(key string) uint8 {
	c := []byte(key)
	return c[0]
}
