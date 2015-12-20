package sandblast

import (
	"bytes"
	"fmt"
	"golang.org/x/net/html"
	henc "html"
	"io"
	"strings"
)

type element struct {
	tag         string
	childs      []*element
	content     string
	collapse    bool
	originalTag string
	linkPart    float32
	hrefs       []string
}

const (
	_LINK_START = "\x11"
	_LINK_END   = "\x13"
)

func (el *element) Clone() *element {
	r := &element{}
	r.tag = el.tag
	r.content = el.content
	r.collapse = el.collapse
	r.originalTag = el.originalTag
	r.linkPart = el.linkPart
	r.hrefs = make([]string, len(el.hrefs))
	copy(r.hrefs, el.hrefs)
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
	return &element{tag, nil, content, false, "", 0.0, nil}
}

func newChildElement(tag string, childs []*element) *element {
	return &element{tag, childs, "", false, "", 0.0, nil}
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
	if len(e.hrefs) > 0 {
		fmt.Fprintf(out, ":%s", strings.Join(e.hrefs, ","))
	}
	out.Write([]byte{'>'})
	if e.childs == nil {
		var lctxt linkContext
		fmt.Fprintf(out, "[%s(%d)]\n", lctxt.convertLinks(e.content, true), len(e.content))
	} else {
		fmt.Fprintf(out, "[%d]\n", len(e.childs))
		sep := false
		for _, child := range e.childs {
			if child != nil {
				child.debugStringEx(out, depth+1)
				sep = false
			} else {
				if !sep {
					out.Write([]byte{'\n'})
					sep = true
				}

				fmt.Fprintf(out, "%s%v\n", makeIndent(depth+1), child)
			}
		}
	}
}

func (e *element) String(flags Flags) string {
	if e == nil {
		return "<nil>"
	}
	out := bytes.NewBuffer([]byte{})
	var lctxt linkContext
	e.stringEx(out, flags, &lctxt)
	if flags&KeepLinks != 0 {
		if len(lctxt.hrefs) > 0 {
			io.WriteString(out, "\n")
		}
		for i := range lctxt.hrefs {
			fmt.Fprintf(out, "\t[%d] %s\n", i, lctxt.hrefs[i])
		}
	}
	if lctxt.cnt != len(lctxt.hrefs) {
		fmt.Fprintf(out, "LINK COUNT INCONSISTENCY (%d %d)\n", lctxt.cnt, len(lctxt.hrefs))
	}
	return string(out.Bytes())
}

func (e *element) stringEx(out *bytes.Buffer, flags Flags, lctxt *linkContext) {
	if e.childs == nil {
		if len(e.hrefs) > 0 {
			ctnt := lctxt.convertLinks(e.content, flags&KeepLinks != 0)
			io.WriteString(out, strings.TrimSpace(ctnt))
			lctxt.push(e.hrefs)
		} else {
			io.WriteString(out, strings.TrimSpace(e.content))
		}
	} else {
		out.Write([]byte{'\n'})
		sep := true
		for _, child := range e.childs {
			if child != nil {
				child.stringEx(out, flags, lctxt)
			} else {
				if !sep {
					out.Write([]byte{'\n'})
					sep = true
				}
			}
		}
	}

	out.Write([]byte{'\n'})
}

type linkContext struct {
	cnt   int
	hrefs []string
}

func (lctxt *linkContext) convertLinks(s string, keep bool) string {
	in := []byte(s)
	out := make([]byte, 0, len(in))
	for _, ch := range in {
		switch ch {
		case _LINK_START[0]:
			// nothing
		case _LINK_END[0]:
			if keep {
				out = append(out, []byte(fmt.Sprintf(" [%d]", lctxt.cnt))...)
			}
			lctxt.cnt++
		default:
			out = append(out, ch)
		}
	}
	return string(out)
}

func (lctxt *linkContext) push(hrefs []string) {
	lctxt.hrefs = append(lctxt.hrefs, hrefs...)
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
	return (nlinks >= len(e.childs)-2) || (nlinks > int(float32(len(e.childs))*0.75))
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
func pushTextEx(childs []*element, ts string, hrefs []string, tsLinkPart float32) bool {
	if childs == nil || len(childs) == 0 {
		return false
	}
	last := childs[len(childs)-1]
	if last.tag != "~text" {
		return false
	}
	newLinkPart := float32(len(ts))*tsLinkPart + float32(len(last.content))*last.linkPart
	last.content += " " + ts
	last.linkPart = newLinkPart / float32(len(last.content))
	last.hrefs = append(last.hrefs, hrefs...)
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

	added := pushTextEx(childs, string(ts), nil, 0.0)
	if !added {
		childs = append(childs, newContentElement("~text", string(ts)))
	}
	return childs
}

func pushElement(childs []*element, child *element) []*element {
	if !child.collapse {
		added := false
		if child.tag == "~text" {
			added = pushTextEx(childs, child.content, child.hrefs, child.linkPart)
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
				added = pushTextEx(childs, cc.content, cc.hrefs, cc.linkPart)
			}
			if !added {
				childs = append(childs, cc)
			}
		}
		return childs
	}
}
