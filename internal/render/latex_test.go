package render

import "testing"

func TestToLatex(t *testing.T) {
	cases := []struct{ in, want string }{
		{"2x+2", "2x + 2"},
		{"2x+2=0", "2x + 2 = 0"},
		{"x^2+3x*2", "x^{2} + 3x \\cdot 2"},
		{"x^(2y)+pi", "x^{2y} + \\pi"},
		{`sqrt(x^2+1)`, `\sqrt{x^{2} + 1}`},
		{`sqrt(sqrt(x))`, `\sqrt{\sqrt{x}}`},
		{"-2x-1", "-2x - 1"},
		{"a/b", "a / b"},
		{"x^2", "x^{2}"},
		{"x^(2+1)+y", "x^{2 + 1} + y"},
		{"3*-2", "3 \\cdot -2"},
		{"inf+1", "\\infty + 1"},
		{"", ""},
	}
	for _, c := range cases {
		if got := ToLatex(c.in); got != c.want {
			t.Errorf("ToLatex(%q) = %q, se esperaba %q", c.in, got, c.want)
		}
	}
}