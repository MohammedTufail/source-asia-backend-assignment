// In-memory product store with full concurrency safety via sync.RWMutex.
// Read operations (list, get) acquire a shared read-lock so they run
// in parallel; write operations (create, add media) acquire an exclusive
// write-lock. The store keeps a separate SKU index for O(1) duplicate
// detection and a stable insertion-order slice for consistent pagination.

package store

import (
	"crypto/rand"
	"fmt"
	"sync"
	"time"

	"source-asia-backend-assignment/part2/models"
)

// Store is the single source of truth for all product data.
// It is safe to use from multiple goroutines simultaneously.
type Store struct {
	mu       sync.RWMutex
	products map[string]*models.Product // keyed by product ID
	skuIndex map[string]string          // SKU → product ID for fast duplicate checks
	order    []string                   // insertion-order IDs for stable pagination
}

// New returns an initialised, empty Store ready for use.
func New() *Store {
	return &Store{
		products: make(map[string]*models.Product),
		skuIndex: make(map[string]string),
		order:    make([]string, 0),
	}
}

// newID generates a random hex ID using crypto/rand — no external dependency needed.
func newID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

// Create adds a new product to the store.
// Returns a descriptive error if the SKU already exists (use as 409 Conflict).
func (s *Store) Create(name, sku string, imageURLs, videoURLs []string) (*models.Product, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.skuIndex[sku]; exists {
		return nil, fmt.Errorf("product with SKU %q already exists", sku)
	}

	if imageURLs == nil {
		imageURLs = []string{}
	}
	if videoURLs == nil {
		videoURLs = []string{}
	}

	p := &models.Product{
		ID:        newID(),
		Name:      name,
		SKU:       sku,
		ImageURLs: imageURLs,
		VideoURLs: videoURLs,
		CreatedAt: time.Now().UTC(),
	}

	s.products[p.ID] = p
	s.skuIndex[sku] = p.ID
	s.order = append(s.order, p.ID)
	return p, nil
}

// ListOptions controls pagination for the List method.
type ListOptions struct {
	Limit  int // number of items to return; capped at MaxLimit
	Offset int // zero-based starting index
}

const (
	DefaultLimit = 20
	MaxLimit     = 100
)

// List returns a page of lightweight ProductListItems without loading full
// media URL arrays. Only fields needed for a grid/list UI are included.
// See models.ProductListItem and models.Product.ToListItem() for the projection.
func (s *Store) List(opts ListOptions) (items []models.ProductListItem, total int) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	total = len(s.order)

	if opts.Offset >= total {
		return []models.ProductListItem{}, total
	}

	end := opts.Offset + opts.Limit
	if end > total {
		end = total
	}

	page := s.order[opts.Offset:end]
	items = make([]models.ProductListItem, 0, len(page))
	for _, id := range page {
		if p, ok := s.products[id]; ok {
			// ToListItem() projects only lightweight fields — it does NOT
			// copy ImageURLs/VideoURLs, keeping the list response fast.
			items = append(items, p.ToListItem())
		}
	}
	return items, total
}

// GetByID returns the full product (including all media URLs) by ID.
// Returns nil if no product with that ID exists.
func (s *Store) GetByID(id string) *models.Product {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.products[id]
}

// AddMedia appends new image and video URLs to an existing product.
// Returns an error if the product ID is not found.
func (s *Store) AddMedia(productID string, imageURLs, videoURLs []string) (*models.Product, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	p, ok := s.products[productID]
	if !ok {
		return nil, fmt.Errorf("product not found")
	}

	p.ImageURLs = append(p.ImageURLs, imageURLs...)
	p.VideoURLs = append(p.VideoURLs, videoURLs...)
	return p, nil
}

// Count returns the total number of products in the store.
func (s *Store) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.products)
}