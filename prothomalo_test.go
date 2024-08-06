package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestProthomAloClient_GetHeadlines(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`
			<html>
				<body>
					<h3 class="headline-title">
						<a href="/news/test-headline">
							<span>Test Headline</span>
						</a>
					</h3>
					<h3 class="headline-title">
						<a href="/news/another-headline">
							<span>Another Headline</span>
						</a>
					</h3>
				</body>
			</html>
		`))
	}))
	defer server.Close()

	// Create a ProthomAloClient with the mock server URL
	client := NewProthomAloClient(server.URL, NewCachingHTTPClient(0, "test-agent"))

	// Get headlines
	response, err := client.GetHeadlines()
	if err != nil {
		t.Fatalf("Error getting headlines: %v", err)
	}

	// Check the response
	if len(response.Headlines) != 2 {
		t.Errorf("Expected 2 headlines, got %d", len(response.Headlines))
	}

	expectedHeadlines := []struct {
		title string
		url   string
	}{
		{"Test Headline", server.URL + "/news/test-headline"},
		{"Another Headline", server.URL + "/news/another-headline"},
	}

	for i, expected := range expectedHeadlines {
		if i >= len(response.Headlines) {
			t.Errorf("Missing expected headline at index %d", i)
			continue
		}

		if response.Headlines[i].Title != expected.title {
			t.Errorf("Expected headline title '%s', got '%s'", expected.title, response.Headlines[i].Title)
		}

		if response.Headlines[i].URL != expected.url {
			t.Errorf("Expected headline URL '%s', got '%s'", expected.url, response.Headlines[i].URL)
		}
	}

	// Check source info
	if response.Source.Name != "ProthomAlo" {
		t.Errorf("Expected source name 'ProthomAlo', got '%s'", response.Source.Name)
	}
}
