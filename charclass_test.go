package reaot

import (
	"testing"
	"unicode"
)

var classLowerAlpha = (*rangeTableClass)(&unicode.RangeTable{
	R16: []unicode.Range16{
		{'a', 'z', 1},
	},
	LatinOffset: 1,
})

var classAlpha = (*rangeTableClass)(&unicode.RangeTable{
	R16: []unicode.Range16{
		{'A', 'Z', 1},
		{'a', 'z', 1},
	},
	LatinOffset: 1,
})

var classDigit = (*rangeTableClass)(&unicode.RangeTable{
	R16: []unicode.Range16{
		{'0', '9', 1},
	},
	LatinOffset: 1,
})

func TestNegateCharClass_LowerAlpha(t *testing.T) {
	notLowerAlpha := NegateCharClass(classLowerAlpha)
	if _, ok := notLowerAlpha.(*rangeTableClass); !ok {
		t.Errorf("notLowerAlpha should be of type *rangeTableClass, but actually of type %T", notLowerAlpha)
	}
	for i := '\000'; i <= 0xFFFFF; i++ { // Test only up to 0xFFFFF due to long-running test
		if notLowerAlpha.Contains(i) != !unicode.Is((*unicode.RangeTable)(classLowerAlpha), i) {
			t.Errorf("notLowerAlpha.Contains(0x%x) should be %t, but actually not", i, !unicode.Is((*unicode.RangeTable)(classLowerAlpha), i))
		}
	}
}

func TestNegateCharClass_Alpha(t *testing.T) {
	notAlpha := NegateCharClass(classAlpha)
	if _, ok := notAlpha.(*rangeTableClass); !ok {
		t.Errorf("notAlpha should be of type *rangeTableClass, but actually of type %T", notAlpha)
	}
	for i := '\000'; i <= 0xFFFFF; i++ { // Test only up to 0xFFFFF due to long-running test
		if notAlpha.Contains(i) != !unicode.Is((*unicode.RangeTable)(classAlpha), i) {
			t.Errorf("notAlpha.Contains(0x%x) should be %t, but actually not", i, !unicode.Is((*unicode.RangeTable)(classAlpha), i))
		}
	}
}

func TestNegateCharClass_Lm(t *testing.T) {
	notLm := NegateCharClass((*rangeTableClass)(unicode.Lm))
	if _, ok := notLm.(funcClass); !ok {
		t.Errorf("notLm should be of type funcClass, but actually of type %T", notLm)
	}
	for i := '\000'; i <= 0xFFFFF; i++ { // Test only up to 0xFFFFF due to long-running test
		if notLm.Contains(i) != !unicode.Is(unicode.Lm, i) {
			t.Errorf("notLm.Contains(0x%x) should be %t, but actually not", i, !unicode.Is(unicode.Lm, i))
		}
	}
}

func TestNegateCharClass_Nd(t *testing.T) {
	// Nd should be negated by RangeTable level, because it only contains ranges with Stride = 1
	notNd := NegateCharClass((*rangeTableClass)(unicode.Nd))
	if _, ok := notNd.(*rangeTableClass); !ok {
		t.Errorf("notNd should be of type *rangeTableClass, but actually of type %T", notNd)
	}
	for i := '\000'; i <= 0xFFFFF; i++ { // Test only up to 0xFFFFF due to long-running test
		if notNd.Contains(i) != !unicode.Is(unicode.Nd, i) {
			t.Errorf("notNd.Contains(0x%x) should be %t, but actually not", i, !unicode.Is(unicode.Nd, i))
		}
	}
}

func TestMergeCharClass_AlphaNum(t *testing.T) {
	alphanum := MergeCharClass(classLowerAlpha, classAlpha, classDigit)
	if _, ok := alphanum.(*rangeTableClass); !ok {
		t.Errorf("alphanum should be of type *rangeTableClass, but actually of type %T", alphanum)
	}
	for i := '\000'; i <= 0xFFFFF; i++ { // Test only up to 0xFFFFF due to long-running test
		if alphanum.Contains(i) != ('A' <= i && i <= 'Z' || 'a' <= i && i <= 'z' || '0' <= i && i <= '9') {
			t.Errorf("alphanum.Contains(0x%x) should be %t, but actually not", i, ('A' <= i && i <= 'Z' || 'a' <= i && i <= 'z' || '0' <= i && i <= '9'))
		}
	}
}
