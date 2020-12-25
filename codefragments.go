package yarex

import (
	"io"
)

type codeFragments struct {
	minReq   int    // minimum number of characters to match this code fragmant
	code     string // fragment of Go code
	follower *codeFragments
}

func (cf *codeFragments) codeLength() int {
	if cf == nil {
		return 0
	}
	return len(cf.code) + cf.follower.codeLength()
}

func (cf *codeFragments) prepend(s string) *codeFragments {
	return &codeFragments{
		minReq:   cf.minReq,
		code:     s,
		follower: cf,
	}
}

func (cf *codeFragments) WriteTo(w io.Writer) (int64, error) {
	var acc int64
	for i := cf; i != nil; i = i.follower {
		n, err := io.WriteString(w, i.code)
		acc += int64(n)
		if err != nil {
			return acc, err
		}
	}
	return acc, nil
}
