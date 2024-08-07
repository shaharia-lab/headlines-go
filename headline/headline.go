package headline

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

type NewsClient interface {
	GetHeadlines() (Response, error)
	Name() string
	SourceInfo() SourceInfo
}

type NewsItem struct {
	Title string `json:"title"`
	URL   string `json:"url"`
}

type SourceInfo struct {
	Name     string `json:"name"`
	Logo     string `json:"logo"`
	Homepage string `json:"homepage"`
}

type Response struct {
	Source    SourceInfo `json:"source"`
	Headlines []NewsItem `json:"headlines"`
}

type CachedResponse struct {
	Body      []Response
	Timestamp time.Time
}

type CachingHTTPClient struct {
	client    *http.Client
	cache     sync.Map
	userAgent string
}

func NewCachingHTTPClient(timeout time.Duration, userAgent string) *CachingHTTPClient {
	return &CachingHTTPClient{
		client: &http.Client{
			Timeout: timeout,
		},
		userAgent: userAgent,
	}
}

func (c *CachingHTTPClient) Get(url string) (*http.Response, error) {
	if cachedBody, ok := c.cache.Load(url); ok {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(cachedBody.(string))),
		}, nil
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", c.userAgent)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		resp.Body.Close()
		return nil, err
	}
	resp.Body.Close()

	c.cache.Store(url, string(body))

	return &http.Response{
		StatusCode: resp.StatusCode,
		Body:       io.NopCloser(bytes.NewReader(body)),
	}, nil
}

func completeURL(baseURL, relativeURL string) string {
	if relativeURL == "" {
		return ""
	}
	if strings.HasPrefix(relativeURL, "http") {
		return relativeURL
	}
	baseURL = strings.TrimRight(baseURL, "/")
	if strings.HasPrefix(relativeURL, "/") {
		return baseURL + relativeURL
	}
	return baseURL + "/" + relativeURL
}
