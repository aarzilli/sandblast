package sandblast

import (
	"golang.org/x/net/html"
	"strings"
)

type nodeKind int
const (
	_K_SUPPRESSED = nodeKind(iota)
	_K_TODESTRUCTURE
	_K_CONTAINER
	_K_KOTCONTAINER
	_K_FORMATTING
	_K_INLINE
)

var elements = map[string]nodeKind{
	/* Suppressed */
	"head": _K_SUPPRESSED,
	"base": _K_SUPPRESSED, "link": _K_SUPPRESSED, "meta": _K_SUPPRESSED, "title": _K_SUPPRESSED,
	"script": _K_SUPPRESSED, "noscript": _K_SUPPRESSED, "style": _K_SUPPRESSED,
	"input": _K_SUPPRESSED, "label": _K_SUPPRESSED, "textarea": _K_SUPPRESSED, "button": _K_SUPPRESSED,
	"isindex": _K_SUPPRESSED,
	"object": _K_SUPPRESSED, "applet": _K_SUPPRESSED, "img": _K_SUPPRESSED, "map": _K_SUPPRESSED,
	"address": _K_SUPPRESSED,
	"basefont": _K_SUPPRESSED,
	"colgroup": _K_SUPPRESSED, "col": _K_SUPPRESSED, "caption": _K_SUPPRESSED,
	"br": _K_SUPPRESSED, "hr": _K_SUPPRESSED,
	"canvas": _K_SUPPRESSED,
	"audio": _K_SUPPRESSED, "video": _K_SUPPRESSED, "source": _K_SUPPRESSED, "track": _K_SUPPRESSED, "embed": _K_SUPPRESSED,
	"datalist": _K_SUPPRESSED, "keygen": _K_SUPPRESSED, "output": _K_SUPPRESSED,
	"command": _K_SUPPRESSED, "progress": _K_SUPPRESSED,
	"ruby": _K_SUPPRESSED, "rt": _K_SUPPRESSED, "rp": _K_SUPPRESSED,

	/* To Destructure */
	"html": _K_TODESTRUCTURE,
	"tbody": _K_TODESTRUCTURE, "thread": _K_TODESTRUCTURE, "tfoot": _K_TODESTRUCTURE, "tr": _K_TODESTRUCTURE, "th": _K_TODESTRUCTURE,
	"form": _K_TODESTRUCTURE, "fieldset": _K_TODESTRUCTURE,
	"optgroup": _K_TODESTRUCTURE,
	"iframe": _K_TODESTRUCTURE,
	"legend": _K_TODESTRUCTURE, "bdo": _K_TODESTRUCTURE,
	"abbr": _K_TODESTRUCTURE, "acronym": _K_TODESTRUCTURE,
	"figure": _K_TODESTRUCTURE, "figcaption": _K_TODESTRUCTURE,

	/* Container */
	"div": _K_CONTAINER, "span": _K_CONTAINER,
	"select": _K_CONTAINER, "option": _K_CONTAINER,
	"table": _K_CONTAINER, "td": _K_CONTAINER,
	"dir": _K_CONTAINER, "dl": _K_CONTAINER, "dt": _K_CONTAINER, "dd": _K_CONTAINER,
	"menu": _K_CONTAINER,
	"ul": _K_CONTAINER, "ol": _K_CONTAINER, "li": _K_CONTAINER,
	"blockquote": _K_CONTAINER, "p": _K_CONTAINER, "cite": _K_CONTAINER, "pre": _K_CONTAINER,
	"h4": _K_CONTAINER, "h5": _K_CONTAINER, "h6": _K_CONTAINER,
	"header": _K_CONTAINER, "hgroup": _K_CONTAINER, "main": _K_CONTAINER, "article": _K_CONTAINER, "aside": _K_CONTAINER, "footer": _K_CONTAINER, "details": _K_CONTAINER, "summary": _K_CONTAINER, 
	"nav": _K_CONTAINER, "section": _K_CONTAINER,
	"dialog": _K_CONTAINER,
	
	/* Keep Original Tag Container */
	"h1": _K_KOTCONTAINER, "h2": _K_KOTCONTAINER, "h3": _K_KOTCONTAINER,

	/* Formatting */
	"tt": _K_FORMATTING,
	"small": _K_FORMATTING, "big": _K_FORMATTING,
	"s": _K_FORMATTING, "strike": _K_FORMATTING,
	"center": _K_FORMATTING,
	"dfn": _K_FORMATTING, "del": _K_FORMATTING,
	"kbd": _K_FORMATTING, "samp": _K_FORMATTING, "var": _K_FORMATTING, "code": _K_FORMATTING,
	"q'": _K_FORMATTING, "ins": _K_FORMATTING,
	"sub": _K_FORMATTING, "sup": _K_FORMATTING,
	"font": _K_FORMATTING,
	"mark": _K_FORMATTING, "time": _K_FORMATTING, "bdi": _K_FORMATTING, "wbr": _K_FORMATTING,
	"meter": _K_FORMATTING,

	/* Inline */
	"strong": _K_INLINE, "b": _K_INLINE,
	"i": _K_INLINE, "em": _K_INLINE,
	"u": _K_INLINE,
	"a": _K_INLINE,

}

func getNodeKind(node *html.Node) nodeKind {
	kind, ok := elements[strings.ToLower(node.Data)]
	if !ok {
		kind, _ = elements["div"]
	}
	return kind
}
