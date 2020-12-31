package yarex_test

//go:generate cmd/yarexgen/yarexgen generate_test.go

import (
	"testing"

	"github.com/Maki-Daisuke/go-yarex"
)

func testMatchStrings(t *testing.T, restr string, tests []string) {
	opRe := yarex.MustCompileOp(restr)
	if !yarex.IsOpMatcher(opRe) {
		t.Errorf("%v should be OpTree matcher, but isn't", opRe)
	}
	compRe := yarex.MustCompile(restr) // should be compiled version
	if !yarex.IsCompiledMatcher(compRe) {
		t.Errorf("%v should be Compiled matcher, but isn't", compRe)
	}
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
	pattern := "foo bar" //yarexgen
	testMatchStrings(t, pattern, []string{
		"foo bar",
		"foo  bar",
		"hogefoo barfuga",
		"foo barf",
		"Afoo bar",
		"foo ba",
	})
}

func TestMatchFooOrBar(t *testing.T) {
	pattern := "foo|bar" //yarexgen
	testMatchStrings(t, pattern, []string{
		"foo bar",
		"hogefoo barfuga",
		"foo baz",
		"bar f",
		"foba",
		"",
	})
}

func TestMatchFooOrBarOrBaz(t *testing.T) {
	pattern := "foo|bar|baz" //yarexgen
	testMatchStrings(t, pattern, []string{
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

func TestMatchBacktracking(t *testing.T) {
	pattern := "(?:foo|fo)oh" //yarexgen
	testMatchStrings(t, pattern, []string{
		"fooh",
		"foooh",
		"foh",
		"fooooooooooh",
		"fooooooooofoooh",
		"",
	})
}

func TestMatchZeroOrMore(t *testing.T) {
	pattern := "fo*oh" //yarexgen
	testMatchStrings(t, pattern, []string{
		"fooh",
	})
}

func TestMatchOneOrMore(t *testing.T) {
	pattern := "fo+oh" //yarexgen
	testMatchStrings(t, pattern, []string{
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
	pattern := "fo{2,5}oh" //yarexgen
	testMatchStrings(t, pattern, []string{
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
	pattern := "fo?oh" //yarexgen
	testMatchStrings(t, pattern, []string{
		"fooh",
		"foh",
		"fh",
		"fooooooooooh",
		"fooooooooofooh",
		"",
		"fo",
		"oh",
	})
	pattern = "fo*oh?" //yarexgen
	testMatchStrings(t, pattern, []string{
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
	pattern := "." //yarexgen
	testMatchStrings(t, pattern, []string{
		"aiueo",
		"\n",
		"",
		" ",
		"\b",
	})
	pattern = ".+x" //yarexgen
	testMatchStrings(t, pattern, []string{
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
	pattern := "^foo bar" //yarexgen
	testMatchStrings(t, pattern, []string{
		"foo bar",
		"foo  bar",
		"hogefoo barfuga",
		"foo barf",
		"Afoo bar",
		"foo ba",
		"\nfoo bar",
	})
	pattern = "(^|A)*foo bar" //yarexgen
	testMatchStrings(t, pattern, []string{
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
	pattern := `(hoge)\1fuga` //yarexgen
	testMatchStrings(t, pattern, []string{
		"hogehogefuga",
		"AAAhogehogefugaBBB",
		"hogefuga",
		"hoge",
		"fuga",
	})
}

func TestMatchClass(t *testing.T) {
	pattern := "[0aB]" //yarexgen
	testMatchStrings(t, pattern, []string{
		"foo",      // false
		"foo  bar", // true
		"FOO BAR",  // true
		"AAAAAA",   // false
		"012345",   // true
		"\000hoge", // false
		"\000hage", // true
	})
	pattern = "[A-Z0-9][a-z]" //yarexgen
	testMatchStrings(t, pattern, []string{
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
	pattern := `^["]{0,1}([^"]*)["]{0,1}[ ]*<(sip|tel|sips):(([^@]*)@){0,1}([^>^:]*|\[[a-fA-F0-9:]*\]):{0,1}([0-9]*){0,1}>(;.*){0,1}$` //yarexgen
	testMatchStrings(t, pattern, []string{
		"\"display_name\"<sip:0312341234@10.0.0.1:5060>;user=phone;hogehoge",
		"<sip:0312341234@10.0.0.1>",
		"\"display_name\"<sip:0312341234@10.0.0.1>",
		"<sip:whois.this>;user=phone",
		"\"0333334444\"<sip:[2001:30:fe::4:123]>;user=phone",
	})
}
