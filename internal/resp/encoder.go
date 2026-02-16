package resp

import (
	"bufio"
	"errors"
	"io"
	"strconv"
)

type Encoder struct {
	writer *bufio.Writer
}

func NewEncoder(w io.Writer) *Encoder {
	bw, ok := w.(*bufio.Writer)
	if !ok {
		bw = bufio.NewWriter(w)
	}
	return &Encoder{
		writer: bw,
	}
}

func (e *Encoder) Encode(v Value) error {
	err := e.encode(v)
	if err != nil {
		return err
	}
	return e.writer.Flush()
}

func (e *Encoder) encode(v Value) error {
	bytecode := v.Type
	var err error

	switch bytecode {
	case TypeSimpleString, TypeError, TypeInteger:
		err = e.encodeBytes(v)
	case TypeBulkString:
		err = e.encodeBytesWithCount(v)
	case TypeArray:
		err = e.encodeArray(v)
	default:
		return errors.New("Not implemented")
	}

	return err

}

func (e *Encoder) encodeBytes(v Value) error {
	if _, err := e.writer.Write([]byte{v.Type}); err != nil {
		return err
	}
	if _, err := e.writer.Write(v.Bytes); err != nil {
		return err
	}
	_, err := e.writer.Write([]byte("\r\n"))
	return err
}

func (e *Encoder) encodeBytesWithCount(v Value) error {
	if _, err := e.writer.Write([]byte{v.Type}); err != nil {
		return err
	}

	var n int
	if v.Bytes == nil {
		n = -1
	} else {
		n = len(v.Bytes)
	}
	if _, err := e.writer.Write([]byte(strconv.Itoa(n))); err != nil {
		return err
	}
	if _, err := e.writer.Write([]byte("\r\n")); err != nil {
		return err
	}

	if v.Bytes != nil {
		if _, err := e.writer.Write(v.Bytes); err != nil {
			return err
		}
		_, err := e.writer.Write([]byte("\r\n"))
		return err
	}
	return nil
}

func (e *Encoder) encodeArray(v Value) error {
	if _, err := e.writer.Write([]byte{v.Type}); err != nil {
		return err
	}

	var n int
	if v.Array == nil {
		n = -1
	} else {
		n = len(v.Array)
	}
	if _, err := e.writer.Write([]byte(strconv.Itoa(n))); err != nil {
		return err
	}
	if _, err := e.writer.Write([]byte("\r\n")); err != nil {
		return err
	}

	for i := range n {
		if err := e.encode(v.Array[i]); err != nil {
			return err
		}
	}
	return nil
}
