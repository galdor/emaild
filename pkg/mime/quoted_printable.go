package mime

import (
	"bytes"
	"fmt"
)

// RFC 2045 6.7. Quoted-Printable Content-Transfer-Encoding

func QuotedPrintableEncode(data string) string {
	// Quoted-printable encoding can in theory be applied to any data. However
	// its behaviour depends on the nature of these data: RFC 2045 clearly
	// states that line breaks must be represented as CRLF sequences, but also
	// that for binary data, CR and LF must always be encoded. We only ever use
	// quoted-printable encoding for textual data, so we should never be
	// impacted.

	const maxLineLength = 76

	var buf bytes.Buffer
	lineLen := 0

	writeSoftBreak := func() {
		buf.WriteString("=\r\n")
		lineLen = 0
	}

	writeEncodedByte := func(c byte) {
		if lineLen+3 > maxLineLength {
			writeSoftBreak()
		}

		fmt.Fprintf(&buf, "=%02X", c)
		lineLen += 3
	}

	for i := 0; i < len(data); i++ {
		c := data[i]

		switch {
		case c >= 33 && c <= 60 || c >= 62 && c <= 126:
			if lineLen+1 > maxLineLength {
				writeSoftBreak()
			}
			buf.WriteByte(c)
			lineLen++

		case (c == '\t' || c == ' '):
			if i == len(data)-1 || data[i+1] == '\r' || data[i+1] == '\n' {
				writeEncodedByte(c)
			} else {
				buf.WriteByte(c)
				lineLen++
			}

		case c == '\n':
			buf.WriteString("\r\n")
			lineLen = 0

		case c == '\r' && i < len(data)-1 && data[i+1] == '\n':
			buf.WriteString("\r\n")
			i++
			lineLen = 0

		default:
			writeEncodedByte(c)
		}
	}

	return buf.String()
}
