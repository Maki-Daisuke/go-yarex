package reaot

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

type CharClass interface {
	Contains(r rune) bool
	String() string
}

type AsciiMaskClass struct {
	Hi uint64
	Lo uint64
}

func (c AsciiMaskClass) Contains(r rune) bool {
	if r > 127 {
		return false
	}
	if r < 64 {
		return (c.Lo & (1 << r)) != 0
	}
	return (c.Hi & (1 << (r - 64))) != 0
}

func (_ AsciiMaskClass) String() string {
	// tentative implementation
	return ":AsciiMask:"
}

type CompAsciiMaskClass struct {
	AsciiMaskClass
}

func (c CompAsciiMaskClass) Contains(r rune) bool {
	if r > 127 {
		return true
	}
	if r < 64 {
		return (c.Lo & (1 << r)) == 0
	}
	return (c.Hi & (1 << (r - 64))) == 0
}

func (_ CompAsciiMaskClass) String() string {
	// tentative implementation
	return ":NegAsciiMask:"
}

// toAsciiMaskClass returns input as-is if impossible to convert to asciiMaskClass
func toAsciiMaskClass(c CharClass) CharClass {
	switch rtc := c.(type) {
	case *RangeTableClass:
		rt := (*unicode.RangeTable)(rtc)
		if len(rt.R32) != 0 {
			return c
		}
		if rt.R16[len(rt.R16)-1].Hi > 127 {
			return c
		}
		var mask AsciiMaskClass
		for _, r := range rt.R16 {
			for i := r.Lo; i <= r.Hi; i += r.Stride {
				if i > 127 {
					panic(fmt.Errorf("(THIS SHOULD NOT HAPPEN) i (=%d) exceeds ASCII range", i))
				}
				if i < 64 {
					mask.Lo |= 1 << i
				} else {
					mask.Hi |= 1 << (i - 64)
				}
			}
		}
		return mask
	}
	return c
}

type RangeTableClass unicode.RangeTable

func (rt *RangeTableClass) Contains(r rune) bool {
	return unicode.Is((*unicode.RangeTable)(rt), r)
}

func (c *RangeTableClass) HasOnlySingleChar() (rune, bool) {
	rt := (*unicode.RangeTable)(c)
	if len(rt.R16) == 1 && len(rt.R32) == 0 && rt.R16[0].Lo == rt.R16[0].Hi {
		return rune(rt.R16[0].Lo), true
	}
	if len(rt.R16) == 0 && len(rt.R32) == 1 && rt.R32[0].Lo == rt.R32[0].Hi {
		return rune(rt.R32[0].Lo), true
	}
	return 0, false
}

func (rt *RangeTableClass) String() string {
	var buf strings.Builder
	for _, r := range rt.R16 {
		if r.Lo == r.Hi {
			buf.WriteString(regexp.QuoteMeta(string([]rune{rune(r.Lo)})))
		} else {
			buf.WriteString(regexp.QuoteMeta(string([]rune{rune(r.Lo)})))
			buf.WriteString("-")
			buf.WriteString(regexp.QuoteMeta(string([]rune{rune(r.Hi)})))
		}
	}
	for _, r := range rt.R32 {
		if r.Lo == r.Hi {
			buf.WriteString(regexp.QuoteMeta(string([]rune{rune(r.Lo)})))
		} else {
			buf.WriteString(regexp.QuoteMeta(string([]rune{rune(r.Lo)})))
			buf.WriteString("-")
			buf.WriteString(regexp.QuoteMeta(string([]rune{rune(r.Hi)})))
		}
	}
	return buf.String()
}

type CompClass struct{ CharClass }

func (c CompClass) Contains(r rune) bool {
	return !c.CharClass.Contains(r)
}

func (nc CompClass) String() string {
	return "^" + nc.CharClass.String()
}

type CompositeClass []CharClass

func (cc CompositeClass) Contains(r rune) bool {
	for _, c := range ([]CharClass)(cc) {
		if c.Contains(r) {
			return true
		}
	}
	return false
}

func (cc CompositeClass) String() string {
	var buf strings.Builder
	for _, c := range ([]CharClass)(cc) {
		buf.WriteString(c.String())
	}
	return buf.String()
}

func NegateCharClass(c CharClass) CharClass {
	switch x := c.(type) {
	case *RangeTableClass:
		neg := negateRangeTable((*unicode.RangeTable)(x))
		if neg != nil {
			return (*RangeTableClass)(neg)
		}
	case AsciiMaskClass:
		return CompAsciiMaskClass{x}
	}
	return CompClass{c}
}

// negateRangeTable can only negate RageTables in which Stride = 1, and returns nil
// for those with Stride != 1
func negateRangeTable(in *unicode.RangeTable) *unicode.RangeTable {
	out := &unicode.RangeTable{R16: []unicode.Range16{}, R32: []unicode.Range32{}, LatinOffset: 0}
	// Note: Range16 and Range32 both represents characters from Low to Hi *inclusively*
	var start uint64 = 0
	for _, r := range in.R16 {
		if r.Stride != 1 {
			return nil
		}
		if start < uint64(r.Lo) {
			out.R16 = append(out.R16, unicode.Range16{uint16(start), r.Lo - 1, 1})
		}
		start = uint64(r.Hi) + 1
	}
	if start <= 0xFFFF {
		out.R16 = append(out.R16, unicode.Range16{uint16(start), 0xFFFF, 1})
	}
	for _, r := range out.R16 {
		if r.Hi <= unicode.MaxLatin1 {
			out.LatinOffset++
		}
	}
	start = 0
	for _, r := range in.R32 {
		if r.Stride != 1 {
			return nil
		}
		if start < uint64(r.Lo) {
			out.R32 = append(out.R32, unicode.Range32{uint32(start), r.Lo - 1, 1})
		}
		start = uint64(r.Hi) + 1
	}
	if start <= 0xFFFFFFFF {
		out.R32 = append(out.R32, unicode.Range32{uint32(start), 0xFFFFFFFF, 1})
	}
	return out
}

