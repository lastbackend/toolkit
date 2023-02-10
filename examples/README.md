# Examples

### 1. GRPC service

----

Follow these run example:

Run the GRPC service:

```console
 go run service/main.go
```

if you see it:
```console
...
server listening at [::]:50005
...
```

_Great, you have succeeded!_


### 2. HTTP Gateway

----

Follow these run example:

Run the Gateway server:

```console
 go run gateway/main.go
```

Run gRPC Hello World Server

```console
 go run helloworld/main.go
```

Make a request to the Hello World Server

```console
curl -i -X POST http://localhost:8080/hello -H 'Content-Type: application/json' --data '{"name":"world"}'

Response:
HTTP/1.1 200 OK
Content-Type: application/json
Date: Wed, 01 Feb 2023 20:40:03 GMT
Content-Length: 24

{"message":"Hello world"} 
```

Call custom handler

```console
curl -i -X GET http://localhost:8080/health -H 'Content-Type: application/json' 

Response:
HTTP/1.1 200 OK
Content-Type: application/json
Date: Wed, 01 Feb 2023 20:44:57 GMT
Content-Length: 15

{"alive": true}
```


### 2. WSS Proxy

----

Follow these run example:

Run the WSS server:

```console
 go run wss/main.go
```

Run gRPC Hello World Server

```console
 go run helloworld/main.go
```

Make a request to the Hello World Server

```console
curl -i -X POST http://localhost:8080/hello -H 'Content-Type: application/json' --data '{"name":"world"}'

Response:
HTTP/1.1 200 OK
Content-Type: application/json
Date: Wed, 01 Feb 2023 20:40:03 GMT
Content-Length: 24

{"message":"Hello world"} 
```

Call wss events handler

```console
websocat ws://localhost:8008/events 
{"type":"HelloWorld","payload":{"name":"world"}}
{"message":"Hello world"} 
```
