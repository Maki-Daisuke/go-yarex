package reaot

import "testing"

func TestParseFooBar(t *testing.T) {
	re, err := parse("foo bar")
	if err != nil {
		t.Errorf("want %v, but got %v", nil, err)
	}
	seq, ok := re.(*AstSeq)
	if !ok {
		t.Fatalf("want *AstSeq, but got %T", re)
	}
	if seq.String() != "(?:foo bar)" {
		t.Errorf("want %q, but got %q", "(?:foo bar)", seq)
	}
}

func TestParseFooOrBar(t *testing.T) {
	re, err := parse("foo|bar")
	if err != nil {
		t.Errorf("want %v, but got %v", nil, err)
	}
	alt, ok := re.(*AstAlt)
	if !ok {
		t.Fatalf("want *AstAlt, but got %v of type %T", re, re)
	}
	seq, ok := (alt.opts[0]).(*AstSeq)
	if !ok {
		t.Fatalf("want *AstSeq, but got %v of type %T", alt.opts[0], alt.opts[0])
	}
	if seq.String() != "(?:foo)" {
		t.Errorf("want %q, but got %q", "(?:foo)", seq)
	}
	seq, ok = (alt.opts[1]).(*AstSeq)
	if !ok {
		t.Fatalf("want *AstSeq, but got %v of type %T", alt.opts[1], alt.opts[1])
	}
	if seq.String() != "(?:bar)" {
		t.Errorf("want %q, but got %q", "(?:bar)", seq)
	}
}
