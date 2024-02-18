package imf

// RFC 5322 Internet Message Format

type Message struct {
	Header []Field
	Body   []byte
}

type Field interface {
	FieldName() string
	WriteValue(*HeaderWriter) error
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

func NewMailbox(localPart, domain string) *Mailbox {
	return &Mailbox{
		AddressSpecification: AddressSpecification{
			LocalPart: localPart,
			Domain:    domain,
		},
	}
}

func NewNamedMailbox(displayName, localPart, domain string) *Mailbox {
	mailbox := NewMailbox(localPart, domain)
	mailbox.DisplayName = displayName
	return mailbox
}
