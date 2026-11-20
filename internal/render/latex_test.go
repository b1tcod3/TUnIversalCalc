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
		{`raiz(3, 8)`, `\sqrt[3]{8}`},
		{`raiz(9)`, `\sqrt{9}`},
		{`raiz(2, x^2+1)`, `\sqrt{x^{2} + 1}`},
		{`raiz(5, raiz(2, x))`, `\sqrt[5]{\sqrt{x}}`},
		{"-2x-1", "-2x - 1"},
		{"a/b", "a / b"},
		{"x^2", "x^{2}"},
		{"x^(2+1)+y", "x^{2 + 1} + y"},
		{"3*-2", "3 \\cdot -2"},
		{"inf+1", "\\infty + 1"},
		{"2pi", "2\\pi"},
		{"spider", "spider"},
		{"sqrtx", "sqrtx"},
		{"", ""},
	}
	for _, c := range cases {
		if got := ToLatex(c.in); got != c.want {
			t.Errorf("ToLatex(%q) = %q, se esperaba %q", c.in, got, c.want)
		}
	}
}
