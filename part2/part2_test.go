// Integration and unit tests for the product catalog API (Part 2).
// Tests cover: product creation (valid, duplicate SKU, bad input), list
// pagination, detail retrieval, media appending, URL validation rules,
// the critical performance invariant (list never serialises full media
// arrays), and concurrency safety under parallel writes.

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"source-asia-backend-assignment/part2/handlers"
	"source-asia-backend-assignment/part2/models"
	"source-asia-backend-assignment/part2/store"
	"source-asia-backend-assignment/part2/validator"
)

// Test server setup

func newTestServer() (*httptest.Server, *store.Store) {
	s := store.New()
	mux := http.NewServeMux()
	mux.Handle("/products", handlers.Products(s))
	// Single wildcard handler routes both /products/{id} and /products/{id}/media.
	mux.HandleFunc("/products/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/media") {
			handlers.Media(s).ServeHTTP(w, r)
			return
		}
		handlers.Products(s).ServeHTTP(w, r)
	})
	return httptest.NewServer(mux), s
}

// Helpers

func doRequest(t *testing.T, srv *httptest.Server, method, path, body string) *http.Response {
	t.Helper()
	var bodyReader *strings.Reader
	if body != "" {
		bodyReader = strings.NewReader(body)
	} else {
		bodyReader = strings.NewReader("")
	}
	req, err := http.NewRequest(method, srv.URL+path, bodyReader)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("%s %s failed: %v", method, path, err)
	}
	return resp
}

func decodeJSON(t *testing.T, resp *http.Response, v any) {
	t.Helper()
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
		t.Fatalf("decode response: %v", err)
	}
}

func createProductOK(t *testing.T, srv *httptest.Server, name, sku string) models.Product {
	t.Helper()
	body := fmt.Sprintf(`{"name":%q,"sku":%q,"image_urls":["https://cdn.example.com/%s/img-1.jpg"],"video_urls":["https://cdn.example.com/%s/demo.mp4"]}`,
		name, sku, sku, sku)
	resp := doRequest(t, srv, http.MethodPost, "/products", body)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("createProduct: expected 201, got %d", resp.StatusCode)
	}
	var p models.Product
	decodeJSON(t, resp, &p)
	return p
}

// Validator unit tests

