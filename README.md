# mirroring-web-crawler

This is a recursive web crawler in Golang.

Challenge's description:

The crawler should be a command-line tool that accepts a starting URL and a destination directory. The crawler will then download the page at the URL, save it in the destination directory, and then recursively proceed to any valid links in this page.

A valid link is the value of an href attribute in an <a> tag the resolves to urls that are children of the initial URL. For example, given initial URL `https://start.url/abc`, URLs that resolve to `https://start.url/abc/foo` and `https://start.url/abc/foo/bar` are valid URLs, but ones that resolve to `https://another.domain` or to `https://start.url/baz` are not valid URLs, and should be skipped.

Additionally, the crawler should:
- Correctly handle being interrupted by Ctrl-C
- Perform work in parallel where reasonable
- Support resume functionality by checking the destination directory for downloaded pages and skip downloading and processing where not necessary
- Provide “happy-path” test coverage

## considerations

- I could implement a worker pool instead of spanning a new go routine for each URL to be downloaded, but I did this way for the sake of time.
- As per challenge's description, the tool only download links (`<a href=...>`), so no assets (css, javascript) are downloaded.
- `wget` provide a flag `--convert-links` which transforms all links to local ones, that's a nice feature.

## qualities

- low [cyclomatic complexity](https://en.wikipedia.org/wiki/Cyclomatic_complexity) (<= 5)
- no linters
- no [vulnerabilities](https://go.dev/blog/vuln) found
- unit test coverage greater than 96%. The few lines not covered are the ones that access external resources, so, naturally, we mock them.

## Makefile help

```
make help
```

## running it

```
make run URL=https://blog.cleancoder.com/ DEST_DIR=saved
```

or 

```
go run cmd/main.go -u https://blog.cleancoder.com/ -d saved
```

Running it again you'll see log messages stating that files already exist.

## unit tests

```
make test
```

### unit test coverage report

```
make coverage
```

## integration test

```
make int-test
```

## linter

```
make lint
```

## check for vulnerabilities

```
make vul-check
```