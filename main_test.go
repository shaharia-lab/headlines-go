package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/shaharia-lab/headlines/headline"
)

// MockNewsClient is a mock implementation of the NewsClient interface for testing
type MockNewsClient struct {
	headlines []headline.NewsItem
}

func (m *MockNewsClient) GetHeadlines() (headline.Response, error) {
	return headline.Response{
		Source:    headline.SourceInfo{Name: "Mock Source", Logo: "http://mock.com/logo.png", Homepage: "http://mock.com"},
		Headlines: m.headlines,
	}, nil
}

func (m *MockNewsClient) Name() string {
	return "MockClient"
}

func (m *MockNewsClient) SourceInfo() headline.SourceInfo {
	return headline.SourceInfo{Name: "Mock Source", Logo: "http://mock.com/logo.png", Homepage: "http://mock.com"}
}

func TestHeadlinesHandler(t *testing.T) {
	// Create mock news clients
	mockClient1 := &MockNewsClient{
		headlines: []headline.NewsItem{{Title: "Test 1", URL: "http://test1.com"}},
	}
	mockClient2 := &MockNewsClient{
		headlines: []headline.NewsItem{{Title: "Test 2", URL: "http://test2.com"}},
	}

	sources := []headline.NewsClient{mockClient1, mockClient2}

	// Create a request to pass to our handler
	req, err := http.NewRequest("GET", "/api/headlines", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()
	handler := headlinesHandler(sources)

	// Call the handler
	handler.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check the response body
	var response []headline.Response
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Could not parse response body: %v", err)
	}

	if len(response) != 2 {
		t.Errorf("Expected 2 sources in response, got %d", len(response))
	}

	if response[0].Headlines[0].Title != "Test 1" || response[1].Headlines[0].Title != "Test 2" {
		t.Error("Unexpected headlines in response")
	}

	// Check the cache header
	if cacheHeader := rr.Header().Get("X-Cache"); cacheHeader != "MISS" {
		t.Errorf("Expected X-Cache header to be MISS, got %s", cacheHeader)
	}
}
