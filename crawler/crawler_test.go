// Copyright (c) 2023 Tiago Melo. All rights reserved.
// Use of this source code is governed by the MIT License that can be found in
// the LICENSE file.
package crawler

import (
	"bytes"
	"errors"
	"io"
	"io/fs"
	"log"
	"net/http"
	goUrl "net/url"
	"os"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/require"
	"github.com/tcmsoftware/mirroring-web-crawler/fixtures"
)

func Test_getUrl(t *testing.T) {
	testCases := []struct {
		name          string
		mockedGet     func(httpClient *http.Client, url string) (*http.Response, error)
		expectedError error
	}{
		{
			name: "happy path",
			mockedGet: func(httpClient *http.Client, url string) (*http.Response, error) {
				return new(http.Response), nil
			},
		},
		{
			name: "error",
			mockedGet: func(httpClient *http.Client, url string) (*http.Response, error) {
				return nil, errors.New("random error")
			},
			expectedError: errors.New("making get request to some url: random error"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			get = tc.mockedGet
			c := New("someurl", "somedir", 0, nil)
			resp, err := getUrl(c.httpClient, "some url")
			if err != nil {
				checkIfErrorIsExpected(t, err, tc.expectedError)
				require.Equal(t, tc.expectedError.Error(), err.Error())
			} else {
				checkIfErrorIsNotExpected(t, err, tc.expectedError)
				require.NotNil(t, resp)
			}
		})
	}
}

func Test_getNextUrls(t *testing.T) {
	testCases := []struct {
		name             string
		fixture          string
		alreadyVisited   []string
		expectedMextUrls []string
	}{
		{
			name:    "existing links are from same domain and not yet visited",
			fixture: fixtures.AllLinksFromSameDomain,
			expectedMextUrls: []string{
				"someurl/some_section/2023/01/19/page1.html",
				"someurl/some_section/2023/02/13/page2.html",
			},
		},
		{
			name:    "existing links are from same domain, but one was already visited",
			fixture: fixtures.AllLinksFromSameDomain,
			alreadyVisited: []string{
				"someurl/some_section/2023/01/19/page1.html",
			},
			expectedMextUrls: []string{
				"someurl/some_section/2023/02/13/page2.html",
			},
		},
		{
			name:    "one link without href",
			fixture: fixtures.OneLinkWithoutHref,
			expectedMextUrls: []string{
				"someurl/some_section/2023/02/13/page2.html",
			},
		},
		{
			name:    "mixed domains",
			fixture: fixtures.LiksWithMixedDomains,
			expectedMextUrls: []string{
				"someurl/some_section/2023/02/13/page2.html",
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c := New("someurl", "somedir", 0, nil)
			for _, u := range tc.alreadyVisited {
				c.visitedUrls[u] = true
			}
			doc, err := fixtures.HtmlToDoc(tc.fixture)
			require.Nil(t, err)
			nextUrls := getNextUrls(c, doc)
			require.Equal(t, tc.expectedMextUrls, nextUrls)
		})
	}
}

func Test_parseResponse(t *testing.T) {
	testCases := []struct {
		name                             string
		mockGoqueryNewDocumentFromReader func(r io.Reader) (*goquery.Document, error)
		expectedError                    error
	}{
		{
			name: "happy path",
			mockGoqueryNewDocumentFromReader: func(r io.Reader) (*goquery.Document, error) {
				return new(goquery.Document), nil
			},
		},
		{
			name: "error",
			mockGoqueryNewDocumentFromReader: func(r io.Reader) (*goquery.Document, error) {
				return nil, errors.New("random error")
			},
			expectedError: errors.New("parsing response from url someurl: random error"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			goqueryNewDocumentFromReader = tc.mockGoqueryNewDocumentFromReader
			resp := &http.Response{
				Body: io.NopCloser(bytes.NewReader([]byte(""))),
			}
			doc, err := parseResponse("someurl", resp)
			if err != nil {
				checkIfErrorIsExpected(t, err, tc.expectedError)
				require.Equal(t, tc.expectedError.Error(), err.Error())
			} else {
				checkIfErrorIsNotExpected(t, err, tc.expectedError)
				require.NotNil(t, doc)
			}
		})
	}
}

func Test_getPagePath(t *testing.T) {
	testCases := []struct {
		name          string
		url           string
		mockParseUrl  func(rawURL string) (*goUrl.URL, error)
		expectedPath  string
		expectedError error
	}{
		{
			name:         "adds index.html for root path",
			url:          "https://blog.cleancoder.com/",
			expectedPath: "destDir/blog.cleancoder.com/index.html",
		},
		{
			name:         "adds uri",
			url:          "https://blog.cleancoder.com/uncle-bob/2019/02/01/somePage.html",
			expectedPath: "destDir/blog.cleancoder.com/uncle-bob/2019/02/01/somePage.html",
		},
		{
			name: "error",
			url:  "https://blog.cleancoder.com/",
			mockParseUrl: func(rawURL string) (*goUrl.URL, error) {
				return nil, errors.New("random error")
			},
			expectedError: errors.New("parsing url https://blog.cleancoder.com/: random error"),
		},
	}
	originalParseUrl := parseUrl
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.mockParseUrl != nil {
				parseUrl = tc.mockParseUrl
			} else {
				parseUrl = originalParseUrl
			}
			path, err := getPagePath("destDir", tc.url)
			if err != nil {
				checkIfErrorIsExpected(t, err, tc.expectedError)
				require.Equal(t, tc.expectedError.Error(), err.Error())
			} else {
				checkIfErrorIsNotExpected(t, err, tc.expectedError)
				require.Equal(t, tc.expectedPath, path)
			}
		})
	}
}

