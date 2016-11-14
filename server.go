package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
)

type Server interface {
	Serve()
	Shutdown()
}

type MapServer struct {
	wg   *sync.WaitGroup
	m    *Map
	ack  bool
	up   bool
	host string
	port int
}

func (ms *MapServer) execute(buf []byte) (string, error) {
	line := string(buf[:]) // TODO: termination character
	line = strings.Replace(line, "\r\n", "", 1)
	parts := strings.Split(line, " ")
	command := strings.ToLower(parts[0])
	switch command {
	case "put":
		if len(parts) != 3 {
			return "", errors.New("KO=Bad command, format: PUT <key> <value>")
		}
		ms.m.Put(parts[1], []byte(parts[2]))
		return fmt.Sprintf("OK=%d", len(parts[2])), nil
	case "get":
		if len(parts) != 2 {
			return "", errors.New("KO=Bad command, format: GET <key>")
		}
		value := ms.m.Get(parts[1])
		if value != nil {
			return fmt.Sprintf("OK=%s", string(value)), nil
		}
		return "KO=null", nil
	case "del":
		if len(parts) != 2 {
			return "", errors.New("KO=Bad command, format: DEL <key>")
		}
		ms.m.Delete(parts[1])
		return fmt.Sprintf("OK=%s", parts[1]), nil
	case "size":
		if len(parts) != 1 {
			return "", errors.New("KO=Bad command, format: SIZE")
		}
		return fmt.Sprintf("OK=%d", ms.m.Size()), nil
	case "clear":
		if len(parts) != 1 {
			return "", errors.New("KO=Bad command, format: CLEAR")
		}
		ms.m.Clear()
		return fmt.Sprintf("OK=%d", ms.m.Size()), nil
	default:
		return "", errors.New("KO=Unrecognized command: <PUT|GET|SIZE|CLEAR> [<key> [value]]")
	}
}

func (ms *MapServer) rewind(buf []byte) {
	for i := 0; i < len(buf); i++ {
		buf[i] = 0
	}
}

type UDPMapServer struct {
	MapServer
	conn *net.UDPConn
}

func NewUDPMapServer(host string, port int, wg *sync.WaitGroup, m *Map, ack bool) (*UDPMapServer, error) {
	addr := net.UDPAddr{
		Port: port,
		IP:   net.ParseIP(host),
	}
	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		return nil, err
	}
	us := UDPMapServer{
		conn: conn,
		MapServer: MapServer{
			wg:   wg,
			m:    m,
			ack:  ack,
			up:   true,
			host: host,
			port: port,
		},
	}
	return &us, nil
}

func (us *UDPMapServer) Serve() {
	log.Printf("info: bootstrapping the UDP Server loop: %s:%d\n", us.host, us.port)
	defer us.wg.Done()
	defer us.conn.Close()
	var buf [1024]byte
	for us.up {
		l, r, err := us.conn.ReadFromUDP(buf[:])
		if err != nil {
			continue
		}
		log.Printf("info: received %d: %s from: %v\n", l, string(buf[:]), r)
		outcome, err := us.execute(buf[:l])
		if err != nil {
			us.conn.WriteToUDP([]byte(err.Error()), r)
		} else {
			us.conn.WriteToUDP([]byte(outcome), r)
		}
		us.rewind(buf[:])
	}
	log.Println("info: shutting down the UDP Server...")
}

func (us *UDPMapServer) Shutdown() {
	us.up = false
	us.conn.Close()
}

type TCPMapServer struct {
	MapServer
	conn *net.TCPListener
}

func NewTCPMapServer(host string, port int, wg *sync.WaitGroup, m *Map, ack bool) (*TCPMapServer, error) {
	addr := net.TCPAddr{
		Port: port,
		IP:   net.ParseIP(host),
	}
	conn, err := net.ListenTCP("tcp", &addr)
	if err != nil {
		return nil, err
	}
	ts := TCPMapServer{
		conn: conn,
		MapServer: MapServer{
			wg:   wg,
			m:    m,
			ack:  ack,
			up:   true,
			host: host,
			port: port,
		},
	}
	return &ts, nil
}

func (ts *TCPMapServer) Serve() {
	log.Printf("info: bootstrapping the TCP Server loop: %s:%d\n", ts.host, ts.port)
	defer ts.wg.Done()
	defer ts.conn.Close()
	for ts.up {
		conn, err := ts.conn.AcceptTCP()
		if err != nil {
			log.Printf("error: accepting connection: %s\n", err.Error())
			continue
		}
		go func(conn *net.TCPConn, m *Map, ack bool) {
			var buf [65535]byte
			for {
				l, err := conn.Read(buf[:])
				if err != nil {
					log.Printf("error: not able to read: %s\n", err.Error())
					return
				}
				log.Printf("info: received %d: %s\n", l, string(buf[:]))
				if ts.checkExit(buf[:l]) {
					conn.Close()
					return
				}
				outcome, err := ts.execute(buf[:l])
				if err != nil {
					conn.Write([]byte(err.Error()))
				} else {
					conn.Write([]byte(outcome))
				}

				ts.rewind(buf[:])
			}
		}(conn, ts.m, ts.ack)
	}
}

