package headline

import (
	"fmt"
	"io"
	"strings"

	"golang.org/x/net/html"
)

// MZaminClient is a client to fetch headlines from mzamin.com
type MZaminClient struct {
	URL        string
	HTTPClient *CachingHTTPClient
}

// NewMZaminClient creates a new MZaminClient
func NewMZaminClient(url string, client *CachingHTTPClient) *MZaminClient {
	return &MZaminClient{
		URL:        url,
		HTTPClient: client,
	}
}

// SourceInfo returns information about the news source
func (c *MZaminClient) SourceInfo() SourceInfo {
	return SourceInfo{
		Name:     "মানবজমিন",
		Logo:     "https://mzamin.com/assets/images/logo.png",
		Homepage: "https://mzamin.com/",
	}
}

// GetHeadlines fetches the headlines from mzamin.com
func (c *MZaminClient) GetHeadlines() (Response, error) {
	resp, err := c.HTTPClient.Get(c.URL)
	if err != nil {
		return Response{Source: c.SourceInfo()}, fmt.Errorf("failed to fetch the website: %v", err)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Response{Source: c.SourceInfo()}, fmt.Errorf("failed to read the response body: %v", err)
	}

	items, err := c.extractMZaminNewsItems(string(body))
	if err != nil {
		return Response{Source: c.SourceInfo()}, fmt.Errorf("failed to extract news items: %v", err)
	}

	return Response{
		Source:    c.SourceInfo(),
		Headlines: items,
	}, nil
}

func (c *MZaminClient) extractMZaminNewsItems(htmlContent string) ([]NewsItem, error) {
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %v", err)
	}

	var newsItems []NewsItem
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "h1" {
			for _, a := range n.Attr {
				if a.Key == "class" && a.Val == "display-3" {
					title, url := c.extractMZaminTitleAndURL(n, c.URL)
					if title != "" && url != "" {
						newsItems = append(newsItems, NewsItem{Title: title, URL: url})
					}
					break
				}
			}
		} else if n.Type == html.ElementNode && n.Data == "h3" {
			title, url := c.extractMZaminTitleAndURL(n, c.URL)
			if title != "" && url != "" {
				newsItems = append(newsItems, NewsItem{Title: title, URL: url})
			}
		}
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			f(child)
		}
	}
	f(doc)

	return newsItems, nil
}

func (c *MZaminClient) extractMZaminTitleAndURL(n *html.Node, baseURL string) (string, string) {
	var title, url string
	for child := n.FirstChild; child != nil; child = child.NextSibling {
		if child.Type == html.ElementNode && child.Data == "a" {
			for _, attr := range child.Attr {
				if attr.Key == "href" {
					url = attr.Val
					break
				}
			}
			title = extractText(child)
			break
		} else if child.Type == html.ElementNode && child.Data == "span" {
			spanText := extractText(child)
			if spanText != "" {
				title += spanText + " "
			}
		}
	}

	url = completeURL(baseURL, url)

	return strings.TrimSpace(title), url
}

func extractText(n *html.Node) string {
	var text string
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.TextNode {
			text += c.Data
		} else if c.Type == html.ElementNode {
			text += extractText(c)
		}
	}
	return strings.TrimSpace(text)
}
