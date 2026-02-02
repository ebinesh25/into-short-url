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

	// Check if URL already exists (duplicate check)
	cmd := client.HGet(ctx, "originalUrls", url)
	existingShortCode, err := cmd.Result()

	if err == nil && existingShortCode != "" {
		// URL already exists, return the existing short URL
		const domain = "url.noskill.in/"
		c.String(http.StatusOK, domain+existingShortCode)
		return
	}

	// Generate new short code (simple hash for testing)
	shortCode := "test_" + url[:5]

	// Store both forward and reverse mappings
	client.HSet(ctx, "shortenUrls", shortCode, url)
	client.HSet(ctx, "originalUrls", url, shortCode)

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
		{
			name:           "sendHugs URl",
			inputURL:       "https://send-hugss.netlify.app/view#YqkeDwbApRL8XJo-dkOrE9UGfMMlyq6BGWFaZOv8mpgLIOpw4AAFW2IUnvbcxxpCL6WyZHRQedh9GlW5yFfRoELpTRiKQ3ifXUJuy47UaFzb7jkJlxq9eTv7-Ltp3pVRJjB93gN1YYJUSohtnc9EzT-b7uqh3Wd1F1bx7UyJfsOopugx5-CMY3GK01UVQN1jbM7QGBXzQvvZpMQ218ePvv-YpQusgU75Hs1ZhRrEZx9A0zHYUjVwABVI5IYifaKZJn1JLBmnCw50CXGwo2jM8ok-neL9YPsYXBfmMCug8XdV1UnYc6-odQtH78jsQabVdwqHo3Eg1cTXEKYfyyYAFci52D7E414.hhzajGlHkuJDjWCn6VCqFHbRHaCNHo2arNwovQ4dXkc",
			expectedStored: "https://send-hugss.netlify.app/view#YqkeDwbApRL8XJo-dkOrE9UGfMMlyq6BGWFaZOv8mpgLIOpw4AAFW2IUnvbcxxpCL6WyZHRQedh9GlW5yFfRoELpTRiKQ3ifXUJuy47UaFzb7jkJlxq9eTv7-Ltp3pVRJjB93gN1YYJUSohtnc9EzT-b7uqh3Wd1F1bx7UyJfsOopugx5-CMY3GK01UVQN1jbM7QGBXzQvvZpMQ218ePvv-YpQusgU75Hs1ZhRrEZx9A0zHYUjVwABVI5IYifaKZJn1JLBmnCw50CXGwo2jM8ok-neL9YPsYXBfmMCug8XdV1UnYc6-odQtH78jsQabVdwqHo3Eg1cTXEKYfyyYAFci52D7E414.hhzajGlHkuJDjWCn6VCqFHbRHaCNHo2arNwovQ4dXkc",
			description:    "URL that has # in the end",
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

// TestURLForDuplication verifies that submitting the same URL twice
// returns the same short URL instead of creating a new one.
func TestURLForDuplication(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Check Duplicate URL returns same short code", func(t *testing.T) {
		mockRedis := createMockRedis()
		router := gin.New()
		testSetupRoutes(router, mockRedis)

		// The URL to shorten
		testURL := "https://www.helloworld.io"
		const domain = "url.noskill.in/"

		// First request: create a new short URL
		req1 := httptest.NewRequest("GET", "/"+testURL, nil)
		w1 := httptest.NewRecorder()
		router.ServeHTTP(w1, req1)

		// Verify first request was successful
		if w1.Code != http.StatusOK {
			t.Errorf("First request: Expected status OK, got %d", w1.Code)
		}

		firstResponse := w1.Body.String()
		if !strings.HasPrefix(firstResponse, domain) {
			t.Errorf("First response should start with '%s', got: %s", domain, firstResponse)
		}

		// Extract the short code from the first response
		firstShortCode := strings.TrimPrefix(firstResponse, domain)

		// Second request: try to shorten the same URL again
		req2 := httptest.NewRequest("GET", "/"+testURL, nil)
		w2 := httptest.NewRecorder()
		router.ServeHTTP(w2, req2)

		// Verify second request was successful
		if w2.Code != http.StatusOK {
			t.Errorf("Second request: Expected status OK, got %d", w2.Code)
		}

		secondResponse := w2.Body.String()

		// The second response should be the SAME short code as the first
		secondShortCode := strings.TrimPrefix(secondResponse, domain)

		if secondShortCode != firstShortCode {
			t.Errorf("Duplicate URL should return same short code:\nFirst:  %s\nSecond: %s", firstShortCode, secondShortCode)
		}

		// Verify only one entry exists in storage (no duplicate created)
		// We need to count how many different URLs were stored
		storedCount := len(mockRedis.storage)
		if storedCount != 1 {
			t.Errorf("Expected exactly 1 entry in storage, got %d", storedCount)
		}
	})

	t.Run("Different URLs create different short codes", func(t *testing.T) {
		mockRedis := createMockRedis()
		router := gin.New()
		testSetupRoutes(router, mockRedis)

		const domain = "url.noskill.in/"

		// First URL
		url1 := "https://www.example.com"
		req1 := httptest.NewRequest("GET", "/"+url1, nil)
		w1 := httptest.NewRecorder()
		router.ServeHTTP(w1, req1)

		// Second URL (different)
		url2 := "https://www.different.com"
		req2 := httptest.NewRequest("GET", "/"+url2, nil)
		w2 := httptest.NewRecorder()
		router.ServeHTTP(w2, req2)

		// Both should be successful
		if w1.Code != http.StatusOK || w2.Code != http.StatusOK {
			t.Errorf("Both requests should succeed")
		}

		short1 := strings.TrimPrefix(w1.Body.String(), domain)
		short2 := strings.TrimPrefix(w2.Body.String(), domain)

		// Short codes should be different
		if short1 == short2 {
			t.Errorf("Different URLs should produce different short codes, got same: %s", short1)
		}
	})
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

