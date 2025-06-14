# Leader based replication - in Golang

![basic-architecture](basic-architecture.png)

Run the application:
```go
go run main.go
```

Write example:

```sh
curl -X POST http://localhost:8080/write \
  -H "Content-Type: application/json" \
  -d '{"key": "user:1", "value": "Jessica"}'
```

Read from the Leader:
```sh
curl "http://localhost:8080/read?key=user:1"
```

Read from specific follower(simulates some delay). If you try this right after writing you may get a 404 if replication hasn’t completed yet:

```sh
curl "http://localhost:8080/follower-read?key=user:1&replica=1"
```

Read with repair - this simulates an eventual consistency repair mechanism:
```sh
curl "http://localhost:8080/read-with-repair?key=user:1"
```

Example logs for POST read-with-repair when replicas are called before replication is complete:

```sh
➜  go-replication-simulation (main) go run main.go                                                                      ✭
2025/06/14 16:05:52 Application running at http://localhost:8080
2025/06/14 16:06:47 Leader wrote key=user:1 value=Jessica
2025/06/14 16:06:47 [ASYNC] Replicated key=user:1 to follower 1
2025/06/14 16:06:47 [ASYNC] Replicated key=user:1 to follower 2
2025/06/14 16:06:48 [REPAIR] Repaired leader with key=user:1
2025/06/14 16:06:48 [REPAIR] Repaired follower 1 with key=user:1
```
Quorum read/writes. Simulates realistic durability guarantees (W replicas). A precursor to stronger consistency and failover resilience.:
```sh
curl -X POST http://localhost:8080/write-with-quorum \
  -H "Content-Type: application/json" \
  -d '{"key":"user:123","value":"Jessica","w":2}'
```
Logs:
```sh
➜  go-replication-simulation (main) go run main.go                                                                      ✭
2025/06/14 16:13:32 Application running at http://localhost:8080
2025/06/14 16:14:10 Leader wrote key=user:123 value=Jessica
2025/06/14 16:14:11 [QUORUM] Replicated key=user:123 to follower 1
2025/06/14 16:14:11 [QUORUM] Replicated key=user:123 to follower 2
```