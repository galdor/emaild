package smtp

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

type LineReader struct {
	Keyword string
	Code    int

	data []byte
}

func NewLineReader(data []byte) (*LineReader, error) {
	space := bytes.IndexByte(data, ' ')
	if space == -1 {
		return nil, fmt.Errorf("truncated line")
	}

	if space == 0 {
		return nil, fmt.Errorf("empty keyword")
	}

	r := LineReader{
		Keyword: string(data[:space]),

		data: data[space+1:],
	}

	isCode := true
	for _, c := range r.Keyword {
		if c < '0' || c > '9' {
			isCode = false
		}
	}

	if isCode {
		r.Code, _ = strconv.Atoi(r.Keyword)
	}

	return &r, nil
}

func (r *LineReader) Skip(n int) {
	r.data = r.data[n:]
}

func (r *LineReader) SkipString(s string) bool {
	if len(r.data) < len(s) {
		return false
	}

	if string(r.data[:len(s)]) != s {
		return false
	}

	r.Skip(len(s))

	return true
}

func (r *LineReader) SkipStringCaseInsensitive(s string) bool {
	if len(r.data) < len(s) {
		return false
	}

	if !strings.EqualFold(string(r.data[:len(s)]), s) {
		return false
	}

	r.Skip(len(s))

	return true
}

func (r *LineReader) ReadAll() []byte {
	data := r.data
	r.data = nil
	return data
}

func (r *LineReader) ReadUntilWhitespace() []byte {
	for i := 0; i < len(r.data); i++ {
		if r.data[i] == ' ' || r.data[i] == '\t' {
			s := r.data[:i]
			r.Skip(i)
			return s
		}
	}

	data := r.data
	r.data = nil
	return data
}