func Test_saveToDisk(t *testing.T) {
	testCases := []struct {
		name                  string
		mockOsCreate          func(name string) (*os.File, error)
		mockGetDocHtml        func(doc *goquery.Document) (string, error)
		mockWriteStringToFile func(f *os.File, data string) (int, error)
		expectedError         error
	}{
		{
			name: "happy path",
			mockOsCreate: func(name string) (*os.File, error) {
				return new(os.File), nil
			},
			mockGetDocHtml: func(doc *goquery.Document) (string, error) {
				return "something", nil
			},
			mockWriteStringToFile: func(f *os.File, data string) (int, error) {
				bytesWritten := 10
				return bytesWritten, nil
			},
		},
		{
			name: "error creating file",
			mockOsCreate: func(name string) (*os.File, error) {
				return nil, errors.New("random error")
			},
			expectedError: errors.New("creating file for someurl: random error"),
		},
		{
			name: "error getting doc html",
			mockOsCreate: func(name string) (*os.File, error) {
				return new(os.File), nil
			},
			mockGetDocHtml: func(doc *goquery.Document) (string, error) {
				return "", errors.New("random error")
			},
			expectedError: errors.New("converting someurl to HTML: random error"),
		},
		{
			name: "error writing file",
			mockOsCreate: func(name string) (*os.File, error) {
				return new(os.File), nil
			},
			mockGetDocHtml: func(doc *goquery.Document) (string, error) {
				return "something", nil
			},
			mockWriteStringToFile: func(f *os.File, data string) (int, error) {
				return 0, errors.New("random error")
			},
			expectedError: errors.New("writing HTML file for someurl: random error"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			osCreate = tc.mockOsCreate
			getDocHtml = tc.mockGetDocHtml
			writeStringToFile = tc.mockWriteStringToFile
			err := saveToDisk("someurl", "somepath", new(goquery.Document))
			if err != nil {
				checkIfErrorIsExpected(t, err, tc.expectedError)
				require.Equal(t, tc.expectedError.Error(), err.Error())
			} else {
				checkIfErrorIsNotExpected(t, err, tc.expectedError)
			}
		})
	}
}

func Test_savePage(t *testing.T) {
	testCases := []struct {
		name            string
		mockGetPagePath func(destDir string, url string) (string, error)
		mockOsStat      func(name string) (fs.FileInfo, error)
		mockOsMkdirAll  func(path string, perm fs.FileMode) error
		mockSaveToDisk  func(url string, pagePath string, doc *goquery.Document) error
		expectedError   error
	}{
		{
			name: "happy path",
			mockGetPagePath: func(destDir, url string) (string, error) {
				return "path", nil
			},
			mockOsStat: func(name string) (fs.FileInfo, error) {
				return nil, errors.New("random error")
			},
			mockOsMkdirAll: func(path string, perm fs.FileMode) error {
				return nil
			},
			mockSaveToDisk: func(url, pagePath string, doc *goquery.Document) error {
				return nil
			},
		},
		{
			name: "file already exists",
			mockGetPagePath: func(destDir, url string) (string, error) {
				return "path", nil
			},
			mockOsStat: func(name string) (fs.FileInfo, error) {
				return nil, nil
			},
		},
		{
			name: "error getting page path",
			mockGetPagePath: func(destDir, url string) (string, error) {
				return "", errors.New("random error")
			},
			expectedError: errors.New("random error"),
		},
		{
			name: "error creating dir",
			mockGetPagePath: func(destDir, url string) (string, error) {
				return "path", nil
			},
			mockOsStat: func(name string) (fs.FileInfo, error) {
				return nil, errors.New("random error")
			},
			mockOsMkdirAll: func(path string, perm fs.FileMode) error {
				return errors.New("random error")
			},
			expectedError: errors.New("creating directory for someurl: random error"),
		},
		{
			name: "error saving to disk",
			mockGetPagePath: func(destDir, url string) (string, error) {
				return "path", nil
			},
			mockOsStat: func(name string) (fs.FileInfo, error) {
				return nil, errors.New("random error")
			},
			mockOsMkdirAll: func(path string, perm fs.FileMode) error {
				return nil
			},
			mockSaveToDisk: func(url, pagePath string, doc *goquery.Document) error {
				return errors.New("random error")
			},
			expectedError: errors.New("random error"),
		},
	}
	log := log.New(os.Stdout, "UNIT TEST :", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			getPagePath = tc.mockGetPagePath
			osStat = tc.mockOsStat
			osMkdirAll = tc.mockOsMkdirAll
			saveToDisk = tc.mockSaveToDisk
			err := savePage("destDir", "someurl", new(goquery.Document), log)
			if err != nil {
				checkIfErrorIsExpected(t, err, tc.expectedError)
				require.Equal(t, tc.expectedError.Error(), err.Error())
			} else {
				checkIfErrorIsNotExpected(t, err, tc.expectedError)
			}
		})
	}
}

