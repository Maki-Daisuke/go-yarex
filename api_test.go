package yarex_test

import (
	"reflect"
	"regexp"
	"testing"

	"github.com/Maki-Daisuke/go-yarex"
)

func testAPIs(t *testing.T, restr string, tests []string) {
	stdRe := regexp.MustCompile(restr)
	opRe := yarex.MustCompileOp(restr)
	compRe := yarex.MustCompile(restr)
	for _, str := range tests {
		r := stdRe.FindString(str)
		if opRe.FindString(str) != r {
			t.Errorf("(OpTree) %v.FindString(%q) returned %q, but expected %q", opRe, str, opRe.FindString(str), r)
		}
		if compRe.FindString(str) != r {
			t.Errorf("(Compiled) %v.FindString(%q) returned %q, but expected %q", opRe, str, compRe.FindString(str), r)
		}
		loc := stdRe.FindStringIndex(str)
		if !reflect.DeepEqual(opRe.FindStringIndex(str), loc) {
			t.Errorf("(OpTree) %v.FindStringIndex(%q) returned %v, but expected %v", opRe, str, opRe.FindStringIndex(str), loc)
		}
		if !reflect.DeepEqual(compRe.FindStringIndex(str), loc) {
			t.Errorf("(Compiled) %v.FindStringIndex(%q) returned %v, but expected %v", opRe, str, compRe.FindStringIndex(str), loc)
		}
	}
}

func TestAPI(t *testing.T) {
	re := "foo bar" //yarexgen
	testAPIs(t, re, []string{
		"foo bar",
		"foo  bar",
		"hogefoo barfuga",
		"foo barf",
		"Afoo bar",
		"foo ba",
	})

	re = "foo|bar" //yarexgen
	testAPIs(t, re, []string{
		"foo bar",
		"hogefoo barfuga",
		"foo baz",
		"bar f",
		"foba",
		"",
	})

	re = "(?:foo|fo)oh" //yarexgen
	testAPIs(t, re, []string{
		"fooh",
		"foooh",
		"foh",
		"fooooooooooh",
		"fooooooooofoooh",
		"",
	})

	re = "fo*oh" //yarexgen
	testAPIs(t, re, []string{
		"fooh",
		"foh",
		"fh",
		"fooooooooooh",
		"fooooooooofoooh",
		"",
		"fo",
		"oh",
	})

	re = "fo+oh" //yarexgen
	testAPIs(t, re, []string{
		"fooh",
		"foh",
		"fh",
		"fooooooooooh",
		"fooooooooofoooh",
		"",
		"fo",
		"oh",
	})

	re = "fo{2,5}oh" //yarexgen
	testAPIs(t, re, []string{
		"fooh",
		"foh",
		"fh",
		"fooooooooooh",
		"fooooooooofoooh",
		"",
		"fo",
		"oh",
	})

	re = "fo?oh" //yarexgen
	testAPIs(t, re, []string{
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
	testAPIs(t, re, []string{
		"ABfooh",
		"foo",
		"fh",
		"foooohoooooo",
		"foooooooooooCD",
		"",
		"fo",
		"oh",
	})

	re = "." //yarexgen
	testAPIs(t, re, []string{
		"aiueo",
		"\n",
		"",
		" ",
		"\b",
	})

	re = ".+x" //yarexgen
	testAPIs(t, re, []string{
		"",
		"x",
		"xx",
		"aaaaax",
		"\nx",
		"xx\nx",
		"xxxxxa",
	})

	re = "^foo bar" //yarexgen
	testAPIs(t, re, []string{
		"foo bar",
		"foo  bar",
		"hogefoo barfuga",
		"foo barf",
		"Afoo bar",
		"foo ba",
		"\nfoo bar",
	})

	re = "(^|A)*foo bar" //yarexgen
	testAPIs(t, re, []string{
		"foo bar",
		"foo  bar",
		"hogefoo barfuga",
		"foo barf",
		"Afoo bar",
		"AAfoo bar",
		"AAAAfoo bar",
		"AABAAfoo bar",
	})

	re = "[0aB]" //yarexgen
	testAPIs(t, re, []string{
		"foo",      // false
		"foo  bar", // true
		"FOO BAR",  // true
		"AAAAAA",   // false
		"012345",   // true
		"\000hoge", // false
		"\000hage", // true
	})

	re = "[A-Z0-9][a-z]" //yarexgen
	testAPIs(t, re, []string{
		"absksdjhasd",
		"alsdAAA",
		"asl;k3as7djj",
		"Aiiiiiiii9",
		"foo BAR",
		"FOO bar",
		"FOObar",
		"fooBARbaz",
	})

	re = `^["]{0,1}([^"]*)["]{0,1}[ ]*<(sip|tel|sips):(([^@]*)@){0,1}([^>^:]*|\[[a-fA-F0-9:]*\]):{0,1}([0-9]*){0,1}>(;.*){0,1}$` //yarexgen
	testAPIs(t, re, []string{
		"\"display_name\"<sip:0312341234@10.0.0.1:5060>;user=phone;hogehoge",
		"<sip:0312341234@10.0.0.1>",
		"\"display_name\"<sip:0312341234@10.0.0.1>",
		"<sip:whois.this>;user=phone",
		"\"0333334444\"<sip:[2001:30:fe::4:123]>;user=phone",
	})
}
