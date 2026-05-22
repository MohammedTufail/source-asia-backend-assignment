# Source Asia — Backend Assignment

A two-part HTTP service written in **Go 1.26** using only the standard library- zero external dependencies. Both parts run as independent binaries on separate ports and share a single Go module.

---

## Table of Contents

1. [Repository Structure](#repository-structure)
2. [Quick Start](#quick-start)
3. [Running Tests](#running-tests)
4. [Part 1 — Rate-limited API](#part-1--rate-limited-api)
   - [Design Decisions](#rate-limiting-design-decisions)
   - [Endpoints](#part-1-endpoints)
   - [PowerShell Examples](#part-1-powershell-examples)
   - [curl Examples](#part-1-curl-examples)
   - [Production Limitations](#part-1-production-limitations)
5. [Part 2 — Product Catalog API](#part-2--product-catalog-api)
   - [Data Model](#data-model)
   - [Endpoints](#part-2-endpoints)
   - [Validation Rules](#validation-rules)
   - [PowerShell Examples](#part-2-powershell-examples)
   - [curl Examples](#part-2-curl-examples)
   - [Production Limitations](#part-2-production-limitations)
6. [AI Tool Usage](#ai-tool-usage)
7. [Test Summary](#test-summary)

---

## Repository Structure

```
source-asia-assignment/
├── go.mod                          Module definition (no external dependencies)
├── README.md
│
├── part1/                          Part 1 — Rate-limited API  (port 8081)
│   ├── main.go                     Entry point; wires the server together
│   ├── ratelimiter.go              Rolling-window rate limiter (concurrency-safe)
│   ├── handlers.go                 HTTP handlers: POST /request, GET /stats
│   └── ratelimiter_test.go         13 tests — unit, integration, concurrency
│   └── testing_log_part1.md
|   └── video.mp4
|
└── part2/                          Part 2 — Product Catalog API  (port 8082)
    ├── main.go                     Entry point; wires the server together
    ├── part2_test.go               25 tests — validator, store, HTTP, perf, concurrency
    ├── models/
    │   └── product.go              Product + ProductListItem types; ToListItem() projection
    ├── store/
    │   └── store.go                In-memory RWMutex-protected store; SKU index; pagination
    ├── handlers/
    │   ├── helpers.go              Shared writeJSON() + error envelope
    │   ├── products.go             POST /products · GET /products · GET /products/{id}
    │   └── media.go                POST /products/{id}/media
    |── validator/
    |   └── validator.go            URL scheme/length/structure checks; array-size limits
    |── testing_log_part2.md
    └── video.mp4


```

---

## Quick Start

### Prerequisites

- **Go 1.22+** — verify with `go version`

### Run Part 1 (Rate-limited API - port 8081)

```bash
go run ./part1
# Output: Part 1 — Rate-limited API running on http://localhost:8081
```

### Run Part 2 (Product Catalog API - port 8082)

```bash
go run ./part2
# Output: Part 2 — Product Catalog API running on http://localhost:8082
```

### Run Both Simultaneously

```bash
# Terminal 1
go run ./part1

# Terminal 2
go run ./part2
```

Or with built binaries:

```bash
go build -o bin/part1 ./part1 && go build -o bin/part2 ./part2

# Windows
start bin\part1.exe
start bin\part2.exe

# Linux / macOS
./bin/part1 &
./bin/part2
```

### Override Port

```bash
# Linux / macOS
PORT=8081 go run ./part1
PORT=8082 go run ./part2

# Windows PowerShell
$env:PORT="8081"; go run ./part1
$env:PORT="8082"; go run ./part2
```

---

## Running Tests

```bash
# Run all tests across both parts
go test ./...

# Verbose output
go test ./... -v

# With Go's built-in race detector (detects data races under concurrency)
go test -race ./...

# Run only Part 1 tests
go test ./part1/... -v

# Run only Part 2 tests
go test ./part2/... -v

# Run a specific test by name
go test ./part2 -run TestPerformanceInvariant -v
go test ./part2 -run TestConcurrentCreate -v
```

**Expected output (all passing):**

```
ok    source-asia-assignment/part1    0.017s
ok    source-asia-assignment/part2    0.018s
```

---

## Part 1 — Rate-limited API

### Rate Limiting Design Decisions

| Decision             | Choice                            | Reason                                                                                                                                                                                                                                       |
| -------------------- | --------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Window type**      | Rolling (sliding) window          | A fixed window allows a burst of 10 requests at the boundary — 5 at the end of window N and 5 at the start of N+1. A rolling window prevents this entirely.                                                                                  |
| **Storage**          | `[]time.Time` timestamps per user | Storing actual timestamps makes rolling-window semantics exact. Expired entries are evicted lazily on every `Allow()` call, so no background goroutine is needed.                                                                            |
| **Concurrency**      | Single `sync.Mutex` in `Allow()`  | The entire check-and-record must be atomic. A read-write lock would not help because `Allow()` always writes. All parallel calls for the same `user_id` are serialised here, making it impossible to exceed the limit under concurrent load. |
| **Success code**     | `201 Created`                     | Each accepted request creates a new entry in the rate-limit window — semantically a resource creation.                                                                                                                                       |
| **Rejected counter** | Cumulative (never reset)          | Cumulative totals are more useful for auditing abuse patterns. A per-window counter would obscure total historical pressure on the system.                                                                                                   |

---

### Part 1 Endpoints

#### `POST /request`

Accepts a user request and enforces the per-user rate limit.

**Request body:**

```json
{
  "user_id": "alice",
  "payload": { "any": "valid JSON value" }
}
```

| Field     | Type     | Rules                                 |
| --------- | -------- | ------------------------------------- |
| `user_id` | string   | Required, non-empty                   |
| `payload` | any JSON | Required, must not be null or missing |

**Responses:**

| Status                  | When                                                                | Body                                                                                 |
| ----------------------- | ------------------------------------------------------------------- | ------------------------------------------------------------------------------------ |
| `201 Created`           | Request accepted within limit                                       | `{"status":"accepted","user_id":"alice","message":"Request accepted successfully."}` |
| `400 Bad Request`       | Missing/empty `user_id`, missing `payload`, invalid JSON            | `{"error":"missing_user_id","message":"..."}`                                        |
| `429 Too Many Requests` | User has exceeded 5 requests in the current 1-minute rolling window | `{"error":"rate_limit_exceeded","message":"..."}`                                    |

---

#### `GET /stats`

Returns a snapshot of per-user request statistics.

**Response — `200 OK`:**

```json
{
  "users": {
    "alice": {
      "accepted": 5,
      "window_accepted": 3,
      "rejected_cumulative": 2
    }
  },
  "note": "window_accepted reflects requests in the current 1-minute rolling window. rejected_cumulative is the total rejected count since server start."
}
```

**Response schema:**

| Field                 | Type | Description                                                   |
| --------------------- | ---- | ------------------------------------------------------------- |
| `accepted`            | int  | Total accepted requests since server start                    |
| `window_accepted`     | int  | Accepted requests within the current 1-minute rolling window  |
| `rejected_cumulative` | int  | Total rejections since server start — cumulative, never reset |

---

### Part 1 PowerShell Examples

> These are the exact commands verified on Windows. Replace `8081` with your `PORT` if overridden.

```powershell
# ── 1. Send a valid request ────────────────────────────────────────────────
Invoke-RestMethod `
  -Method POST `
  -Uri "http://localhost:8081/request" `
  -ContentType "application/json" `
  -Body '{"user_id":"alice","payload":{"action":"buy","item":"widget"}}' |
ConvertTo-Json

# Expected: 201, body shows status: "accepted"


# ── 2. Hit the rate limit (send 6 requests for the same user) ─────────────
for ($i = 1; $i -le 6; $i++) {
  try {
    Invoke-RestMethod `
      -Method POST `
      -Uri "http://localhost:8081/request" `
      -ContentType "application/json" `
      -Body '{"user_id":"user-limit","payload":1}'
    Write-Host "Request $i : 201 Accepted"
  } catch {
    Write-Host "Request $i : $($_.Exception.Response.StatusCode.value__) — $($_.ErrorDetails.Message)"
  }
}
# Expected: requests 1-5 → 201, request 6 → 429 rate_limit_exceeded


# ── 3. Missing user_id → 400 ───────────────────────────────────────────────
try {
  Invoke-RestMethod `
    -Method POST `
    -Uri "http://localhost:8081/request" `
    -ContentType "application/json" `
    -Body '{"payload":"no user"}'
} catch {
  $_.ErrorDetails.Message
}
# Expected: {"error":"missing_user_id","message":"user_id is required and must be a non-empty string."}


# ── 4. Empty user_id → 400 ─────────────────────────────────────────────────
try {
  Invoke-RestMethod `
    -Method POST `
    -Uri "http://localhost:8081/request" `
    -ContentType "application/json" `
    -Body '{"user_id":"","payload":1}'
} catch {
  $_.ErrorDetails.Message
}
# Expected: {"error":"missing_user_id","message":"..."}


# ── 5. Invalid JSON → 400 ──────────────────────────────────────────────────
try {
  Invoke-RestMethod `
    -Method POST `
    -Uri "http://localhost:8081/request" `
    -ContentType "application/json" `
    -Body 'not valid json'
} catch {
  $_.ErrorDetails.Message
}
# Expected: {"error":"invalid_json","message":"..."}


# ── 6. Get stats ───────────────────────────────────────────────────────────
Invoke-RestMethod `
  -Method GET `
  -Uri "http://localhost:8081/stats" |
ConvertTo-Json -Depth 10
# Expected: per-user breakdown with accepted, window_accepted, rejected_cumulative
```

**Actual output observed (after running the tests above):**

```json
{
  "users": {
    "concurrent-user": {
      "accepted": 5,
      "rejected_cumulative": 15,
      "window_accepted": 0
    },
    "user-1": {
      "accepted": 1,
      "rejected_cumulative": 0,
      "window_accepted": 0
    },
    "user-limit": {
      "accepted": 5,
      "rejected_cumulative": 1,
      "window_accepted": 0
    }
  },
  "note": "window_accepted reflects requests in the current 1-minute rolling window. rejected_cumulative is the total rejected count since server start."
}
```

> `window_accepted` shows `0` because these requests were sent more than 1 minute before the `/stats` call — they have expired from the rolling window. The `accepted` and `rejected_cumulative` counts are permanent.

---

### Part 1 curl Examples

```bash
# Send a valid request
curl -s -X POST http://localhost:8081/request \
  -H "Content-Type: application/json" \
  -d '{"user_id":"alice","payload":{"action":"buy","item":"widget"}}'

# Hit the rate limit (run 6 times rapidly — 6th should return 429)
for i in $(seq 1 6); do
  curl -s -o /dev/null -w "Request $i: HTTP %{http_code}\n" \
    -X POST http://localhost:8081/request \
    -H "Content-Type: application/json" \
    -d '{"user_id":"bob","payload":1}'
done

# Missing user_id → 400
curl -s -X POST http://localhost:8081/request \
  -H "Content-Type: application/json" \
  -d '{"payload":"no user id here"}'

# Empty user_id → 400
curl -s -X POST http://localhost:8081/request \
  -H "Content-Type: application/json" \
  -d '{"user_id":"","payload":1}'

# Invalid JSON → 400
curl -s -X POST http://localhost:8081/request \
  -H "Content-Type: application/json" \
  -d 'not valid json'

# Get stats
curl -s http://localhost:8081/stats | python3 -m json.tool
```

---

### Part 1 Production Limitations

| Limitation                        | Impact                                                                                        | Production Fix                                                                                |
| --------------------------------- | --------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------- |
| **Single instance only**          | Rate-limit state lives in one process — horizontal scaling is impossible without coordination | Move state to Redis; use atomic `INCR` + `EXPIRE` or a Lua script for the check-and-increment |
| **Memory only**                   | All counters reset on restart                                                                 | Persist to Redis (with TTL) or a time-series store                                            |
| **No user eviction**              | Memory grows unboundedly for large or constantly-changing user sets                           | Background goroutine that prunes entries with no activity in the last N minutes               |
| **No authentication on `/stats`** | Any caller can enumerate all user IDs and their traffic patterns                              | Add API key, JWT, or IP-allowlist middleware                                                  |
| **No distributed tracing**        | Hard to debug request attribution across services                                             | Integrate OpenTelemetry with trace/span IDs                                                   |

---

## Part 2 — Product Catalog API

### Data Model

All product data is held in a single `Store` struct protected by a `sync.RWMutex`:

```
Store
├── products   map[string]*Product    — id → full Product (all fields including URL slices)
├── skuIndex   map[string]string      — sku → id (O(1) duplicate detection on create)
└── order      []string               — insertion-order product IDs (stable pagination)
```

**Why three structures?**

- `products` is the source of truth — full data including all image/video URL slices.
- `skuIndex` avoids scanning all products on every `POST /products` just to check uniqueness.
- `order` gives a deterministic page order. Go maps have random iteration; without this slice, `offset=20` on one request might return different products than the next.

**List vs Detail — the key performance design:**

|                                    | `GET /products` (list)                                                                   | `GET /products/{id}` (detail)                |
| ---------------------------------- | ---------------------------------------------------------------------------------------- | -------------------------------------------- |
| Lock type                          | `RLock` (shared — multiple readers run in parallel)                                      | `RLock` (shared)                             |
| Data read                          | `order[offset:offset+limit]` → project via `ToListItem()`                                | `products[id]` — full struct                 |
| Media URLs serialised              | ❌ **Never.** Only `image_count`, `video_count`, `thumbnail_url` (first image)           | ✅ Full `image_urls` and `video_urls` arrays |
| Cost at 1,000 products × 10 images | Reads 20 lightweight structs; 9,980 products and all 10,000 image URLs are never touched | Reads exactly 1 product                      |

This guarantee is enforced in a single place: `models.Product.ToListItem()`. The list handler calls this projection function and never accesses `p.ImageURLs` or `p.VideoURLs` directly. The detail handler returns `*Product` as-is.

---

### Part 2 Endpoints

#### `POST /products`

Creates a new product. `image_urls` and `video_urls` are optional on creation — media can be added later via `POST /products/{id}/media`.

**Request body:**

```json
{
  "name": "Widget A",
  "sku": "SKU-001",
  "image_urls": [
    "https://cdn.example.com/products/sku-001/img-1.jpg",
    "https://cdn.example.com/products/sku-001/img-2.jpg"
  ],
  "video_urls": ["https://cdn.example.com/products/sku-001/demo.mp4"]
}
```

**Success — `201 Created`:**

```json
{
  "id": "59a75a00-731d-90e2-b5af-4eaf2e3d5fd6",
  "name": "Widget A",
  "sku": "SKU-001",
  "image_urls": [
    "https://cdn.example.com/products/sku-001/img-1.jpg",
    "https://cdn.example.com/products/sku-001/img-2.jpg"
  ],
  "video_urls": ["https://cdn.example.com/products/sku-001/demo.mp4"],
  "created_at": "2026-05-22T22:51:35Z"
}
```

| Status            | When                                                               |
| ----------------- | ------------------------------------------------------------------ |
| `201 Created`     | Product created successfully                                       |
| `400 Bad Request` | Empty `name`, empty `sku`, invalid URL, or too many URLs per array |
| `409 Conflict`    | A product with this `sku` already exists                           |

> **Why 409 for duplicate SKU?** HTTP 409 Conflict is semantically correct: the conflict is between the incoming request and the _current state of the server_ (an existing product owns that SKU). HTTP 400 means the request itself is malformed — which it isn't.

---

#### `GET /products`

Paginated list for a UI grid. **Never returns `image_urls` or `video_urls` arrays** — only counts and an optional thumbnail.

**Query parameters:**

| Parameter | Default | Maximum | Description               |
| --------- | ------- | ------- | ------------------------- |
| `limit`   | `20`    | `100`   | Number of items per page  |
| `offset`  | `0`     | —       | Zero-based starting index |

**Response — `200 OK`:**

```json
{
  "data": [
    {
      "id": "59a75a00-731d-90e2-b5af-4eaf2e3d5fd6",
      "name": "Widget A",
      "sku": "SKU-001",
      "image_count": 2,
      "video_count": 1,
      "thumbnail_url": "https://cdn.example.com/products/sku-001/img-1.jpg",
      "created_at": "2026-05-22T22:51:35Z"
    }
  ],
  "total": 1001,
  "limit": 20,
  "offset": 0,
  "has_more": true
}
```

| Field      | Description                                                        |
| ---------- | ------------------------------------------------------------------ |
| `data`     | Array of lightweight list items — no URL arrays                    |
| `total`    | Total number of products in the store (for UI pagination controls) |
| `limit`    | The limit used for this response                                   |
| `offset`   | The offset used for this response                                  |
| `has_more` | `true` if there are more products beyond this page                 |

---

#### `GET /products/{id}`

Returns the full product including all stored `image_urls` and `video_urls`.

**`200 OK`** — full product (same shape as `POST /products` success response).

**`404 Not Found`:**

```json
{
  "error": "not_found",
  "message": "No product found with the given ID."
}
```

---

#### `POST /products/{id}/media`

Appends new image and/or video URLs to an existing product. At least one array must contain at least one URL.

**Request body:**

```json
{
  "image_urls": ["https://cdn.example.com/products/sku-001/img-3.jpg"],
  "video_urls": ["https://cdn.example.com/products/sku-001/demo-v2.mp4"]
}
```

**`200 OK`** — returns the full updated product with all appended URLs included.

| Status            | When                                           |
| ----------------- | ---------------------------------------------- |
| `200 OK`          | URLs appended successfully                     |
| `400 Bad Request` | Both arrays empty, or any URL fails validation |
| `404 Not Found`   | Product ID does not exist                      |

---

### Validation Rules

| Rule                       | Detail                                                                             |
| -------------------------- | ---------------------------------------------------------------------------------- |
| `name`                     | Required; rejected if empty or whitespace-only                                     |
| `sku`                      | Required; rejected if empty or whitespace-only; must be unique across all products |
| URL scheme                 | Must be `http://` or `https://`. Rejected: `ftp://`, `//relative`, bare hostnames  |
| URL length                 | Maximum **2048 characters** per URL                                                |
| URL structure              | Parsed via `net/url.ParseRequestURI`; host segment must be non-empty               |
| Max URLs per array         | **20 URLs** per `image_urls` or `video_urls` array per single request              |
| Media endpoint requirement | At least one of `image_urls` or `video_urls` must be non-empty                     |

---

### Part 2 PowerShell Examples

> These are the exact commands verified on Windows. The server runs on port `8082` in these examples — adjust to `8082` (default) if you haven't set a custom PORT.

```powershell
# ── 1. Create a product ────────────────────────────────────────────────────
try {
  Invoke-RestMethod `
    -Method POST `
    -Uri "http://localhost:8082/products" `
    -ContentType "application/json" `
    -Body '{
      "name": "Widget A",
      "sku": "SKU-001",
      "image_urls": [
        "https://cdn.example.com/products/sku-001/img1.jpg"
      ],
      "video_urls": [
        "https://cdn.example.com/products/sku-001/demo.mp4"
      ]
    }' |
  ConvertTo-Json -Depth 10
} catch {
  $_.ErrorDetails.Message
}
# Expected: 201 — full product object with assigned "id" and "created_at"


# ── 2. Duplicate SKU → 409 ─────────────────────────────────────────────────
try {
  Invoke-RestMethod `
    -Method POST `
    -Uri "http://localhost:8082/products" `
    -ContentType "application/json" `
    -Body '{"name":"Duplicate Widget","sku":"SKU-001"}'
} catch {
  $_.ErrorDetails.Message
}
# Expected: {"error":"duplicate_sku","message":"product with SKU \"SKU-001\" already exists"}


# ── 3. Empty name → 400 ────────────────────────────────────────────────────
try {
  Invoke-RestMethod `
    -Method POST `
    -Uri "http://localhost:8082/products" `
    -ContentType "application/json" `
    -Body '{"name":"","sku":"SKU-EMPTY"}'
} catch {
  $_.ErrorDetails.Message
}
# Expected: {"error":"validation_error","message":"name is required and must be a non-empty string"}


# ── 4. Empty SKU → 400 ─────────────────────────────────────────────────────
try {
  Invoke-RestMethod `
    -Method POST `
    -Uri "http://localhost:8082/products" `
    -ContentType "application/json" `
    -Body '{"name":"Widget","sku":""}'
} catch {
  $_.ErrorDetails.Message
}
# Expected: {"error":"validation_error","message":"sku is required and must be a non-empty string"}


# ── 5. Invalid URL → 400 ───────────────────────────────────────────────────
try {
  Invoke-RestMethod `
    -Method POST `
    -Uri "http://localhost:8082/products" `
    -ContentType "application/json" `
    -Body '{"name":"Widget","sku":"SKU-BADURL","image_urls":["not-a-url"]}'
} catch {
  $_.ErrorDetails.Message
}
# Expected: {"error":"validation_error","message":"image_urls[0]: URL is not valid: ..."}


# ── 6. List products (default pagination) ─────────────────────────────────
Invoke-RestMethod `
  -Method GET `
  -Uri "http://localhost:8082/products" |
ConvertTo-Json -Depth 10
# Expected: data array with image_count/video_count — NO image_urls or video_urls arrays


# ── 7. List products with explicit limit + offset ─────────────────────────
Invoke-RestMethod `
  -Method GET `
  -Uri "http://localhost:8082/products?limit=3&offset=0" |
ConvertTo-Json -Depth 10


# ── 8. Get product by ID (full detail) ────────────────────────────────────
# Replace <id> with the id returned from step 1
Invoke-RestMethod `
  -Method GET `
  -Uri "http://localhost:8082/products/<id>" |
ConvertTo-Json -Depth 10
# Expected: full product with image_urls and video_urls arrays


# ── 9. Unknown ID → 404 ────────────────────────────────────────────────────
try {
  Invoke-RestMethod `
    -Method GET `
    -Uri "http://localhost:8082/products/unknown-id"
} catch {
  $_.ErrorDetails.Message
}
# Expected: {"error":"not_found","message":"No product found with the given ID."}


# ── 10. Append media to a product ─────────────────────────────────────────
# Replace <id> with a real product id
try {
  Invoke-RestMethod `
    -Method POST `
    -Uri "http://localhost:8082/products/<id>/media" `
    -ContentType "application/json" `
    -Body '{
      "image_urls": ["https://cdn.example.com/new-image.jpg"],
      "video_urls": ["https://cdn.example.com/new-video.mp4"]
    }' |
  ConvertTo-Json -Depth 10
} catch {
  $_.ErrorDetails.Message
}
# Expected: 200 — full updated product showing the new URLs appended


# ── 11. Empty media body → 400 ─────────────────────────────────────────────
try {
  Invoke-RestMethod `
    -Method POST `
    -Uri "http://localhost:8082/products/<id>/media" `
    -ContentType "application/json" `
    -Body '{"image_urls":[],"video_urls":[]}'
} catch {
  $_.ErrorDetails.Message
}
# Expected: {"error":"validation_error","message":"at least one of image_urls or video_urls must be provided..."}


# ── 12. Seed 1,000 products and verify list is fast ────────────────────────
for ($i = 1; $i -le 1000; $i++) {
  $body = @{
    name       = "Product-$i"
    sku        = "PERF-SKU-$i"
    image_urls = @(
      "https://cdn.example.com/$i/img1.jpg",
      "https://cdn.example.com/$i/img2.jpg",
      "https://cdn.example.com/$i/img3.jpg",
      "https://cdn.example.com/$i/img4.jpg",
      "https://cdn.example.com/$i/img5.jpg",
      "https://cdn.example.com/$i/img6.jpg",
      "https://cdn.example.com/$i/img7.jpg",
      "https://cdn.example.com/$i/img8.jpg",
      "https://cdn.example.com/$i/img9.jpg",
      "https://cdn.example.com/$i/img10.jpg"
    )
    video_urls = @()
  } | ConvertTo-Json -Depth 5

  Invoke-RestMethod `
    -Method POST `
    -Uri "http://localhost:8082/products" `
    -ContentType "application/json" `
    -Body $body | Out-Null
}

# Measure list performance — should complete in under 5ms
Measure-Command {
  Invoke-RestMethod `
    -Method GET `
    -Uri "http://localhost:8082/products?limit=20"
}
# Observed result: TotalMilliseconds ~1.95 ms with 1,001 products × 10 images stored
```

**Actual performance result observed (1,001 products, ~10,000 image URLs stored):**

```
TotalMilliseconds : 1.9525
```

The list query reads 20 lightweight structs and never touches the 10,000 stored image URL strings.

---

### Part 2 curl Examples

```bash
# Create a product
curl -s -X POST http://localhost:8082/products \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Widget A",
    "sku": "SKU-001",
    "image_urls": [
      "https://cdn.example.com/products/sku-001/img-1.jpg",
      "https://cdn.example.com/products/sku-001/img-2.jpg"
    ],
    "video_urls": [
      "https://cdn.example.com/products/sku-001/demo.mp4"
    ]
  }' | python3 -m json.tool

# Duplicate SKU → 409
curl -s -w "\nHTTP %{http_code}" -X POST http://localhost:8082/products \
  -H "Content-Type: application/json" \
  -d '{"name":"Widget B","sku":"SKU-001"}'

# Invalid URL → 400
curl -s -w "\nHTTP %{http_code}" -X POST http://localhost:8082/products \
  -H "Content-Type: application/json" \
  -d '{"name":"Bad","sku":"SKU-BAD","image_urls":["not-a-url"]}'

# List products (no media arrays in response)
curl -s "http://localhost:8082/products?limit=20&offset=0" | python3 -m json.tool

# Get full product detail (replace <id>)
curl -s http://localhost:8082/products/<id> | python3 -m json.tool

# Unknown ID → 404
curl -s -w "\nHTTP %{http_code}" http://localhost:8082/products/no-such-id

# Append media to a product
curl -s -X POST http://localhost:8082/products/<id>/media \
  -H "Content-Type: application/json" \
  -d '{
    "image_urls": ["https://cdn.example.com/products/sku-001/img-3.jpg"],
    "video_urls": ["https://cdn.example.com/products/sku-001/demo-v2.mp4"]
  }' | python3 -m json.tool

# Empty media arrays → 400
curl -s -w "\nHTTP %{http_code}" -X POST http://localhost:8082/products/<id>/media \
  -H "Content-Type: application/json" \
  -d '{"image_urls":[],"video_urls":[]}'

# Seed 1,000 products for performance testing
for i in $(seq 1 1000); do
  curl -s -o /dev/null -X POST http://localhost:8082/products \
    -H "Content-Type: application/json" \
    -d "{
      \"name\":\"Product $i\",
      \"sku\":\"SEED-$(printf '%04d' $i)\",
      \"image_urls\":[
        \"https://cdn.example.com/$i/img-1.jpg\",
        \"https://cdn.example.com/$i/img-2.jpg\",
        \"https://cdn.example.com/$i/img-3.jpg\",
        \"https://cdn.example.com/$i/img-4.jpg\",
        \"https://cdn.example.com/$i/img-5.jpg\",
        \"https://cdn.example.com/$i/img-6.jpg\",
        \"https://cdn.example.com/$i/img-7.jpg\",
        \"https://cdn.example.com/$i/img-8.jpg\",
        \"https://cdn.example.com/$i/img-9.jpg\",
        \"https://cdn.example.com/$i/img-10.jpg\"
      ]
    }"
done

# Verify performance — should be near-instant
time curl -s "http://localhost:8082/products?limit=20" | python3 -m json.tool
```

---

### Part 2 Production Limitations

| Limitation                 | Impact                                                                  | Production Fix                                                     |
| -------------------------- | ----------------------------------------------------------------------- | ------------------------------------------------------------------ |
| **In-memory only**         | All products lost on restart; no persistence                            | PostgreSQL with `products` + `product_media` tables (schema below) |
| **Single instance**        | Cannot scale horizontally                                               | Stateless HTTP service backed by a shared database                 |
| **No search or filtering** | List endpoint supports pagination only — no filter by name/SKU/category | Full-text search via PostgreSQL `tsvector` or Elasticsearch        |
| **No media deduplication** | Same URL can be appended multiple times via `POST /media`               | Unique constraint on `(product_id, url)` in the database           |
| **No soft deletes**        | No `DELETE /products/{id}` endpoint                                     | Add `deleted_at` column; filter in all queries                     |
| **No authentication**      | All endpoints are publicly writable                                     | API key or OAuth2 middleware                                       |

#### PostgreSQL + CDN design

```sql
-- Core product fields only — no URLs in this table
CREATE TABLE products (
  id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  name       TEXT        NOT NULL,
  sku        TEXT        NOT NULL UNIQUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- All media stored in a separate table — this is the critical design choice.
-- List queries NEVER join this table; they use COUNT(*) subqueries or
-- pre-aggregated counters. Detail queries join it for one product at a time.
CREATE TABLE product_media (
  id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  product_id UUID        NOT NULL REFERENCES products(id) ON DELETE CASCADE,
  media_type TEXT        NOT NULL CHECK (media_type IN ('image', 'video')),
  url        TEXT        NOT NULL,
  position   INT         NOT NULL DEFAULT 0,
  UNIQUE (product_id, url)
);

CREATE INDEX ON product_media (product_id);
```

**List query** — never reads `product_media` rows:

```sql
SELECT
  p.id, p.name, p.sku, p.created_at,
  COUNT(*) FILTER (WHERE m.media_type = 'image') AS image_count,
  COUNT(*) FILTER (WHERE m.media_type = 'video') AS video_count,
  MIN(m.url) FILTER (WHERE m.media_type = 'image' AND m.position = 0) AS thumbnail_url
FROM products p
LEFT JOIN product_media m ON m.product_id = p.id
GROUP BY p.id
ORDER BY p.created_at DESC
LIMIT $1 OFFSET $2;
```

**Detail query** — loads all media for exactly one product:

```sql
SELECT p.*, m.media_type, m.url, m.position
FROM products p
LEFT JOIN product_media m ON m.product_id = p.id
WHERE p.id = $1
ORDER BY m.media_type, m.position;
```

**CDN:** Store only the path segment (e.g. `/products/sku-001/img-1.jpg`) in the database; prepend the CDN base URL at serialisation time. This makes CDN migrations (e.g. switching providers) a single config change with zero data migration.

---

## AI Tool Usage

Claude (claude.ai) was used to assist with:

- Reviewing the overall file and package structure for clarity
- Suggesting test case coverage (edge cases, concurrency scenarios)
- Reviewing README formatting

All implementation logic, design decisions, and code were written and verified by the author.

---

## Test Summary

### Part 1 — 13 tests

| Test                                      | What it verifies                                                    |
| ----------------------------------------- | ------------------------------------------------------------------- |
| `TestRateLimiter_AllowsUpToFive`          | Exactly 5 requests allowed, 6th rejected                            |
| `TestRateLimiter_IndependentUsers`        | alice's limit does not affect bob                                   |
| `TestRateLimiter_ConcurrentSafety`        | 50 goroutines → exactly 5 accepted (race safety)                    |
| `TestRateLimiter_StatsRejectedCumulative` | Rejected counter accumulates correctly                              |
| `TestRateLimiter_WindowExpiry`            | Old timestamps evicted; fresh requests allowed after window expires |
| `TestHandler_AcceptsValidRequest`         | Valid body → 201                                                    |
| `TestHandler_RejectsMissingUserID`        | Missing `user_id` → 400                                             |
| `TestHandler_RejectsEmptyUserID`          | Empty string `user_id` → 400                                        |
| `TestHandler_RejectsMissingPayload`       | Missing `payload` → 400                                             |
| `TestHandler_RejectsInvalidJSON`          | Malformed JSON → 400                                                |
| `TestHandler_RateLimit`                   | 5 × 201 then 429                                                    |
| `TestHandler_StatsEndpoint`               | `/stats` returns correct `window_accepted` count                    |
| `TestHandler_ConcurrentRequests`          | 20 HTTP goroutines → exactly 5 accepted                             |

### Part 2 — 25 tests

| Test                                                    | What it verifies                                          |
| ------------------------------------------------------- | --------------------------------------------------------- |
| `TestValidator_AcceptsHTTPS`                            | `https://` URL passes                                     |
| `TestValidator_AcceptsHTTP`                             | `http://` URL passes                                      |
| `TestValidator_RejectsFTP`                              | `ftp://` scheme rejected                                  |
| `TestValidator_RejectsNoScheme`                         | Bare hostname rejected                                    |
| `TestValidator_RejectsTooLong`                          | URL > 2048 chars rejected                                 |
| `TestValidator_RejectsEmpty`                            | Empty string URL rejected                                 |
| `TestValidator_RejectsSliceOverLimit`                   | > 20 URLs per array rejected                              |
| `TestStore_DuplicateSKU`                                | Second create with same SKU returns error                 |
| `TestStore_ListDoesNotLoadMediaArrays`                  | List items have counts, not URL slices                    |
| `TestStore_PaginationOffset`                            | `offset=7, limit=3` returns correct page                  |
| `TestStore_OffsetBeyondTotal`                           | Offset past end returns empty slice                       |
| `TestHTTP_CreateProduct_201`                            | Valid create → 201 with assigned ID                       |
| `TestHTTP_CreateProduct_EmptyName`                      | Empty name → 400                                          |
| `TestHTTP_CreateProduct_EmptySKU`                       | Empty SKU → 400                                           |
| `TestHTTP_CreateProduct_DuplicateSKU_409`               | Duplicate SKU → 409                                       |
| `TestHTTP_CreateProduct_InvalidURL`                     | Bad URL in array → 400                                    |
| `TestHTTP_ListProducts_DefaultPagination`               | Default limit/offset; no URL arrays in items              |
| `TestHTTP_ListProducts_Pagination`                      | Custom limit/offset; correct `has_more`                   |
| `TestHTTP_GetProduct_FullMedia`                         | Detail returns full `image_urls` + `video_urls`           |
| `TestHTTP_GetProduct_NotFound`                          | Unknown ID → 404                                          |
| `TestHTTP_AddMedia_AppendsURLs`                         | Media appended; count increases correctly                 |
| `TestHTTP_AddMedia_EmptyBody_400`                       | Empty arrays → 400                                        |
| `TestHTTP_AddMedia_NotFound`                            | Unknown product ID → 404                                  |
| `TestPerformanceInvariant_ListDoesNotSerialiseAllMedia` | 1,000 products × 10 images; list reads 20 items in < 50ms |
| `TestConcurrentCreate`                                  | 55 goroutines (5 dup + 50 unique SKU); exactly 1 dup wins |

```
go test ./...       → ok  part1  ok  part2  (38 tests, all passing)
go test -race ./... → ok  part1  ok  part2  (race detector: no issues)
```
