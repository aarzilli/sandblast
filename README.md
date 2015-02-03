Library that uses Readability-like heuristics to extract text from an HTML document.

Example:
```go
import "golang.org/x/net/html"
…
node, err := html.Parse(bytes.NewReader(raw_html))
if err != nil {
	log.Fatal("Parsing error: ", err)
}
title, text := sandblast.Extract(node)
fmt.Printf("Title: %s\n%s", title, text)
…
```
See also `example/extract.go`, a command line utility to extract text from a URL.
