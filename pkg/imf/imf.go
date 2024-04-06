package imf

import (
	"fmt"
	"testing"
)

// RFC 5322 Internet Message Format

const (
	MaxLineLength         = 78
	ExtendedMaxLineLength = 998
)

type Message struct {
	Header []*Field
	Body   Body // optional
}

type Body []byte

func (b Body) String() string {
	return fmt.Sprintf("#<body %dB>", len(b))
}

type Field struct {
	Raw   string
	Name  string
	Value FieldValue
	Error string
}

func (f *Field) String() string {
	if f.HasError() {
		return fmt.Sprintf("#<invalid-field %s %q>", f.Name, f.Error)
	}

	return fmt.Sprintf("#<field %s %v>", f.Name, f.Value)
}

func (f *Field) SetError(format string, args ...interface{}) {
	f.Error = fmt.Sprintf(format, args...)
}

func (f *Field) HasError() bool {
	return f.Error != ""
}

type FieldValue interface {
	Decode(*DataDecoder) error
	Encode(*DataEncoder) error

	testGenerate(*TestMessageGenerator)
	testCheck(*testing.T, *TestMessageGenerator, FieldValue)
}

type Address interface{} // Mailbox or Group

type Addresses []Address

func (addrs Addresses) String() string {
	return MustEncodeInlineData(func(e *DataEncoder) error {
		e.WriteAddressList(addrs)
		return nil
	})
}

type SpecificAddress struct {
	LocalPart string
	Domain    Domain
}

func (spec SpecificAddress) String() string {
	return MustEncodeInlineData(func(e *DataEncoder) error {
		e.WriteSpecificAddress(spec)
		return nil
	})
}

type Mailbox struct {
	SpecificAddress
	DisplayName *string
}

type Mailboxes []*Mailbox

func (mb Mailbox) String() string {
	return MustEncodeInlineData(func(e *DataEncoder) error {
		e.WriteMailbox(&mb)
		return nil
	})
}

func (mbs Mailboxes) String() string {
	return MustEncodeInlineData(func(e *DataEncoder) error {
		e.WriteMailboxList(mbs)
		return nil
	})
}

type Group struct {
	DisplayName string
	Mailboxes   []*Mailbox
}

func (g Group) String() string {
	return MustEncodeInlineData(func(e *DataEncoder) error {
		e.WriteGroup(&g)
		return nil
	})
}

type MessageId struct {
	Left  string
	Right Domain
}

type MessageIds []MessageId

func (id MessageId) String() string {
	return MustEncodeInlineData(func(e *DataEncoder) error {
		e.WriteMessageId(id)
		return nil
	})
}

func (ids MessageIds) String() string {
	return MustEncodeInlineData(func(e *DataEncoder) error {
		e.WriteMessageIdList(ids)
		return nil
	})
}

type ReceivedToken interface{} // SpecificAddress, Domain or string

type ReceivedTokens []ReceivedToken

func (ts ReceivedTokens) String() string {
	return MustEncodeInlineData(func(e *DataEncoder) error {
		e.WriteReceivedTokens(ts)
		return nil
	})
}

type Domain string

func (d Domain) String() string {
	return MustEncodeInlineData(func(e *DataEncoder) error {
		e.WriteDomain(d)
		return nil
	})
}
