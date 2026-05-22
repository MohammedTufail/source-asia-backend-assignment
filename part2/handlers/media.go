// HTTP handler for POST /products/{id}/media — appends new image and/or
// video URLs to an existing product. At least one URL array must be
// non-empty; both are validated before any mutation occurs. Returns the
// full updated product on success so the client has the current state.

package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"source-asia-backend-assignment/part2/store"
	"source-asia-backend-assignment/part2/validator"
)

// mediaRequest is the expected JSON body for POST /products/{id}/media.
type mediaRequest struct {
	ImageURLs []string `json:"image_urls"`
	VideoURLs []string `json:"video_urls"`
}

// Media returns an http.Handler for POST /products/{id}/media.
// The id is extracted from the URL path; both arrays are validated in full
// before any write is attempted, ensuring atomicity from the client's view.
func Media(s *store.Store) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, errResp("method_not_allowed", ""))
			return
		}

		// Path pattern: /products/{id}/media
		// Extract the id segment between the two slashes.
		path := strings.TrimPrefix(r.URL.Path, "/products/")
		path = strings.TrimSuffix(path, "/media")
		id := path

		if id == "" {
			writeJSON(w, http.StatusBadRequest, errResp("missing_id", "Product ID is required."))
			return
		}

		// Verify product exists before parsing the body.
		if s.GetByID(id) == nil {
			writeJSON(w, http.StatusNotFound, errResp("404 - not_found", "No product found with the given ID."))
			return
		}

		var req mediaRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, errResp("invalid_json", "Request body must be valid JSON."))
			return
		}

		if err := validator.ValidateMediaInput(req.ImageURLs, req.VideoURLs); err != nil {
			writeJSON(w, http.StatusBadRequest, errResp("validation_error", err.Error()))
			return
		}

		updated, err := s.AddMedia(id, req.ImageURLs, req.VideoURLs)
		if err != nil {
			// Should only happen in a race where the product was deleted between
			// the existence check above and the write — extremely unlikely in
			// single-instance in-memory mode but handled defensively.
			writeJSON(w, http.StatusNotFound, errResp("400 - not_found", err.Error()))
			return
		}

		writeJSON(w, http.StatusOK, updated)
	})
}