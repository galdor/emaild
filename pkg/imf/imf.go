package imf

import (
	"fmt"
	"testing"
)

// RFC 5322 Internet Message Format

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
	Read(*DataReader) error
	Write(*DataWriter) error

	testGenerate(*TestMessageGenerator)
	testCheck(*testing.T, *TestMessageGenerator, FieldValue)
}

type Address interface{} // Mailbox or Group

type Addresses []Address

func (addrs Addresses) String() string {
	return MustWriteInlineData(func(w *DataWriter) error {
		w.WriteAddressList(addrs)
		return nil
	})
}

type SpecificAddress struct {
	LocalPart string
	Domain    Domain
}

func (spec SpecificAddress) String() string {
	return MustWriteInlineData(func(w *DataWriter) error {
		w.WriteSpecificAddress(spec)
		return nil
	})
}

type Mailbox struct {
	SpecificAddress
	DisplayName *string
}

type Mailboxes []*Mailbox

func (mb Mailbox) String() string {
	return MustWriteInlineData(func(w *DataWriter) error {
		w.WriteMailbox(&mb)
		return nil
	})
}

func (mbs Mailboxes) String() string {
	return MustWriteInlineData(func(w *DataWriter) error {
		w.WriteMailboxList(mbs)
		return nil
	})
}

type Group struct {
	DisplayName string
	Mailboxes   []*Mailbox
}

func (g Group) String() string {
	return MustWriteInlineData(func(w *DataWriter) error {
		w.WriteGroup(&g)
		return nil
	})
}

type MessageId struct {
	Left  string
	Right Domain
}

type MessageIds []MessageId

func (id MessageId) String() string {
	return MustWriteInlineData(func(w *DataWriter) error {
		w.WriteMessageId(id)
		return nil
	})
}

func (ids MessageIds) String() string {
	return MustWriteInlineData(func(w *DataWriter) error {
		w.WriteMessageIdList(ids)
		return nil
	})
}

type ReceivedToken interface{} // SpecificAddress, Domain or string

type ReceivedTokens []ReceivedToken

func (ts ReceivedTokens) String() string {
	return MustWriteInlineData(func(w *DataWriter) error {
		w.WriteReceivedTokens(ts)
		return nil
	})
}

type Domain string

func (d Domain) String() string {
	return MustWriteInlineData(func(w *DataWriter) error {
		w.WriteDomain(d)
		return nil
	})
}
