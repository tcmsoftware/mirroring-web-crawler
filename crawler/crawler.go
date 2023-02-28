// Copyright (c) 2023 Tiago Melo. All rights reserved.
// Use of this source code is governed by the MIT License that can be found in
// the LICENSE file.
package crawler

import (
	"io/fs"
	"log"
	"net/http"
	goUrl "net/url"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/pkg/errors"
)

// For ease of unit testing, so
// we can inject everything we need to.
var (
	get = func(httpClient *http.Client, url string) (*http.Response, error) {
		return httpClient.Get(url)
	}
	goqueryNewDocumentFromReader = goquery.NewDocumentFromReader
	parseUrl                     = goUrl.Parse
	osStat                       = os.Stat
	osMkdirAll                   = os.MkdirAll
	osCreate                     = os.Create
	getDocHtml                   = func(doc *goquery.Document) (string, error) {
		return doc.Html()
	}
	writeStringToFile = func(f *os.File, data string) (int, error) {
		return f.WriteString(data)
	}
	getUrl = func(httpClient *http.Client, url string) (*http.Response, error) {
		resp, err := get(httpClient, url)
		if err != nil {
			return nil, errors.Wrapf(err, "making get request to %v", url)
		}
		return resp, nil
	}
	parseResponse = func(url string, resp *http.Response) (*goquery.Document, error) {
		defer resp.Body.Close()
		doc, err := goqueryNewDocumentFromReader(resp.Body)
		if err != nil {
			return nil, errors.Wrapf(err, "parsing response from url %v", url)
		}
		return doc, nil
	}
	getPagePath = func(destDir, url string) (string, error) {
		parsedUrl, err := parseUrl(url)
		if err != nil {
			return "", errors.Wrapf(err, "parsing url %v", url)
		}
		if strings.HasSuffix(parsedUrl.Path, "/") {
			return path.Join(destDir, parsedUrl.Host, parsedUrl.Path, "index.html"), nil
		} else {
			return path.Join(destDir, parsedUrl.Host, parsedUrl.Path), nil
		}
	}
	saveToDisk = func(url string, pagePath string, doc *goquery.Document) error {
		file, err := osCreate(pagePath)
		if err != nil {
			return errors.Wrapf(err, "creating file for %s", url)
		}
		defer file.Close()
		html, err := getDocHtml(doc)
		if err != nil {
			return errors.Wrapf(err, "converting %s to HTML", url)
		}
		_, err = writeStringToFile(file, html)
		if err != nil {
			return errors.Wrapf(err, "writing HTML file for %s", url)
		}
		return nil
	}
	savePage = func(destDir, url string, doc *goquery.Document, log *log.Logger) error {
		pagePath, err := getPagePath(destDir, url)
		if err != nil {
			return err
		}
		if _, err := osStat(pagePath); err == nil {
			log.Printf("%s already exists, skipping\n", url)
			return nil
		}
		err = osMkdirAll(path.Dir(pagePath), fs.ModePerm)
		if err != nil {
			return errors.Wrapf(err, "creating directory for %s", url)
		}
		if err := saveToDisk(url, pagePath, doc); err != nil {
			return err
		}
		return nil
	}
	processUrl = func(httpClient *http.Client, destDir, url string, log *log.Logger) (*goquery.Document, error) {
		response, err := getUrl(httpClient, url)
		if err != nil {
			return nil, err
		}
		doc, err := parseResponse(url, response)
		if err != nil {
			return nil, err
		}
		err = savePage(destDir, url, doc, log)
		if err != nil {
			return nil, err
		}
		return doc, nil
	}
	fromSameDomain = func(startUrl, link string) bool {
		return strings.HasPrefix(link, "/") || strings.HasPrefix(link, startUrl)
	}
	getAbsoluteUrl = func(startUrl, link string) string {
		if strings.HasPrefix(link, "/") {
			return startUrl + link
		}
		return link
	}
	getNextUrls = func(c *Crawler, doc *goquery.Document) []string {
		var nextUrls []string
		doc.Find("a[href]").Each(func(i int, s *goquery.Selection) {
			href, exists := s.Attr("href")
			if exists && fromSameDomain(c.startUrl, href) {
				childUrl := getAbsoluteUrl(c.startUrl, href)
				c.mu.Lock()
				if !c.visitedUrls[childUrl] {
					c.visitedUrls[childUrl] = true
					nextUrls = append(nextUrls, childUrl)
				}
				c.mu.Unlock()
			}
		})
		return nextUrls
	}
)

// Crawler is a recursive web crawler.
type Crawler struct {
	log         *log.Logger
	startUrl    string
	destDir     string
	httpClient  *http.Client
	visitedUrls map[string]bool
	mu          sync.Mutex
}

// New creates a new Crawler.
func New(startUrl string, destDir string, timeout time.Duration, log *log.Logger) *Crawler {
	return &Crawler{
		startUrl: startUrl,
		destDir:  destDir,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		log:         log,
		visitedUrls: make(map[string]bool),
	}
}

// Run runs the crawler.
// All urls are fetched concurrently.
func (c *Crawler) Run() error {
	c.visitedUrls[c.startUrl] = true
	urls := []string{c.startUrl}
	for len(urls) > 0 {
		var wg sync.WaitGroup
		nextUrls := []string{}
		for _, url := range urls {
			wg.Add(1)
			go func(url string) {
				defer wg.Done()
				doc, err := processUrl(c.httpClient, c.destDir, url, c.log)
				if err != nil {
					c.log.Println(err)
					return
				}
				nextUrls = append(nextUrls, getNextUrls(c, doc)...)
			}(url)
		}
		wg.Wait()
		urls = nextUrls
	}
	return nil
}
