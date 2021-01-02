package yarex_test

import (
	"testing"

	"github.com/Maki-Daisuke/go-yarex"
)

//go:generate cmd/yarexgen/yarexgen sip_pattern_compile_test.go

// This benchmark was barrowed from https://qiita.com/marnie_ms4/items/7014563083ca1d824905
var sipPattern = `^["]{0,1}([^"]*)["]{0,1}[ ]*<(sip|tel|sips):(([^@]*)@){0,1}([^>^:]*|\[[a-fA-F0-9:]*\]):{0,1}([0-9]*){0,1}>(;.*){0,1}$` //yarexgen
var sipReYa = yarex.MustCompile(sipPattern)
var testStrings = []string{"\"display_name\"<sip:0312341234@10.0.0.1:5060>;user=phone;hogehoge",
	"<sip:0312341234@10.0.0.1>",
	"\"display_name\"<sip:0312341234@10.0.0.1>",
	"<sip:whois.this>;user=phone",
	"\"0333334444\"<sip:[2001:30:fe::4:123]>;user=phone",
}

func BenchmarkSipPattern_Compiled(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, s := range testStrings {
			sipReYa.MatchString(s)
		}
	}
}
