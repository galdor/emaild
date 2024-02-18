package imf

import (
	"fmt"
	"time"
)

// RFC 5322 3.6. Field Definitions

// Return-Path
type ReturnPathField AddressSpecification

func (f ReturnPathField) FieldName() string {
	return "ReturnPath"
}

func (f ReturnPathField) WriteValue(w *HeaderWriter) error {
	w.WriteRune('<')

	if f.LocalPart == "" && f.Domain == "" {
		w.WriteAddressSpecification(AddressSpecification(f))
	}

	w.WriteRune('>')

	return nil
}

// Received
type ReceivedField struct {
	Tokens []ReceivedToken
	Date   time.Time
}

func (f ReceivedField) FieldName() string {
	return "Received"
}

func (f ReceivedField) WriteValue(w *HeaderWriter) error {
	for i, token := range f.Tokens {
		if i > 0 {
			w.WriteRune(' ')
		}

		switch value := token.(type) {
		case string:
			w.WriteString(value)
		case AddressSpecification:
			w.WriteAddressSpecification(value)
		default:
			panic(fmt.Sprintf("unhandled received token %#v (%T)",
				token, token))
		}
	}

	w.WriteRune(';')

	w.WriteDateTime(f.Date)

	return nil
}

// Resent-Date
type ResentDateField time.Time

func (f ResentDateField) FieldName() string {
	return "Resent-Date"
}

func (f ResentDateField) WriteValue(w *HeaderWriter) error {
	w.WriteDateTime(time.Time(f))
	return nil
}

// Resent-From
type ResentFromField []*Mailbox

func (f ResentFromField) FieldName() string {
	return "Resent-From"
}

func (f ResentFromField) WriteValue(w *HeaderWriter) error {
	if len(f) == 0 {
		return fmt.Errorf("invalid empty mailbox list")
	}

	return w.WriteMailboxList(f)
}

// Resent-Sender
type ResentSenderField Mailbox

func (f ResentSenderField) FieldName() string {
	return "Resent-Sender"
}

func (f ResentSenderField) WriteValue(w *HeaderWriter) error {
	mailbox := Mailbox(f)
	return w.WriteMailbox(&mailbox)
}

// Resent-To
type ResentToField []Address

func (f ResentToField) FieldName() string {
	return "Resent-To"
}

func (f ResentToField) WriteValue(w *HeaderWriter) error {
	if len(f) == 0 {
		return fmt.Errorf("invalid address list")
	}

	return w.WriteAddressList(f)
}

// Resent-Cc
type ResentCcField []Address

func (f ResentCcField) FieldName() string {
	return "Resent-Cc"
}

func (f ResentCcField) WriteValue(w *HeaderWriter) error {
	if len(f) == 0 {
		return fmt.Errorf("invalid address list")
	}

	return w.WriteAddressList(f)
}

// Resent-Bcc
type ResentBccField []Address

func (f ResentBccField) FieldName() string {
	return "Resent-Bcc"
}

func (f ResentBccField) WriteValue(w *HeaderWriter) error {
	// The Resent-Bcc field can be empty

	return w.WriteAddressList(f)
}

// Resent-Message-ID
type ResentMessageIdField MessageId

func (f ResentMessageIdField) FieldName() string {
	return "Resent-Message-ID"
}

func (f ResentMessageIdField) WriteValue(w *HeaderWriter) error {
	return w.WriteMessageId(MessageId(f))
}

// Date
type DateField time.Time

func (f DateField) FieldName() string {
	return "Date"
}

func (f DateField) WriteValue(w *HeaderWriter) error {
	w.WriteDateTime(time.Time(f))
	return nil
}

// From
type FromField []*Mailbox

func (f FromField) FieldName() string {
	return "From"
}

func (f FromField) WriteValue(w *HeaderWriter) error {
	if len(f) == 0 {
		return fmt.Errorf("invalid empty mailbox list")
	}

	return w.WriteMailboxList(f)
}

// Sender
type SenderField Mailbox

func (f SenderField) FieldName() string {
	return "Sender"
}

func (f SenderField) WriteValue(w *HeaderWriter) error {
	mailbox := Mailbox(f)
	return w.WriteMailbox(&mailbox)
}

// Reply-To
type ReplyToField []Address

func (f ReplyToField) FieldName() string {
	return "Reply-To"
}

func (f ReplyToField) WriteValue(w *HeaderWriter) error {
	if len(f) == 0 {
		return fmt.Errorf("invalid empty address list")
	}

	return w.WriteAddressList(f)
}

// To
type ToField []Address

func (f ToField) FieldName() string {
	return "To"
}

func (f ToField) WriteValue(w *HeaderWriter) error {
	if len(f) == 0 {
		return fmt.Errorf("invalid empty address list")
	}

	return w.WriteAddressList(f)
}

// Cc
type CcField []Address

func (f CcField) FieldName() string {
	return "Cc"
}

func (f CcField) WriteValue(w *HeaderWriter) error {
	if len(f) == 0 {
		return fmt.Errorf("invalid empty address list")
	}

	return w.WriteAddressList(f)
}

// Bcc
type BccField []Address

func (f BccField) FieldName() string {
	return "Bcc"
}

func (f BccField) WriteValue(w *HeaderWriter) error {
	// The Bcc field can be empty

	return w.WriteAddressList(f)
}

// Message-ID
type MessageIdField MessageId

func (f MessageIdField) FieldName() string {
	return "Message-ID"
}

func (f MessageIdField) WriteValue(w *HeaderWriter) error {
	return w.WriteMessageId(MessageId(f))
}

// In-Reply-To
type InReplyToField []MessageId

func (f InReplyToField) FieldName() string {
	return "In-Reply-To"
}

func (f InReplyToField) WriteValue(w *HeaderWriter) error {
	if len(f) == 0 {
		return fmt.Errorf("invalid empty message id list")
	}

	return w.WriteMessageIdList(f)
}

// References
type ReferencesField []MessageId

func (f ReferencesField) FieldName() string {
	return "References"
}

func (f ReferencesField) WriteValue(w *HeaderWriter) error {
	if len(f) == 0 {
		return fmt.Errorf("invalid empty message id list")
	}

	return w.WriteMessageIdList(f)
}

// Subject
type SubjectField string

func (f SubjectField) FieldName() string {
	return "Subject"
}

func (f SubjectField) WriteValue(w *HeaderWriter) error {
	w.WriteUnstructured(string(f))
	return nil
}

// Comments
type CommentsField string

func (f CommentsField) FieldName() string {
	return "Comments"
}

func (f CommentsField) WriteValue(w *HeaderWriter) error {
	w.WriteUnstructured(string(f))
	return nil
}

type KeywordsField []string

func (f KeywordsField) FieldName() string {
	return "Keywords"
}

func (f KeywordsField) WriteValue(w *HeaderWriter) error {
	if len(f) == 0 {
		return fmt.Errorf("invalid empty phrase list")
	}

	return w.WritePhraseList(f)
}

// Optional fields
type OptionalField struct {
	Name  string
	Value string
}

func (f OptionalField) FieldName() string {
	return f.Name
}

func (f OptionalField) WriteValue(w *HeaderWriter) error {
	w.WriteUnstructured(f.Value)
	return nil
}
