package ui

import (
	"strings"
	"testing"
)

func TestVisualWidth(t *testing.T) {
	cases := []struct {
		in   string
		want int
	}{
		{"", 0},
		{"abc", 3},
		{"\x1b[31mabc\x1b[0m", 3},
		{"漢字", 4},
		{"á", 1},
	}
	for _, c := range cases {
		if got := visualWidth(c.in); got != c.want {
			t.Errorf("visualWidth(%q) = %d, se esperaba %d", c.in, got, c.want)
		}
	}
}

func TestTruncateVisual(t *testing.T) {
	cases := []struct {
		in   string
		max  int
		want string
	}{
		{"abcdef", 4, "abc…"},
		{"abc", 4, "abc"},
		{"abcdef", 0, ""},
		{"漢字", 3, "漢…"},
	}
	for _, c := range cases {
		if got := truncateVisual(c.in, c.max); got != c.want {
			t.Errorf("truncateVisual(%q, %d) = %q, se esperaba %q", c.in, c.max, got, c.want)
		}
	}

	styled := "\x1b[31mabcdef\x1b[0m"
	got := truncateVisual(styled, 4)
	if visualWidth(got) > 4 || !strings.Contains(got, "abc") {
		t.Errorf("truncateVisual ANSI = %q (ancho %d)", got, visualWidth(got))
	}
}

func TestPadOrTrimVisual(t *testing.T) {
	if got := padOrTrimVisual("x", 3); got != "x  " {
		t.Errorf("pad = %q", got)
	}
	if got := padOrTrimVisual("abcdef", 3); got != "abc" {
		t.Errorf("trim = %q", got)
	}
	if got := padOrTrimVisual("漢a", 3); got != "漢a" || visualWidth(got) != 3 {
		t.Errorf("trim ancho = %q (%d)", got, visualWidth(got))
	}
}
