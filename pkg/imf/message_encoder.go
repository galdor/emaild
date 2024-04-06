package imf

import "fmt"

type MessageEncoder struct {
	MaxLineLength int // 0 if no maximum line length

	msg *Message
}

func NewMessageEncoder(msg *Message) *MessageEncoder {
	return &MessageEncoder{
		MaxLineLength: MaxLineLength, // be strict by default

		msg: msg,
	}
}

func (e *MessageEncoder) Encode() ([]byte, error) {
	// Note that we do not alter the body to try to respect the line length
	// limit. There is no defined mechanism for this purpose, soft breaks are
	// part of MIME, so this is to be handled in a future higher level layer.

	dd := NewDataEncoder(e.MaxLineLength)

	for _, field := range e.msg.Header {
		if err := dd.WriteField(field); err != nil {
			return nil, fmt.Errorf("cannot encode field %q: %w", field.Name, err)
		}
	}

	if len(e.msg.Body) > 0 {
		dd.WriteString("\r\n")
		dd.buf.Write(e.msg.Body)
	}

	return dd.Bytes(), nil
}
