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
	pkgname     string
	stateCount  uint
	varCount    uint
	repeatCount uint
	funcs       map[string]funcResult
}

func NewGoGenerator(pkg string) *GoGenerator {
	return &GoGenerator{pkg, 0, 0, 0, map[string]funcResult{}}
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

func (gg *GoGenerator) newRepeatID() uint {
	gg.repeatCount++
	return gg.repeatCount
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
	case AstNotNewline:
		return &codeFragments{follower.minReq + 1, fmt.Sprintf(`
			if len(str)-p < %d {
				return false
			}
			if str[p] == '\n' {
				return false
			}
			p += 1
		`, follower.minReq+1), follower}
	case *AstSeq:
		return gg.generateSeq(funcID, r.seq, follower)
	case *AstAlt:
		return gg.generateAlt(funcID, r.opts, follower)
	case *AstRepeat:
		return gg.generateRepeat(funcID, r.re, r.min, r.max, follower)
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

func (gg *GoGenerator) generateAlt(funcID string, opts []Ast, follower *codeFragments) *codeFragments {
	switch len(opts) {
	case 0:
		return follower
	case 1:
		return gg.generateSeq(funcID, opts, follower)
	}

	origMinReq := follower.minReq

	followerState := gg.newState()
	follower = follower.prepend(fmt.Sprintf(`
		fallthrough
	case %d:
	`, followerState))
	follower = gg.generateAst(funcID, opts[len(opts)-1], follower)
	minReq := follower.minReq
	stateLastOpt := gg.newState()
	follower = follower.prepend(fmt.Sprintf("case %d:\n", stateLastOpt))

	states := make([]uint, len(opts)-1)
	for i := len(opts) - 2; i >= 0; i-- {
		follower = follower.prepend(fmt.Sprintf("state = %d\n", followerState))
		follower.minReq = origMinReq
		follower = gg.generateAst(funcID, opts[i], follower)
		if follower.minReq < minReq {
			minReq = follower.minReq
		}
		s := gg.newState()
		follower = follower.prepend(fmt.Sprintf("case %d:\n", s))
		states[i] = s
	}

	tries := make([]string, len(states))
	for i, s := range states {
		tries[i] = fmt.Sprintf(`%s(%d, c, p, onSuccess)`, funcID, s)
	}
	follower = follower.prepend(fmt.Sprintf(`
		if %s {
			return true
		}
		state = %d
	`, strings.Join(tries, " || "), stateLastOpt))
	follower.minReq = minReq
	return follower
}

func (gg *GoGenerator) generateRepeat(funcID string, re Ast, min, max int, follower *codeFragments) *codeFragments {
	if min > 0 {
		return gg.generateAst(funcID, re, gg.generateRepeat(funcID, re, min-1, max-1, follower))
	}
	if max == 0 {
		return follower
	}
	if max > 0 {
		follower = gg.generateRepeat(funcID, re, 0, max-1, follower)
		followerState := gg.newState()
		follower = follower.prepend(fmt.Sprintf(`
			fallthrough
		case %d:
		`, followerState))
		minReq := follower.minReq
		follower = gg.generateAst(funcID, re, follower)
		altState := gg.newState()
		follower = follower.prepend(fmt.Sprintf(`
			if %s(%d, c, p, onSuccess) {
				return true
			}
			state = %d
		case %d:
		`, funcID, altState, followerState, altState))
		follower.minReq = minReq
		return follower
	}
	// Here, we need to compile infinite-loop regexp
	startState := gg.newState()
	repeatState := gg.newState()
	followerState := gg.newState()
	follower = follower.prepend(fmt.Sprintf(`
		state = %d
	case %d:
	`, startState, followerState))
	minReq := follower.minReq
	follower = gg.generateAst(funcID, re, follower)
	if canMatchZeroWidth(re) { // If re can matches zero-width string, we need zero-width check
		repeatID := gg.newRepeatID()
		follower = follower.prepend(fmt.Sprintf(`
			prev := ctx.FindVal(yarex.ContextKey{'r', %d})
			if prev == p { // This means zero-width matching occurs.
				state = %d // So, terminate repeating.
				continue
			}
			ctx2 := ctx.With(yarex.ContextKey{'r', %d}, p)
			if %s(%d, uintptr(unsafe.Pointer(&ctx)), p, onSuccess) {
				return true
			}
			state = %d
		case %d:
		`, repeatID, followerState, repeatID, funcID, repeatState, followerState, repeatState))
	} else { // We can skip zero-width check for optimization
		follower = follower.prepend(fmt.Sprintf(`
			if %s(%d, c, p, onSuccess) {
				return true
			}
			state = %d
		case %d:
		`, funcID, repeatState, followerState, repeatState))
	}
	follower.minReq = minReq
	return follower.prepend(fmt.Sprintf(`
		fallthrough
	case %d:
	`, startState))
}
