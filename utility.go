package main

import (
	"log"
	"sync"
	"time"

	"github.com/shaharia-lab/headlines-go/headline"
)

var (
	headlinesCache sync.Map
	cacheDuration  = 1 * time.Minute
)

func getHeadlines(sources []headline.NewsClient) []headline.Response {
	var wg sync.WaitGroup
	results := make([]headline.Response, len(sources))

	for i, source := range sources {
		wg.Add(1)
		go func(index int, s headline.NewsClient) {
			defer wg.Done()
			items, err := s.GetHeadlines()
			if err != nil {
				log.Printf("Error fetching headlines from %s: %v", s.Name(), err)
				results[index] = headline.Response{Source: s.SourceInfo(), Headlines: nil}
			} else {
				results[index] = items
			}
		}(i, source)
	}

	wg.Wait()
	return results
}

func getCachedHeadlines() ([]headline.Response, bool) {
	if cachedResp, ok := headlinesCache.Load("headlines"); ok {
		cached := cachedResp.(headline.CachedResponse)
		if time.Since(cached.Timestamp) < cacheDuration {
			return cached.Body, true
		}
	}
	return nil, false
}

func cacheHeadlines(headlines []headline.Response) {
	headlinesCache.Store("headlines", headline.CachedResponse{
		Body:      headlines,
		Timestamp: time.Now(),
	})
}
