package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"golang.org/x/net/html"
)

var wg sync.WaitGroup
var mu sync.Mutex

// store visited links
var visited = make(map[string]bool)

// store discovered links
var jobs = make(chan string, 100)

type Crawler struct {
	StartingUrl string
	baseUrl     string
}

func NewCrawler(startingUrl string) (*Crawler, error) {
	sUrl, err := url.Parse(startingUrl)
	if err != nil {
		return &Crawler{StartingUrl: startingUrl}, err
	}

	baseUrl := sUrl.Scheme + "://" + sUrl.Host
	return &Crawler{
		StartingUrl: startingUrl,
		baseUrl:     baseUrl,
	}, nil
}

func (c *Crawler) Crawl(url string) error {
	defer wg.Done()

	// check if link is visited
	mu.Lock()
	if visited[url] {
		mu.Unlock()
		return nil
	}
	visited[url] = true
	mu.Unlock()
	// visit the url with the http package
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println(err)
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	links, err := c.parseLinks(strings.NewReader(string(body)))
	if err != nil {
		return err
	}

	// add the links to the channel
	for _, link := range links {
		jobs <- link
	}

	return nil
}

func (c *Crawler) parseLinks(r io.Reader) ([]string, error) {
	links := []string{}
	doc, err := html.Parse(r)
	if err != nil {
		fmt.Println(err)
		return links, err
	}

	for n := range doc.Descendants() {
		// fmt.Printf("%v: %s\n", n.Type, n.Data)
		if n.Type == html.ElementNode && n.Data == "a" {

			for _, attr := range n.Attr {
				if attr.Key == "href" {
					ulink := attr.Val

					if !isLinkValid(ulink) {
						continue
					}

					link := addBaseUrlToLink(ulink, c.baseUrl)
					links = append(links, link)
				}
			}

		}
	}

	return links, nil
}

func isLinkValid(link string) bool {
	return !strings.HasPrefix(link, "#")
}

func addBaseUrlToLink(link string, baseUrl string) string {
	if strings.HasPrefix(link, "/") {
		return baseUrl + link
	}
	return link
}
