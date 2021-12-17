package yarex

import (
	"fmt"
	"io"
	"regexp"
	"strings"
)

var reNotWord = regexp.MustCompile(`\W`)

type charClassResult struct {
	id   string
	code *codeFragments
}

type GoGenerator struct {
	pkgname      string
	useUtf8      bool
	stateCount   uint
	idPrefix     string
	idCount      uint
	repeatCount  uint
	funcs        map[string]*codeFragments
	charClasses  map[string]charClassResult
	useCharClass bool
}

func NewGoGenerator(file string, pkg string) *GoGenerator {
	gg := &GoGenerator{}
	gg.pkgname = pkg
	gg.idPrefix = fmt.Sprintf("yarexGen_%s", reNotWord.ReplaceAllString(file, "_"))
	gg.funcs = map[string]*codeFragments{}
	gg.charClasses = map[string]charClassResult{}
	return gg
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
		code := gg.generateFunc(r, ast)
		gg.funcs[r] = code
	}
	return nil
}

func (gg *GoGenerator) WriteTo(w io.Writer) (int64, error) {
	var acc int64
	importUtf8 := ""
	if gg.useUtf8 {
		importUtf8 = `"unicode/utf8"`
	}
	n, err := fmt.Fprintf(w, `package %s

	import (
		"strconv"
		"unsafe"
		%s
		"github.com/Maki-Daisuke/go-yarex"
	)

	`, gg.pkgname, importUtf8)
	acc += int64(n)
	if err != nil {
		return acc, err
	}

	for _, cr := range gg.charClasses {
		n, err := fmt.Fprintf(w, "var %s = ", cr.id)
		acc += int64(n)
		if err != nil {
			return acc, err
		}
		m, err := cr.code.WriteTo(w)
		acc += m
		if err != nil {
			return acc, err
		}
		n, err = fmt.Fprintf(w, "\n")
		acc += int64(n)
		if err != nil {
			return acc, err
		}
	}

	for _, f := range gg.funcs {
		n, err := f.WriteTo(w)
		acc += n
		if err != nil {
			return acc, err
		}
	}

	return acc, nil
}

func (gg *GoGenerator) newState() uint {
	gg.stateCount++
	return gg.stateCount
}

func (gg *GoGenerator) newId() string {
	gg.idCount++
	return fmt.Sprintf("%s%d", gg.idPrefix, gg.idCount)
}

func (gg *GoGenerator) newRepeatID() uint {
	gg.repeatCount++
	return gg.repeatCount
}

func (gg *GoGenerator) generateFunc(re string, ast Ast) *codeFragments {
	funcID := gg.newId()
	gg.stateCount = 0
	gg.useCharClass = false
	follower := gg.generateAst(funcID, ast, &codeFragments{0, fmt.Sprintf(`
			onSuccess(ctx.Push(yarex.ContextKey{'c', 0}, p))
			return true
		default:
			// This should not happen.
			panic("state" + strconv.Itoa(state) + "is not defined")
		}
	}
}
var _ = yarex.RegisterCompiledRegexp(%q, %t, %d, %s)
	`, re, canOnlyMatchAtBegining(ast), minRequiredLengthOfAst(ast), funcID), nil})

	varDecl := ""
	if gg.useCharClass {
		varDecl = `
	var (
		r rune
		size int
	)
		`
	}

	return follower.prepend(fmt.Sprintf(`
func %s (state int, ctx yarex.MatchContext, p int, onSuccess func(yarex.MatchContext)) bool {
	%s
	str := *(*string)(unsafe.Pointer(ctx.Str))
	for{
		switch state {
		case 0:
	`, funcID, varDecl))
}

