package reaot

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

func opCompile(re Regexp) OpTree {
	return (&opCompiler{}).compile(re, OpSuccess{})
}

type opCompiler struct {
	repeatCount uint
}

func (oc *opCompiler) compile(re Regexp, follower OpTree) OpTree {
	switch r := re.(type) {
	case ReLit:
		str := string(r)
		return &OpStr{
			OpBase: OpBase{
				minReq:   follower.minimumReq() + len(str),
				follower: follower,
			},
			str: str,
		}
	case *ReSeq:
		return oc.compileSeq(r.seq, follower)
	case *ReAlt:
		return oc.compileAlt(r.opts, follower)
	case ReNotNewline:
		return &OpNotNewLine{
			OpBase: OpBase{
				minReq:   follower.minimumReq() + 1,
				follower: follower,
			},
		}
	case *ReRepeat:
		return oc.compileRepeat(r.re, r.min, r.max, follower)
	case *ReCap:
		return oc.compileCapture(r.re, r.index, follower)
	case ReBackRef:
		return &OpBackRef{
			OpBase: OpBase{
				minReq:   follower.minimumReq(),
				follower: follower,
			},
			index: uint(r),
		}
	case ReAssertBegin:
		return &OpAssertBegin{
			OpBase: OpBase{
				minReq:   follower.minimumReq(),
				follower: follower,
			},
		}
	case ReAssertEnd:
		return &OpAssertEnd{
			OpBase: OpBase{
				minReq:   follower.minimumReq(),
				follower: follower,
			},
		}
	case ReCharClass:
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

func (oc *opCompiler) compileSeq(seq []Regexp, follower OpTree) OpTree {
	if len(seq) == 0 {
		return follower
	}
	follower = oc.compileSeq(seq[1:], follower)
	return oc.compile(seq[0], follower)
}

func (oc *opCompiler) compileAlt(opts []Regexp, follower OpTree) OpTree {
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

func (oc *opCompiler) compileRepeat(re Regexp, min, max int, follower OpTree) OpTree {
	if min > 0 {
		return oc.compile(re, oc.compileRepeat(re, min-1, max-1, follower))
	}
	if max > 0 {
		left := oc.compile(re, oc.compileRepeat(re, 0, max-1, follower))
		return newOpAlt(left, follower)
	}
	if max < 0 { // This means repeat infinitely
		oc.repeatCount++
		self := &OpRepeat{
			OpBase: OpBase{
				minReq: follower.minimumReq(),
			},
			index: oc.repeatCount,
			alt:   follower,
		}
		self.follower = oc.compile(re, self) // self-reference makes infinite loop
		return self
	}
	// If you are here, that means min==0 && max==0
	return follower
}

func (oc *opCompiler) compileCapture(re Regexp, index uint, follower OpTree) OpTree {
	follower = &OpCaptureEnd{
		OpBase: OpBase{
			minReq:   follower.minimumReq(),
			follower: follower,
		},
		index: index,
	}
	follower = oc.compile(re, follower)
	return &OpCaptureStart{
		OpBase: OpBase{
			minReq:   follower.minimumReq(),
			follower: follower,
		},
		index: index,
	}
}
