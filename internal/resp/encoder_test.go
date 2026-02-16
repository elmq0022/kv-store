package resp_test

import (
	"bytes"
	"testing"

	"github.com/elmq0022/kv-store/internal/resp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncoder(t *testing.T) {
	tests := []struct {
		name  string
		value resp.Value
		want  string
	}{
		// Simple strings
		{
			name: "simple string",
			value: resp.Value{
				Type:  '+',
				Bytes: []byte("OK"),
			},
			want: "+OK\r\n",
		},
		{
			name: "simple string hello world",
			value: resp.Value{
				Type:  '+',
				Bytes: []byte("hello world"),
			},
			want: "+hello world\r\n",
		},

		// Errors
		{
			name: "simple error message",
			value: resp.Value{
				Type:  '-',
				Bytes: []byte("Error message"),
			},
			want: "-Error message\r\n",
		},

		// Integers
		{
			name: "integer",
			value: resp.Value{
				Type:  ':',
				Bytes: []byte("42"),
			},
			want: ":42\r\n",
		},
		{
			name: "integer zero",
			value: resp.Value{
				Type:  ':',
				Bytes: []byte("0"),
			},
			want: ":0\r\n",
		},
		{
			name: "negative integer",
			value: resp.Value{
				Type:  ':',
				Bytes: []byte("-1"),
			},
			want: ":-1\r\n",
		},

		// Bulk strings
		{
			name: "bulk string",
			value: resp.Value{
				Type:  '$',
				Bytes: []byte("ping"),
			},
			want: "$4\r\nping\r\n",
		},
		{
			name: "bulk string with spaces",
			value: resp.Value{
				Type:  '$',
				Bytes: []byte("hello world"),
			},
			want: "$11\r\nhello world\r\n",
		},
		{
			name: "bulk empty string",
			value: resp.Value{
				Type:  '$',
				Bytes: []byte(""),
			},
			want: "$0\r\n\r\n",
		},
		{
			name: "null bulk string",
			value: resp.Value{
				Type:  '$',
				Bytes: nil,
			},
			want: "$-1\r\n",
		},

		// Arrays
		{
			name: "ping array",
			value: resp.Value{
				Type: '*',
				Array: []resp.Value{
					{Type: '$', Bytes: []byte("ping")},
				},
			},
			want: "*1\r\n$4\r\nping\r\n",
		},
		{
			name: "echo hello world array",
			value: resp.Value{
				Type: '*',
				Array: []resp.Value{
					{Type: '$', Bytes: []byte("echo")},
					{Type: '$', Bytes: []byte("hello world")},
				},
			},
			want: "*2\r\n$4\r\necho\r\n$11\r\nhello world\r\n",
		},
		{
			name: "get key array",
			value: resp.Value{
				Type: '*',
				Array: []resp.Value{
					{Type: '$', Bytes: []byte("get")},
					{Type: '$', Bytes: []byte("key")},
				},
			},
			want: "*2\r\n$3\r\nget\r\n$3\r\nkey\r\n",
		},
		{
			name: "null array",
			value: resp.Value{
				Type:  '*',
				Array: nil,
			},
			want: "*-1\r\n",
		},
		{
			name: "empty array",
			value: resp.Value{
				Type:  '*',
				Array: []resp.Value{},
			},
			want: "*0\r\n",
		},
		{
			name: "array with mixed types",
			value: resp.Value{
				Type: '*',
				Array: []resp.Value{
					{Type: '+', Bytes: []byte("OK")},
					{Type: '-', Bytes: []byte("ERR unknown")},
					{Type: ':', Bytes: []byte("100")},
					{Type: '$', Bytes: []byte("hello")},
					{Type: '$', Bytes: nil},
				},
			},
			want: "*5\r\n+OK\r\n-ERR unknown\r\n:100\r\n$5\r\nhello\r\n$-1\r\n",
		},
		{
			name: "nested array",
			value: resp.Value{
				Type: '*',
				Array: []resp.Value{
					{
						Type: '*',
						Array: []resp.Value{
							{Type: '$', Bytes: []byte("a")},
						},
					},
					{
						Type: '*',
						Array: []resp.Value{
							{Type: '$', Bytes: []byte("b")},
						},
					},
				},
			},
			want: "*2\r\n*1\r\n$1\r\na\r\n*1\r\n$1\r\nb\r\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			enc := resp.NewEncoder(&buf)
			err := enc.Encode(tt.value)
			require.NoError(t, err)
			assert.Equal(t, tt.want, buf.String())
		})
	}
}

