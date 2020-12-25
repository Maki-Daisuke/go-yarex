package yarex

import (
	"fmt"
	"strings"
)

// MustCompileOp is identical to MustCompile, but ignores compiled version of regexp
// and returns OpTree version.
func MustCompileOp(ptn string) *Regexp {
	ast, err := parse(ptn)
	if err != nil {
		panic(err)
	}
	ast = optimizeAst(ast)
	op := opCompile(ast)
	return &Regexp{ptn, opExecer{op}}
}

func DumpAst(re Ast) string {
	var buf strings.Builder
	dumpAux(re, 0, &buf)
	return buf.String()
}

func indent(n int) string {
	return strings.Repeat("  ", n)
}

func dumpAux(re Ast, n int, buf *strings.Builder) {
	switch v := re.(type) {
	case AstLit:
		fmt.Fprintf(buf, "%sLit{%q}", indent(n), string(v))
		return
	case *AstSeq:
		if len(v.seq) == 0 {
			fmt.Fprintf(buf, "%sSeq{ }", indent(n))
			return
		}
		fmt.Fprintf(buf, "%sSeq{\n", indent(n))
		for _, r := range v.seq {
			dumpAux(r, n+1, buf)
			buf.WriteRune('\n')
		}
		fmt.Fprintf(buf, "%s}", indent(n))
		return
	case *AstAlt:
		if len(v.opts) == 0 {
			fmt.Fprintf(buf, "%sAlt{ }", indent(n))
			return
		}
		fmt.Fprintf(buf, "%sAlt{\n", indent(n))
		for _, r := range v.opts {
			dumpAux(r, n+1, buf)
			buf.WriteRune('\n')
		}
		fmt.Fprintf(buf, "%s}", indent(n))
		return
	case AstNotNewline:
		fmt.Fprintf(buf, "%sNotNewLine", indent(n))
		return
	case *AstRepeat:
		fmt.Fprintf(buf, "%sRepeat(min=%d,max=%d){\n", indent(n), v.min, v.max)
		dumpAux(v.re, n+1, buf)
		fmt.Fprintf(buf, "\n%s}", indent(n))
		return
	case *AstCap:
		fmt.Fprintf(buf, "%sCapture(index=%d){\n", indent(n), v.index)
		dumpAux(v.re, n+1, buf)
		fmt.Fprintf(buf, "\n%s}", indent(n))
		return
	case AstBackRef:
		fmt.Fprintf(buf, "%sBackRef(index=%d)", indent(n), int(v))
		return
	case AstAssertBegin:
		fmt.Fprintf(buf, "%sAssertBegin", indent(n))
		return
	case AstAssertEnd:
		fmt.Fprintf(buf, "%sAssertEnd", indent(n))
		return
	case AstCharClass:
		fmt.Fprintf(buf, "%sCharClass%s", indent(n), v)
		return
	}
	panic(fmt.Errorf("IMPLEMENT DUMP for %T", re))
}
