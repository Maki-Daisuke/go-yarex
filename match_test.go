package yarex_test

//go:generate cmd/yarexgen/yarexgen match_test.go

import (
	"regexp"
	"testing"

	"github.com/Maki-Daisuke/go-yarex"
)

func testMatchStrings(t *testing.T, restr string, tests []string) {
	ast, err := yarex.Parse(restr)
	if err != nil {
		t.Fatalf("want nil, but got %s", err)
	}
	stdRe := regexp.MustCompile(restr)
	ast = yarex.OptimizeAst(ast)
	opRe := yarex.MustCompileOp(restr)
	compRe := yarex.MustCompile(restr)
	if !yarex.IsCompiledMatcher(compRe) {
		t.Errorf("%v should be Compiled matcher, but isn't", compRe)
	}
	for _, str := range tests {
		match := stdRe.MatchString(str)
		if yarex.AstMatch(ast, str) != match {
			if match {
				t.Errorf("(Interp) %v should match against %q, but didn't", ast, str)
			} else {
				t.Errorf("(Interp) %v shouldn't match against %q, but did", ast, str)
			}
		}
		if opRe.MatchString(str) != match {
			if match {
				t.Errorf("(OpTree) %v should match against %q, but didn't", opRe, str)
			} else {
				t.Errorf("(OpTree) %v shouldn't match against %q, but did", opRe, str)
			}
		}
		if compRe.MatchString(str) != match {
			if match {
				t.Errorf("(Compiled) %v should match against %q, but didn't", compRe, str)
			} else {
				t.Errorf("(Compiled) %v shouldn't match against %q, but did", compRe, str)
			}
		}
	}
}

func TestMatchFooBar(t *testing.T) {
	re := "foo bar" //yarexgen
	testMatchStrings(t, re, []string{
		"foo bar",
		"foo  bar",
		"hogefoo barfuga",
		"foo barf",
		"Afoo bar",
		"foo ba",
	})
}

func TestMatchFooOrBar(t *testing.T) {
	re := "foo|bar" //yarexgen
	testMatchStrings(t, re, []string{
		"foo bar",
		"hogefoo barfuga",
		"foo baz",
		"bar f",
		"foba",
		"",
	})
}

func TestMatchBacktracking(t *testing.T) {
	re := "(?:foo|fo)oh" //yarexgen
	testMatchStrings(t, re, []string{
		"fooh",
		"foooh",
		"foh",
		"fooooooooooh",
		"fooooooooofoooh",
		"",
	})
}

func TestMatchZeroOrMore(t *testing.T) {
	re := "fo*oh" //yarexgen
	testMatchStrings(t, re, []string{
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
	re := "fo+oh" //yarexgen
	testMatchStrings(t, re, []string{
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

func TestMatchQuantifier(t *testing.T) {
	re := "fo{2,5}oh" //yarexgen
	testMatchStrings(t, re, []string{
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
	re := "fo?oh" //yarexgen
	testMatchStrings(t, re, []string{
		"fooh",
		"foh",
		"fh",
		"fooooooooooh",
		"fooooooooofooh",
		"",
		"fo",
		"oh",
	})
	re = "fo*oh?" //yarexgen
	testMatchStrings(t, re, []string{
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
	re := "." //yarexgen
	testMatchStrings(t, re, []string{
		"aiueo",
		"\n",
		"",
		" ",
		"\b",
	})
	re = ".+x" //yarexgen
	testMatchStrings(t, re, []string{
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
	re := "^foo bar" //yarexgen
	testMatchStrings(t, re, []string{
		"foo bar",
		"foo  bar",
		"hogefoo barfuga",
		"foo barf",
		"Afoo bar",
		"foo ba",
		"\nfoo bar",
	})
	re = "(^|A)*foo bar" //yarexgen
	testMatchStrings(t, re, []string{
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
	pattern := `(hoge)\1fuga` //yarexgen
	ast, err := yarex.Parse(pattern)
	if err != nil {
		t.Fatalf("want nil, but got %s", err)
	}
	ast = yarex.OptimizeAst(ast)
	opRe := yarex.MustCompileOp(pattern)
	compRe := yarex.MustCompileOp(pattern)
	for _, test := range tests {
		if yarex.AstMatch(ast, test.str) != test.result {
			if test.result {
				t.Errorf("(Interp) %v should match against %q, but didn't", ast, test.str)
			} else {
				t.Errorf("(Interp) %v shouldn't match against %q, but did", ast, test.str)
			}
		}
		if opRe.MatchString(test.str) != test.result {
			if test.result {
				t.Errorf("(OpTree) %v should match against %q, but didn't", opRe, test.str)
			} else {
				t.Errorf("(OpTree) %v shouldn't match against %q, but did", opRe, test.str)
			}
		}
		if compRe.MatchString(test.str) != test.result {
			if test.result {
				t.Errorf("(Compiled) %v should match against %q, but didn't", compRe, test.str)
			} else {
				t.Errorf("(Compiled) %v shouldn't match against %q, but did", compRe, test.str)
			}
		}
	}
}

func TestMatchClass(t *testing.T) {
	re := "[0aB]" //yarexgen
	testMatchStrings(t, re, []string{
		"foo",      // false
		"foo  bar", // true
		"FOO BAR",  // true
		"AAAAAA",   // false
		"012345",   // true
		"\000hoge", // false
		"\000hage", // true
	})
	re = "[A-Z0-9][a-z]" //yarexgen
	testMatchStrings(t, re, []string{
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

func TestSipAddress(t *testing.T) {
	re := `^["]{0,1}([^"]*)["]{0,1}[ ]*<(sip|tel|sips):(([^@]*)@){0,1}([^>^:]*|\[[a-fA-F0-9:]*\]):{0,1}([0-9]*){0,1}>(;.*){0,1}$` //yarexgen
	testMatchStrings(t, re, []string{
		"\"display_name\"<sip:0312341234@10.0.0.1:5060>;user=phone;hogehoge",
		"<sip:0312341234@10.0.0.1>",
		"\"display_name\"<sip:0312341234@10.0.0.1>",
		"<sip:whois.this>;user=phone",
		"\"0333334444\"<sip:[2001:30:fe::4:123]>;user=phone",
	})
}
