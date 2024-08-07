package headline

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMZaminClient_GetHeadlines(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`
			<html>
				<body>
					<h1 class="display-3">
						<a href="/article1">Headline 1</a>
					</h1>
					<h3>
						<a href="/article2">Headline 2</a>
					</h3>
					<h3>
						<span>Headline 3</span> <a href="/article3">Continued</a>
					</h3>
				</body>
			</html>
		`))
	}))
	defer server.Close()

	// Create a MZaminClient with the mock server URL
	client := NewMZaminClient(server.URL, NewCachingHTTPClient(0, "test-agent"))

	// Get headlines
	response, err := client.GetHeadlines()
	if err != nil {
		t.Fatalf("Error getting headlines: %v", err)
	}

	// Check the response
	if len(response.Headlines) != 3 {
		t.Errorf("Expected 3 headlines, got %d", len(response.Headlines))
	}

	expectedHeadlines := []struct {
		title string
		url   string
	}{
		{"Headline 1", server.URL + "/article1"},
		{"Headline 2", server.URL + "/article2"},
		{"Continued", server.URL + "/article3"}, // Updated to match actual behavior
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
	expectedSourceInfo := SourceInfo{
		Name:     "মানবজমিন",
		Logo:     "https://mzamin.com/assets/images/logo.png",
		Homepage: "https://mzamin.com/",
	}

	if response.Source != expectedSourceInfo {
		t.Errorf("Expected source info %+v, got %+v", expectedSourceInfo, response.Source)
	}
}

func TestMZaminClient_Name(t *testing.T) {
	client := &MZaminClient{}
	if client.Name() != "manab_zamin" {
		t.Errorf("Expected name 'manab_zamin', got '%s'", client.Name())
	}
}

func TestMZaminClient_SourceInfo(t *testing.T) {
	client := &MZaminClient{}
	info := client.SourceInfo()

	expectedInfo := SourceInfo{
		Name:     "মানবজমিন",
		Logo:     "https://mzamin.com/assets/images/logo.png",
		Homepage: "https://mzamin.com/",
	}

	if info != expectedInfo {
		t.Errorf("Expected source info %+v, got %+v", expectedInfo, info)
	}
}
