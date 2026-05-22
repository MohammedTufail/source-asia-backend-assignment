// Centralises all input validation rules for the product catalog API.
// URL validation enforces http/https scheme, a 2048-character maximum
// length, and basic structural sanity — no external HTTP calls are made.
// Array limits (max 20 URLs per request) and required-field checks are
// also defined here so handler code stays focused on request routing.

package validator

import (
	"fmt"
	"net/url"
	"strings"
)

const (
	// MaxURLLength is the maximum allowed length for any single URL string.
	MaxURLLength = 2048

	// MaxURLsPerRequest is the maximum number of URLs accepted in a single
	// image_urls or video_urls array within one request body.
	MaxURLsPerRequest = 20
)

// ValidateURL returns an error if the URL string is not acceptable.
// Rules:
//   - Must not be empty.
//   - Must not exceed MaxURLLength characters.
//   - Must be parseable by net/url.
//   - Scheme must be "http" or "https".
//   - Host must be non-empty (e.g. rejects "https://" alone).
func ValidateURL(raw string) error {
	if raw == "" {
		return fmt.Errorf("URL must not be empty")
	}
	if len(raw) > MaxURLLength {
		return fmt.Errorf("URL exceeds maximum length of %d characters", MaxURLLength)
	}

	parsed, err := url.ParseRequestURI(raw)
	if err != nil {
		return fmt.Errorf("URL is not valid: %w", err)
	}

	scheme := strings.ToLower(parsed.Scheme)
	if scheme != "http" && scheme != "https" {
		return fmt.Errorf("URL scheme must be http or https, got %q", parsed.Scheme)
	}

	if parsed.Host == "" {
		return fmt.Errorf("URL must include a host")
	}

	return nil
}

// ValidateURLSlice validates every URL in a slice and enforces the per-request
// array size limit. Returns a descriptive error referencing the field name
// so handlers can pass it directly to the client.
func ValidateURLSlice(fieldName string, urls []string) error {
	if len(urls) > MaxURLsPerRequest {
		return fmt.Errorf("%s exceeds maximum of %d URLs per request", fieldName, MaxURLsPerRequest)
	}
	for i, u := range urls {
		if err := ValidateURL(u); err != nil {
			return fmt.Errorf("%s[%d]: %w", fieldName, i, err)
		}
	}
	return nil
}

// ValidateProductInput checks the name and SKU fields for the create-product
// request and validates both URL arrays. Returns the first validation error
// found, or nil if all inputs are acceptable.
func ValidateProductInput(name, sku string, imageURLs, videoURLs []string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("name is required and must be a non-empty string")
	}
	if strings.TrimSpace(sku) == "" {
		return fmt.Errorf("sku is required and must be a non-empty string")
	}
	if err := ValidateURLSlice("image_urls", imageURLs); err != nil {
		return err
	}
	if err := ValidateURLSlice("video_urls", videoURLs); err != nil {
		return err
	}
	return nil
}

// ValidateMediaInput checks that at least one URL is provided and that both
// arrays are valid. Used by POST /products/{id}/media.
func ValidateMediaInput(imageURLs, videoURLs []string) error {
	if len(imageURLs) == 0 && len(videoURLs) == 0 {
		return fmt.Errorf("at least one of image_urls or video_urls must be provided with at least one URL")
	}
	if err := ValidateURLSlice("image_urls", imageURLs); err != nil {
		return err
	}
	if err := ValidateURLSlice("video_urls", videoURLs); err != nil {
		return err
	}
	return nil
}