package reaot

import "testing"

func TestMatchFooBar(t *testing.T) {
	re, err := parse("foo bar")
	if err != nil {
		t.Fatalf("want nil, but got %s", err)
	}
	tests := []struct {
		str    string
		result bool
	}{
		{"foo bar", true}, {"foo  bar", false}, {"hogefoo barfuga", true}, {"foo barf", true}, {"Afoo bar", true}, {"foo ba", false},
	}
	for _, test := range tests {
		if Match(re, test.str) != test.result {
			if test.result {
				t.Errorf("%v should match against %q, but didn't", re, test.str)
			} else {
				t.Errorf("%v shouldn't match against %q, but did", re, test.str)
			}
		}
	}
}

func TestMatchFooOrBar(t *testing.T) {
	re, err := parse("foo|bar")
	if err != nil {
		t.Fatalf("want nil, but got %s", err)
	}
	tests := []struct {
		str    string
		result bool
	}{
		{"foo bar", true},
		{"hogefoo barfuga", true},
		{"foo baz", true},
		{"bar f", true},
		{"foba", false},
		{"", false},
	}
	for _, test := range tests {
		if Match(re, test.str) != test.result {
			if test.result {
				t.Errorf("%v should match against %q, but didn't", re, test.str)
			} else {
				t.Errorf("%v shouldn't match against %q, but did", re, test.str)
			}
		}
	}
}

func TestMatchBacktracking(t *testing.T) {
	re, err := parse("(foo|fo)oh")
	if err != nil {
		t.Fatalf("want nil, but got %s", err)
	}
	tests := []struct {
		str    string
		result bool
	}{
		{"fooh", true},
		{"foooh", true},
		{"foh", false},
		{"fooooooooooh", false},
		{"fooooooooofoooh", true},
		{"", false},
	}
	for _, test := range tests {
		if Match(re, test.str) != test.result {
			if test.result {
				t.Errorf("%v should match against %q, but didn't", re, test.str)
			} else {
				t.Errorf("%v shouldn't match against %q, but did", re, test.str)
			}
		}
	}
}

func TestMatchZeroOrMore(t *testing.T) {
	re, err := parse("fo*oh")
	if err != nil {
		t.Fatalf("want nil, but got %s", err)
	}
	tests := []struct {
		str    string
		result bool
	}{
		{"fooh", true},
		{"foh", true},
		{"fh", false},
		{"fooooooooooh", true},
		{"fooooooooofoooh", true},
		{"", false},
		{"fo", false},
		{"oh", false},
	}
	for _, test := range tests {
		if Match(re, test.str) != test.result {
			if test.result {
				t.Errorf("%v should match against %q, but didn't", re, test.str)
			} else {
				t.Errorf("%v shouldn't match against %q, but did", re, test.str)
			}
		}
	}
}

func TestMatchOneOrMore(t *testing.T) {
	re, err := parse("fo+oh")
	if err != nil {
		t.Fatalf("want nil, but got %s", err)
	}
	tests := []struct {
		str    string
		result bool
	}{
		{"fooh", true},
		{"foh", false},
		{"fh", false},
		{"fooooooooooh", true},
		{"fooooooooofoooh", true},
		{"", false},
		{"fo", false},
		{"oh", false},
	}
	for _, test := range tests {
		if Match(re, test.str) != test.result {
			if test.result {
				t.Errorf("%v should match against %q, but didn't", re, test.str)
			} else {
				t.Errorf("%v shouldn't match against %q, but did", re, test.str)
			}
		}
	}
}
