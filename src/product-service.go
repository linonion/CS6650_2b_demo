package main

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Product matches the Product schema in api.yaml.
type Product struct {
	ProductID    int32  `json:"product_id"`
	SKU          string `json:"sku"`
	Manufacturer string `json:"manufacturer"`
	CategoryID   int32  `json:"category_id"`
	Weight       int32  `json:"weight"`
	SomeOtherID  int32  `json:"some_other_id"`
}

// ProductInput uses pointers to detect missing required fields.
type ProductInput struct {
	ProductID    *int32  `json:"product_id"`
	SKU          *string `json:"sku"`
	Manufacturer *string `json:"manufacturer"`
	CategoryID   *int32  `json:"category_id"`
	Weight       *int32  `json:"weight"`
	SomeOtherID  *int32  `json:"some_other_id"`
}

// APIError matches the Error schema in api.yaml.
type APIError struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// Seed products, similar to the albums in the starter code.
var products = map[int32]Product{
	1: {
		ProductID:    1,
		SKU:          "APPLE-001",
		Manufacturer: "Apple",
		CategoryID:   10,
		Weight:       200,
		SomeOtherID:  100,
	},
	2: {
		ProductID:    2,
		SKU:          "BOOK-001",
		Manufacturer: "Example Publisher",
		CategoryID:   20,
		Weight:       500,
		SomeOtherID:  200,
	},
	3: {
		ProductID:    3,
		SKU:          "MUG-001",
		Manufacturer: "Example Manufacturer",
		CategoryID:   30,
		Weight:       350,
		SomeOtherID:  300,
	},
}

// Protect the shared map from concurrent requests.
var productsMu sync.RWMutex

func main() {
	router := gin.New()
	router.Use(prometheusMiddleware())
	router.Use(gin.LoggerWithFormatter(func(p gin.LogFormatterParams) string {
		entry := struct {
			Timestamp  string  `json:"timestamp"`
			Method     string  `json:"method"`
			Endpoint   string  `json:"endpoint"`
			Status     int     `json:"status"`
			DurationMS float64 `json:"duration_ms"`
		}{
			Timestamp:  p.TimeStamp.Format("2006-01-02T15:04:05Z07:00"),
			Method:     p.Method,
			Endpoint:   p.Path,
			Status:     p.StatusCode,
			DurationMS: float64(p.Latency.Microseconds()) / 1000,
		}

		data, _ := json.Marshal(entry)
		return string(data) + "\n"
	}))

	// Return the specified Error JSON for unexpected server panics.
	router.Use(gin.CustomRecovery(func(c *gin.Context, recovered any) {
		c.AbortWithStatusJSON(http.StatusInternalServerError, APIError{
			Error:   "INTERNAL_SERVER_ERROR",
			Message: "Internal server error",
		})
	}))

	router.GET("/products/:productId", getProduct)
	router.POST("/products/:productId/details", updateProductDetails)
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	if err := router.Run(":8080"); err != nil {
		panic(err)
	}
}

// Parse the product ID from the URL.
func parseProductID(raw string) (int32, error) {
	id, err := strconv.ParseInt(raw, 10, 32)
	if err != nil || id < 1 {
		return 0, errors.New("productId must be a positive 32-bit integer")
	}
	return int32(id), nil
}

// GET /products/:productId
func getProduct(c *gin.Context) {
	id, err := parseProductID(c.Param("productId"))
	if err != nil {
		c.JSON(http.StatusNotFound, APIError{
			Error:   "PRODUCT_NOT_FOUND",
			Message: "Product not found",
			Details: err.Error(),
		})
		return
	}

	productsMu.RLock()
	product, exists := products[id]
	productsMu.RUnlock()

	if !exists {
		c.JSON(http.StatusNotFound, APIError{
			Error:   "PRODUCT_NOT_FOUND",
			Message: "Product not found",
		})
		return
	}

	c.JSON(http.StatusOK, product)
}

// Validate all required fields and constraints from api.yaml.
func validateProduct(input ProductInput, id int32) (Product, error) {
	if input.ProductID == nil ||
		input.SKU == nil ||
		input.Manufacturer == nil ||
		input.CategoryID == nil ||
		input.Weight == nil ||
		input.SomeOtherID == nil {
		return Product{}, errors.New("all six product fields are required")
	}

	if *input.ProductID != id {
		return Product{}, errors.New("product_id must match the URL productId")
	}

	if *input.ProductID < 1 || *input.CategoryID < 1 ||
		*input.SomeOtherID < 1 {
		return Product{}, errors.New("IDs must be positive integers")
	}

	if *input.Weight < 0 {
		return Product{}, errors.New("weight cannot be negative")
	}

	skuLength := utf8.RuneCountInString(*input.SKU)
	if skuLength < 1 || skuLength > 100 {
		return Product{}, errors.New("sku must contain 1 to 100 characters")
	}

	manufacturerLength := utf8.RuneCountInString(*input.Manufacturer)
	if manufacturerLength < 1 || manufacturerLength > 200 {
		return Product{}, errors.New("manufacturer must contain 1 to 200 characters")
	}

	return Product{
		ProductID:    *input.ProductID,
		SKU:          *input.SKU,
		Manufacturer: *input.Manufacturer,
		CategoryID:   *input.CategoryID,
		Weight:       *input.Weight,
		SomeOtherID:  *input.SomeOtherID,
	}, nil
}

// POST /products/:productId/details
// Update an existing product; do not create a new one.
func updateProductDetails(c *gin.Context) {
	id, err := parseProductID(c.Param("productId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, APIError{
			Error:   "INVALID_INPUT",
			Message: "Invalid product ID",
			Details: err.Error(),
		})
		return
	}

	// Require a JSON request body.
	if !strings.HasPrefix(
		strings.ToLower(c.GetHeader("Content-Type")),
		"application/json",
	) {
		c.JSON(http.StatusBadRequest, APIError{
			Error:   "INVALID_INPUT",
			Message: "Content-Type must be application/json",
		})
		return
	}

	var input ProductInput
	decoder := json.NewDecoder(c.Request.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&input); err != nil {
		c.JSON(http.StatusBadRequest, APIError{
			Error:   "INVALID_INPUT",
			Message: "Invalid JSON body",
			Details: err.Error(),
		})
		return
	}

	// Reject additional JSON values after the first object.
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		c.JSON(http.StatusBadRequest, APIError{
			Error:   "INVALID_INPUT",
			Message: "Request body must contain one JSON object",
		})
		return
	}

	product, err := validateProduct(input, id)
	if err != nil {
		c.JSON(http.StatusBadRequest, APIError{
			Error:   "INVALID_INPUT",
			Message: "Invalid product data",
			Details: err.Error(),
		})
		return
	}

	productsMu.Lock()
	if _, exists := products[id]; !exists {
		productsMu.Unlock()

		c.JSON(http.StatusNotFound, APIError{
			Error:   "PRODUCT_NOT_FOUND",
			Message: "Product not found",
		})
		return
	}

	products[id] = product
	productsMu.Unlock()

	// 204 responses must not contain a response body.
	c.Status(http.StatusNoContent)
}
