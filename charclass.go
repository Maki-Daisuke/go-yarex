package reaot

import "unicode"

type CharClass interface {
	Contains(r rune) bool
}

type rangeTableClass unicode.RangeTable

func (rt *rangeTableClass) Contains(r rune) bool {
	return unicode.Is((*unicode.RangeTable)(rt), r)
}

type funcClass func(r rune) bool

func (fc funcClass) Contains(r rune) bool {
	return fc(r)
}

func NegateCharClass(c CharClass) CharClass {
	if rt, ok := c.(*rangeTableClass); ok {
		neg := negateRangeTable((*unicode.RangeTable)(rt))
		if neg != nil {
			return (*rangeTableClass)(neg)
		}
	}
	return funcClass(func(r rune) bool {
		return !c.Contains(r)
	})
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
	return FuncClass(func(r rune) bool {
		for _, c := range cs {
			if c.Contains(r) {
				return true
			}
		}
		return false
	})
}
