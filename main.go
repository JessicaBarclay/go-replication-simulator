package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"go-replication-simulation/store"
)

var leaderStore = store.NewStore()

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
	// In real life you'd do HTTP POST to follower nodes here
	time.Sleep(500 * time.Millisecond) // Simulate lag
	log.Printf("[ASYNC] Replicating key=%s to followers", req.Key)
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

func main() {
	http.HandleFunc("/write", writeHandler)
	http.HandleFunc("/read", readHandler)

	log.Println("Leader running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
