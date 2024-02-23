package imf

import (
	"errors"
	"fmt"

	"github.com/galdor/emaild/pkg/utils"
)

type DataReader struct {
	buf []byte
}

func NewDataReader(data []byte) *DataReader {
	r := DataReader{
		buf: data,
	}

	return &r
}

func (r *DataReader) Empty() bool {
	return len(r.buf) == 0
}

func (r *DataReader) Skip(n int) {
	if len(r.buf) < n {
		utils.Panicf("cannot skip %d bytes in a %d byte buffer", n, len(r.buf))
	}

	r.buf = r.buf[n:]
}

func (r *DataReader) StartsWithByte(c byte) bool {
	return len(r.buf) > 0 && r.buf[0] == c
}

func (r *DataReader) SkipByte(c byte) bool {
	if !r.StartsWithByte(c) {
		return false
	}

	r.Skip(1)
	return true
}

func (r *DataReader) MaybeSkipFWS() error {
	if len(r.buf) == 0 {
		return nil
	}

	if r.buf[0] != '\r' {
		return nil
	}

	if len(r.buf) < 3 {
		return fmt.Errorf("truncated fws sequence")
	}

	if r.buf[1] != '\n' {
		return fmt.Errorf("missing lf character after cr character")
	}

	if !IsWSP(r.buf[2]) {
		return fmt.Errorf("invalid fws sequence: missing whitespace after " +
			"crlf sequence")
	}

	r.buf = r.buf[3:]

	return nil
}

func (r *DataReader) SkipCFWS() error {
	for len(r.buf) > 0 {
		if IsWSP(r.buf[0]) || r.buf[0] == '\r' || r.buf[0] == '\n' {
			r.buf = r.buf[1:]
			continue
		}

		if r.buf[0] == '(' {
			// TODO read/skip the comment and continue
			return errors.New("unhandled comment")
		}

		break
	}

	return nil
}

func (r *DataReader) ReadWhile(fn func(byte) bool) []byte {
	for i := 0; i < len(r.buf); i++ {
		if !fn(r.buf[i]) {
			data := r.buf[:i]
			r.buf = r.buf[i:]
			return data
		}
	}

	data := r.buf
	r.buf = nil
	return data
}

func (r *DataReader) ReadUnstructured() (string, error) {
	// RFC 5322 3.2.5. Miscellaneous Tokens

	var value []byte

	for len(r.buf) > 0 {
		if err := r.MaybeSkipFWS(); err != nil {
			return "", err
		}

		// TODO check VCHAR
		value = append(value, r.buf[0])
		r.buf = r.buf[1:]
	}

	return string(value), nil
}