func TestEncoderErrors(t *testing.T) {
	tests := []struct {
		name    string
		value   resp.Value
		wantErr string
	}{
		{
			name: "unknown type",
			value: resp.Value{
				Type:  '!',
				Bytes: []byte("bad"),
			},
			wantErr: "Not implemented",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			enc := resp.NewEncoder(&buf)
			err := enc.Encode(tt.value)
			assert.ErrorContains(t, err, tt.wantErr)
		})
	}
}

func TestEncoderParserRoundtrip(t *testing.T) {
	// Encode a Value, then parse the output, and verify we get the same Value back.
	tests := []struct {
		name  string
		value resp.Value
	}{
		{
			name:  "simple string",
			value: resp.Value{Type: '+', Bytes: []byte("OK")},
		},
		{
			name:  "error",
			value: resp.Value{Type: '-', Bytes: []byte("Error message")},
		},
		{
			name:  "integer",
			value: resp.Value{Type: ':', Bytes: []byte("42")},
		},
		{
			name:  "bulk string",
			value: resp.Value{Type: '$', Bytes: []byte("hello world")},
		},
		{
			name:  "empty bulk string",
			value: resp.Value{Type: '$', Bytes: []byte("")},
		},
		{
			name:  "null bulk string",
			value: resp.Value{Type: '$', Bytes: nil},
		},
		{
			name: "array",
			value: resp.Value{
				Type: '*',
				Array: []resp.Value{
					{Type: '$', Bytes: []byte("set")},
					{Type: '$', Bytes: []byte("key")},
					{Type: '$', Bytes: []byte("value")},
				},
			},
		},
		{
			name:  "null array",
			value: resp.Value{Type: '*', Array: nil},
		},
		{
			name:  "empty array",
			value: resp.Value{Type: '*', Array: []resp.Value{}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			enc := resp.NewEncoder(&buf)
			err := enc.Encode(tt.value)
			require.NoError(t, err)

			got, err := resp.NewDecoder(&buf).Decode()
			require.NoError(t, err)
			assert.Equal(t, tt.value, got)
		})
	}
}

func TestParseEncodeRoundtrip(t *testing.T) {
	// Parse raw RESP bytes, then re-encode, and verify the output matches the original.
	tests := []struct {
		name string
		raw  string
	}{
		{"simple string", "+OK\r\n"},
		{"error", "-Error message\r\n"},
		{"integer", ":42\r\n"},
		{"bulk string", "$4\r\nping\r\n"},
		{"empty bulk string", "$0\r\n\r\n"},
		{"null bulk string", "$-1\r\n"},
		{"ping command", "*1\r\n$4\r\nping\r\n"},
		{"echo hello world", "*2\r\n$4\r\necho\r\n$11\r\nhello world\r\n"},
		{"get key", "*2\r\n$3\r\nget\r\n$3\r\nkey\r\n"},
		{"null array", "*-1\r\n"},
		{"empty array", "*0\r\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed, err := resp.NewDecoder(bytes.NewReader([]byte(tt.raw))).Decode()
			require.NoError(t, err)

			var buf bytes.Buffer
			enc := resp.NewEncoder(&buf)
			err = enc.Encode(parsed)
			require.NoError(t, err)
			assert.Equal(t, tt.raw, buf.String())
		})
	}
}
