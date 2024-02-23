package imf

import (
	"fmt"
	"time"

	"github.com/galdor/emaild/pkg/utils"
)

// RFC 5322 3.6. FieldValue Definitions

// Return-Path
type ReturnPathFieldValue AddressSpecification

func (f *ReturnPathFieldValue) Read(r *DataReader) error {
	// TODO
	return nil
}

func (f ReturnPathFieldValue) Write(w *DataWriter) error {
	w.WriteRune('<')

	if f.LocalPart == "" && f.Domain == "" {
		w.WriteAddressSpecification(AddressSpecification(f))
	}

	w.WriteRune('>')

	return nil
}

// Received
type ReceivedFieldValue struct {
	Tokens []ReceivedToken
	Date   time.Time
}

func (f *ReceivedFieldValue) Read(r *DataReader) error {
	// TODO
	return nil
}

func (f ReceivedFieldValue) Write(w *DataWriter) error {
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
			utils.Panicf("unhandled received token %#v (%T)", token, token)
		}
	}

	w.WriteRune(';')

	w.WriteDateTime(f.Date)

	return nil
}

// Resent-Date
type ResentDateFieldValue time.Time

func (f *ResentDateFieldValue) Read(r *DataReader) error {
	// TODO
	return nil
}

func (f ResentDateFieldValue) Write(w *DataWriter) error {
	w.WriteDateTime(time.Time(f))
	return nil
}

// Resent-From
type ResentFromFieldValue []*Mailbox

func (f *ResentFromFieldValue) Read(r *DataReader) error {
	// TODO
	return nil
}

func (f ResentFromFieldValue) Write(w *DataWriter) error {
	if len(f) == 0 {
		return fmt.Errorf("invalid empty mailbox list")
	}

	return w.WriteMailboxList(f)
}

// Resent-Sender
type ResentSenderFieldValue Mailbox

func (f *ResentSenderFieldValue) Read(r *DataReader) error {
	// TODO
	return nil
}

func (f ResentSenderFieldValue) Write(w *DataWriter) error {
	mailbox := Mailbox(f)
	return w.WriteMailbox(&mailbox)
}

// Resent-To
type ResentToFieldValue []Address

func (f *ResentToFieldValue) Read(r *DataReader) error {
	// TODO
	return nil
}

func (f ResentToFieldValue) Write(w *DataWriter) error {
	if len(f) == 0 {
		return fmt.Errorf("invalid address list")
	}

	return w.WriteAddressList(f)
}

// Resent-Cc
type ResentCcFieldValue []Address

func (f *ResentCcFieldValue) Read(r *DataReader) error {
	// TODO
	return nil
}

func (f ResentCcFieldValue) Write(w *DataWriter) error {
	if len(f) == 0 {
		return fmt.Errorf("invalid address list")
	}

	return w.WriteAddressList(f)
}

// Resent-Bcc
type ResentBccFieldValue []Address

func (f *ResentBccFieldValue) Read(r *DataReader) error {
	// TODO
	return nil
}

func (f ResentBccFieldValue) Write(w *DataWriter) error {
	// The Resent-Bcc field can be empty

	return w.WriteAddressList(f)
}

// Resent-Message-ID
type ResentMessageIdFieldValue MessageId

func (f *ResentMessageIdFieldValue) Read(r *DataReader) error {
	// TODO
	return nil
}

func (f ResentMessageIdFieldValue) Write(w *DataWriter) error {
	return w.WriteMessageId(MessageId(f))
}

// Date
type DateFieldValue time.Time

func (f *DateFieldValue) Read(r *DataReader) error {
	// TODO
	return nil
}

func (f DateFieldValue) Write(w *DataWriter) error {
	w.WriteDateTime(time.Time(f))
	return nil
}

// From
type FromFieldValue []*Mailbox

func (f *FromFieldValue) Read(r *DataReader) error {
	// TODO mailbox-list
	return nil
}

func (f FromFieldValue) Write(w *DataWriter) error {
	if len(f) == 0 {
		return fmt.Errorf("invalid empty mailbox list")
	}

	return w.WriteMailboxList(f)
}

// Sender
type SenderFieldValue Mailbox

func (f *SenderFieldValue) Read(r *DataReader) error {
	// TODO mailbox
	return nil
}

func (f SenderFieldValue) Write(w *DataWriter) error {
	mailbox := Mailbox(f)
	return w.WriteMailbox(&mailbox)
}

