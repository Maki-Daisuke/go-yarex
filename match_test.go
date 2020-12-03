package reaot

import (
	"regexp"
	"testing"
)

func testMatchStrings(t *testing.T, restr string, tests []string) {
	re, err := parse(restr)
	if err != nil {
		t.Fatalf("want nil, but got %s", err)
	}
	goRe := regexp.MustCompile(restr)
	for _, str := range tests {
		match := goRe.MatchString(str)
		if Match(re, str) != match {
			if match {
				t.Errorf("%v should match against %q, but didn't", re, str)
			} else {
				t.Errorf("%v shouldn't match against %q, but did", re, str)
			}
		}
	}
}

func TestMatchFooBar(t *testing.T) {
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
	testMatchStrings(t, "foo|bar", []string{
		"foo bar",
		"hogefoo barfuga",
		"foo baz",
		"bar f",
		"foba",
		"",
	})
}

func TestMatchBacktracking(t *testing.T) {
	testMatchStrings(t, "(foo|fo)oh", []string{
		"fooh",
		"foooh",
		"foh",
		"fooooooooooh",
		"fooooooooofoooh",
		"",
	})
}

func TestMatchZeroOrMore(t *testing.T) {
	testMatchStrings(t, "fo*oh", []string{
		"fooh",
		"foh",
		"fh",
		"fooooooooooh",
		"fooooooooofoooh",
		"",
		"fo",
		"oh",
	})
}

func TestMatchOneOrMore(t *testing.T) {
	testMatchStrings(t, "fo+oh", []string{
		"fooh",
		"foh",
		"fh",
		"fooooooooooh",
		"fooooooooofoooh",
		"",
		"fo",
		"oh",
	})
}

func TestMatchOpt(t *testing.T) {
	testMatchStrings(t, "fo?oh", []string{
		"fooh",
		"foh",
		"fh",
		"fooooooooooh",
		"fooooooooofooh",
		"",
		"fo",
		"oh",
	})
	testMatchStrings(t, "fo*oh?", []string{
		"ABfooh",
		"foo",
		"fh",
		"foooohoooooo",
		"foooooooooooCD",
		"",
		"fo",
		"oh",
	})
}

func TestMatchWildcard(t *testing.T) {
	testMatchStrings(t, ".", []string{
		"aiueo",
		"\n",
		"",
		" ",
		"\b",
	})
	testMatchStrings(t, ".+x", []string{
		"",
		"x",
		"xx",
		"aaaaax",
		"\nx",
		"xx\nx",
		"xxxxxa",
	})
}

func TestMatchBegin(t *testing.T) {
	testMatchStrings(t, "^foo bar", []string{
		"foo bar",
		"foo  bar",
		"hogefoo barfuga",
		"foo barf",
		"Afoo bar",
		"foo ba",
		"\nfoo bar",
	})
	testMatchStrings(t, "(^|A)*foo bar", []string{
		"foo bar",
		"foo  bar",
		"hogefoo barfuga",
		"foo barf",
		"Afoo bar",
		"AAfoo bar",
		"AAAAfoo bar",
		"AABAAfoo bar",
	})
}

func TestMatchBackRef(t *testing.T) {
	// Here, we cannot use testMatchStrings, because Go's regexp does not
	// support back-reference.
	tests := []struct {
		str    string
		result bool
	}{
		{"hogehogefuga", true},
		{"AAAhogehogefugaBBB", true},
		{"hogefuga", false},
		{"hoge", false},
		{"fuga", false},
	}
	re, err := parse(`(hoge)\1fuga`)
	if err != nil {
		t.Fatalf("want nil, but got %s", err)
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

func TestMatchClass(t *testing.T) {
	testMatchStrings(t, "[0aB]", []string{
		"foo",      // false
		"foo  bar", // true
		"FOO BAR",  // true
		"AAAAAA",   // false
		"012345",   // true
		"\000hoge", // false
		"\000hage", // true
	})
	testMatchStrings(t, "[A-Z0-9][a-z]", []string{
		"absksdjhasd",
		"alsdAAA",
		"asl;k3as7djj",
		"Aiiiiiiii9",
		"foo BAR",
		"FOO bar",
		"FOObar",
		"fooBARbaz",
	})
}
