package parser

import (
	"bufio"
	"errors"
	"io"
	"strconv"
)

const (
	typeBulkString   byte = '$'
	typeArray        byte = '*'
	typeSimpleString byte = '+'
	typeInteger      byte = ':'
	typeError        byte = '-'
)

type Value struct {
	Type  byte
	Bytes []byte
	Array []Value
}

func Parse(r io.Reader) (Value, error) {
	br := bufio.NewReader(r)
	var err error

	bytecode, err := br.ReadByte()
	if err != nil {
		return Value{}, err
	}

	val := Value{}
	val.Type = bytecode

	switch bytecode {
	case typeBulkString:
		val.Bytes, err = parseBulkString(br)
	case typeSimpleString:
		val.Bytes, err = parseSimpleString(br)
	case typeInteger:
		val.Bytes, err = parseInteger(br)
	case typeError:
		val.Bytes, err = parseError(br)
	case typeArray:
		val.Array, err = parseArray(br)
	default:
		return Value{}, errors.New("bytecode not implemented")
	}

	return val, err
}

func parseBulkString(br *bufio.Reader) ([]byte, error) {
	b, err := br.ReadBytes('\n')
	if err != nil {
		return nil, err
	}
	if len(b) < 2 {
		return nil, errors.New("bad parse")
	}

	b = b[:len(b)-2]
	nWant, err := strconv.Atoi(string(b))
	if err != nil {
		return nil, err
	}

	if nWant == -1 {
		return nil, nil
	}

	p := make([]byte, nWant)
	_, err = io.ReadFull(br, p)
	if err != nil {
		return nil, err
	}
	code, err := br.ReadByte()
	if code != '\r' || err != nil {
		return nil, errors.New("nope")
	}
	code, err = br.ReadByte()
	if code != '\n' || err != nil {
		return nil, errors.New("nope")
	}

	return p, nil
}

func parseSimpleString(br *bufio.Reader) ([]byte, error) {
	b, err := br.ReadBytes('\n')
	if err != nil {
		return nil, err
	}
	if len(b) < 2 {
		return nil, errors.New("bad parse")
	}
	return b[:len(b)-2], nil
}

func parseInteger(br *bufio.Reader) ([]byte, error) {
	b, err := br.ReadBytes('\n')
	if err != nil {
		return nil, err
	}
	if len(b) < 2 {
		return nil, errors.New("bad parse")
	}
	return b[:len(b)-2], nil
}

func parseError(br *bufio.Reader) ([]byte, error) {
	b, err := br.ReadBytes('\n')
	if err != nil {
		return nil, err
	}
	if len(b) < 2 {
		return nil, errors.New("bad parse")
	}
	return b[:len(b)-2], nil
}

func parseArray(br *bufio.Reader) ([]Value, error) {
	return nil, nil
}
