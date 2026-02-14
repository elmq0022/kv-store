package parser

import (
	"bufio"
	"io"
)

type Value struct {
	Type  byte
	Bytes []byte
	Array []Value
}

func Parse(r io.Reader) Value {
	br := bufio.NewReader(r)
	_ = br
	return Value{}
}
