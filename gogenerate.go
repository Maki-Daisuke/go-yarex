package yarex

import (
	"fmt"
	"io"
	"strings"
)

const (
	funcName = "yarexCompiledRegex"
)

type funcResult struct {
	id       string
	headOnly bool
	minReq   int
	code     *codeFragments
}

type GoGenerator struct {
	pkgname    string
	stateCount uint
	varCount   uint
	funcs      map[string]funcResult
}

func NewGoGenerator(pkg string) *GoGenerator {
	return &GoGenerator{pkg, 0, 0, map[string]funcResult{}}
}

func (gg *GoGenerator) Add(rs ...string) error {
	for _, r := range rs {
		if _, ok := gg.funcs[r]; ok {
			continue
		}
		ast, err := parse(r)
		if err != nil {
			return err
		}
		ast = optimizeAst(ast)
		id := fmt.Sprintf("%s%d", funcName, gg.newVar())
		code := gg.generateFunc(id, ast)
		gg.funcs[r] = funcResult{id, canOnlyMatchAtBegining(ast), code.minReq, code}
	}
	return nil
}

func (gg *GoGenerator) WriteTo(w io.Writer) (int64, error) {
	var acc int64
	n, err := fmt.Fprintf(w, `package %s
	
	import (
		"fmt"
		"unsafe"
		"github.com/Maki-Daisuke/go-yarex"
	)
	`, gg.pkgname)
	acc += int64(n)
	if err != nil {
		return acc, err
	}

	for _, f := range gg.funcs {
		n, err := f.code.WriteTo(w)
		acc += n
		if err != nil {
			return acc, err
		}
	}

	n, err = io.WriteString(w, "func init(){\n")
	acc += int64(n)
	if err != nil {
		return acc, err
	}

	for r, f := range gg.funcs {
		n, err = fmt.Fprintf(w, "\tyarex.RegisterCompiledRegexp(%q, %t, %d, %s)\n", r, f.headOnly, f.minReq, f.id)
		acc += int64(n)
		if err != nil {
			return acc, err
		}
	}

	n, err = io.WriteString(w, "}\n")
	acc += int64(n)
	if err != nil {
		return acc, err
	}

	return acc, nil
}

func (gg *GoGenerator) newState() uint {
	gg.stateCount++
	return gg.stateCount
}

func (gg *GoGenerator) newVar() uint {
	gg.varCount++
	return gg.varCount
}

func (gg *GoGenerator) generateFunc(funcID string, re Ast) *codeFragments {
	cf := &codeFragments{0, fmt.Sprintf(
		`func %s (state int, c uintptr, p int, onSuccess func(*yarex.MatchContext)) bool {
			ctx := (*yarex.MatchContext)(unsafe.Pointer(c))
			str := ctx.Str
			for{
				switch state {
				case 0:
		`, funcID),
		gg.generateAst(funcID, re, &codeFragments{0, `
					c := ctx.With(yarex.ContextKey{'c', 0}, p)
					onSuccess(&c)
					return true
				default:
					// This should not happen.
					panic(fmt.Errorf("state%d is not defined", state))
				}
			}
		}
		`, nil}),
	}
	cf.minReq = cf.follower.minReq
	return cf
}

func (gg *GoGenerator) generateAst(funcID string, re Ast, follower *codeFragments) *codeFragments {
	switch r := re.(type) {
	case AstLit:
		return gg.generateLit(string(r), follower)
	case *AstSeq:
		return gg.generateSeq(funcID, r.seq, follower)
	default:
		panic(fmt.Errorf("Please implement compiler for %T", re))
	}
}

func (gg *GoGenerator) generateLit(str string, follower *codeFragments) *codeFragments {
	if len(str) == 0 {
		return follower
	}
	minReq := follower.minReq + len(str)
	var buf strings.Builder
	fmt.Fprintf(&buf, `if len(str)-p < %d {
		return false
	}
	`, minReq)
	fmt.Fprintf(&buf, "if !(str[p] == %d", str[0])
	for i := 1; i < len(str); i++ {
		fmt.Fprintf(&buf, "&& str[p+%d] == %d", i, str[i])
	}
	fmt.Fprintf(&buf, `) {
		return false
	}
	p += %d
	`, len(str))
	return &codeFragments{minReq, buf.String(), follower}
}

func (gg *GoGenerator) generateSeq(funcID string, seq []Ast, follower *codeFragments) *codeFragments {
	if len(seq) == 0 {
		return follower
	}
	follower = gg.generateSeq(funcID, seq[1:], follower)
	return gg.generateAst(funcID, seq[0], follower)
}
