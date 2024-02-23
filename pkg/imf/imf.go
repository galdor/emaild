package imf

import "fmt"

// RFC 5322 Internet Message Format

type Message struct {
	Header []*Field
	Body   []byte // optional
}

type Field struct {
	Raw   string
	Name  string
	Value FieldValue
}

func (f *Field) String() string {
	return fmt.Sprintf("#<field %q %v>", f.Name, f.Value)
}

type FieldValue interface {
	Read(*DataReader) error
	Write(*DataWriter) error
}

type Address interface{} // Mailbox or Group

type AddressSpecification struct {
	LocalPart string
	Domain    string
}

type Mailbox struct {
	AddressSpecification
	DisplayName string // optional
}

type Group struct {
	DisplayName string
	Mailboxes   []*Mailbox
}

type MessageId struct {
	Left  string
	Right string
}

type ReceivedToken interface{} // string or AddressSpecification
