package headline

import (
	"fmt"
	"io"
	"strings"

	"golang.org/x/net/html"
)

// ProthomAloClient is a client to fetch headlines from prothomalo.com
type ProthomAloClient struct {
	URL        string
	HTTPClient *CachingHTTPClient
}

// NewProthomAloClient creates a new ProthomAloClient
func NewProthomAloClient(url string, client *CachingHTTPClient) *ProthomAloClient {
	return &ProthomAloClient{
		URL:        url,
		HTTPClient: client,
	}
}

// SourceInfo returns information about the news source
func (c *ProthomAloClient) SourceInfo() SourceInfo {
	return SourceInfo{
		Name:     "ProthomAlo",
		Logo:     "https://encrypted-tbn0.gstatic.com/images?q=tbn:ANd9GcSUTX3amtUek4Ia80_rbqUkfwS6sYaeSUdqwg&s",
		Homepage: "https://www.prothomalo.com",
	}
}

// GetHeadlines fetches the headlines from prothomalo.com
func (c *ProthomAloClient) GetHeadlines() (Response, error) {
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

	items, err := c.extractNewsItems(string(body))

	return Response{
		Source:    c.SourceInfo(),
		Headlines: items,
	}, nil
}

func (c *ProthomAloClient) extractNewsItems(htmlContent string) ([]NewsItem, error) {
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %v", err)
	}

	var newsItems []NewsItem
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "h3" {
			for _, a := range n.Attr {
				if a.Key == "class" && strings.Contains(a.Val, "headline-title") {
					title, url := c.extractTitleAndURL(n)
					if title != "" && url != "" {
						newsItems = append(newsItems, NewsItem{Title: title, URL: url})
					}
					break
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	return newsItems, nil
}

func (c *ProthomAloClient) extractTitleAndURL(n *html.Node) (string, string) {
	var title, url string
	for child := n.FirstChild; child != nil; child = child.NextSibling {
		if child.Type == html.ElementNode && child.Data == "a" {
			for _, attr := range child.Attr {
				if attr.Key == "href" {
					url = attr.Val
					break
				}
			}
			for cc := child.FirstChild; cc != nil; cc = cc.NextSibling {
				if cc.Type == html.ElementNode && cc.Data == "span" {
					title = extractText(cc)
					break
				}
			}
			break
		}
	}
	return title, completeURL(c.URL, url)
}
