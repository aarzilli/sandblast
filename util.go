package sandblast

import (
	"golang.org/x/net/html"
	"strings"
	"unicode"
)

func collapseWhitespace(in []rune) []rune {
	var b []rune = make([]rune, len(in))
	d := 0
	spaceSeen := true
	for s := range in {
		if spaceSeen {
			if !unicode.IsSpace(in[s]) {
				spaceSeen = false
				b[d] = in[s]
				d++
			}
		} else {
			if unicode.IsSpace(in[s]) {
				b[d] = ' '
				d++
				spaceSeen = true
			} else {
				b[d] = in[s]
				d++
			}
		}
	}
	return b[:d]
}

type cleanAsciiArtStateFn func(int) cleanAsciiArtStateFn

func cleanAsciiArt(in []rune) []rune {
	b := make([]rune, 0, len(in))
	start := 0
	count := 0

	var baseSpace, baseNormal, maybeAsciiArt cleanAsciiArtStateFn

	isAsciiArt := func(r rune) bool {
		return !unicode.In(r, unicode.Ll, unicode.Lu, unicode.Lt, unicode.Lm, unicode.Lo, unicode.Nd, unicode.Nl, unicode.No)
	}

	baseSpace = func(s int) cleanAsciiArtStateFn {
		//println("baseSpace <", string(in[s]), ">", isAsciiArt(in[s]))
		if unicode.IsSpace(in[s]) {
			b = append(b, in[s])
			return baseSpace
		} else if isAsciiArt(in[s]) {
			start = s
			count = 1
			return maybeAsciiArt
		} else {
			b = append(b, in[s])
			return baseNormal
		}
	}

	baseNormal = func(s int) cleanAsciiArtStateFn {
		//println("baseNormal <", string(in[s]), ">",)
		b = append(b, in[s])
		if unicode.IsSpace(in[s]) {
			return baseSpace
		} else {
			return baseNormal
		}
	}

	maybeAsciiArt = func(s int) cleanAsciiArtStateFn {
		//println("maybeAsciiArt <", string(in[s]), ">")
		if isAsciiArt(in[s]) && !unicode.IsSpace(in[s]) {
			count++
			return maybeAsciiArt
		} else if unicode.IsSpace(in[s]) {
			//println("exiting", count)
			if count > 3 {
				b = append(b, in[s])
			} else {
				b = append(b, in[start:s+1]...)
			}
			return baseSpace
		} else {
			//println("exiting (to normal)")
			b = append(b, in[start:s+1]...)
			return baseNormal
		}
	}

	state := baseSpace

	for s := range in {
		state = state(s)
	}

	return b
}

func cleanControl(in []rune) []rune {
	var b []rune = nil

	for s := range in {
		if unicode.IsControl(in[s]) && !unicode.IsSpace(in[s]) {
			if b == nil {
				b = make([]rune, 0, len(in))
				b = append(b, in[:s]...)
			}
		} else {
			if b != nil {
				b = append(b, in[s])
			}
		}
		s++
	}

	if b == nil {
		return in
	}
	return b
}

func getAttribute(node *html.Node, name string) string {
	for i := range node.Attr {
		if strings.ToLower(node.Attr[i].Key) == "href" {
			return node.Attr[i].Val
		}
	}
	return ""
}
