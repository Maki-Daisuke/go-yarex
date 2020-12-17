package yarex

import (
	"regexp"
	"testing"
)

func testMatchStrings(t *testing.T, restr string, tests []string) {
	ast, err := parse(restr)
	if err != nil {
		t.Fatalf("want nil, but got %s", err)
	}
	ast = optimizeAst(ast)
	yaRe := MustCompile(restr)
	goRe := regexp.MustCompile(restr)
	for _, str := range tests {
		match := goRe.MatchString(str)
		if astMatch(ast, str) != match {
			if match {
				t.Errorf("(Interp) %v should match against %q, but didn't", ast, str)
			} else {
				t.Errorf("(Interp) %v shouldn't match against %q, but did", ast, str)
			}
		}
		if yaRe.MatchString(str) != match {
			if match {
				t.Errorf("(OpTree) %v should match against %q, but didn't", yaRe, str)
			} else {
				t.Errorf("(OpTree) %v shouldn't match against %q, but did", yaRe, str)
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

func TestMatchQuantifier(t *testing.T) {
	testMatchStrings(t, "fo{2,5}oh", []string{
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
	pattern := `(hoge)\1fuga`
	ast, err := parse(pattern)
	if err != nil {
		t.Fatalf("want nil, but got %s", err)
	}
	ast = optimizeAst(ast)
	yaRe := MustCompile(pattern)
	for _, test := range tests {
		if astMatch(ast, test.str) != test.result {
			if test.result {
				t.Errorf("(Interp) %v should match against %q, but didn't", ast, test.str)
			} else {
				t.Errorf("(Interp) %v shouldn't match against %q, but did", ast, test.str)
			}
		}
		if yaRe.MatchString(test.str) != test.result {
			if test.result {
				t.Errorf("(OpTree) %v should match against %q, but didn't", yaRe, test.str)
			} else {
				t.Errorf("(OpTree) %v shouldn't match against %q, but did", yaRe, test.str)
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

func TestSipAddress(t *testing.T) {
	testMatchStrings(t, `^["]{0,1}([^"]*)["]{0,1}[ ]*<(sip|tel|sips):(([^@]*)@){0,1}([^>^:]*|\[[a-fA-F0-9:]*\]):{0,1}([0-9]*){0,1}>(;.*){0,1}$`, []string{
		"\"display_name\"<sip:0312341234@10.0.0.1:5060>;user=phone;hogehoge",
		"<sip:0312341234@10.0.0.1>",
		"\"display_name\"<sip:0312341234@10.0.0.1>",
		"<sip:whois.this>;user=phone",
		"\"0333334444\"<sip:[2001:30:fe::4:123]>;user=phone",
	})
}
