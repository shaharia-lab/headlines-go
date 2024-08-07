package headline

import (
	"fmt"
	"io"
	"strings"

	"golang.org/x/net/html"
)

type DailyStarBanglaClient struct {
	URL        string
	HTTPClient *CachingHTTPClient
}

func NewDailyStarBanglaClient(url string, client *CachingHTTPClient) *DailyStarBanglaClient {
	return &DailyStarBanglaClient{
		URL:        url,
		HTTPClient: client,
	}
}

func (c *DailyStarBanglaClient) SourceInfo() SourceInfo {
	return SourceInfo{
		Name:     "Daily Star Bangla",
		Logo:     "https://bangla.thedailystar.net/sites/all/themes/sloth/logo-bn.png",
		Homepage: "https://bangla.thedailystar.net/",
	}
}

func (c *DailyStarBanglaClient) GetHeadlines() (Response, error) {
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

	headlines, err := c.extractDailyStarBanglaHeadlines(string(body), c.URL)
	if err != nil {
		return Response{Source: c.SourceInfo()}, err
	}

	return Response{
		Source:    c.SourceInfo(),
		Headlines: headlines,
	}, nil
}

func (c *DailyStarBanglaClient) extractDailyStarBanglaHeadlines(htmlContent, baseURL string) ([]NewsItem, error) {
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %v", err)
	}

	var headlines []NewsItem
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "div" {
			class := getAttr(n, "class")
			if strings.Contains(class, "panel-pane pane-home-top-v7 no-title block") ||
				strings.Contains(class, "panel-pane pane-category-news no-title block") {
				c.extractHeadlinesFromSection(n, &headlines, baseURL)
			}
		}
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			f(child)
		}
	}
	f(doc)

	return headlines, nil
}

func (c *DailyStarBanglaClient) extractHeadlinesFromSection(n *html.Node, headlines *[]NewsItem, baseURL string) {
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "div" && strings.Contains(getAttr(n, "class"), "card-content") {
			var title, url string
			for child := n.FirstChild; child != nil; child = child.NextSibling {
				if child.Type == html.ElementNode && child.Data == "h3" && getAttr(child, "class") == "title" {
					for a := child.FirstChild; a != nil; a = a.NextSibling {
						if a.Type == html.ElementNode && a.Data == "a" {
							title = extractText(a)
							url = getAttr(a, "href")
							break
						}
					}
					if title != "" && url != "" {
						*headlines = append(*headlines, NewsItem{
							Title: title,
							URL:   completeURL(baseURL, url),
						})
					}
					break
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(n)
}

func getAttr(n *html.Node, key string) string {
	for _, attr := range n.Attr {
		if attr.Key == key {
			return attr.Val
		}
	}
	return ""
}
