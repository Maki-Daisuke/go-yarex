package yarex

func newOpAlt(left, right OpTree) *OpAlt {
	min := left.minimumReq()
	if right.minimumReq() < min {
		min = right.minimumReq()
	}
	return &OpAlt{
		OpBase: OpBase{
			minReq:   min,
			follower: left,
		},
		alt: right,
	}
}

func opCompile(re Ast) OpTree {
	return (&opCompiler{}).compile(re, OpSuccess{})
}

type opCompiler struct {
	repeatCount uint
}

func (oc *opCompiler) compile(re Ast, follower OpTree) OpTree {
	switch r := re.(type) {
	case AstLit:
		str := string(r)
		return &OpStr{
			OpBase: OpBase{
				minReq:   follower.minimumReq() + len(str),
				follower: follower,
			},
			str: str,
		}
	case *AstSeq:
		return oc.compileSeq(r.seq, follower)
	case *AstAlt:
		return oc.compileAlt(r.opts, follower)
	case AstNotNewline:
		return &OpNotNewLine{
			OpBase: OpBase{
				minReq:   follower.minimumReq() + 1,
				follower: follower,
			},
		}
	case *AstRepeat:
		return oc.compileRepeat(r.re, r.min, r.max, follower)
	case *AstCap:
		return oc.compileCapture(r.re, r.index, follower)
	case AstBackRef:
		return &OpBackRef{
			OpBase: OpBase{
				minReq:   follower.minimumReq(),
				follower: follower,
			},
			key: ContextKey{'c', uint(r)},
		}
	case AstAssertBegin:
		return &OpAssertBegin{
			OpBase: OpBase{
				minReq:   follower.minimumReq(),
				follower: follower,
			},
		}
	case AstAssertEnd:
		return &OpAssertEnd{
			OpBase: OpBase{
				minReq:   follower.minimumReq(),
				follower: follower,
			},
		}
	case AstCharClass:
		return &OpClass{
			OpBase: OpBase{
				minReq:   follower.minimumReq() + 1,
				follower: follower,
			},
			cls: (CharClass)(r),
		}
	}
	panic("EXECUTION SHOULD NOT REACH HERE")
}

func (oc *opCompiler) compileSeq(seq []Ast, follower OpTree) OpTree {
	if len(seq) == 0 {
		return follower
	}
	follower = oc.compileSeq(seq[1:], follower)
	return oc.compile(seq[0], follower)
}

func (oc *opCompiler) compileAlt(opts []Ast, follower OpTree) OpTree {
	if len(opts) == 0 {
		panic("THIS SHOULD NOT HAPPEN")
	}
	left := oc.compile(opts[0], follower)
	if len(opts) == 1 {
		return left
	}
	right := oc.compileAlt(opts[1:], follower)
	return newOpAlt(left, right)
}

func (oc *opCompiler) compileRepeat(re Ast, min, max int, follower OpTree) OpTree {
	if min > 0 {
		return oc.compile(re, oc.compileRepeat(re, min-1, max-1, follower))
	}
	if max > 0 {
		left := oc.compile(re, oc.compileRepeat(re, 0, max-1, follower))
		return newOpAlt(left, follower)
	}
	if max < 0 { // This means repeat infinitely
		if !canMatchZeroWidth(re) { // If re does not match zero-width string, we can optimize by skipping zero-width check
			self := &OpAlt{
				OpBase: OpBase{
					minReq: follower.minimumReq(),
				},
				alt: follower,
			}
			self.follower = oc.compile(re, self) // self-reference makes infinite loop
			return self
		}
		oc.repeatCount++
		self := &OpRepeat{
			OpBase: OpBase{
				minReq: follower.minimumReq(),
			},
			key: ContextKey{'r', oc.repeatCount},
			alt: follower,
		}
		self.follower = oc.compile(re, self) // self-reference makes infinite loop
		return self
	}
	// If you are here, that means min==0 && max==0
	return follower
}

func canMatchZeroWidth(re Ast) bool {
	switch r := re.(type) {
	case AstBackRef, AstAssertBegin, AstAssertEnd:
		return true
	case AstNotNewline, AstCharClass:
		return false
	case AstLit:
		return len(string(r)) == 0
	case *AstSeq:
		for _, s := range r.seq {
			if !canMatchZeroWidth(s) {
				return false
			}
		}
		return true
	case *AstAlt:
		for _, o := range r.opts {
			if canMatchZeroWidth(o) {
				return true
			}
		}
		return false
	case *AstRepeat:
		return r.min == 0 || canMatchZeroWidth(r.re)
	case *AstCap:
		return canMatchZeroWidth(r.re)
	}
	panic("EXECUTION SHOULD NOT REACH HERE")
}

func (oc *opCompiler) compileCapture(re Ast, index uint, follower OpTree) OpTree {
	follower = &OpCaptureEnd{
		OpBase: OpBase{
			minReq:   follower.minimumReq(),
			follower: follower,
		},
		key: ContextKey{'c', index},
	}
	follower = oc.compile(re, follower)
	return &OpCaptureStart{
		OpBase: OpBase{
			minReq:   follower.minimumReq(),
			follower: follower,
		},
		key: ContextKey{'c', index},
	}
}
