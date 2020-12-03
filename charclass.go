package reaot

import "unicode"

type CharClass interface {
	Contains(r rune) bool
}

type rangeTableClass unicode.RangeTable

func (rt *rangeTableClass) Contains(r rune) bool {
	return unicode.Is((*unicode.RangeTable)(rt), r)
}

type negClass struct{ CharClass }

func (nc negClass) Contains(r rune) bool {
	return !nc.CharClass.Contains(r)
}

type compositeClass []CharClass

func (cc compositeClass) Contains(r rune) bool {
	for _, c := range ([]CharClass)(cc) {
		if c.Contains(r) {
			return true
		}
	}
	return false
}

func NegateCharClass(c CharClass) CharClass {
	if rt, ok := c.(*rangeTableClass); ok {
		neg := negateRangeTable((*unicode.RangeTable)(rt))
		if neg != nil {
			return (*rangeTableClass)(neg)
		}
	}
	return negClass{c}
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
		if rc, ok := c.(*rangeTableClass); ok {
			merged := mergeRangeTable(acc, (*unicode.RangeTable)(rc))
			if merged != nil {
				acc = merged
				continue
			}
		}
		out = append(out, c)
	}
	if len(acc.R16) > 0 || len(acc.R32) > 0 {
		out = append(out, (*rangeTableClass)(acc))
	}
	if len(out) == 1 {
		return out[0]
	}
	return compositeClass(out)
}

// flattenCharClass recursively flatten compositeClass into a list of CharClasses
func flattenCharClass(cs ...CharClass) []CharClass {
	out := []CharClass{}
	for _, i := range cs {
		if c, ok := i.(compositeClass); ok {
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
