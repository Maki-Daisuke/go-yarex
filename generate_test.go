package yarex_test

//go:generate cmd/yarexgen/yarexgen generate_test.go

import (
	"testing"

	"github.com/Maki-Daisuke/go-yarex"
)

func testMatchStrings(t *testing.T, restr string, tests []string) {
	opRe := yarex.MustCompileOp(restr)
	compRe := yarex.MustCompile(restr) // should be compiled version
	for _, str := range tests {
		match := opRe.MatchString(str)
		if compRe.MatchString(str) != match {
			if match {
				t.Errorf("(AOT-C) %q should match against %q, but didn't", compRe, str)
			} else {
				t.Errorf("(AOT-C) %q shouldn't match against %q, but did", compRe, str)
			}
		}
	}
}

func TestMatchFooBar(t *testing.T) {
	yarex.MustCompile("foo bar") //yarexgen
	testMatchStrings(t, "foo bar", []string{
		"foo bar",
		"foo  bar",
		"hogefoo barfuga",
		"foo barf",
		"Afoo bar",
		"foo ba",
	})
}

func TestMatchFooOrBar(t *testing.T) {
	yarex.MustCompile("foo|bar") //yarexgen
	testMatchStrings(t, "foo|bar", []string{
		"foo bar",
		"hogefoo barfuga",
		"foo baz",
		"bar f",
		"foba",
		"",
	})
}

func TestMatchFooOrBarOrBaz(t *testing.T) {
	yarex.MustCompile("foo|bar|baz") //yarexgen
	testMatchStrings(t, "foo|bar|baz", []string{
		"foo bar",
		"hogefoo barfuga",
		"foo baz",
		"bar f",
		"foba",
		"ba",
		"baz",
		"fobabaf",
		"fbboaaorz",
	})
}
