package sandblast

import (
	"bytes"
	"fmt"
	"strings"
	"golang.org/x/net/html"
)

func extractEx(node *html.Node, destructive bool) (title, text string, simplified, flattened, cleaned *element, err error) {
	root := findRoot(node)
	if root == nil {
		err = fmt.Errorf("Could not find root")
		return
	}
	
	title = getTitle(root)
	simplified, flattened, cleaned = extractTextEx(root, destructive)
	text = cleaned.String()
	return
}

func ExtractEx(node *html.Node) (title, text string, simplified, flattened, cleaned *element, err error) {
	title, text, simplified, flattened, cleaned, err = extractEx(node, false)
	return
}

func Extract(node *html.Node) (title, text string, err error) {
	title, text, _, _, _, err = extractEx(node, true)
	return
}

func findRoot(node *html.Node) *html.Node {
	if node == nil {
		return nil
	}
	if node.Type == html.DocumentNode {
		return findRoot(node.FirstChild)
	}
	for node != nil {
		if (node.Type == html.ElementNode) && (strings.ToLower(node.Data) == "html") {
			return node
		}
		node = node.NextSibling
	}
	return nil
}

func getTitle(root *html.Node) string {
	head := findChild(root, "head")
	title := findChild(head, "title")
	if title == nil {
		return ""
	}
	return strings.TrimSpace(findContent(title.FirstChild))
	
}

func findChild(root *html.Node, name string) *html.Node {
	if root == nil {
		return nil
	}
	name = strings.ToLower(name)
	child := root.FirstChild
	for child != nil {
		if (child.Type == html.ElementNode) && (strings.ToLower(child.Data) == name) {
			return child
		}
		child = child.NextSibling
	}
	return nil
}

func findContent(node *html.Node) string {
	if node == nil {
		return ""
	}
	out := bytes.NewBuffer([]byte{})
	for node != nil {
		if node.Type == html.TextNode {
			out.Write([]byte(node.Data))
		}
		node = node.NextSibling
	}
	return string(out.Bytes())
}