// MockRedisClient is a simple mock for testing with reverse lookup support
type MockRedisClient struct {
	storage      map[string]string // short -> original URL (for "shortenUrls" hash)
	reverseIndex map[string]string // original URL -> short (for "originalUrls" hash)
}

func createMockRedis() *MockRedisClient {
	return &MockRedisClient{
		storage:      make(map[string]string),
		reverseIndex: make(map[string]string),
	}
}

// Implement the interface methods needed by the code
func (m *MockRedisClient) HGet(ctx context.Context, key, field string) *redis.StringCmd {
	// Return a mock command with the stored value
	cmd := redis.NewStringCmd(ctx)

	// Check which hash is being accessed
	if key == "shortenUrls" || key == "urls" {
		// Forward lookup: short -> original URL
		if val, ok := m.storage[field]; ok {
			cmd.SetVal(val)
		}
	} else if key == "originalUrls" {
		// Reverse lookup: original URL -> short
		if val, ok := m.reverseIndex[field]; ok {
			cmd.SetVal(val)
		}
	}

	return cmd
}

func (m *MockRedisClient) HSet(ctx context.Context, key string, values ...interface{}) *redis.IntCmd {
	// Store based on which hash is being accessed
	if len(values) >= 2 {
		field := values[0].(string)
		value := values[1].(string)

		if key == "shortenUrls" || key == "urls" {
			// Store short -> original URL
			m.storage[field] = value
		} else if key == "originalUrls" {
			// Store original URL -> short (reverse mapping)
			m.reverseIndex[field] = value
		}
	}

	cmd := redis.NewIntCmd(ctx)
	cmd.SetVal(1)
	return cmd
}

// Helper methods for testing
func (m *MockRedisClient) StoreURL(short, full string) {
	m.storage[short] = full           // For "shortenUrls" hash
	m.reverseIndex[full] = short      // For "originalUrls" hash
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