// Reply-To
type ReplyToFieldValue []Address

func (f *ReplyToFieldValue) Read(r *DataReader) error {
	// TODO address-list
	return nil
}

func (f ReplyToFieldValue) Write(w *DataWriter) error {
	if len(f) == 0 {
		return fmt.Errorf("invalid empty address list")
	}

	return w.WriteAddressList(f)
}

// To
type ToFieldValue []Address

func (f *ToFieldValue) Read(r *DataReader) error {
	// TODO address-list
	return nil
}

func (f ToFieldValue) Write(w *DataWriter) error {
	if len(f) == 0 {
		return fmt.Errorf("invalid empty address list")
	}

	return w.WriteAddressList(f)
}

// Cc
type CcFieldValue []Address

func (f *CcFieldValue) Read(r *DataReader) error {
	// TODO address-list
	return nil
}

func (f CcFieldValue) Write(w *DataWriter) error {
	if len(f) == 0 {
		return fmt.Errorf("invalid empty address list")
	}

	return w.WriteAddressList(f)
}

// Bcc
type BccFieldValue []Address

func (f *BccFieldValue) Read(r *DataReader) error {
	// TODO address-list
	return nil
}

func (f BccFieldValue) Write(w *DataWriter) error {
	// The Bcc field can be empty
	return w.WriteAddressList(f)
}

// Message-ID
type MessageIdFieldValue MessageId

func (f *MessageIdFieldValue) Read(r *DataReader) error {
	// TODO
	return nil
}

func (f MessageIdFieldValue) Write(w *DataWriter) error {
	return w.WriteMessageId(MessageId(f))
}

// In-Reply-To
type InReplyToFieldValue []MessageId

func (f *InReplyToFieldValue) Read(r *DataReader) error {
	// TODO
	return nil
}

func (f InReplyToFieldValue) Write(w *DataWriter) error {
	if len(f) == 0 {
		return fmt.Errorf("invalid empty message id list")
	}

	return w.WriteMessageIdList(f)
}

// References
type ReferencesFieldValue []MessageId

func (f *ReferencesFieldValue) Read(r *DataReader) error {
	// TODO
	return nil
}

func (f ReferencesFieldValue) Write(w *DataWriter) error {
	if len(f) == 0 {
		return fmt.Errorf("invalid empty message id list")
	}

	return w.WriteMessageIdList(f)
}

// Subject
type SubjectFieldValue string

func (f SubjectFieldValue) String() string {
	return fmt.Sprintf("%q", string(f))
}

func (f *SubjectFieldValue) Read(r *DataReader) error {
	value, err := r.ReadUnstructured()
	if err != nil {
		return err
	}

	*f = SubjectFieldValue(value)
	return nil
}

func (f SubjectFieldValue) Write(w *DataWriter) error {
	w.WriteUnstructured(string(f))
	return nil
}

// Comments
type CommentsFieldValue string

func (f CommentsFieldValue) String() string {
	return fmt.Sprintf("%q", string(f))
}

func (f *CommentsFieldValue) Read(r *DataReader) error {
	value, err := r.ReadUnstructured()
	if err != nil {
		return err
	}

	*f = CommentsFieldValue(value)
	return nil
}

func (f CommentsFieldValue) Write(w *DataWriter) error {
	w.WriteUnstructured(string(f))
	return nil
}

// Keywords
type KeywordsFieldValue []string

func (f *KeywordsFieldValue) Read(r *DataReader) error {
	// TODO phrase *("," phrase)
	//
	// With obsolete syntax, some phrases can be empty and should probably be
	// removed.
	return nil
}

func (f KeywordsFieldValue) Write(w *DataWriter) error {
	if len(f) == 0 {
		return fmt.Errorf("invalid empty phrase list")
	}

	return w.WritePhraseList(f)
}

// Optional fields
type OptionalFieldValue string

func (f OptionalFieldValue) String() string {
	return fmt.Sprintf("%q", string(f))
}

func (f *OptionalFieldValue) Read(r *DataReader) error {
	value, err := r.ReadUnstructured()
	if err != nil {
		return err
	}

	*f = OptionalFieldValue(value)
	return nil
}

func (f OptionalFieldValue) Write(w *DataWriter) error {
	w.WriteUnstructured(string(f))
	return nil
}
