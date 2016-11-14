# dmap (distributed map)

## Rationale
dmap is a Golang implementation of a fast concurrent map, accessible from the network. Clients can put/get/remove their key/value tuples using UDP, TCP or HTTP: client API allows to pick the transport transparently.

## Architecture
The concurrent map implementation is not lock-free but it's sharded and disciplinate the contention with RW (Read/Write) locks. UDP, TCP and HTTP servers expose the concurrent map operations over the network. For the UDP and TCP versions, a lightweight text protocol has been developed getting inspiration from Redis'; instead, for the HTTP version the JSON data format has been used on top of the HTTP protocol itself.
Hashing was initially used to address the shards, then from a performance benchmark it came up that it was too slow so the shard addressing is realized just picking the first byte of the key (0 to 255 nuber to address one of the 256 shards); this is not going to bring an homogeneous distribution and so contention avoidance (shards are used to spread out the entries avoiding contention), and therefore is something to be worked on. 

### Wire Protocol
The concurrent map provides 5 main operations.

- *Put*. To store data as a key/value pair.
- *Get*. To retrieve data corresponding to the key.
- *Delete*. To delete data corresponding to the key.
- *Clear*. To clear the data stored into the map.
- *Size*. To retrieve the size of the map. 

The wire protocol should be implementing these five operations then. The Client API implements this interface and provides network RPC calls to deal with the distributed map.

#### UDP/TCP Details

- *Put*. Request ```PUT <key> <value>```, and response ```OK=<X>``` where X is the number of written bytes.
- *Get*. ```GET <key>```, and response ```OK=<value>```.
- *Delete*. ```DEL <key>```, and response ```OK=<key>``` confirming that the key has been removed.
- *Clear*. ```CLEAR```, and response ```OK=<size>``` to confirm the clean up.
- *Size*.  ```SIZE```, and response ```OK=<size>``` to return the actual size.

In case of any error, the response is: ```KO=<error_messsage>```.

#### REST Endpoints Details

- *Put*. ```POST /api/v1/map``` with a body ```{ "key": "<key>", "value": "<value>" }```
- *Get*. ```GET /api/v1/map?key=<key>```
- *Delete*. ```DELETE /api/v1/map?key=<key>```
- *Clear*. ```DELETE /api/v1/map?key=*```
- *Size*. ```GET /api/v1/map?key=*```

The response is in JSON and in the format: ```{ "outcome": "KO", "error": "<error_message>" }```, in case of error, or ```{ "outcome": "KO", "<size>|<value>|": "<X>" }``` in case of success, and according to the service invoked.

## Run the Server
Golang compiles to executable binaries. So, it should be enough to run:

```
$> ./dmap
go build && ./dmap 
2016/11/14 21:33:04 info: bootstrapping the UDP Server loop: localhost:12345
2016/11/14 21:33:04 info: bootstrapping the TCP Server loop: localhost:12346
2016/11/14 21:33:04 info: bootstrapping the HTTP Server loop: localhost:8080
2016/11/14 21:33:05 info: received 13: PUT key value from: [::1]:50534
2016/11/14 21:33:05 OK=5
```

this will bring up the UDP, TCP and HTTP server all accessing the same concurrent map: no matter which transport is used, the information will be stored in the same shards.

### Telnet Client
Once started the concurrent map server, Telnet can be used to use the services provided over TCP:

```
telnet localhost 12346
Trying ::1...
Connected to localhost.
Escape character is '^]'.
GET key
OK=value
```

from the server perspective:

```
[developer@localhost dmap]$ go build && ./dmap 
2016/11/14 21:33:04 info: bootstrapping the UDP Server loop: localhost:12345
2016/11/14 21:33:04 info: bootstrapping the TCP Server loop: localhost:12346
2016/11/14 21:33:04 info: bootstrapping the HTTP Server loop: localhost:8080
2016/11/14 21:33:05 info: received 13: PUT key value from: [::1]:50534
2016/11/14 21:33:05 OK=5
2016/11/14 21:34:20 info: received 9: GET key
```

the GET shows up into the served operations.

## Run the Tests
dmap comes up with tests for three three conenctors (UDP, TCP and HTTP) and the Client API. For example, if only the HTTP Client API wants to be tested:

```
go test -run=TestHTTPMapClient -v
=== RUN   TestHTTPMapClient
2016/11/14 19:58:22 info: bootstrapping the HTTP Server loop: localhost:8080
2016/11/14 19:58:23 info: serving POST  map[key:key12345 value:value12345]
2016/11/14 19:58:23 map[outcome:OK wrote:10]
2016/11/14 19:58:23 info: serving GET map[key:[key12345]]
2016/11/14 19:58:23 map[outcome:OK value:value12345]
2016/11/14 19:58:23 info: serving GET map[key:[*]]
2016/11/14 19:58:23 map[outcome:OK size:1]
2016/11/14 19:58:23 info: serving DELETE map[key:[*]]
2016/11/14 19:58:23 map[outcome:OK size:0]
2016/11/14 19:58:23 info: serving GET map[key:[*]]
2016/11/14 19:58:23 map[outcome:OK size:0]
--- PASS: TestHTTPMapClient (2.01s)
PASS
ok  	dmap	2.015s
```

