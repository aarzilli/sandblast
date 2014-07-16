package sandblast

import (
	"testing"
)

func TestCleanControl(t *testing.T) {
	tf := func(in, target string) {
		out := string(cleanControl([]rune(in)))
		if out != target {
			t.Errorf("Error cleaning control character on <%s>\n\tgot <%s>\n\texpected <%s>\n", in, out, target)
		}
	}
	tf("test\ntest\002 test", "test\ntest test")
}

func TestCollapseWhitespace(t *testing.T) {
	tf := func(in, target string) {
		out := string(collapseWhitespace([]rune(in)))
		if out != target {
			t.Errorf("Error collapsing whitespace on <%s>\n\tgot <%s>\n\texpected <%s>\n", in, out, target)
		}
	}
	tf("   \n\ttest\ntest\ttest  ", "test test test ")
}

func TestCleanAsciiArt(t *testing.T) {
	tf := func(in, target string) {
		out := string(cleanAsciiArt([]rune(in)))
		if out != target {
			t.Errorf("Error cleaning ASCII art on <%s>\n\tgot <%s>\n\texpected <%s>\n", in, out, target)
		}
	}
	tf("test ===== test === test", "test  test === test")
}

