package main

import (
	"reflect"
	"testing"
	"time"
)

func TestGetAndCacheHeadlines(t *testing.T) {
	// Clear the cache before testing
	headlinesCache.Range(func(key, value interface{}) bool {
		headlinesCache.Delete(key)
		return true
	})

	testHeadlines := []Response{
		{
			Source: SourceInfo{Name: "Test Source", Logo: "http://example.com/logo.png", Homepage: "http://example.com"},
			Headlines: []NewsItem{
				{Title: "Test Headline 1", URL: "http://example.com/1"},
				{Title: "Test Headline 2", URL: "http://example.com/2"},
			},
		},
	}

	// Test caching
	cacheHeadlines(testHeadlines)

	// Test retrieval
	cachedHeadlines, isCached := getCachedHeadlines()
	if !isCached {
		t.Error("Expected headlines to be cached")
	}

	if !reflect.DeepEqual(cachedHeadlines, testHeadlines) {
		t.Error("Cached headlines do not match original headlines")
	}

	// Test cache expiration
	cacheDuration = 1 * time.Millisecond
	time.Sleep(2 * time.Millisecond)

	_, isCached = getCachedHeadlines()
	if isCached {
		t.Error("Expected cache to be expired")
	}
}

func TestGetHeadlines(t *testing.T) {
	mockClient1 := &MockNewsClient{
		headlines: []NewsItem{{Title: "Test 1", URL: "http://test1.com"}},
	}
	mockClient2 := &MockNewsClient{
		headlines: []NewsItem{{Title: "Test 2", URL: "http://test2.com"}},
	}

	sources := []NewsClient{mockClient1, mockClient2}

	results := getHeadlines(sources)

	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}

	if results[0].Headlines[0].Title != "Test 1" || results[1].Headlines[0].Title != "Test 2" {
		t.Error("Unexpected headlines in results")
	}
}

func TestCompleteURL(t *testing.T) {
	testCases := []struct {
		baseURL     string
		relativeURL string
		expectedURL string
	}{
		{"http://example.com", "/path", "http://example.com/path"},
		{"http://example.com/", "path", "http://example.com/path"},
		{"http://example.com", "http://other.com", "http://other.com"},
		{"http://example.com", "", ""},
	}

	for _, tc := range testCases {
		result := completeURL(tc.baseURL, tc.relativeURL)
		if result != tc.expectedURL {
			t.Errorf("completeURL(%s, %s) = %s; want %s", tc.baseURL, tc.relativeURL, result, tc.expectedURL)
		}
	}
}

// MockNewsClient is a mock implementation of the NewsClient interface for testing
type MockNewsClient struct {
	headlines []NewsItem
}

func (m *MockNewsClient) GetHeadlines() (Response, error) {
	return Response{
		Source:    SourceInfo{Name: "Mock Source", Logo: "http://mock.com/logo.png", Homepage: "http://mock.com"},
		Headlines: m.headlines,
	}, nil
}

func (m *MockNewsClient) Name() string {
	return "MockClient"
}

func (m *MockNewsClient) SourceInfo() SourceInfo {
	return SourceInfo{Name: "Mock Source", Logo: "http://mock.com/logo.png", Homepage: "http://mock.com"}
}
