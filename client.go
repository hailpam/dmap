package dmap

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
)

type Client interface {
	Dial() error
	Close() error
	Put(string, []byte) error
	Get(string) ([]byte, error)
	Delete(string) error
	Size() (int, error)
	Clear() error
}

type MapClient struct {
	conn net.Conn
	host string
	port int
}

func (mc *MapClient) Close() error {
	mc.conn.Write([]byte("CLOSE"))
	return mc.conn.Close()
}

func (mc *MapClient) Put(key string, value []byte) error {
	command := fmt.Sprintf("PUT %s %s", key, string(value[:]))
	_, err := mc.conn.Write([]byte(command))
	if err != nil {
		return err
	}
	var buf [1024]byte
	l, err := mc.conn.Read(buf[:])
	if err != nil {
		return nil
	}
	_, err = mc.parse(buf[:], l)
	if err != nil {
		return err
	}
	return nil
}

func (mc *MapClient) Get(key string) ([]byte, error) {
	command := fmt.Sprintf("GET %s", key)
	_, err := mc.conn.Write([]byte(command))
	if err != nil {
		return nil, err
	}
	var buf [1024]byte
	l, err := mc.conn.Read(buf[:])
	if err != nil {
		return nil, nil
	}
	b, err := mc.parse(buf[:], l)
	if err != nil {
		return nil, err
	}
	return []byte(b), nil
}

func (mc *MapClient) Size() (int, error) {
	_, err := mc.conn.Write([]byte("SIZE"))
	if err != nil {
		return 0, err
	}
	var buf [1024]byte
	l, err := mc.conn.Read(buf[:])
	if err != nil {
		return 0, err
	}
	b, err := mc.parse(buf[:], l)
	if err != nil {
		return 0, err
	}
	i, err := strconv.Atoi(b)
	if err != nil {
		return -1, err
	}
	return i, nil
}

func (mc *MapClient) Clear() error {
	_, err := mc.conn.Write([]byte("CLEAR"))
	if err != nil {
		return err
	}
	var buf [1024]byte
	l, err := mc.conn.Read(buf[:])
	if err != nil {
		return err
	}
	_, err = mc.parse(buf[:], l)
	if err != nil {
		return err
	}
	return nil
}

func (mc *MapClient) Delete(key string) error {
	command := fmt.Sprintf("DEL %s", key)
	_, err := mc.conn.Write([]byte(command))
	if err != nil {
		return err
	}
	var buf [1024]byte
	l, err := mc.conn.Read(buf[:])
	if err != nil {
		return nil
	}
	_, err = mc.parse(buf[:], l)
	if err != nil {
		return err
	}
	return nil
}

func (mc *MapClient) parse(buf []byte, length int) (string, error) {
	res := string(buf[:length])
	log.Println(res)
	parts := strings.Split(res, "=")
	if len(parts) != 2 {
		return "", errors.New("Unexpected response: " + res)
	}
	if parts[0] != "OK" {
		return "", errors.New(parts[1])
	}
	return parts[1], nil
}

type UDPMapClient struct {
	MapClient
}

func NewUDPMapClient(host string, port int) *UDPMapClient {
	uc := &UDPMapClient{
		MapClient: MapClient{
			host: host,
			port: port,
		},
	}
	return uc
}

func (uc *UDPMapClient) Dial() error {
	conn, err := net.Dial("udp", fmt.Sprintf("%s:%d", uc.host, uc.port))
	if err != nil {
		return err
	}
	uc.conn = conn
	return nil
}

type TCPMapClient struct {
	MapClient
}

func NewTCPMapClient(host string, port int) *TCPMapClient {
	tc := &TCPMapClient{
		MapClient: MapClient{
			host: host,
			port: port,
		},
	}
	return tc
}

func (uc *TCPMapClient) Dial() error {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", uc.host, uc.port))
	if err != nil {
		return err
	}
	uc.conn = conn
	return nil
}

type HTTPMapClient struct {
	MapClient
	client *http.Client
}

func NewHTTPMapClient(host string, port int) *HTTPMapClient {
	client := &http.Client{}
	hc := &HTTPMapClient{
		client: client,
		MapClient: MapClient{
			host: host,
			port: port,
		},
	}
	return hc
}

func (hc *HTTPMapClient) Dial() error {
	// nothing to do, as per HTTP client semantic
	return nil
}

func (hc *HTTPMapClient) Close() error {
	// nothing to do, as per HTTP client semantic
	return nil
}

func (hc *HTTPMapClient) Put(key string, value []byte) error {
	url := fmt.Sprintf("http://%s:%d/api/v1/map", hc.host, hc.port)
	body := fmt.Sprintf("{ \"key\": \"%s\", \"value\": \"%s\" }", key, string(value))
	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(body)))
	if err != nil {
		return nil
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	resp, err := hc.client.Do(req)
	if err != nil {
		return err
	}
	json, err := hc.parseBody(resp)
	if err != nil {
		return err
	}
	if json["outcome"].(string) == "KO" {
		return errors.New(json["error"].(string))
	}
	return nil
}

func (hc *HTTPMapClient) Get(key string) ([]byte, error) {
	url := fmt.Sprintf("http://%s:%d/api/v1/map?key=%s", hc.host, hc.port, key)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	resp, err := hc.client.Do(req)
	if err != nil {
		return nil, err
	}
	json, err := hc.parseBody(resp)
	if err != nil {
		return nil, err
	}
	if json["outcome"].(string) == "KO" {
		return []byte(""), nil
	}
	return []byte(json["value"].(string)), nil
}

func (hc *HTTPMapClient) Delete(key string) error {
	url := fmt.Sprintf("http://%s:%d/api/v1/map?key=%s", hc.host, hc.port, key)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	resp, err := hc.client.Do(req)
	if err != nil {
		return err
	}
	_, err = hc.parseBody(resp)
	if err != nil {
		return err
	}
	return nil
}

func (hc *HTTPMapClient) Clear() error {
	url := fmt.Sprintf("http://%s:%d/api/v1/map?key=*", hc.host, hc.port)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	resp, err := hc.client.Do(req)
	if err != nil {
		return err
	}
	_, err = hc.parseBody(resp)
	if err != nil {
		return err
	}
	return nil
}

func (hc *HTTPMapClient) Size() (int, error) {
	url := fmt.Sprintf("http://%s:%d/api/v1/map?key=*", hc.host, hc.port)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return -1, err
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	resp, err := hc.client.Do(req)
	if err != nil {
		return -1, err
	}
	json, err := hc.parseBody(resp)
	if err != nil {
		return -1, err
	}
	if json["outcome"].(string) == "KO" {
		return -1, nil
	}
	return int(json["size"].(float64)), nil
}

func (hc *HTTPMapClient) parseBody(res *http.Response) (map[string]interface{}, error) {
	if res.StatusCode != 200 {
		return nil, errors.New(res.Status)
	}
	c, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	var j map[string]interface{}
	err = json.Unmarshal(c, &j)
	if err != nil {
		return nil, err
	}
	log.Println(j)
	return j, nil
}
