package reaot

import "testing"

func TestParseFooBar(t *testing.T) {
	re, err := parse("foo bar")
	if err != nil {
		t.Errorf("want %v, but got %v", nil, err)
	}
	seq, ok := re.(*ReSeq)
	if !ok {
		t.Fatalf("want *ReSeq, but got %T", re)
	}
	if seq.String() != "(foo bar)" {
		t.Errorf("want %q, but got %q", "(foo bar)", seq)
	}
}

func TestParseFooOrBar(t *testing.T) {
	re, err := parse("foo|bar")
	if err != nil {
		t.Errorf("want %v, but got %v", nil, err)
	}
	alt, ok := re.(*ReAlt)
	if !ok {
		t.Fatalf("want *ReAlt, but got %v of type %T", re, re)
	}
	seq, ok := (alt.opts[0]).(*ReSeq)
	if !ok {
		t.Fatalf("want *ReSeq, but got %v of type %T", alt.opts[0], alt.opts[0])
	}
	if seq.String() != "(foo)" {
		t.Errorf("want %q, but got %q", "foo", seq)
	}
	seq, ok = (alt.opts[1]).(*ReSeq)
	if !ok {
		t.Fatalf("want *ReSeq, but got %v of type %T", alt.opts[1], alt.opts[1])
	}
	if seq.String() != "(bar)" {
		t.Errorf("want %q, but got %q", "bar", seq)
	}
}
