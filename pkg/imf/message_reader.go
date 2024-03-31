package imf

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	"github.com/galdor/emaild/pkg/utils"
)

var ErrLineTooLong = errors.New("line too long")

type MessageReader struct {
	MixedEOL      bool
	MaxLineLength int

	buf  []byte
	line []byte
	body bool
	msg  Message
}

func NewMessageReader() *MessageReader {
	// RFC 5322 2.1.1. says that the maximum line length should be 78 characters
	// (which in the context of this RFC means bytes), but in practice
	// absolutely no email server respects it, so we fall back to the mandatory
	// 998 byte limit.

	r := MessageReader{
		MixedEOL:      false,
		MaxLineLength: 998,
	}

	return &r
}

func (r *MessageReader) Read(data []byte) error {
	r.buf = append(r.buf, data...)

	if r.body {
		return nil
	}

	for {
		eol := bytes.IndexByte(r.buf, '\n')
		if eol == -1 {
			// We want to fail as soon as we exceed the maximum line length even
			// if we have not seen the end of the line yet.

			if len(r.buf) > r.MaxLineLength {
				return ErrLineTooLong
			}

			break
		}

		if !r.MixedEOL && (eol == 0 || r.buf[eol-1] != '\r') {
			return fmt.Errorf("invalid '\n' character")
		}

		// We compute the index of the EOL sequence (either \r\n, or possibly \n
		// if MixedEOL is true) so that we can check if this is an empty line.
		var eolStart int
		if eol > 0 && r.buf[eol-1] == '\r' {
			eolStart = eol - 1
		} else {
			eolStart = eol
		}

		if eolStart == 0 {
			// An empty line marks the beginning of the body

			r.body = true

			if err := r.maybeProcessLine(); err != nil {
				return err
			}

			return nil
		}

		line := r.buf[:eol+1]
		r.buf = r.buf[eol+1:]

		if len(line) > r.MaxLineLength {
			return ErrLineTooLong
		}

		if !IsWSP(line[0]) {
			// If the line does not start with a whitespace character, this is
			// not the continuation of a folded field, meaning that the previous
			// line has been entirely read.

			if err := r.maybeProcessLine(); err != nil {
				return err
			}
		}

		r.line = append(r.line, line...)
	}

	return nil
}

func (r *MessageReader) Close() (*Message, error) {
	// If there is no body, there may still be a line in the current line
	// buffer.
	if err := r.maybeProcessLine(); err != nil {
		return nil, err
	}

	if r.body && len(r.buf) > 0 {
		// TODO decode r.buf into r.msg.Body
		r.msg.Body = r.buf
		r.buf = nil
	}

	return &r.msg, nil
}

func (r *MessageReader) ReadAll(data []byte) (*Message, error) {
	if err := r.Read(data); err != nil {
		return nil, err
	}

	return r.Close()
}

func (r *MessageReader) maybeProcessLine() error {
	if len(r.line) == 0 {
		return nil
	}

	// We want to keep folded lines as they are in the raw representation of
	// each field, but the final EOL sequence does not impact parsing. We make
	// sure to support both possible EOL sequences (\n and \r\n). If MixedEOL is
	// false, the Read() method made sure the line ends with \r\n so there is no
	// point in checking again here.

	if r.line[len(r.line)-1] == '\n' {
		r.line = r.line[:len(r.line)-1]

		if r.line[len(r.line)-1] == '\r' {
			r.line = r.line[:len(r.line)-1]

		}
	}

	field := Field{
		Raw: string(r.line),
	}

	rr := NewDataReader(r.line)
	rr.MixedEOL = r.MixedEOL

	// Field name
	field.Name = string(rr.ReadWhile(IsFieldChar))
	if len(field.Name) == 0 {
		return fmt.Errorf("empty field name")
	}

	// Colon separator
	if _, err := rr.ReadFWS(); err != nil {
		return err
	}

	if !rr.SkipByte(':') {
		return fmt.Errorf("missing colon after field name %q", field.Name)
	}

	// Field value
	if _, err := rr.ReadFWS(); err != nil {
		return err
	}

	switch strings.ToLower(field.Name) {
	case "return-path":
		field.Value = &ReturnPathFieldValue{}
	case "received":
		field.Value = &ReceivedFieldValue{}
	case "resent-date":
		field.Value = &ResentDateFieldValue{}
	case "resent-from":
		field.Value = &ResentFromFieldValue{}
	case "resent-sender":
		field.Value = &ResentSenderFieldValue{}
	case "resent-to":
		field.Value = &ResentToFieldValue{}
	case "resent-cc":
		field.Value = &ResentCcFieldValue{}
	case "resent-bcc":
		field.Value = &ResentBccFieldValue{}
	case "resent-message-id":
		field.Value = &ResentMessageIdFieldValue{}
	case "date":
		field.Value = &DateFieldValue{}
	case "from":
		field.Value = &FromFieldValue{}
	case "sender":
		field.Value = &SenderFieldValue{}
	case "reply-to":
		field.Value = &ReplyToFieldValue{}
	case "to":
		field.Value = &ToFieldValue{}
	case "cc":
		field.Value = &CcFieldValue{}
	case "bcc":
		field.Value = &BccFieldValue{}
	case "message-id":
		field.Value = &MessageIdFieldValue{}
	case "in-reply-to":
		field.Value = &InReplyToFieldValue{}
	case "references":
		field.Value = &ReferencesFieldValue{}
	case "subject":
		field.Value = utils.Ref(SubjectFieldValue(""))
	case "comments":
		field.Value = utils.Ref(CommentsFieldValue(""))
	case "keywords":
		field.Value = &KeywordsFieldValue{}
	default:
		field.Value = utils.Ref(OptionalFieldValue(""))
	}

	defer func() {
		r.msg.Header = append(r.msg.Header, &field)
		r.line = nil
	}()

	if err := field.Value.Read(rr); err != nil {
		field.SetError("invalid value: %v", err)
		return nil
	}

	if _, err := rr.ReadCFWS(); err != nil {
		field.SetError("invalid trailing data: %v", err)
		return nil
	}

	if !rr.Empty() {
		field.SetError("invalid trailing data")
		return nil
	}

	return nil
}
