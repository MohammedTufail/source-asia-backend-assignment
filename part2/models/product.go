// Defines the core data types for the product catalog.
// The design separates the full Product (used on detail pages) from the
// lightweight ProductListItem (used on list pages) so that the list endpoint
// never needs to load or serialise thousands of media URLs — only counts
// and an optional thumbnail are included in list responses.

package models

import "time"

// Product is the canonical in-memory representation of a product.
// image_urls and video_urls are stored as plain string slices; no external
// CDN or storage integration is required.
type Product struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	SKU       string    `json:"sku"`
	ImageURLs []string  `json:"image_urls"`
	VideoURLs []string  `json:"video_urls"`
	CreatedAt time.Time `json:"created_at"`
}

// ProductListItem is the trimmed shape returned by GET /products.
// It intentionally omits ImageURLs and VideoURLs and replaces them with
// counts plus an optional thumbnail, keeping the list response fast
// even when products have hundreds of stored media URLs.
type ProductListItem struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	SKU          string    `json:"sku"`
	ImageCount   int       `json:"image_count"`
	VideoCount   int       `json:"video_count"`
	ThumbnailURL string    `json:"thumbnail_url,omitempty"` // first image_url if available
	CreatedAt    time.Time `json:"created_at"`
}

// ToListItem converts a full Product to its lightweight list representation.
// This is the only place where the projection is defined, keeping the logic
// in one spot so future changes (e.g. adding a new field) are made once.
func (p *Product) ToListItem() ProductListItem {
	item := ProductListItem{
		ID:         p.ID,
		Name:       p.Name,
		SKU:        p.SKU,
		ImageCount: len(p.ImageURLs),
		VideoCount: len(p.VideoURLs),
		CreatedAt:  p.CreatedAt,
	}
	if len(p.ImageURLs) > 0 {
		item.ThumbnailURL = p.ImageURLs[0]
	}
	return item
}