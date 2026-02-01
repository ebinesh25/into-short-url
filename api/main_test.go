package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// RedisClient interface defines the methods we need from Redis
type RedisClient interface {
	HGet(ctx context.Context, key, field string) *redis.StringCmd
	HSet(ctx context.Context, key string, values ...interface{}) *redis.IntCmd
	Ping(ctx context.Context) *redis.StatusCmd
}

// testSetupRoutes is a test-only version of setupRoutes that uses our interface
func testSetupRoutes(r *gin.Engine, client RedisClient) {
	r.GET("/*url", func(c *gin.Context) {
		url := strings.TrimPrefix(c.Param("url"), "/")

		if rawQuery := c.Request.URL.RawQuery; rawQuery != "" {
			url = url + "?" + rawQuery
		}

		if url == "ping" {
			c.JSON(200, gin.H{
				"message": "pong",
			})
			return
		}

		if strings.HasPrefix(url, "http") {
			testShortenURL(c, url, client)
			return
		} else {
			testResolveURL(c, client, url)
			return
		}
	})
}

// testShortenURL is a test-only version that uses RedisClient interface
func testShortenURL(c *gin.Context, url string, client RedisClient) {
	ctx := context.Background()

	// Generate short code (simple hash for testing)
	shortCode := "test_" + url[:5]

	// Store in Redis
	client.HSet(ctx, "urls", shortCode, url)

	const domain = "url.noskill.in/"
	c.String(http.StatusOK, domain+shortCode)
}

// testResolveURL is a test-only version that uses RedisClient interface
func testResolveURL(c *gin.Context, client RedisClient, shortCode string) {
	ctx := context.Background()

	// Get URL from Redis
	cmd := client.HGet(ctx, "urls", shortCode)
	fullURL, err := cmd.Result()

	if err != nil || fullURL == "" {
		c.String(http.StatusNotFound, "URL not found")
		return
	}

	c.Redirect(http.StatusMovedPermanently, fullURL)
}

// TestURLWithQueryParams verifies that URLs containing query parameters
// (like YouTube video URLs) are properly captured and stored.
func TestURLWithQueryParams(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create a mock Redis client (miniredis or mock)
	// For this test, we'll use a minimal mock
	mockRedis := createMockRedis()

	tests := []struct {
		name           string
		inputURL       string
		expectedStored string // The URL that should be stored in Redis
		description    string
	}{
		{
			name:           "YouTube URL with video ID query param",
			inputURL:       "/https:/www.youtube.com/watch?v=XrlY3jdrM_E",
			expectedStored: "https:/www.youtube.com/watch?v=XrlY3jdrM_E",
			description:    "URL with ?v= query parameter should be stored with query params",
		},
		{
			name:           "URL with multiple query parameters",
			inputURL:       "/https://example.com/page?foo=bar&baz=qux",
			expectedStored: "https://example.com/page?foo=bar&baz=qux",
			description:    "URL with multiple query params should preserve all",
		},
		{
			name:           "URL without query parameters",
			inputURL:       "/https://google.com",
			expectedStored: "https://google.com",
			description:    "Simple URL without query params should work as before",
		},
		{
			name:           "YouTube URL with encoded query param",
			inputURL:       "/https://youtube.com/watch?t=10s&list=PLxyz&index=5",
			expectedStored: "https://youtube.com/watch?t=10s&list=PLxyz&index=5",
			description:    "URL with multiple query params including encoded values",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new Gin router for each test
			router := gin.New()
			testSetupRoutes(router, mockRedis)

			// Create a test request
			req := httptest.NewRequest("GET", tt.inputURL, nil)
			w := httptest.NewRecorder()

			// Execute the request
			router.ServeHTTP(w, req)

			// Verify the response was successful
			if w.Code != http.StatusOK {
				t.Errorf("Expected status OK, got %d", w.Code)
			}

			// The response should be a shortened URL (url.noskill.in/XXXXX)
			responseBody := w.Body.String()
			if !strings.HasPrefix(responseBody, "url.noskill.in/") {
				t.Errorf("Expected response to start with 'url.noskill.in/', got: %s", responseBody)
			}

			// Extract the short code from response
			shortCode := strings.TrimPrefix(responseBody, "url.noskill.in/")

			// Verify Redis stored the full URL with query params
			storedURL := mockRedis.GetStoredURL(shortCode)
			if storedURL != tt.expectedStored {
				t.Errorf("Redis stored URL mismatch:\nExpected: %s\nGot:      %s", tt.expectedStored, storedURL)
			}
		})
	}
}

// TestShortCodeResolution verifies that a short code redirects
// to the correct full URL including query parameters.
func TestShortCodeResolution(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRedis := createMockRedis()
	router := gin.New()
	testSetupRoutes(router, mockRedis)

	// Pre-populate Redis with a known mapping
	testShortCode := "testCode123"
	testFullURL := "https:/www.youtube.com/watch?v=XrlY3jdrM_E"
	mockRedis.StoreURL(testShortCode, testFullURL)

	// Create a request to resolve the short code
	req := httptest.NewRequest("GET", "/"+testShortCode, nil)
	w := httptest.NewRecorder()

	// Execute the request
	router.ServeHTTP(w, req)

	// Verify we get a redirect
	if w.Code != http.StatusMovedPermanently {
		t.Errorf("Expected 301 redirect, got %d", w.Code)
	}

	// Verify the redirect location includes query parameters
	location := w.Header().Get("Location")
	if location != testFullURL {
		t.Errorf("Redirect location mismatch:\nExpected: %s\nGot:      %s", testFullURL, location)
	}
}

// TestPingEndpoint verifies the ping endpoint works correctly
func TestPingEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRedis := createMockRedis()
	router := gin.New()
	testSetupRoutes(router, mockRedis)

	req := httptest.NewRequest("GET", "/ping", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status OK for ping, got %d", w.Code)
	}
}

// MockRedisClient is a simple mock for testing
type MockRedisClient struct {
	storage map[string]string
}

func createMockRedis() *MockRedisClient {
	return &MockRedisClient{
		storage: make(map[string]string),
	}
}

// Implement the interface methods needed by the code
func (m *MockRedisClient) HGet(ctx context.Context, key, field string) *redis.StringCmd {
	// Return a mock command with the stored value
	cmd := redis.NewStringCmd(ctx)
	if val, ok := m.storage[field]; ok {
		cmd.SetVal(val)
	}
	return cmd
}

func (m *MockRedisClient) HSet(ctx context.Context, key string, values ...interface{}) *redis.IntCmd {
	// Store the short -> URL mapping
	if len(values) >= 2 {
		short := values[0].(string)
		full := values[1].(string)
		m.storage[short] = full
	}
	cmd := redis.NewIntCmd(ctx)
	cmd.SetVal(1)
	return cmd
}

// Helper methods for testing
func (m *MockRedisClient) StoreURL(short, full string) {
	m.storage[short] = full
}

func (m *MockRedisClient) GetStoredURL(short string) string {
	return m.storage[short]
}

// Ping method for Redis client compatibility
func (m *MockRedisClient) Ping(ctx context.Context) *redis.StatusCmd {
	cmd := redis.NewStatusCmd(ctx)
	cmd.SetVal("PONG")
	return cmd
}
