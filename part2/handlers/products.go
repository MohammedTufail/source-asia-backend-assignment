// HTTP handlers for the three core product endpoints:
//   POST   /products          — create a new product (201 or 409 on duplicate SKU)
//   GET    /products          — paginated list with lightweight items (no media arrays)
//   GET    /products/{id}     — full product detail including all image and video URLs
//
// Pagination defaults: limit=20, offset=0; maximum limit is 100.
// All responses and errors are JSON; the error shape is {"error":"...", "message":"..."}.

package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"source-asia-backend-assignment/part2/store"
	"source-asia-backend-assignment/part2/validator"
)

// createProductRequest is the expected JSON body for POST /products.
type createProductRequest struct {
	Name      string   `json:"name"`
	SKU       string   `json:"sku"`
	ImageURLs []string `json:"image_urls"`
	VideoURLs []string `json:"video_urls"`
}

// listResponse is the envelope returned by GET /products.
type listResponse struct {
	Data       interface{} `json:"data"`
	Total      int         `json:"total"`
	Limit      int         `json:"limit"`
	Offset     int         `json:"offset"`
	HasMore    bool        `json:"has_more"`
}

// Products returns an http.Handler that routes /products and /products/{id}.
// Using a single handler with prefix routing avoids the need for an external
// router dependency while keeping the code in one place.
func Products(s *store.Store) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/products", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			createProduct(s, w, r)
		case http.MethodGet:
			listProducts(s, w, r)
		default:
			writeJSON(w, http.StatusMethodNotAllowed, errResp("method_not_allowed", ""))
		}
	})
	mux.HandleFunc("/products/", func(w http.ResponseWriter, r *http.Request) {
		// Strip "/products/" prefix to get the id segment.
		id := strings.TrimPrefix(r.URL.Path, "/products/")

		// Reject sub-paths like /products/123/extra (media has its own handler).
		if strings.Contains(id, "/") {
			writeJSON(w, http.StatusNotFound, errResp("not_found", ""))
			return
		}

		if r.Method != http.MethodGet {
			writeJSON(w, http.StatusMethodNotAllowed, errResp("method_not_allowed", ""))
			return
		}
		getProduct(s, w, r, id)
	})
	return mux
}

// createProduct handles POST /products.
func createProduct(s *store.Store, w http.ResponseWriter, r *http.Request) {
	var req createProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errResp("invalid_json", "Request body must be valid JSON."))
		return
	}

	if err := validator.ValidateProductInput(req.Name, req.SKU, req.ImageURLs, req.VideoURLs); err != nil {
		writeJSON(w, http.StatusBadRequest, errResp("validation_error", err.Error()))
		return
	}

	product, err := s.Create(req.Name, req.SKU, req.ImageURLs, req.VideoURLs)
	if err != nil {
		// Store returns an error only on duplicate SKU — treat as 409 Conflict.
		writeJSON(w, http.StatusConflict, errResp("409 - duplicate_sku", err.Error()))
		return
	}

	writeJSON(w, http.StatusCreated, product)
}

// listProducts handles GET /products with limit/offset pagination.
func listProducts(s *store.Store, w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	limit := store.DefaultLimit
	if v := q.Get("limit"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 {
			writeJSON(w, http.StatusBadRequest, errResp("invalid_parameter", "limit must be a positive integer"))
			return
		}
		if n > store.MaxLimit {
			n = store.MaxLimit
		}
		limit = n
	}

	offset := 0
	if v := q.Get("offset"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 0 {
			writeJSON(w, http.StatusBadRequest, errResp("invalid_parameter", "offset must be a non-negative integer"))
			return
		}
		offset = n
	}

	items, total := s.List(store.ListOptions{Limit: limit, Offset: offset})

	writeJSON(w, http.StatusOK, listResponse{
		Data:    items,
		Total:   total,
		Limit:   limit,
		Offset:  offset,
		HasMore: offset+len(items) < total,
	})
}

// getProduct handles GET /products/{id}.
func getProduct(s *store.Store, w http.ResponseWriter, r *http.Request, id string) {
	if id == "" {
		writeJSON(w, http.StatusBadRequest, errResp("missing_id", "Product ID is required."))
		return
	}

	product := s.GetByID(id)
	if product == nil {
		writeJSON(w, http.StatusNotFound, errResp("not_found", "No product found with the given ID."))
		return
	}

	writeJSON(w, http.StatusOK, product)
}