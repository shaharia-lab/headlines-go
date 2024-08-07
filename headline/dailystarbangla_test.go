package headline

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"golang.org/x/net/html"
)

func TestDailyStarBanglaClient_GetHeadlines(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`
			<html>
				<body>
					<div class="panel-pane pane-home-top-v7 no-title block">
						<div class="card-content">
							<h3 class="title">
								<a href="/article1">Headline 1</a>
							</h3>
						</div>
					</div>
					<div class="panel-pane pane-category-news no-title block">
						<div class="card-content">
							<h3 class="title">
								<a href="/article2">Headline 2</a>
							</h3>
						</div>
					</div>
				</body>
			</html>
		`))
	}))
	defer server.Close()

	// Create a DailyStarBanglaClient with the mock server URL
	client := NewDailyStarBanglaClient(server.URL, NewCachingHTTPClient(0, "test-agent"))

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
		{"Headline 1", server.URL + "/article1"},
		{"Headline 2", server.URL + "/article2"},
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
		Name:     "Daily Star Bangla",
		Logo:     "https://bangla.thedailystar.net/sites/all/themes/sloth/logo-bn.png",
		Homepage: "https://bangla.thedailystar.net/",
	}

	if response.Source != expectedSourceInfo {
		t.Errorf("Expected source info %+v, got %+v", expectedSourceInfo, response.Source)
	}
}

func TestDailyStarBanglaClient_SourceInfo(t *testing.T) {
	client := &DailyStarBanglaClient{}
	info := client.SourceInfo()

	expectedInfo := SourceInfo{
		Name:     "Daily Star Bangla",
		Logo:     "https://bangla.thedailystar.net/sites/all/themes/sloth/logo-bn.png",
		Homepage: "https://bangla.thedailystar.net/",
	}

	if info != expectedInfo {
		t.Errorf("Expected source info %+v, got %+v", expectedInfo, info)
	}
}

func TestGetAttr(t *testing.T) {
	node := &html.Node{
		Type: html.ElementNode,
		Data: "div",
		Attr: []html.Attribute{
			{Key: "class", Val: "test-class"},
			{Key: "id", Val: "test-id"},
		},
	}

	tests := []struct {
		key      string
		expected string
	}{
		{"class", "test-class"},
		{"id", "test-id"},
		{"nonexistent", ""},
	}

	for _, test := range tests {
		result := getAttr(node, test.key)
		if result != test.expected {
			t.Errorf("getAttr(%s) = %s; want %s", test.key, result, test.expected)
		}
	}
}
