package parser_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/elmq0022/kv-store/parser"
	"github.com/stretchr/testify/assert"
)

func TestParser(t *testing.T) {
	tests := []struct {
		name string
		msg  []byte
		want parser.Value
	}{
		{
			name: "null bulk string",
			msg:  []byte("$-1\r\n"),
			want: parser.Value{
				Type:  '$',
				Bytes: nil,
				Array: nil,
			},
		},
		{
			name: "ping",
			msg:  []byte("*1\r\n$4\r\nping\r\n"),
			want: parser.Value{
				Type:  '*',
				Bytes: nil,
				Array: []parser.Value{
					{
						Type:  '$',
						Bytes: []byte("ping"),
					},
				},
			},
		},
		{
			name: "echo hello world",
			msg:  []byte("*2\r\n$4\r\necho\r\n$11\r\nhello world\r\n"),
			want: parser.Value{
				Type:  '*',
				Bytes: nil,
				Array: []parser.Value{
					{Type: '$', Bytes: []byte("echo")},
					{Type: '$', Bytes: []byte("hello world")},
				},
			},
		},
		{
			name: "get key",
			msg:  []byte("*2\r\n$3\r\nget\r\n$3\r\nkey\r\n"),
			want: parser.Value{
				Type:  '*',
				Bytes: nil,
				Array: []parser.Value{
					{Type: '$', Bytes: []byte("get")},
					{Type: '$', Bytes: []byte("key")},
				},
			},
		},
		{
			name: "simple string",
			msg:  []byte("+OK\r\n"),
			want: parser.Value{
				Type:  '+',
				Bytes: []byte("OK"),
				Array: nil,
			},
		},
		{
			name: "simple error message",
			msg:  []byte("-Error message\r\n"),
			want: parser.Value{
				Type:  '-',
				Bytes: []byte("Error message"),
				Array: nil,
			},
		},
		{
			name: "bulk empty string",
			msg:  []byte("$0\r\n\r\n"),
			want: parser.Value{
				Type:  '$',
				Bytes: []byte(""),
				Array: nil,
			},
		},
		{
			name: "simple string hello world",
			msg:  []byte("+hello world\r\n"),
			want: parser.Value{
				Type:  '+',
				Bytes: []byte("hello world"),
			},
		},
		{
			name: "integer",
			msg:  []byte(":42\r\n"),
			want: parser.Value{
				Type:  ':',
				Bytes: []byte("42"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parser.Parse(bytes.NewReader(tt.msg))
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParserErrors(t *testing.T) {
	tests := []struct {
		name    string
		msg     []byte
		wantErr string
	}{
		{
			name:    "unknown type prefix",
			msg:     []byte("!invalid\r\n"),
			wantErr: "unknown type prefix: !",
		},
		{
			name:    "bulk string truncated data",
			msg:     []byte("$10\r\nhello\r\n"),
			wantErr: "unexpected EOF",
		},
		{
			name:    "bulk string negative length",
			msg:     []byte("$-2\r\n"),
			wantErr: "invalid length for bulk string",
		},
		{
			name:    "array negative size",
			msg:     []byte("*-2\r\n"),
			wantErr: "invalid int for array size",
		},
		{
			name:    "missing CRLF terminator",
			msg:     []byte("+OK\n"),
			wantErr: "expected CRLF line terminator",
		},
		{
			name:    "null array",
			msg:     []byte("*-1\r\n"),
			wantErr: "",
		},
		{
			name:    "line exceeds max length",
			msg:     []byte("+" + strings.Repeat("a", 65*1024) + "\r\n"),
			wantErr: "line length exceeds maximum",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parser.Parse(bytes.NewReader(tt.msg))
			if tt.wantErr == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.wantErr)
			}
		})
	}
}
