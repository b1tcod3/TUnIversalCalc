package engine

import "testing"

func TestLexTokens(t *testing.T) {
	toks, err := lex("2x+3.5*sin(y)")
	if err != nil {
		t.Fatalf("lex: %v", err)
	}
	var got []string
	for _, tk := range toks {
		if tk.Kind != TokEOF {
			got = append(got, tk.Text)
		}
	}
	want := []string{"2", "x", "+", "3.5", "*", "sin", "(", "y", ")"}
	if len(got) != len(want) {
		t.Fatalf("tokens = %v, se esperaba %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("token %d = %q, se esperaba %q", i, got[i], want[i])
		}
	}
}

func TestLexConstantsAndFunctions(t *testing.T) {
	cases := []struct {
		in   string
		kind TokKind
		text string
	}{
		{"pi", TokConst, "pi"},
		{"inf", TokConst, "inf"},
		{"sqrt(x)", TokFunc, "sqrt"},
		{"ABS(2)", TokFunc, "abs"},
	}
	for _, c := range cases {
		toks, err := lex(c.in)
		if err != nil {
			t.Fatalf("lex(%q): %v", c.in, err)
		}
		if toks[0].Kind != c.kind || toks[0].Text != c.text {
			t.Errorf("lex(%q) → %v (%v), se esperaba %v (%v)", c.in, toks[0].Text, toks[0].Kind, c.text, c.kind)
		}
	}
}

func TestLexErrors(t *testing.T) {
	for _, in := range []string{"a@b", "1.2.3", "ln(x)", "foo(x)"} {
		if _, err := lex(in); err == nil {
			t.Errorf("lex(%q) debe fallar", in)
		}
	}
}

func TestLexMultiLetterVars(t *testing.T) {
	toks, err := lex("2alpha+beta gamma")
	if err != nil {
		t.Fatalf("lex: %v", err)
	}
	var got []string
	for _, tk := range toks {
		if tk.Kind != TokEOF {
			got = append(got, tk.Text)
		}
	}
	want := []string{"2", "alpha", "+", "beta", "gamma"}
	if len(got) != len(want) {
		t.Fatalf("tokens = %v, se esperaba %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("token %d = %q, se esperaba %q", i, got[i], want[i])
		}
	}

	// Dos letras siguen siendo un producto implícito para no romper 4ab^2.
	toks, err = lex("4ab^2")
	if err != nil {
		t.Fatalf("lex: %v", err)
	}
	got = got[:0]
	for _, tk := range toks {
		if tk.Kind != TokEOF {
			got = append(got, tk.Text)
		}
	}
	want = []string{"4", "a", "b", "^", "2"}
	if len(got) != len(want) {
		t.Fatalf("tokens = %v, se esperaba %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("token %d = %q, se esperaba %q", i, got[i], want[i])
		}
	}
}

func TestLexParensAndEqual(t *testing.T) {
	toks, err := lex("(x+1)=0")
	if err != nil {
		t.Fatalf("lex: %v", err)
	}
	if toks[0].Kind != TokLParen || toks[4].Kind != TokRParen || toks[5].Kind != TokEq {
		t.Fatalf("paréntesis/igual mal lexeados: %+v", toks)
	}
}
