package reaot

import (
	"regexp"
	"testing"
)

// This benchmark was barrowed from https://qiita.com/marnie_ms4/items/7014563083ca1d824905
var sipPattern = `^["]{0,1}([^"]*)["]{0,1}[ ]*<(sip|tel|sips):(([^@]*)@){0,1}([^>^:]*|\[[a-fA-F0-9:]*\]):{0,1}([0-9]*){0,1}>(;.*){0,1}$`
var sipReRe, _ = parse(sipPattern)
var sipReOpt = optimize(sipReRe)
var sipReOp = opCompile(sipReOpt)
var sipReStd = regexp.MustCompile(sipPattern)
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

func BenchmarkSipPattern_Rere(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, s := range testStrings {
			Match(sipReRe, s)
		}
	}
}

func BenchmarkSipPattern_Optimized(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, s := range testStrings {
			Match(sipReOpt, s)
		}
	}
}

func BenchmarkSipPattern_Optree(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, s := range testStrings {
			MatchOpTree(sipReOp, s)
		}
	}
}
