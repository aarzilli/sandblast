package sandblast

import (
	"bytes"
	"fmt"
	"code.google.com/p/go.net/html"
	henc "html"
	"strings"
)

type element struct {
	tag string
	childs []*element
	content string
	collapse bool
	originalTag string
	linkPart float32
}

func (el *element) Clone() *element {
	r := &element{}
	r.tag = el.tag
	r.content = el.content
	r.collapse = el.collapse
	r.originalTag = el.originalTag
	r.linkPart = el.linkPart
	if el.childs != nil {
		r.childs = make([]*element, len(el.childs))
		for i := range r.childs {
			r.childs[i] = el.childs[i].Clone()
		}
	} else {
		r.childs = nil
	}
	return r
}

func newContentElement(tag, content string) *element {
	return &element{ tag, nil, content, false, "", 0.0 }
}

func newChildElement(tag string, childs []*element) *element {
	return &element{ tag, childs, "", false, "", 0.0 }
}

// Returns a representation of the element suitable for debugging the library
func (e *element) DebugString() string {
	out := bytes.NewBuffer([]byte{})
	e.debugStringEx(out, 0)
	return string(out.Bytes())
}

func (e *element) debugStringEx(out *bytes.Buffer, depth int) {
	out.Write([]byte(makeIndent(depth)))
	
	fmt.Fprintf(out, "<%s", e.tag)
	if e.originalTag != "" {
		fmt.Fprintf(out, ":%s", e.originalTag)
	}
	if e.linkPart > 0.001 {
		fmt.Fprintf(out, ":%g", e.linkPart)
	}
	out.Write([]byte{ '>' })
	if e.childs == nil {
		fmt.Fprintf(out, "[%s(%d)]\n", e.content, len(e.content))
	} else {
		fmt.Fprintf(out, "[%d]\n", len(e.childs))
		sep := false
		for _, child := range e.childs {
			if child != nil {
				child.debugStringEx(out, depth+1)
				sep =  false
			} else {
				if !sep {
					out.Write([]byte{ '\n' })
					sep = true
				}
				
				fmt.Fprintf(out, "%s%v\n", makeIndent(depth+1), child)
			}
		}
	}
}

func (e *element) String() string {
	if e == nil {
		return "<nil>"
	}
	out := bytes.NewBuffer([]byte{})
	e.stringEx(out)
	return string(out.Bytes())
}

func (e *element) stringEx(out *bytes.Buffer) {
	if e.childs == nil {
		out.Write([]byte(strings.TrimSpace(e.content)))
	} else {
		out.Write([]byte{ '\n' })
		sep := true
		for _, child := range e.childs {
			if child != nil {
				child.stringEx(out)
			} else {
				if !sep {
					out.Write([]byte{ '\n' })
					sep = true
				}
			}
		}
	}
	
	out.Write([]byte{ '\n' })
}

func (e *element) isHeader() bool {
	if e.tag != "~textdiv" && e.tag != "~text" {
		return false
	}
	return e.originalTag == "h"
}

func (e *element) isLinkList() bool {
	if e.childs == nil {
		return false
	}
	
	if len(e.childs) < 5 {
		return false
	}
	
	if e.tag == "select" {
		return true
	}
	
	nlinks := 0
	for _, child := range e.childs {
		if child.tag != "~text" && child.tag != "~textdiv" {
			return false
		}
		if child.linkPart > 0.70 {
			nlinks++
		}
	}
	if nlinks < 2 {
		return false
	}
	return (nlinks >= len(e.childs) - 2) || (nlinks > int(float32(len(e.childs)) * 0.75))
}

func (e *element) isLinkBlob() bool {
	if e.tag != "~text" && e.tag != "~textdiv" {
		return false
	}
	return e.linkPart > 0.7
}

func (e *element) okText() bool {
	return e != nil && e.tag == "~textblock" && len(e.content) > 50
}

/* Fuses a text element to the last text element in childs.
If this is not possible (for example because childs doesn't end with a text element) returns false
*/
func pushTextEx(childs []*element, ts string, tsLinkPart float32) bool {
	if childs == nil || len(childs) == 0 {
		return false
	}
	last := childs[len(childs) - 1]
	if last.tag != "~text" {
		return false
	}
	newLinkPart := float32(len(ts)) * tsLinkPart + float32(len(last.content)) * last.linkPart;
	last.content += " " + ts;
	last.linkPart = newLinkPart / float32(len(last.content))
	return true
}

// Adds a new text element to childs
func pushText(childs []*element, node *html.Node) []*element {
	ts := []rune(henc.UnescapeString(node.Data))
	ts = collapseWhitespace(ts)
	ts = cleanAsciiArt(ts)
	ts = cleanControl(ts)
	
	if len(ts) <= 0 {
		return childs
	}
	
	added := pushTextEx(childs, string(ts), 0.0)
	if !added {
		childs = append(childs, newContentElement("~text", string(ts)))
	}
	return childs
}

func pushElement(childs []*element, child *element) []*element {
	if !child.collapse {
		added := false
		if child.tag == "~text" {
			added = pushTextEx(childs, child.content, child.linkPart)
		}
		if !added {
			childs = append(childs, child)
		}
		return childs
	} else {
		// collapsing
		for _, cc := range child.childs {
			added := false
			if cc.tag == "~text" {
				added = pushTextEx(childs, cc.content, cc.linkPart)
			}
			if !added {
				childs = append(childs, cc)
			}
		}
		return childs
	}
}

