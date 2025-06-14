package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"go-replication-simulation/store"
)

var (
	leaderStore    = store.NewStore()
	followerStores = []*store.Store{
		store.NewStore(), // follower 1
		store.NewStore(), // follower 2
	}
)

type WriteRequest struct {
	Key   string `json:"key"`
	Value string `json:"value"`
	W     int    `json:"w"` // quorum size (optional, default = 1)
}

func writeHandler(w http.ResponseWriter, r *http.Request) {
	var req WriteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	leaderStore.Set(req.Key, req.Value)
	log.Printf("Leader wrote key=%s value=%s", req.Key, req.Value)

	// Simulate async replication to followers
	go replicateToFollowers(req)

	w.WriteHeader(http.StatusNoContent)
}

func replicateToFollowers(req WriteRequest) {
	for i, follower := range followerStores {
		go func(idx int, f *store.Store) {
			time.Sleep(time.Duration(200+idx*300) * time.Millisecond) // staggered delay
			f.Set(req.Key, req.Value)
			log.Printf("[ASYNC] Replicated key=%s to follower %d", req.Key, idx+1)
		}(i, follower)
	}
}

func readWithRepairHandler(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")

	// Collect records from all replicas
	type versioned struct {
		storeIdx int
		record   store.Record
		found    bool
	}

	allVersions := []versioned{}

	// Check leader
	leaderRec, ok := leaderStore.Get(key)
	allVersions = append(allVersions, versioned{storeIdx: -1, record: leaderRec, found: ok})

	// Check followers
	for i, follower := range followerStores {
		rec, ok := follower.Get(key)
		allVersions = append(allVersions, versioned{storeIdx: i, record: rec, found: ok})
	}

	// Find the freshest record
	var latest store.Record
	var foundLatest bool
	for _, v := range allVersions {
		if !v.found {
			continue
		}
		if !foundLatest || v.record.Timestamp.After(latest.Timestamp) {
			latest = v.record
			foundLatest = true
		}
	}

	if !foundLatest {
		http.Error(w, "not found in any replica", http.StatusNotFound)
		return
	}

	// Repair stale replicas
	for _, v := range allVersions {
		if !v.found || v.record.Timestamp.Before(latest.Timestamp) {
			if v.storeIdx == -1 {
				// Repair leader
				leaderStore.Set(latest.Key, latest.Value)
				log.Printf("[REPAIR] Repaired leader with key=%s", latest.Key)
			} else {
				followerStores[v.storeIdx].Set(latest.Key, latest.Value)
				log.Printf("[REPAIR] Repaired follower %d with key=%s", v.storeIdx+1, latest.Key)
			}
		}
	}

	// Return the latest value to client
	json.NewEncoder(w).Encode(latest)
}

func readHandler(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	rec, ok := leaderStore.Get(key)
	if !ok {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(rec)
}

func followerReadHandler(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	replicaStr := r.URL.Query().Get("replica")

	replicaNum, err := strconv.Atoi(replicaStr)
	if err != nil || replicaNum < 1 || replicaNum > len(followerStores) {
		http.Error(w, "invalid replica number", http.StatusBadRequest)
		return
	}

	idx := replicaNum - 1
	rec, ok := followerStores[idx].Get(key)
	if !ok {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(rec)
}

func writeWithQuorumHandler(w http.ResponseWriter, r *http.Request) {
	var req WriteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if req.W == 0 {
		req.W = 1 // default write quorum
	}

	// Always write to leader first
	leaderStore.Set(req.Key, req.Value)
	log.Printf("Leader wrote key=%s value=%s", req.Key, req.Value)

	// Synchronously write to followers to reach quorum
	acks := 1 // already have 1 from leader
	errCh := make(chan error, len(followerStores))

	for i, follower := range followerStores {
		go func(idx int, f *store.Store) {
			time.Sleep(time.Duration(200+idx*300) * time.Millisecond) // simulate network lag
			f.Set(req.Key, req.Value)
			log.Printf("[QUORUM] Replicated key=%s to follower %d", req.Key, idx+1)
			errCh <- nil
		}(i, follower)
	}

	timeout := time.After(2 * time.Second)
	for acks < req.W {
		select {
		case <-errCh:
			acks++
		case <-timeout:
			http.Error(w, "quorum write timeout", http.StatusGatewayTimeout)
			return
		}
	}

	w.WriteHeader(http.StatusNoContent)
}

func main() {
	http.HandleFunc("/write", writeHandler)
	http.HandleFunc("/read", readHandler)
	http.HandleFunc("/follower-read", followerReadHandler)
	http.HandleFunc("/read-with-repair", readWithRepairHandler)
	http.HandleFunc("/write-with-quorum", writeWithQuorumHandler)

	log.Println("Application running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
