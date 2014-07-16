package main

import (
	"os"
	"fmt"
	"bytes"
	"github.com/aarzilli/sandblast"
	"code.google.com/p/go.net/html"
	"log"
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: extract <url> [debug]\n")
	os.Exit(1)
}

func main() {
	if len(os.Args) < 2 {
		usage()
	}
	
	url := os.Args[1]
	isDebug := false
	if len(os.Args) >= 3 {
		if os.Args[2] == "debug" {
			isDebug = true
		} else {
			usage()
		}
	}
	
	rawhtml, _, _, err := sandblast.FetchURL(url)
	if err != nil {
		log.Fatalf("Could not fetch url: %s\n", url)
	}
	
	node, err := html.Parse(bytes.NewReader([]byte(rawhtml)))
	if err != nil {
		log.Fatal("Parsing error: ", err)
	}
	title, text, simplified, flattened, cleaned, err := sandblast.ExtractEx(node)
	if err != nil {
		log.Fatal("Extraction error: ", err)
	}
	
	fmt.Printf("TITLE: %s\n", title)
	if isDebug {
		fmt.Printf("SIMPLIFIED:\n%s\n", simplified.DebugString())
		fmt.Printf("FLATTENED:\n%s\n", flattened.DebugString())
		fmt.Printf("CLEANED:\n%s\n", cleaned.DebugString())
	}
	fmt.Printf("TEXT:\n%s\n", text)
}