func Test_getAbsoluteUrl(t *testing.T) {
	testCases := []struct {
		name           string
		startUrl       string
		link           string
		expectedOutput string
	}{
		{
			name:           "link starts with forward slash",
			startUrl:       "somedomain",
			link:           "/somelink",
			expectedOutput: "somedomain/somelink",
		},
		{
			name:           "link does not start with forward slash",
			startUrl:       "somedomain",
			link:           "somelink",
			expectedOutput: "somelink",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			au := getAbsoluteUrl(tc.startUrl, tc.link)
			require.Equal(t, tc.expectedOutput, au)
		})
	}
}

func Test_processUrl(t *testing.T) {
	testCases := []struct {
		name              string
		mockGetUrl        func(httpClient *http.Client, url string) (*http.Response, error)
		mockParseResponse func(url string, resp *http.Response) (*goquery.Document, error)
		mockSavePage      func(destDir string, url string, doc *goquery.Document, log *log.Logger) error
		expectedError     error
	}{
		{
			name: "happy path",
			mockGetUrl: func(httpClient *http.Client, url string) (*http.Response, error) {
				return new(http.Response), nil
			},
			mockParseResponse: func(url string, resp *http.Response) (*goquery.Document, error) {
				return new(goquery.Document), nil
			},
			mockSavePage: func(destDir, url string, doc *goquery.Document, log *log.Logger) error {
				return nil
			},
		},
		{
			name: "error getting url",
			mockGetUrl: func(httpClient *http.Client, url string) (*http.Response, error) {
				return nil, errors.New("random error")
			},
			expectedError: errors.New("random error"),
		},
		{
			name: "error parsing response",
			mockGetUrl: func(httpClient *http.Client, url string) (*http.Response, error) {
				return new(http.Response), nil
			},
			mockParseResponse: func(url string, resp *http.Response) (*goquery.Document, error) {
				return nil, errors.New("random error")
			},
			expectedError: errors.New("random error"),
		},
		{
			name: "error saving page",
			mockGetUrl: func(httpClient *http.Client, url string) (*http.Response, error) {
				return new(http.Response), nil
			},
			mockParseResponse: func(url string, resp *http.Response) (*goquery.Document, error) {
				return new(goquery.Document), nil
			},
			mockSavePage: func(destDir, url string, doc *goquery.Document, log *log.Logger) error {
				return errors.New("random error")
			},
			expectedError: errors.New("random error"),
		},
	}
	log := log.New(os.Stdout, "UNIT TEST :", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			getUrl = tc.mockGetUrl
			parseResponse = tc.mockParseResponse
			savePage = tc.mockSavePage
			doc, err := processUrl(new(http.Client), "destDir", "someurl", log)
			if err != nil {
				checkIfErrorIsExpected(t, err, tc.expectedError)
				require.Equal(t, tc.expectedError.Error(), err.Error())
			} else {
				checkIfErrorIsNotExpected(t, err, tc.expectedError)
				require.NotNil(t, doc)
			}
		})
	}
}

func TestRun(t *testing.T) {
	testCases := []struct {
		name            string
		mockProcessUrl  func(httpClient *http.Client, destDir string, url string, log *log.Logger) (*goquery.Document, error)
		mockGetNextUrls func(c *Crawler, doc *goquery.Document) []string
	}{
		{
			name: "happy path",
			mockProcessUrl: func(httpClient *http.Client, destDir, url string, log *log.Logger) (*goquery.Document, error) {
				return new(goquery.Document), nil
			},
			mockGetNextUrls: func(c *Crawler, doc *goquery.Document) []string {
				return []string{}
			},
		},
		{
			name: "error processing url",
			mockProcessUrl: func(httpClient *http.Client, destDir, url string, log *log.Logger) (*goquery.Document, error) {
				return nil, errors.New("random error")
			},
			mockGetNextUrls: func(c *Crawler, doc *goquery.Document) []string {
				return []string{}
			},
		},
	}
	log := log.New(os.Stdout, "UNIT TEST :", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			processUrl = tc.mockProcessUrl
			getNextUrls = tc.mockGetNextUrls
			c := New("firsturl", "somedir", 0, log)
			err := c.Run()
			require.Nil(t, err)
		})
	}
}

func checkIfErrorIsExpected(t *testing.T, err, expectedError error) {
	if expectedError == nil {
		t.Fatalf(`expected no error, got "%v"`, err)
	}
}

func checkIfErrorIsNotExpected(t *testing.T, err, expectedError error) {
	if expectedError != nil {
		t.Fatalf(`expected error "%v", got nil`, expectedError.Error())
	}
}
