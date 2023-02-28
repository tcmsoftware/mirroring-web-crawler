// Copyright (c) 2023 Tiago Melo. All rights reserved.
// Use of this source code is governed by the MIT License that can be found in
// the LICENSE file.
package fixtures

import (
	"io"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/pkg/errors"
)

const (
	InitialUrl             = "http://somedomain.com/"
	AllLinksFromSameDomain = `
	<html>
		<head>
		</head>
		<body>
		<a href="/some_section/2023/01/19/page1.html">Page 1</a>
		<a href="/some_section/2023/02/13/page2.html">Page 2</a>
		</body>
	</html>
	`
	OneLinkWithoutHref = `
	<html>
		<head>
		</head>
		<body>
		<a>Page 1</a>
		<a href="/some_section/2023/02/13/page2.html">Page 2</a>
		</body>
	</html>
	`
	LiksWithMixedDomains = `
	<html>
		<head>
		</head>
		<body>
		<a href="http://somesite.com/some_section/2023/01/19/page1.html">Page 1</a>
		<a href="/some_section/2023/02/13/page2.html">Page 2</a>
		</body>
	</html>
	`
)

func parseHtmlPage(htmlPage string) (*goquery.Document, error) {
	r := io.NopCloser(strings.NewReader(htmlPage))
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, errors.Wrapf(err, "parsing html page %v", htmlPage)
	}
	parsedUrl, err := url.Parse(InitialUrl)
	if err != nil {
		return nil, errors.Wrapf(err, "parsing initial url %v", htmlPage)
	}
	doc.Url = parsedUrl
	return doc, nil
}

func HtmlToDoc(htmlPage string) (*goquery.Document, error) {
	doc, err := parseHtmlPage(htmlPage)
	if err != nil {
		return nil, err
	}
	return doc, nil
}
