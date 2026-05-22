// Mounts all product and media handlers onto a single ServeMux and starts
// the HTTP server on :8082 (overridable via PORT environment variable).

package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"source-asia-backend-assignment/part2/handlers"
	"source-asia-backend-assignment/part2/store"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}

	s := store.New()

	mux := http.NewServeMux()

	// /products — handles POST and GET (list).
	mux.Handle("/products", handlers.Products(s))

	// /products/* — single wildcard that routes to media or product detail
	// based on whether the path ends with /media.
	mux.HandleFunc("/products/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/media") {
			handlers.Media(s).ServeHTTP(w, r)
			return
		}
		handlers.Products(s).ServeHTTP(w, r)
	})

	addr := ":" + port
	fmt.Printf("Part 2 — Product Catalog API running on http://localhost%s\n", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}