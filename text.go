package sandblast

import (
	"code.google.com/p/go.net/html"
	"strings"
	"bytes"
)

const _MAX_PROCESSING_DEPTH = 100

func extractTextEx(root *html.Node, destructive bool) (simplified, flattened, cleaned *element) {
	simplified = simplify(root, 0)
	if destructive {
		flattened = flatten(simplified)
	} else {
		x := simplified.Clone()
		//println("Flatten argument:", x.DebugString())
		flattened = flatten(x)
	}
	if destructive {
		cleaned = clean(flattened)
	} else {
		cleaned = clean(flattened.Clone())
	}
	return
}

func simplify(node *html.Node, depth int) *element {
	if depth > _MAX_PROCESSING_DEPTH {
		return nil
	}
	
	switch node.Type {
	case html.ErrorNode:
		return nil
	case html.CommentNode:
		return nil
	case html.DoctypeNode:
		return nil
	case html.DocumentNode:
		return nil

	case html.TextNode:
		return newContentElement("~text", node.Data)
		
	case html.ElementNode:
		// rest
	}
	
	kind := getNodeKind(node)
	if kind == _K_SUPPRESSED {
		return nil
	}
	
	childs := []*element{}
	
	for childn := node.FirstChild; childn != nil; childn = childn.NextSibling {
		if childn.Type == html.TextNode {
			childs = pushText(childs, childn)
		} else {
			child := simplify(childn, depth+1)
			if child != nil {
				childs = pushElement(childs, child)
			}
		}
	}
		
	if len(childs) == 0 {
		return nil
	}
	
	kot := false
	switch kind {
	case _K_KOTCONTAINER:
		kot = true
		fallthrough
	case _K_CONTAINER:
		if len(childs) == 1 {
			if childs[0].tag == "~text" {
				if kot {
					childs[0].originalTag = "h"
				}
				childs[0].tag = "~textdiv"
			}
			return childs[0]
		}
		
	case _K_FORMATTING:
		if len(childs) == 1 {
			return childs[0]
		}
		
	case _K_INLINE:
		if len(childs) <= 0 {
			return nil
		}
		
		if len(childs) == 1 {
			if (childs[0].tag == "~text" || childs[0].tag == "~textdiv") && strings.ToLower(node.Data) == "a" {
				childs[0].linkPart = 1.0
			}
			return childs[0]
		}
		
	case _K_TODESTRUCTURE:
		if strings.ToLower(node.Data) == "tr" {
			linkPart := float32(0.0)
			hasComplexTds := false
			trText := bytes.NewBuffer([]byte{})
			
			for _, child := range childs {
				if !(child.tag == "~textdiv") {
					hasComplexTds = true
					break
				}
				
				trText.Write([]byte(child.content))
				linkPart += float32(len(child.content)) * child.linkPart
			}
			
			if !hasComplexTds {
				r := newContentElement("~textdiv", string(trText.Bytes()))
				r.linkPart = linkPart / float32(len(r.content))
				return r
			}
		}
		
		r := newChildElement("~transient", childs)
		r.collapse = true
		
		return r
	}
	
	return newChildElement(strings.ToLower(node.Data), childs)
}

func flatten(e *element) *element {
	if e == nil {
		return e
	}
	
	if e.isHeader() {
		e.tag = "~header"
		return e
	}
	
	if e.isLinkList() {
		e.tag = "~linklist"
		return e
	}
	
	if e.isLinkBlob() {
		e.tag = "~linkblob"
		return e
	}
	
	if e.childs == nil {
		e.tag = "~textblock"
		return e
	}
	
	childs := make([]*element, 0, len(e.childs))
	
	for i := range e.childs {
		fchild := flatten(e.childs[i])
		if fchild.collapse {
			for _, subchild := range fchild.childs {
				childs = append(childs, subchild)
			}
		} else {
			childs = append(childs, fchild)
		}
	}
	
	e.tag = "~transient"
	e.childs = childs
	e.collapse = true
	return e
}

func clean(e *element) *element {
	if e == nil || e.childs == nil {
		return e
	}
	
	for i := range e.childs {
		if e.childs[i] == nil {
			continue
		}
		
		switch e.childs[i].tag {
		case "~linkblob":
			fallthrough
		case "~linklist":
			e.childs[i] = nil
		
		case "~textblock":
			if len(e.childs[i].content) <= 15 || strings.Index(e.childs[i].content, " ") < 0 {
				e.childs[i] = nil
			}
		}
	}
	
	for i := range e.childs {
		if e.childs[i] == nil {
			continue
		}
		
		var next *element
		if i+1 < len(e.childs) {
			next = e.childs[i+1]
		}
		
		var prev *element
		if i-1 >= 0 {
			prev = e.childs[i-1]
		}
		
		if e.tag == "~header" {
			if !next.okText() {
				e.childs[i] = nil
			}
		} else if !e.childs[i].okText() {
			if !next.okText() && !prev.okText() {
				e.childs[i] = nil
			}
		}
	}
	
	return e
}

func makeIndent(depth int) string {
	b := make([]byte, depth*3)
	for i := range b {
		b[i] = ' '
	}
	return string(b)
}

