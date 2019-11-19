package reaot

import "testing"

func TestParseFooBar(t *testing.T) {
	re, err := parse("foo bar")
	if err != nil {
		t.Errorf("want %v, but got %v", nil, err)
	}
	lit, ok := re.(*ReLit)
	if !ok {
		t.Fatalf("want *ReLit, but got %T", re)
	}
	if lit.str != "foo bar" {
		t.Errorf("want %q, but got %q", "foo bar", lit.str)
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
	lit, ok := (alt.opts[0]).(*ReLit)
	if !ok {
		t.Fatalf("want *ReLit, but got %v of type %T", alt.opts[0], alt.opts[0])
	}
	if lit.str != "foo" {
		t.Errorf("want %q, but got %q", "foo", lit.str)
	}
	lit, ok = (alt.opts[1]).(*ReLit)
	if !ok {
		t.Fatalf("want *ReLit, but got %v of type %T", alt.opts[1], alt.opts[1])
	}
	if lit.str != "bar" {
		t.Errorf("want %q, but got %q", "bar", lit.str)
	}
}