func MergeCharClass(cs ...CharClass) CharClass {
	out := []CharClass{}
	acc := &unicode.RangeTable{[]unicode.Range16{}, []unicode.Range32{}, 1}
	for _, c := range flattenCharClass(cs...) {
		// Try to merge RangeTable and fallback to composite if merge fails.
		if rc, ok := c.(*RangeTableClass); ok {
			merged := mergeRangeTable(acc, (*unicode.RangeTable)(rc))
			if merged != nil {
				acc = merged
				continue
			}
		}
		out = append(out, c)
	}
	if len(acc.R16) > 0 || len(acc.R32) > 0 {
		out = append(out, (*RangeTableClass)(acc))
	}
	if len(out) == 1 {
		return out[0]
	}
	return CompositeClass(out)
}

// flattenCharClass recursively flatten compositeClass into a list of CharClasses
func flattenCharClass(cs ...CharClass) []CharClass {
	out := []CharClass{}
	for _, i := range cs {
		if c, ok := i.(CompositeClass); ok {
			out = append(out, flattenCharClass(([]CharClass)(c)...)...)
		} else {
			out = append(out, i)
		}
	}
	return out
}

// mergeRangeTable can only merge RageTables in which Stride = 1, and returns nil
// for those with Stride != 1
func mergeRangeTable(left, right *unicode.RangeTable) *unicode.RangeTable {
	out := &unicode.RangeTable{[]unicode.Range16{}, []unicode.Range32{}, 0}
	i := 0
	j := 0
	for {
		var next unicode.Range16
		if i < len(left.R16) {
			if j < len(right.R16) {
				if left.R16[i].Lo <= right.R16[j].Lo {
					next = left.R16[i]
					i++
				} else {
					next = right.R16[j]
					j++
				}
			} else {
				next = left.R16[i]
				i++
			}
		} else if j < len(right.R16) {
			next = right.R16[j]
			j++
		} else {
			break
		}
		if len(out.R16) == 0 {
			out.R16 = append(out.R16, next)
			continue
		}
		if next.Stride != 1 || out.R16[len(out.R16)-1].Stride != 1 { // If either Stride is not 1, give up to merge.
			return nil
		}
		if next.Lo <= out.R16[len(out.R16)-1].Hi+1 { // If the next range is overlapping or adjoininig the previus one, merge them.
			if next.Hi > out.R16[len(out.R16)-1].Hi {
				out.R16[len(out.R16)-1].Hi = next.Hi
			}
		} else { // Otherwise, just append the next one.
			out.R16 = append(out.R16, next)
		}
	}
	// Recalculate LatinOffset
	out.LatinOffset = 0
	for _, r := range out.R16 {
		if r.Hi <= unicode.MaxLatin1 {
			out.LatinOffset++
		}
	}
	// Do the same for R16 also.
	i = 0
	j = 0
	for {
		var next unicode.Range32
		if i < len(left.R32) {
			if j < len(right.R32) {
				if left.R32[i].Lo <= right.R32[j].Lo {
					next = left.R32[i]
					i++
				} else {
					next = right.R32[j]
					j++
				}
			} else {
				next = left.R32[i]
				i++
			}
		} else if j < len(right.R32) {
			next = right.R32[j]
			j++
		} else {
			break
		}
		if len(out.R32) == 0 {
			out.R32 = append(out.R32, next)
			continue
		}
		if next.Stride != 1 || out.R16[len(out.R16)-1].Stride != 1 { // If either Stride is not 1, give up to merge.
			return nil
		}
		if next.Lo <= out.R32[len(out.R32)-1].Hi+1 { // If the next range is overlapping or adjoininig the previus one, merge them.
			if next.Hi > out.R32[len(out.R32)-1].Hi {
				out.R32[len(out.R32)-1].Hi = next.Hi
			}
		} else { // Otherwise, just append the next one.
			out.R32 = append(out.R32, next)
		}
	}
	return out
}

func rangeTableFromTo(lo, hi rune) *unicode.RangeTable {
	if lo > hi {
		panic(fmt.Errorf(`lo (%q) is higer than hi (%q)`, lo, hi))
	}
	out := &unicode.RangeTable{make([]unicode.Range16, 0, 1), make([]unicode.Range32, 0, 1), 0}
	if lo <= 0xFFFF {
		if hi <= 0xFFFF {
			out.R16 = append(out.R16, unicode.Range16{Lo: uint16(lo), Hi: uint16(hi), Stride: 1})
			if hi <= unicode.MaxLatin1 {
				out.LatinOffset = 1
			}
		} else {
			out.R16 = append(out.R16, unicode.Range16{Lo: uint16(lo), Hi: 0xFFFF, Stride: 1})
			out.R32 = append(out.R32, unicode.Range32{Lo: 0x10000, Hi: uint32(hi), Stride: 1})
		}
	} else {
		out.R32 = append(out.R32, unicode.Range32{Lo: uint32(lo), Hi: uint32(hi), Stride: 1})
	}
	return out
}
