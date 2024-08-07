package headline

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

var (
	headlinesCache sync.Map
	cacheDuration  = 1 * time.Minute
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

func GetHeadlines(sources []NewsClient) []Response {
	var wg sync.WaitGroup
	results := make([]Response, len(sources))

	for i, source := range sources {
		wg.Add(1)
		go func(index int, s NewsClient) {
			defer wg.Done()
			items, err := s.GetHeadlines()
			if err != nil {
				log.Printf("Error fetching headlines from %s: %v", s.Name(), err)
				results[index] = Response{Source: s.SourceInfo(), Headlines: nil}
			} else {
				results[index] = items
			}
		}(i, source)
	}

	wg.Wait()
	return results
}

func GetCachedHeadlines() ([]Response, bool) {
	if cachedResp, ok := headlinesCache.Load("headlines"); ok {
		cached := cachedResp.(CachedResponse)
		if time.Since(cached.Timestamp) < cacheDuration {
			return cached.Body, true
		}
	}
	return nil, false
}

func CacheHeadlines(headlines []Response) {
	headlinesCache.Store("headlines", CachedResponse{
		Body:      headlines,
		Timestamp: time.Now(),
	})
}
