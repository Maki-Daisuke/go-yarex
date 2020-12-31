package yarex

import (
	"fmt"
	"strings"
	"unicode"
)

func (gg *GoGenerator) generateCharClassAux(cc CharClass, follower *codeFragments) *codeFragments {
	switch c := cc.(type) {
	case AsciiMaskClass:
		return gg.generateAsciiMaskClass(c, follower)
	case CompAsciiMaskClass:
		return gg.generateCompAsciiMaskClass(c, follower)
	case *RangeTableClass:
		return gg.generateRangeTableClass(c, follower)
	case CompClass:
		return gg.generateCompClass(c, follower)
	case CompositeClass:
	}
	panic(fmt.Errorf("Please implement compiler for %T", cc))
}

func (gg *GoGenerator) generateAsciiMaskClass(c AsciiMaskClass, follower *codeFragments) *codeFragments {
	return &codeFragments{1, fmt.Sprintf(`yarex.AsciiMaskClass{Hi: 0x%X, Lo: 0x%X}`, c.Hi, c.Lo), follower}
}

func (gg *GoGenerator) generateCompAsciiMaskClass(c CompAsciiMaskClass, follower *codeFragments) *codeFragments {
	return &codeFragments{1, fmt.Sprintf(`yarex.CompAsciiMaskClass{yarex.AsciiMaskClass{Hi: 0x%X, Lo: 0x%X}}`, c.AsciiMaskClass.Hi, c.AsciiMaskClass.Lo), follower}
}

func (gg *GoGenerator) generateRangeTableClass(c *RangeTableClass, follower *codeFragments) *codeFragments {
	rt := (*unicode.RangeTable)(c)
	var buf strings.Builder
	buf.WriteString("(*yarex.RangeTableClass)(*unicode.RangeTable{\n")
	if rt.R16 != nil {
		buf.WriteString("	R16: []Range16{\n")
		for _, r := range rt.R16 {
			fmt.Fprintf(&buf, "		{0x%04x, 0x%04x, %d},\n", r.Lo, r.Hi, r.Stride)
		}
		buf.WriteString("	},\n")
	}
	if rt.R32 != nil {
		buf.WriteString("	R32: []Range32{\n")
		for _, r := range rt.R32 {
			fmt.Fprintf(&buf, "		{0x%x, 0x%x, %d},\n", r.Lo, r.Hi, r.Stride)
		}
		buf.WriteString("	},\n")
	}
	if rt.LatinOffset != 0 {
		fmt.Fprintf(&buf, "		LatinOffset: %d,\n", rt.LatinOffset)
	}
	buf.WriteString("})\n")
	return &codeFragments{1, buf.String(), follower}
}

func (gg *GoGenerator) generateCompClass(c CompClass, follower *codeFragments) *codeFragments {
	return &codeFragments{
		1,
		`yarex.CompClass{`,
		gg.generateCharClassAux(c.CharClass, follower.prepend("}")),
	}
}

func (gg *GoGenerator) generateCompositeClass(c CompositeClass, follower *codeFragments) *codeFragments {
	follower = follower.prepend(")")
	cs := ([]CharClass)(c)
	follower = gg.generateCharClassAux(cs[len(cs)-1], follower)
	for _, c := range cs[0 : len(cs)-1] {
		follower = follower.prepend(", ")
		follower = gg.generateCharClassAux(c, follower)
	}
	return follower.prepend("yarex.ComopsiteClass(")
}
