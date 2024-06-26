package imf

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	"github.com/galdor/emaild/pkg/utils"
)

var ErrLineTooLong = errors.New("line too long")

type MessageDecoder struct {
	MixedEOL      bool
	MaxLineLength int

	buf  []byte
	line []byte
	body bool
	msg  Message
}

func NewMessageDecoder() *MessageDecoder {
	// RFC 5322 2.1.1. says that the maximum line length should be 78 characters
	// (which in the context of this RFC means bytes), but in practice
	// absolutely no email server respects it, so we fall back to the mandatory
	// 998 byte limit.

	return &MessageDecoder{
		MixedEOL:      false,
		MaxLineLength: ExtendedMaxLineLength, // be tolerant by default
	}
}

func (d *MessageDecoder) Decode(data []byte) error {
	d.buf = append(d.buf, data...)

	for {
		eol := bytes.IndexByte(d.buf, '\n')
		if eol == -1 {
			// We want to fail as soon as we exceed the maximum line length even
			// if we have not seen the end of the line yet.

			if len(d.buf) > d.MaxLineLength {
				return ErrLineTooLong
			}

			break
		}

		if !d.MixedEOL && (eol == 0 || d.buf[eol-1] != '\r') {
			return fmt.Errorf("invalid '\n' character")
		}

		line := d.buf[:eol+1]
		d.buf = d.buf[eol+1:]

		if len(line) > d.MaxLineLength {
			return ErrLineTooLong
		}

		if d.body {
			d.msg.Body = append(d.msg.Body, line...)
		} else {
			// We compute the index of the EOL sequence (either \r\n, or
			// possibly \n if MixedEOL is true) so that we can check if this is
			// an empty line.
			var eolStart int
			if len(line) >= 2 && line[len(line)-2] == '\r' {
				eolStart = len(line) - 2
			} else {
				eolStart = len(line) - 1
			}

			if eolStart == 0 {
				// An empty line marks the beginning of the body
				d.body = true

				// We may still need to process the current line if there is
				// one.
				if err := d.maybeProcessLine(); err != nil {
					return err
				}

				continue
			}

			if !IsWSP(line[0]) {
				// If the line does not start with a whitespace character, this
				// is not the continuation of a folded field, meaning that the
				// previous line has been entirely read.

				if err := d.maybeProcessLine(); err != nil {
					return err
				}
			}

			d.line = append(d.line, line...)
		}
	}

	return nil
}

func (d *MessageDecoder) Close() (*Message, error) {
	// If there is no body, there may still be a line in the current line buffer
	// since we cannot know if the current line is complete until we see the
	// next line or the end of the data (since the header field can be folded on
	// multiple lines).
	if err := d.maybeProcessLine(); err != nil {
		return nil, err
	}

	return &d.msg, nil
}

func (d *MessageDecoder) DecodeAll(data []byte) (*Message, error) {
	if err := d.Decode(data); err != nil {
		return nil, err
	}

	return d.Close()
}

func (d *MessageDecoder) maybeProcessLine() error {
	if len(d.line) == 0 {
		return nil
	}

	// We want to keep folded lines as they are in the raw representation of
	// each field, but the final EOL sequence does not impact parsing. We make
	// sure to support both possible EOL sequences (\n and \r\n). If MixedEOL is
	// false, the Decode() method made sure the line ends with \r\n so there is
	// no point in checking again here.

	if d.line[len(d.line)-1] == '\n' {
		d.line = d.line[:len(d.line)-1]

		if d.line[len(d.line)-1] == '\r' {
			d.line = d.line[:len(d.line)-1]

		}
	}

	field := Field{
		Raw: string(d.line),
	}

	dd := NewDataDecoder(d.line)
	dd.MixedEOL = d.MixedEOL

	// Field name
	field.Name = string(dd.ReadWhile(IsFieldChar))
	if len(field.Name) == 0 {
		return fmt.Errorf("empty field name")
	}

	// Colon separator
	if _, err := dd.ReadFWS(); err != nil {
		return err
	}

	if !dd.SkipByte(':') {
		return fmt.Errorf("missing colon after field name %q", field.Name)
	}

	// Field value
	if _, err := dd.ReadFWS(); err != nil {
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
		d.msg.Header = append(d.msg.Header, &field)
		d.line = nil
	}()

	if err := field.Value.Decode(dd); err != nil {
		field.SetError("invalid value: %v", err)
		return nil
	}

	if _, err := dd.ReadCFWS(); err != nil {
		field.SetError("invalid trailing data: %v", err)
		return nil
	}

	if !dd.Empty() {
		field.SetError("invalid trailing data")
		return nil
	}

	return nil
}