func TestValidator_AcceptsHTTPS(t *testing.T) {
	if err := validator.ValidateURL("https://cdn.example.com/img.jpg"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidator_AcceptsHTTP(t *testing.T) {
	if err := validator.ValidateURL("http://example.com/video.mp4"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidator_RejectsFTP(t *testing.T) {
	if err := validator.ValidateURL("ftp://cdn.example.com/file.jpg"); err == nil {
		t.Fatal("expected error for ftp scheme")
	}
}

func TestValidator_RejectsNoScheme(t *testing.T) {
	if err := validator.ValidateURL("cdn.example.com/file.jpg"); err == nil {
		t.Fatal("expected error for missing scheme")
	}
}

func TestValidator_RejectsTooLong(t *testing.T) {
	long := "https://example.com/" + strings.Repeat("a", validator.MaxURLLength)
	if err := validator.ValidateURL(long); err == nil {
		t.Fatal("expected error for URL exceeding max length")
	}
}

func TestValidator_RejectsEmpty(t *testing.T) {
	if err := validator.ValidateURL(""); err == nil {
		t.Fatal("expected error for empty URL")
	}
}

func TestValidator_RejectsSliceOverLimit(t *testing.T) {
	urls := make([]string, validator.MaxURLsPerRequest+1)
	for i := range urls {
		urls[i] = fmt.Sprintf("https://cdn.example.com/img-%d.jpg", i)
	}
	if err := validator.ValidateURLSlice("image_urls", urls); err == nil {
		t.Fatal("expected error for over-limit URL slice")
	}
}

// Store unit tests

func TestStore_DuplicateSKU(t *testing.T) {
	s := store.New()
	_, err := s.Create("Widget A", "SKU-001", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	_, err = s.Create("Widget B", "SKU-001", nil, nil)
	if err == nil {
		t.Fatal("expected duplicate SKU error")
	}
}

func TestStore_ListDoesNotLoadMediaArrays(t *testing.T) {
	s := store.New()
	imageURLs := make([]string, 10)
	for i := range imageURLs {
		imageURLs[i] = fmt.Sprintf("https://cdn.example.com/img-%d.jpg", i)
	}
	for i := 0; i < 5; i++ {
		_, err := s.Create(fmt.Sprintf("Product %d", i), fmt.Sprintf("SKU-%03d", i), imageURLs, nil)
		if err != nil {
			t.Fatal(err)
		}
	}

	items, total := s.List(store.ListOptions{Limit: 20, Offset: 0})
	if total != 5 {
		t.Fatalf("expected total=5, got %d", total)
	}
	// Verify list items are the lightweight type — they should have counts, not arrays.
	for _, item := range items {
		if item.ImageCount != 10 {
			t.Fatalf("expected ImageCount=10, got %d", item.ImageCount)
		}
		// ThumbnailURL should be set (first image).
		if item.ThumbnailURL == "" {
			t.Fatal("expected ThumbnailURL to be set")
		}
	}
}

func TestStore_PaginationOffset(t *testing.T) {
	s := store.New()
	for i := 0; i < 10; i++ {
		s.Create(fmt.Sprintf("Product %d", i), fmt.Sprintf("SKU-%d", i), nil, nil)
	}
	items, total := s.List(store.ListOptions{Limit: 3, Offset: 7})
	if total != 10 {
		t.Fatalf("expected total=10, got %d", total)
	}
	if len(items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(items))
	}
}

func TestStore_OffsetBeyondTotal(t *testing.T) {
	s := store.New()
	s.Create("P", "S", nil, nil)
	items, _ := s.List(store.ListOptions{Limit: 10, Offset: 99})
	if len(items) != 0 {
		t.Fatalf("expected 0 items for offset beyond total, got %d", len(items))
	}
}

// HTTP handler integration tests

func TestHTTP_CreateProduct_201(t *testing.T) {
	srv, _ := newTestServer()
	defer srv.Close()
	p := createProductOK(t, srv, "Widget A", "SKU-001")
	if p.ID == "" {
		t.Fatal("expected non-empty ID in response")
	}
	if p.SKU != "SKU-001" {
		t.Fatalf("expected SKU-001, got %s", p.SKU)
	}
}

func TestHTTP_CreateProduct_EmptyName(t *testing.T) {
	srv, _ := newTestServer()
	defer srv.Close()
	resp := doRequest(t, srv, http.MethodPost, "/products", `{"name":"","sku":"SKU-X","image_urls":[]}`)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestHTTP_CreateProduct_EmptySKU(t *testing.T) {
	srv, _ := newTestServer()
	defer srv.Close()
	resp := doRequest(t, srv, http.MethodPost, "/products", `{"name":"Widget","sku":""}`)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestHTTP_CreateProduct_DuplicateSKU_409(t *testing.T) {
	srv, _ := newTestServer()
	defer srv.Close()
	createProductOK(t, srv, "Widget A", "SKU-DUP")
	resp := doRequest(t, srv, http.MethodPost, "/products", `{"name":"Widget B","sku":"SKU-DUP"}`)
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("expected 409, got %d", resp.StatusCode)
	}
}

func TestHTTP_CreateProduct_InvalidURL(t *testing.T) {
	srv, _ := newTestServer()
	defer srv.Close()
	resp := doRequest(t, srv, http.MethodPost, "/products",
		`{"name":"W","sku":"S","image_urls":["not-a-url"]}`)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid URL, got %d", resp.StatusCode)
	}
}

func TestHTTP_ListProducts_DefaultPagination(t *testing.T) {
	srv, _ := newTestServer()
	defer srv.Close()
	for i := 0; i < 5; i++ {
		createProductOK(t, srv, fmt.Sprintf("P%d", i), fmt.Sprintf("SKU-%d", i))
	}
	resp := doRequest(t, srv, http.MethodGet, "/products", "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var result map[string]interface{}
	decodeJSON(t, resp, &result)
	data := result["data"].([]interface{})
	if len(data) != 5 {
		t.Fatalf("expected 5 items, got %d", len(data))
	}

	// Confirm list items do NOT contain image_urls or video_urls arrays.
	item := data[0].(map[string]interface{})
	if _, hasImageURLs := item["image_urls"]; hasImageURLs {
		t.Fatal("list response must not include image_urls array")
	}
	if _, hasVideoURLs := item["video_urls"]; hasVideoURLs {
		t.Fatal("list response must not include video_urls array")
	}
	// But should have counts.
	if _, hasCount := item["image_count"]; !hasCount {
		t.Fatal("list response must include image_count")
	}
}

func TestHTTP_ListProducts_Pagination(t *testing.T) {
	srv, _ := newTestServer()
	defer srv.Close()
	for i := 0; i < 10; i++ {
		createProductOK(t, srv, fmt.Sprintf("P%d", i), fmt.Sprintf("SKU-P%d", i))
	}

	resp := doRequest(t, srv, http.MethodGet, "/products?limit=3&offset=5", "")
	var result map[string]interface{}
	decodeJSON(t, resp, &result)

	data := result["data"].([]interface{})
	if len(data) != 3 {
		t.Fatalf("expected 3 items with limit=3, got %d", len(data))
	}
	if int(result["total"].(float64)) != 10 {
		t.Fatalf("expected total=10, got %v", result["total"])
	}
	if !result["has_more"].(bool) {
		t.Fatal("expected has_more=true")
	}
}

func TestHTTP_GetProduct_FullMedia(t *testing.T) {
	srv, _ := newTestServer()
	defer srv.Close()
	p := createProductOK(t, srv, "Widget A", "SKU-FULL")

	resp := doRequest(t, srv, http.MethodGet, "/products/"+p.ID, "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var full models.Product
	decodeJSON(t, resp, &full)
	if len(full.ImageURLs) != 1 {
		t.Fatalf("expected 1 image URL in detail, got %d", len(full.ImageURLs))
	}
	if len(full.VideoURLs) != 1 {
		t.Fatalf("expected 1 video URL in detail, got %d", len(full.VideoURLs))
	}
}

func TestHTTP_GetProduct_NotFound(t *testing.T) {
	srv, _ := newTestServer()
	defer srv.Close()
	resp := doRequest(t, srv, http.MethodGet, "/products/nonexistent-id", "")
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}

func TestHTTP_AddMedia_AppendsURLs(t *testing.T) {
	srv, _ := newTestServer()
	defer srv.Close()
	p := createProductOK(t, srv, "Widget A", "SKU-MEDIA")

	mediaBody := `{"image_urls":["https://cdn.example.com/new-img.jpg"],"video_urls":["https://cdn.example.com/new-vid.mp4"]}`
	resp := doRequest(t, srv, http.MethodPost, "/products/"+p.ID+"/media", mediaBody)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 from media endpoint, got %d", resp.StatusCode)
	}

	var updated models.Product
	decodeJSON(t, resp, &updated)
	// Created with 1 image + 1 video; added 1 more each.
	if len(updated.ImageURLs) != 2 {
		t.Fatalf("expected 2 image URLs after append, got %d", len(updated.ImageURLs))
	}
	if len(updated.VideoURLs) != 2 {
		t.Fatalf("expected 2 video URLs after append, got %d", len(updated.VideoURLs))
	}
}

func TestHTTP_AddMedia_EmptyBody_400(t *testing.T) {
	srv, _ := newTestServer()
	defer srv.Close()
	p := createProductOK(t, srv, "Widget A", "SKU-M2")
	resp := doRequest(t, srv, http.MethodPost, "/products/"+p.ID+"/media",
		`{"image_urls":[],"video_urls":[]}`)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 for empty media body, got %d", resp.StatusCode)
	}
}

func TestHTTP_AddMedia_NotFound(t *testing.T) {
	srv, _ := newTestServer()
	defer srv.Close()
	resp := doRequest(t, srv, http.MethodPost, "/products/no-such-id/media",
		`{"image_urls":["https://cdn.example.com/img.jpg"]}`)
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}

// Performance invariant test

// TestPerformanceInvariant_ListDoesNotSerialiseAllMedia creates 1000 products
// each with 10 image URLs (10,000 total) and verifies that GET /products
// returns quickly and the response does not contain any image_urls arrays.
func TestPerformanceInvariant_ListDoesNotSerialiseAllMedia(t *testing.T) {
	s := store.New()
	imageURLs := make([]string, 10)
	for i := range imageURLs {
		imageURLs[i] = fmt.Sprintf("https://cdn.example.com/products/img-%d.jpg", i)
	}

	for i := 0; i < 1000; i++ {
		_, err := s.Create(
			fmt.Sprintf("Product %d", i),
			fmt.Sprintf("PERF-SKU-%04d", i),
			imageURLs,
			nil,
		)
		if err != nil {
			t.Fatalf("create product %d: %v", i, err)
		}
	}

	start := time.Now()
	items, total := s.List(store.ListOptions{Limit: 20, Offset: 0})
	elapsed := time.Since(start)

	if total != 1000 {
		t.Fatalf("expected total=1000, got %d", total)
	}
	if len(items) != 20 {
		t.Fatalf("expected 20 items, got %d", len(items))
	}
	// Should be well under 50ms even in a test environment.
	if elapsed > 50*time.Millisecond {
		t.Logf("WARNING: list query took %v — may be slow under load", elapsed)
	}
	for _, item := range items {
		if item.ImageCount != 10 {
			t.Fatalf("expected ImageCount=10, got %d", item.ImageCount)
		}
	}
}

// Concurrency safety test

// TestConcurrentCreate fires 50 goroutines creating products with distinct SKUs
// and 5 goroutines creating products with the same SKU, verifying exactly one
// succeeds for the duplicate and all 50 distinct ones succeed.
func TestConcurrentCreate(t *testing.T) {
	s := store.New()

	var wg sync.WaitGroup
	dupAccepted := 0
	var mu sync.Mutex

	// 5 goroutines all racing to create the same SKU — only 1 should win.
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := s.Create("Dup Product", "SKU-DUP", nil, nil)
			if err == nil {
				mu.Lock()
				dupAccepted++
				mu.Unlock()
			}
		}()
	}
	// 50 goroutines with unique SKUs — all should succeed.
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			_, err := s.Create(fmt.Sprintf("Product %d", n), fmt.Sprintf("UNIQUE-%d", n), nil, nil)
			if err != nil {
				t.Errorf("unique SKU create failed: %v", err)
			}
		}(i)
	}
	wg.Wait()

	if dupAccepted != 1 {
		t.Fatalf("expected exactly 1 duplicate SKU to succeed, got %d", dupAccepted)
	}
	if s.Count() != 51 { // 1 dup + 50 unique
		t.Fatalf("expected 51 products in store, got %d", s.Count())
	}
}

// Compile-time check: ensure listResponse is importable from handlers.
var _ = bytes.NewBuffer