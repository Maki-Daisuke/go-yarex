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
	return opCompileAux(re, OpSuccess{})
}

func opCompileAux(re Regexp, follower OpTree) OpTree {
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
		return opCompileSeq(r.seq, follower)
	case *ReAlt:
		return opCompileAlt(r.opts, follower)
	case ReNotNewline:
		return &OpNotNewLine{
			OpBase: OpBase{
				minReq:   follower.minimumReq() + 1,
				follower: follower,
			},
		}
	case *ReRepeat:
		return opCompileRepeat(r.re, r.min, r.max, follower)
	case *ReCap:
		return opCompileCapture(r.re, r.index, follower)
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
	panic("TODO")
}

func opCompileSeq(seq []Regexp, follower OpTree) OpTree {
	if len(seq) == 0 {
		return follower
	}
	follower = opCompileSeq(seq[1:], follower)
	return opCompileAux(seq[0], follower)
}

func opCompileAlt(opts []Regexp, follower OpTree) OpTree {
	if len(opts) == 0 {
		panic("THIS SHOULD NOT HAPPEN")
	}
	left := opCompileAux(opts[0], follower)
	if len(opts) == 1 {
		return left
	}
	right := opCompileAlt(opts[1:], follower)
	return newOpAlt(left, right)
}

func opCompileRepeat(re Regexp, min, max int, follower OpTree) OpTree {
	if min > 0 {
		follower = opCompileRepeat(re, min-1, max-1, follower)
		return opCompileAux(re, follower)
	}
	if max < 0 { // This means repeat infinitely
		self := newOpAlt(follower, follower)
		self.follower = opCompileAux(re, self) // self-reference makes infinite loop
		return self
	}
	if max > 0 {
		left := opCompileAux(re, opCompileRepeat(re, 0, max-1, follower))
		return newOpAlt(left, follower)
	}
	// If you are here, that means min==0 && max==0
	return follower
}

func opCompileCapture(re Regexp, index uint, follower OpTree) OpTree {
	follower = &OpCaptureEnd{
		OpBase: OpBase{
			minReq:   follower.minimumReq(),
			follower: follower,
		},
		index: index,
	}
	follower = opCompileAux(re, follower)
	return &OpCaptureStart{
		OpBase: OpBase{
			minReq:   follower.minimumReq(),
			follower: follower,
		},
		index: index,
	}
}