func (gg *GoGenerator) generateAst(funcID string, re Ast, follower *codeFragments) *codeFragments {
	switch r := re.(type) {
	case AstLit:
		return gg.generateLit(string(r), follower)
	case AstNotNewline:
		gg.useUtf8 = true
		return &codeFragments{follower.minReq + 1, fmt.Sprintf(`
			if len(str)-p < %d {
				return false
			}
			r, size := utf8.DecodeRuneInString(str[p:])
			if size == 0 || r == utf8.RuneError {
				return false
			}
			if r == '\n' {
				return false
			}
			p += size
		`, follower.minReq+1), follower}
	case *AstSeq:
		return gg.generateSeq(funcID, r.seq, follower)
	case *AstAlt:
		return gg.generateAlt(funcID, r.opts, follower)
	case *AstRepeat:
		return gg.generateRepeat(funcID, r.re, r.min, r.max, follower)
	case *AstCap:
		return gg.compileCapture(funcID, r.re, r.index, follower)
	case AstBackRef:
		return gg.compileBackRef(uint(r), follower)
	case AstAssertBegin:
		return follower.prepend(`
			if p != 0 {
				return false
			}
		`)
	case AstAssertEnd:
		return follower.prepend(`
			if p != len(str) {
				return false
			}
		`)
	case AstCharClass:
		return gg.generateCharClass(r.str, r.CharClass, follower)
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
		tries[i] = fmt.Sprintf(`%s(%d, ctx, p, onSuccess)`, funcID, s)
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
	switch r := re.(type) {
	case AstCharClass:
		return gg.generateRepeatCharClass(funcID, r, min, max, follower)
	}
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
			if %s(%d, ctx, p, onSuccess) {
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
			ctx2 := ctx.Push(yarex.ContextKey{'r', %d}, p)
			if %s(%d, ctx2, p, onSuccess) {
				return true
			}
			state = %d
		case %d:
		`, repeatID, followerState, repeatID, funcID, repeatState, followerState, repeatState))
	} else { // We can skip zero-width check for optimization
		follower = follower.prepend(fmt.Sprintf(`
			if %s(%d, ctx, p, onSuccess) {
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

func (gg *GoGenerator) generateRepeatCharClass(funcID string, re AstCharClass, min, max int, follower *codeFragments) *codeFragments {
	gg.generateCharClass(re.str, re.CharClass, follower) // Compile and register CharClass
	ccId := gg.charClasses[re.str].id                    // Get CharClass's identifier
	followerState := gg.newState()
	maxCond := ""
	if max >= 0 {
		maxCond = fmt.Sprintf(`n < %d && `, max)
	}
	return follower.prepend(fmt.Sprintf(`
		stack := (yarex.IntStackPool.Get().(*[]int))
		endPos := len(str) - %d
		n := 0
		for %s p < endPos {
			r, size = utf8.DecodeRuneInString(str[p:])
			if size == 0 || r == utf8.RuneError {
				break
			}
			if !%s.Contains(r) {
				break
			}
			if len(*stack) == n {
				*stack = append(*stack, p)
				*stack = (*stack)[:cap(*stack)]
			} else {
				(*stack)[n] = p
			}
			n++
			p += size
		}
		for n > %d {  // try backtrack
			if %s(%d, ctx, p, onSuccess) {
				yarex.IntStackPool.Put(stack)
				return true
			}
			n--
			p = (*stack)[n]
		}
		yarex.IntStackPool.Put(stack)
		fallthrough
	case %d:
	`, follower.minReq, maxCond, ccId, min, funcID, followerState, followerState))
}

func (gg *GoGenerator) compileCapture(funcID string, re Ast, index uint, follower *codeFragments) *codeFragments {
	follower = follower.prepend(fmt.Sprintf(`
		ctx = ctx.Push(yarex.ContextKey{'c', %d}, p)
	`, index))
	follower = gg.generateAst(funcID, re, follower)
	return follower.prepend(fmt.Sprintf(`
		ctx = ctx.Push(yarex.ContextKey{'c', %d}, p)
	`, index))
}

func (gg *GoGenerator) compileBackRef(index uint, follower *codeFragments) *codeFragments {
	return follower.prepend(fmt.Sprintf(`
		s, ok := ctx.GetCaptured(yarex.ContextKey{'c', %d})
		if !ok {  // There is no captured string with the index. So, failed matching.
			return false
		}
		l := len(s)
		if len(str)-p < l {
			return false
		}
		for i := 0; i < l; i++ {
			if str[p+i] != s[i] {
				return false
			}
		}
		p += l
	`, index))
}

func (gg *GoGenerator) generateCharClass(ptn string, c CharClass, follower *codeFragments) *codeFragments {
	var id string
	if r, ok := gg.charClasses[ptn]; ok {
		id = r.id
	} else {
		id = gg.newId()
		gg.charClasses[ptn] = charClassResult{
			id:   id,
			code: gg.generateCharClassAux(c, nil),
		}
	}
	gg.useCharClass = true
	return &codeFragments{follower.minReq + 1, fmt.Sprintf(`
		if len(str)-p < %d {
			return false
		}
		r, size = utf8.DecodeRuneInString(str[p:])
		if size == 0 || r == utf8.RuneError {
			return false
		}
		if !%s.Contains(r) {
			return false
		}
		p += size
	`, follower.minReq+1, id), follower}
}
