package resp

import (
	"bytes"
	"testing"
)

// --- Decoder benchmarks ---

// Typical SET command: *3\r\n$3\r\nSET\r\n$3\r\nfoo\r\n$3\r\nbar\r\n
var setCmd = []byte("*3\r\n$3\r\nSET\r\n$3\r\nfoo\r\n$3\r\nbar\r\n")

// Typical GET command: *2\r\n$3\r\nGET\r\n$3\r\nfoo\r\n
var getCmd = []byte("*2\r\n$3\r\nGET\r\n$3\r\nfoo\r\n")

func BenchmarkDecodeSet(b *testing.B) {
	for b.Loop() {
		d := NewDecoder(bytes.NewReader(setCmd))
		if _, err := d.Decode(); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDecodeGet(b *testing.B) {
	for b.Loop() {
		d := NewDecoder(bytes.NewReader(getCmd))
		if _, err := d.Decode(); err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark with a larger bulk string value (1 KB)
func BenchmarkDecodeBulkString1KB(b *testing.B) {
	val := bytes.Repeat([]byte("x"), 1024)
	cmd := []byte("$1024\r\n" + string(val) + "\r\n")
	for b.Loop() {
		d := NewDecoder(bytes.NewReader(cmd))
		if _, err := d.Decode(); err != nil {
			b.Fatal(err)
		}
	}
}

// --- Encoder benchmarks ---

var okVal = Value{Type: TypeSimpleString, Bytes: []byte("OK")}

var bulkVal = Value{Type: TypeBulkString, Bytes: []byte("bar")}

var setRespArray = Value{
	Type: TypeArray,
	Array: []Value{
		{Type: TypeBulkString, Bytes: []byte("SET")},
		{Type: TypeBulkString, Bytes: []byte("foo")},
		{Type: TypeBulkString, Bytes: []byte("bar")},
	},
}

func BenchmarkEncodeSimpleString(b *testing.B) {
	var buf bytes.Buffer
	for b.Loop() {
		buf.Reset()
		e := NewEncoder(&buf)
		if err := e.Encode(okVal); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkEncodeBulkString(b *testing.B) {
	var buf bytes.Buffer
	for b.Loop() {
		buf.Reset()
		e := NewEncoder(&buf)
		if err := e.Encode(bulkVal); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkEncodeArray(b *testing.B) {
	var buf bytes.Buffer
	for b.Loop() {
		buf.Reset()
		e := NewEncoder(&buf)
		if err := e.Encode(setRespArray); err != nil {
			b.Fatal(err)
		}
	}
}

// Encode a 1 KB bulk string
func BenchmarkEncodeBulkString1KB(b *testing.B) {
	val := Value{Type: TypeBulkString, Bytes: bytes.Repeat([]byte("x"), 1024)}
	var buf bytes.Buffer
	for b.Loop() {
		buf.Reset()
		e := NewEncoder(&buf)
		if err := e.Encode(val); err != nil {
			b.Fatal(err)
		}
	}
}
