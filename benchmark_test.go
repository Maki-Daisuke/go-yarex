package yarex_test

import (
	"os"
	"regexp"
	"testing"

	"github.com/Maki-Daisuke/go-yarex"
)

//go:generate cmd/yarexgen/yarexgen benchmark_test.go

// This benchmark was barrowed from https://qiita.com/tj8000rpm/items/b92d7617883639a3e714
var sipPattern = `^["]{0,1}([^"]*)["]{0,1}[ ]*<(sip|tel|sips):(([^@]*)@){0,1}([^>^:]*|\[[a-fA-F0-9:]*\]):{0,1}([0-9]*){0,1}>(;.*){0,1}$` //yarexgen
var sipReStd = regexp.MustCompile(sipPattern)
var sipReAst, _ = yarex.Parse(sipPattern)
var sipReOpt = yarex.OptimizeAst(sipReAst)
var sipReOp = yarex.MustCompileOp(sipPattern)

// Initialize sipReComp in TestMain, because it must be initialized after RegisterCompiledRegexp is called.
var sipReComp *yarex.Regexp

func TestMain(m *testing.M) {
	sipReComp = yarex.MustCompile(sipPattern)
	os.Exit(m.Run())
}

var testStrings = []string{"\"display_name\"<sip:0312341234@10.0.0.1:5060>;user=phone;hogehoge",
	"<sip:0312341234@10.0.0.1>",
	"\"display_name\"<sip:0312341234@10.0.0.1>",
	"<sip:whois.this>;user=phone",
	"\"0333334444\"<sip:[2001:30:fe::4:123]>;user=phone",
}

func BenchmarkSipPattern_Standard(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, s := range testStrings {
			sipReStd.MatchString(s)
		}
	}
}

func BenchmarkSipPattern_Ast(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, s := range testStrings {
			yarex.AstMatch(sipReAst, s)
		}
	}
}

func BenchmarkSipPattern_Optimized(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, s := range testStrings {
			yarex.AstMatch(sipReOpt, s)
		}
	}
}

func BenchmarkSipPattern_Optree(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, s := range testStrings {
			sipReOp.MatchString(s)
		}
	}
}

func BenchmarkSipPattern_Compiled(b *testing.B) {
	if !yarex.IsCompiledMatcher(sipReComp) {
		panic("Not compiled!!!!!")
	}
	for i := 0; i < b.N; i++ {
		for _, s := range testStrings {
			sipReComp.MatchString(s)
		}
	}
}
