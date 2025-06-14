# Leader based replication - in Golang

![basic-architecture](basic-architecture.png)

Run the application:
```go
go run main.go
```

Write example:

```http
curl -X POST http://localhost:8080/write \
  -H "Content-Type: application/json" \
  -d '{"key": "user:1", "value": "Jessica"}'
```

Read from the Leader:
```http
curl "http://localhost:8080/read?key=user:1"
```

Read from specific follower(simulates some delay). If you try this right after writing you may get a 404 if replication hasn’t completed yet:

```http
curl "http://localhost:8080/follower-read?key=user:1&replica=1"
```
