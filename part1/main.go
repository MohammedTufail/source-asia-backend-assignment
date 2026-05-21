package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	rl := NewRateLimiter()

	mux := http.NewServeMux()
	mux.HandleFunc("/request", handleRequest(rl))
	mux.HandleFunc("/stats", handleStats(rl))

	addr := ":" + port
	fmt.Printf("Part 1 — Rate-limited API server running on http://localhost%s\n", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}