# HW5 — Product API

## Part II: Product API (Go + Gin)

The API is implemented in `src/product-service.go`, using Gin and an in-memory `map[int32]Product` protected by `sync.RWMutex`. The server starts with products 1, 2, and 3. POST updates an existing product's details and does not create a new product. Changes are stored in memory only and are lost when the server restarts.

### Run locally

```bash
cd src
go run product-service.go
```

The server listens on port `8080`.

### Run with Docker

From `src/`:

```bash
docker build -t hw05-product-api .
docker run --rm -p 8080:8080 --name hw05-product hw05-product-api
```

In another terminal, send requests to `http://localhost:8080`.

### API examples and response codes

**GET existing product — 200 OK (verified locally)**

```bash
curl -i http://localhost:8080/products/1
```

Example response body (initial seed data):

```json
{"product_id":1,"sku":"APPLE-001","manufacturer":"Apple","category_id":10,"weight":200,"some_other_id":100}
```

**GET nonexistent product — 404 Not Found (verified locally)**

```bash
curl -i http://localhost:8080/products/999
```

```json
{"error":"PRODUCT_NOT_FOUND","message":"Product not found"}
```

**POST update existing product — 204 No Content (verified locally)**

```bash
curl -i -X POST http://localhost:8080/products/1/details \
  -H 'Content-Type: application/json' \
  -d '{"product_id":1,"sku":"APPLE-UPDATED","manufacturer":"Apple","category_id":10,"weight":300,"some_other_id":100}'
```

A successful POST returns `204` with **no response body**. Verify the update:

```bash
curl -i http://localhost:8080/products/1
```

**POST nonexistent product — 404 Not Found (verified locally)**

```bash
curl -i -X POST http://localhost:8080/products/999/details \
  -H 'Content-Type: application/json' \
  -d '{"product_id":999,"sku":"TEST-001","manufacturer":"Test","category_id":10,"weight":100,"some_other_id":100}'
```

```json
{"error":"PRODUCT_NOT_FOUND","message":"Product not found"}
```

**POST missing required field — 400 Bad Request (verified locally)**

```bash
curl -i -X POST http://localhost:8080/products/1/details \
  -H 'Content-Type: application/json' \
  -d '{"product_id":1,"sku":"TEST-001","manufacturer":"Test","category_id":10,"some_other_id":100}'
```

```json
{"error":"INVALID_INPUT","message":"Invalid product data","details":"all six product fields are required"}
```

**POST negative weight — 400 Bad Request (verified locally)**

```bash
curl -i -X POST http://localhost:8080/products/1/details \
  -H 'Content-Type: application/json' \
  -d '{"product_id":1,"sku":"TEST-001","manufacturer":"Test","category_id":10,"weight":-5,"some_other_id":100}'
```

**POST mismatched URL/body IDs — 400 Bad Request (verified locally)**

```bash
curl -i -X POST http://localhost:8080/products/1/details \
  -H 'Content-Type: application/json' \
  -d '{"product_id":2,"sku":"TEST-001","manufacturer":"Test","category_id":10,"weight":100,"some_other_id":100}'
```

```json
{"error":"INVALID_INPUT","message":"Invalid product data","details":"product_id must match the URL productId"}
```

**Unexpected server failure — 500 Internal Server Error (implemented; NOT tested)**

The API specification includes a `500` response for both GET and POST. The server uses Gin's custom recovery middleware to return a structured JSON error if a handler panics. Example of the **intended** response (not an observed test result):

```json
{"error":"INTERNAL_SERVER_ERROR","message":"Internal server error"}
```

The current in-memory implementation has no database or external service dependency that can be intentionally made unavailable, and no production endpoint was added solely to force a `500`. Therefore, the `500` response has **not** been directly verified. This remains a testing limitation.

### API design notes

- Required fields: `product_id`, `sku`, `manufacturer`, `category_id`, `weight`, `some_other_id`.
- `product_id`, `category_id`, and `some_other_id` must be positive 32-bit integers; `weight` must be a nonnegative 32-bit integer.
- `sku` length must be 1–100 characters; `manufacturer` length must be 1–200 characters.
- URL `productId` must match the JSON `product_id` in POST requests.
- `204` returns no body; error responses use the `Error` schema (`error`, `message`, optional `details`).

## Part III: AWS ECS deployment

_To be documented after the Terraform deployment is verified._

## Part IV: Locust load testing

_To be documented after the load and stress tests are run; include screenshots and HttpUser vs FastHttpUser comparison._

## Part V: Observability

_To be documented after Prometheus/Grafana and structured request logging are implemented; include dashboard screenshots and a log excerpt._