func (ts *TCPMapServer) checkExit(buf []byte) bool {
	command := string(buf[:])
	if strings.Contains(strings.ToLower(command), "close") {
		return true
	}
	return false
}

func (ts *TCPMapServer) Shutdown() {
	ts.up = false
	ts.conn.Close()
}

type HTTPMapServer struct {
	MapServer
	s *http.Server
}

func NewHTTPMapServer(host string, port int, wg *sync.WaitGroup, m *Map, ack bool) (*HTTPMapServer, error) {
	addr := fmt.Sprintf("%s:%d", host, port)
	s := http.Server{
		Addr: addr,
	}
	hs := HTTPMapServer{
		s: &s,
		MapServer: MapServer{
			wg:   wg,
			m:    m,
			ack:  ack,
			host: host,
			port: port,
		},
	}
	return &hs, nil
}

func (hs *HTTPMapServer) Serve() {
	log.Printf("info: bootstrapping the HTTP Server loop: %s:%d\n", hs.host, hs.port)
	defer hs.wg.Done()
	http.HandleFunc("/api/v1/map", hs.handler)
	http.ListenAndServe(fmt.Sprintf("%s:%d", hs.host, hs.port), nil)
}

func (hs *HTTPMapServer) handler(w http.ResponseWriter, r *http.Request) {
	res := make(map[string]interface{})
	if r.Header.Get("Accept") != "application/json" ||
		r.Header.Get("Content-Type") != "application/json" {
		res["error"] = "Bad content type: only JSON supported"
		buf, _ := json.Marshal(res)
		w.Write(buf[:])
		return
	}
	switch r.Method {
	case "POST":
		hs.postHandler(w, r, res)
	case "GET":
		hs.getHandler(w, r, res)
	case "DELETE":
		hs.deleteHandler(w, r, res)
	default:
		w.Header().Add("Status-Code", "400")
		w.Header().Add("Reason-Phrase", "Unrecognized service request")
		res["error"] = "Bad method: only POST, GET, DELETE accepted"
		buf, _ := json.Marshal(res)
		w.Write(buf[:])
		return
	}
}

// ?key=* (size), or ?key=<key>
func (hs *HTTPMapServer) getHandler(w http.ResponseWriter, r *http.Request, rs map[string]interface{}) {
	qs := r.URL.Query()
	log.Printf("info: serving GET %v\n", qs)
	w.Header().Add("Content-Type", "application/json")
	if qs.Get("key") == "*" {
		rs["outcome"] = "OK"
		rs["size"] = hs.m.Size()
	} else if qs.Get("key") != "" {
		value := hs.m.Get(qs.Get("key"))
		if value == nil {
			rs["outcome"] = "OK"
			rs["value"] = "null"
		} else {
			rs["outcome"] = "OK"
			rs["value"] = string(value)
		}

	} else {
		w.Header().Add("Status-Code", "400")
		w.Header().Add("Reason-Phrase", "Unrecognized service request")
		rs["outcome"] = "KO"
		rs["error"] = "Unrecognized pattern, GET: ?key=* (size), or ?key=<key> allowed"
	}
	buf, _ := json.Marshal(rs)
	w.Write(buf[:])
}

// body
func (hs *HTTPMapServer) postHandler(w http.ResponseWriter, r *http.Request, rs map[string]interface{}) {
	d := json.NewDecoder(r.Body)
	defer r.Body.Close()
	var req map[string]interface{}
	err := d.Decode(&req)
	if err != nil {
		log.Printf("error: not able to read: %s\n", err.Error())
		return
	}
	log.Printf("info: serving POST  %v\n", req)
	key := req["key"].(string)
	value := []byte(req["value"].(string))
	if req["key"] == nil || req["value"] == nil {
		rs["outcome"] = "KO"
		rs["error"] = "Unrecognized JSON: no key/value pair"
		buf, _ := json.Marshal(rs)
		w.Write(buf[:])
		return
	}
	hs.m.Put(key, value)
	rs["outcome"] = "OK"
	rs["wrote"] = len(value)
	buf, _ := json.Marshal(rs)
	w.Write(buf[:])
}

// ?key=* (delete all), or ?key=<key>
func (hs *HTTPMapServer) deleteHandler(w http.ResponseWriter, r *http.Request, rs map[string]interface{}) {
	qs := r.URL.Query()
	log.Printf("info: serving DELETE %v\n", qs)
	w.Header().Add("Content-Type", "application/json")
	if qs.Get("key") == "*" {
		hs.m.Clear()
		rs["outcome"] = "OK"
		rs["size"] = hs.m.Size()
	} else if qs.Get("key") != "" {
		rs["outcome"] = "OK"
		rs["key"] = qs.Get("key")
		hs.m.Delete(qs.Get("key"))
	} else {
		w.Header().Add("Status-Code", "400")
		w.Header().Add("Reason-Phrase", "Unrecognized service request")
		rs["outcome"] = "KO"
		rs["error"] = "Unrecognized pattern, DELETE: ?key=* (clear), or ?key=<key> allowed"
	}
	buf, _ := json.Marshal(rs)
	w.Write(buf[:])
}
