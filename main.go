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

func main() {
	http.HandleFunc("/write", writeHandler)
	http.HandleFunc("/read", readHandler)
	http.HandleFunc("/follower-read", followerReadHandler)

	log.Println("Application running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
