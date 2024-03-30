package imf

import (
	"bytes"
	"errors"
	"fmt"
	"testing"
	"time"
)

// RFC 5322 3.6. FieldValue Definitions

// Return-Path
type ReturnPathFieldValue struct {
	Address *SpecificAddress
}

func (v *ReturnPathFieldValue) String() string {
	if v.Address == nil {
		return "<nil>"
	}

	return fmt.Sprintf("%v", v.Address)
}

func (v *ReturnPathFieldValue) Read(r *DataReader) error {
	spec, err := r.ReadAngleAddress(true)
	if err != nil {
		return err
	}

	v.Address = spec
	return nil
}

func (v ReturnPathFieldValue) Write(w *DataWriter) error {
	w.WriteRune('<')

	if v.Address != nil {
		w.WriteSpecificAddress(*v.Address)
	}

	w.WriteRune('>')

	return nil
}

func (v *ReturnPathFieldValue) testGenerate(g *TestMessageGenerator) {
	g.writeString("<")

	if g.maybe(0.1) {
		v.Address = nil

		if g.maybe(0.25) {
			g.generateCFWS()
		}
	} else {
		v.Address = g.generateSpecificAddress()
	}

	g.writeString(">")
}

func (v ReturnPathFieldValue) testCheck(t *testing.T, g *TestMessageGenerator, ev FieldValue) {
	g.checkSpecificAddress(t, ev.(*ReturnPathFieldValue).Address, v.Address)
}

// Received
type ReceivedFieldValue struct {
	Tokens string // [1]
	Date   time.Time

	// [1] We do not currently parse individual tokens due to the brain damaged
	// specification indicating that each token is either a word, an angle
	// address, a specific address or a domain. Good luck differentiating those.
}

func (v *ReceivedFieldValue) String() string {
	return fmt.Sprintf("%s", v.Date.Format(time.RFC3339))
}

func (v *ReceivedFieldValue) Read(r *DataReader) error {
	// Careful, ';' can be part of one or more tokens since they can be words.
	// We look for it starting from the end of the field since date-time
	// elements cannot contain it.

	r2 := NewDataReader(r.ReadFromChar(';'))

	if _, err := r2.ReadCFWS(); err != nil {
		return err
	}

	if r2.Empty() {
		return fmt.Errorf("missing or empty datetime string")
	}

	date, err := r2.ReadDateTime()
	if err != nil {
		return fmt.Errorf("invalid datetime: %w", err)
	}

	v.Tokens = string(r.ReadAll())
	v.Date = *date

	return nil
}

func (v ReceivedFieldValue) Write(w *DataWriter) error {
	return errors.New("not implemented")
}

func (v *ReceivedFieldValue) testGenerate(g *TestMessageGenerator) {
	// TODO
	panic("not implemented")
}

func (v ReceivedFieldValue) testCheck(t *testing.T, g *TestMessageGenerator, ev FieldValue) {
	// TODO
	panic("not implemented")
}

// Resent-Date
type ResentDateFieldValue time.Time

func (v *ResentDateFieldValue) String() string {
	return fmt.Sprintf("%s", time.Time(*v).Format(time.RFC3339))
}

func (v *ResentDateFieldValue) Read(r *DataReader) error {
	date, err := r.ReadDateTime()
	if err != nil {
		return fmt.Errorf("invalid datetime: %w", err)
	}

	*v = ResentDateFieldValue(*date)
	return nil
}

func (v ResentDateFieldValue) Write(w *DataWriter) error {
	w.WriteDateTime(time.Time(v))
	return nil
}

func (v *ResentDateFieldValue) testGenerate(g *TestMessageGenerator) {
	*v = ResentDateFieldValue(g.generateDate())
}

func (v ResentDateFieldValue) testCheck(t *testing.T, g *TestMessageGenerator, ev FieldValue) {
	g.checkDate(t, time.Time(*ev.(*ResentDateFieldValue)), time.Time(v))
}

// Resent-From
//
// Support addresses and not just mailboxes (RFC 6854).
type ResentFromFieldValue Addresses

func (v ResentFromFieldValue) String() string {
	return fmt.Sprintf("%v", Addresses(v))
}

func (v *ResentFromFieldValue) Read(r *DataReader) error {
	addrs, err := r.ReadAddressList(false)
	if err != nil {
		return err
	}

	*v = ResentFromFieldValue(addrs)
	return nil
}

func (v ResentFromFieldValue) Write(w *DataWriter) error {
	if len(v) == 0 {
		return fmt.Errorf("invalid empty address list")
	}

	return w.WriteAddressList(Addresses(v))
}

func (v *ResentFromFieldValue) testGenerate(g *TestMessageGenerator) {
	*v = ResentFromFieldValue(g.generateAddresses(false))
}

func (v ResentFromFieldValue) testCheck(t *testing.T, g *TestMessageGenerator, ev FieldValue) {
	ev2 := ev.(*ResentFromFieldValue)
	g.checkAddresses(t, Addresses(*ev2), Addresses(v))
}

// Resent-Sender
//
// See RFC 6854, Resent-Sender fields can contain addresses and not just
// mailboxes.
type ResentSenderFieldValue struct {
	Address Address
}

func (v ResentSenderFieldValue) String() string {
	return fmt.Sprintf("%v", v.Address)
}

func (v *ResentSenderFieldValue) Read(r *DataReader) error {
	addr, err := r.ReadAddress()
	if err != nil {
		return err
	}

	v.Address = addr
	return nil
}

func (v ResentSenderFieldValue) Write(w *DataWriter) error {
	return w.WriteAddress(v.Address)
}

func (v *ResentSenderFieldValue) testGenerate(g *TestMessageGenerator) {
	g.generateAddress()
}

func (v ResentSenderFieldValue) testCheck(t *testing.T, g *TestMessageGenerator, ev FieldValue) {
	g.checkAddress(t, ev.(*ResentSenderFieldValue).Address, v.Address)
}

// Resent-To
type ResentToFieldValue Addresses

func (v ResentToFieldValue) String() string {
	return fmt.Sprintf("%v", Addresses(v))
}

func (v *ResentToFieldValue) Read(r *DataReader) error {
	addrs, err := r.ReadAddressList(false)
	if err != nil {
		return err
	}

	*v = ResentToFieldValue(addrs)
	return nil
}

func (v ResentToFieldValue) Write(w *DataWriter) error {
	if len(v) == 0 {
		return fmt.Errorf("invalid address list")
	}

	return w.WriteAddressList(Addresses(v))
}

func (v *ResentToFieldValue) testGenerate(g *TestMessageGenerator) {
	*v = ResentToFieldValue(g.generateAddresses(false))
}

func (v ResentToFieldValue) testCheck(t *testing.T, g *TestMessageGenerator, ev FieldValue) {
	ev2 := ev.(*ResentToFieldValue)
	g.checkAddresses(t, Addresses(*ev2), Addresses(v))
}

// Resent-Cc
type ResentCcFieldValue Addresses

func (v ResentCcFieldValue) String() string {
	return fmt.Sprintf("%v", Addresses(v))
}

func (v *ResentCcFieldValue) Read(r *DataReader) error {
	addrs, err := r.ReadAddressList(false)
	if err != nil {
		return err
	}

	*v = ResentCcFieldValue(addrs)
	return nil
}

func (v ResentCcFieldValue) Write(w *DataWriter) error {
	if len(v) == 0 {
		return fmt.Errorf("invalid address list")
	}

	return w.WriteAddressList(Addresses(v))
}

func (v *ResentCcFieldValue) testGenerate(g *TestMessageGenerator) {
	*v = ResentCcFieldValue(g.generateAddresses(false))
}

func (v ResentCcFieldValue) testCheck(t *testing.T, g *TestMessageGenerator, ev FieldValue) {
	ev2 := ev.(*ResentCcFieldValue)
	g.checkAddresses(t, Addresses(*ev2), Addresses(v))
}

// Resent-Bcc
type ResentBccFieldValue Addresses

func (v ResentBccFieldValue) String() string {
	return fmt.Sprintf("%v", Addresses(v))
}

func (v *ResentBccFieldValue) Read(r *DataReader) error {
	addrs, err := r.ReadAddressList(true)
	if err != nil {
		return err
	}

	*v = ResentBccFieldValue(Addresses(addrs))
	return nil
}

func (v ResentBccFieldValue) Write(w *DataWriter) error {
	// The Resent-Bcc field can be empty

	return w.WriteAddressList(Addresses(v))
}

func (v *ResentBccFieldValue) testGenerate(g *TestMessageGenerator) {
	*v = ResentBccFieldValue(g.generateAddresses(true))
}

func (v ResentBccFieldValue) testCheck(t *testing.T, g *TestMessageGenerator, ev FieldValue) {
	ev2 := ev.(*ResentBccFieldValue)
	g.checkAddresses(t, Addresses(*ev2), Addresses(v))
}

// Resent-Message-ID
type ResentMessageIdFieldValue MessageId

func (v ResentMessageIdFieldValue) String() string {
	return fmt.Sprintf("%v", MessageId(v))
}

func (v *ResentMessageIdFieldValue) Read(r *DataReader) error {
	id, err := r.ReadMessageId()
	if err != nil {
		return err
	}

	*v = ResentMessageIdFieldValue(*id)
	return nil
}

func (v ResentMessageIdFieldValue) Write(w *DataWriter) error {
	return w.WriteMessageId(MessageId(v))
}

func (v *ResentMessageIdFieldValue) testGenerate(g *TestMessageGenerator) {
	*v = ResentMessageIdFieldValue(g.generateMessageId())
}

func (v ResentMessageIdFieldValue) testCheck(t *testing.T, g *TestMessageGenerator, ev FieldValue) {
	ev2 := ev.(*ResentMessageIdFieldValue)
	g.checkMessageId(t, MessageId(*ev2), MessageId(v))
}

// Date
type DateFieldValue time.Time

func (v *DateFieldValue) String() string {
	return fmt.Sprintf("%s", time.Time(*v).Format(time.RFC3339))
}

func (v *DateFieldValue) Read(r *DataReader) error {
	date, err := r.ReadDateTime()
	if err != nil {
		return fmt.Errorf("invalid datetime: %w", err)
	}

	*v = DateFieldValue(*date)
	return nil
}

func (v DateFieldValue) Write(w *DataWriter) error {
	w.WriteDateTime(time.Time(v))
	return nil
}

func (v *DateFieldValue) testGenerate(g *TestMessageGenerator) {
	*v = DateFieldValue(g.generateDate())
}

func (v DateFieldValue) testCheck(t *testing.T, g *TestMessageGenerator, ev FieldValue) {
	g.checkDate(t, time.Time(*ev.(*DateFieldValue)), time.Time(v))
}

// From
//
// See RFC 6854, From fields can contain addresses and not just mailboxes.
type FromFieldValue Addresses

func (v FromFieldValue) String() string {
	return fmt.Sprintf("%v", Addresses(v))
}

func (v *FromFieldValue) Read(r *DataReader) error {
	addrs, err := r.ReadAddressList(false)
	if err != nil {
		return err
	}

	*v = FromFieldValue(addrs)
	return nil
}

func (v FromFieldValue) Write(w *DataWriter) error {
	if len(v) == 0 {
		return fmt.Errorf("invalid empty address list")
	}

	return w.WriteAddressList(Addresses(v))
}

func (v *FromFieldValue) testGenerate(g *TestMessageGenerator) {
	*v = FromFieldValue(g.generateAddresses(false))
}

func (v FromFieldValue) testCheck(t *testing.T, g *TestMessageGenerator, ev FieldValue) {
	ev2 := ev.(*FromFieldValue)
	g.checkAddresses(t, Addresses(*ev2), Addresses(v))
}

// Sender
//
// Can be an address and not just a mailbox (RFC 6854).
type SenderFieldValue struct {
	Address Address
}

func (v SenderFieldValue) String() string {
	return fmt.Sprintf("%v", v.Address)
}

func (v *SenderFieldValue) Read(r *DataReader) error {
	addr, err := r.ReadAddress()
	if err != nil {
		return err
	}

	v.Address = addr
	return nil
}

func (v SenderFieldValue) Write(w *DataWriter) error {
	return w.WriteAddress(v.Address)
}

func (v *SenderFieldValue) testGenerate(g *TestMessageGenerator) {
	g.generateAddress()
}

func (v SenderFieldValue) testCheck(t *testing.T, g *TestMessageGenerator, ev FieldValue) {
	g.checkAddress(t, ev.(*SenderFieldValue).Address, v.Address)
}

// Reply-To
type ReplyToFieldValue Addresses

func (v ReplyToFieldValue) String() string {
	return fmt.Sprintf("%v", Addresses(v))
}

func (v *ReplyToFieldValue) Read(r *DataReader) error {
	addrs, err := r.ReadAddressList(false)
	if err != nil {
		return err
	}

	*v = ReplyToFieldValue(addrs)
	return nil
}

func (v ReplyToFieldValue) Write(w *DataWriter) error {
	if len(v) == 0 {
		return fmt.Errorf("invalid empty address list")
	}

	return w.WriteAddressList(Addresses(v))
}

func (v *ReplyToFieldValue) testGenerate(g *TestMessageGenerator) {
	*v = ReplyToFieldValue(g.generateAddresses(false))
}

func (v ReplyToFieldValue) testCheck(t *testing.T, g *TestMessageGenerator, ev FieldValue) {
	ev2 := ev.(*ReplyToFieldValue)
	g.checkAddresses(t, Addresses(*ev2), Addresses(v))
}

// To
type ToFieldValue Addresses

func (v ToFieldValue) String() string {
	return fmt.Sprintf("%v", Addresses(v))
}

func (v *ToFieldValue) Read(r *DataReader) error {
	addrs, err := r.ReadAddressList(false)
	if err != nil {
		return err
	}

	*v = ToFieldValue(addrs)
	return nil
}

func (v ToFieldValue) Write(w *DataWriter) error {
	if len(v) == 0 {
		return fmt.Errorf("invalid empty address list")
	}

	return w.WriteAddressList(Addresses(v))
}

func (v *ToFieldValue) testGenerate(g *TestMessageGenerator) {
	*v = ToFieldValue(g.generateAddresses(false))
}

func (v ToFieldValue) testCheck(t *testing.T, g *TestMessageGenerator, ev FieldValue) {
	ev2 := ev.(*ToFieldValue)
	g.checkAddresses(t, Addresses(*ev2), Addresses(v))
}

// Cc
type CcFieldValue Addresses

func (v CcFieldValue) String() string {
	return fmt.Sprintf("%v", Addresses(v))
}

func (v *CcFieldValue) Read(r *DataReader) error {
	addrs, err := r.ReadAddressList(false)
	if err != nil {
		return err
	}

	*v = CcFieldValue(addrs)
	return nil
}

func (v CcFieldValue) Write(w *DataWriter) error {
	if len(v) == 0 {
		return fmt.Errorf("invalid empty address list")
	}

	return w.WriteAddressList(Addresses(v))
}

func (v *CcFieldValue) testGenerate(g *TestMessageGenerator) {
	*v = CcFieldValue(g.generateAddresses(false))
}

func (v CcFieldValue) testCheck(t *testing.T, g *TestMessageGenerator, ev FieldValue) {
	ev2 := ev.(*CcFieldValue)
	g.checkAddresses(t, Addresses(*ev2), Addresses(v))
}

// Bcc
type BccFieldValue Addresses

func (v BccFieldValue) String() string {
	return fmt.Sprintf("%v", Addresses(v))
}

func (v *BccFieldValue) Read(r *DataReader) error {
	addrs, err := r.ReadAddressList(true)
	if err != nil {
		return err
	}

	*v = BccFieldValue(addrs)
	return nil
}

func (v BccFieldValue) Write(w *DataWriter) error {
	// The Bcc field can be empty
	return w.WriteAddressList(Addresses(v))
}

func (v *BccFieldValue) testGenerate(g *TestMessageGenerator) {
	*v = BccFieldValue(g.generateAddresses(true))
}

func (v BccFieldValue) testCheck(t *testing.T, g *TestMessageGenerator, ev FieldValue) {
	ev2 := ev.(*BccFieldValue)
	g.checkAddresses(t, Addresses(*ev2), Addresses(v))
}

// Message-ID
type MessageIdFieldValue MessageId

func (v MessageIdFieldValue) String() string {
	return fmt.Sprintf("%v", MessageId(v))
}

func (v *MessageIdFieldValue) Read(r *DataReader) error {
	id, err := r.ReadMessageId()
	if err != nil {
		return err
	}

	*v = MessageIdFieldValue(*id)
	return nil
}

func (v MessageIdFieldValue) Write(w *DataWriter) error {
	return w.WriteMessageId(MessageId(v))
}

func (v *MessageIdFieldValue) testGenerate(g *TestMessageGenerator) {
	*v = MessageIdFieldValue(g.generateMessageId())
}

func (v MessageIdFieldValue) testCheck(t *testing.T, g *TestMessageGenerator, ev FieldValue) {
	ev2 := ev.(*MessageIdFieldValue)
	g.checkMessageId(t, MessageId(*ev2), MessageId(v))
}

// In-Reply-To
type InReplyToFieldValue MessageIds

func (v InReplyToFieldValue) String() string {
	return fmt.Sprintf("%v", MessageIds(v))
}

func (v *InReplyToFieldValue) Read(r *DataReader) error {
	ids, err := r.ReadMessageIdList(true)
	if err != nil {
		return err
	}

	*v = InReplyToFieldValue(ids)
	return nil
}

func (v InReplyToFieldValue) Write(w *DataWriter) error {
	if len(v) == 0 {
		return fmt.Errorf("invalid empty message id list")
	}

	return w.WriteMessageIdList(MessageIds(v))
}

func (v *InReplyToFieldValue) testGenerate(g *TestMessageGenerator) {
	*v = InReplyToFieldValue(g.generateMessageIds())
}

func (v InReplyToFieldValue) testCheck(t *testing.T, g *TestMessageGenerator, ev FieldValue) {
	ev2 := ev.(*InReplyToFieldValue)
	g.checkMessageIds(t, MessageIds(*ev2), MessageIds(v))
}

// References
type ReferencesFieldValue MessageIds

func (v ReferencesFieldValue) String() string {
	return fmt.Sprintf("%v", MessageIds(v))
}

func (v *ReferencesFieldValue) Read(r *DataReader) error {
	ids, err := r.ReadMessageIdList(true)
	if err != nil {
		return err
	}

	*v = ReferencesFieldValue(ids)
	return nil
}

func (v ReferencesFieldValue) Write(w *DataWriter) error {
	if len(v) == 0 {
		return fmt.Errorf("invalid empty message id list")
	}

	return w.WriteMessageIdList(MessageIds(v))
}

func (v *ReferencesFieldValue) testGenerate(g *TestMessageGenerator) {
	*v = ReferencesFieldValue(g.generateMessageIds())
}

func (v ReferencesFieldValue) testCheck(t *testing.T, g *TestMessageGenerator, ev FieldValue) {
	ev2 := ev.(*ReferencesFieldValue)
	g.checkMessageIds(t, MessageIds(*ev2), MessageIds(v))
}

// Subject
type SubjectFieldValue string

func (v SubjectFieldValue) String() string {
	return fmt.Sprintf("%q", string(v))
}

func (v *SubjectFieldValue) Read(r *DataReader) error {
	value, err := r.ReadUnstructured()
	if err != nil {
		return err
	}

	*v = SubjectFieldValue(value)
	return nil
}

func (v SubjectFieldValue) Write(w *DataWriter) error {
	w.WriteUnstructured(string(v))
	return nil
}

func (v *SubjectFieldValue) testGenerate(g *TestMessageGenerator) {
	*v = SubjectFieldValue(g.generateUnstructured())
}

func (v SubjectFieldValue) testCheck(t *testing.T, g *TestMessageGenerator, ev FieldValue) {
	ev2 := ev.(*SubjectFieldValue)
	g.checkUnstructured(t, string(*ev2), string(v))
}

// Comments
type CommentsFieldValue string

func (v CommentsFieldValue) String() string {
	return fmt.Sprintf("%q", string(v))
}

func (v *CommentsFieldValue) Read(r *DataReader) error {
	value, err := r.ReadUnstructured()
	if err != nil {
		return err
	}

	*v = CommentsFieldValue(value)
	return nil
}

func (v CommentsFieldValue) Write(w *DataWriter) error {
	w.WriteUnstructured(string(v))
	return nil
}

func (v *CommentsFieldValue) testGenerate(g *TestMessageGenerator) {
	*v = CommentsFieldValue(g.generateUnstructured())
}

func (v CommentsFieldValue) testCheck(t *testing.T, g *TestMessageGenerator, ev FieldValue) {
	ev2 := ev.(*CommentsFieldValue)
	g.checkUnstructured(t, string(*ev2), string(v))
}

// Keywords
type KeywordsFieldValue []string

func (v KeywordsFieldValue) String() string {
	var buf bytes.Buffer

	for i, phrase := range v {
		if i > 0 {
			buf.WriteString(", ")
		}

		fmt.Fprintf(&buf, "%q", phrase)
	}

	return buf.String()
}

func (v *KeywordsFieldValue) Read(r *DataReader) error {
	phrases, err := r.ReadPhraseList()
	if err != nil {
		return err
	}

	*v = phrases
	return nil
}

func (v KeywordsFieldValue) Write(w *DataWriter) error {
	if len(v) == 0 {
		return fmt.Errorf("invalid empty phrase list")
	}

	return w.WritePhraseList(v)
}

func (v *KeywordsFieldValue) testGenerate(g *TestMessageGenerator) {
	// TODO
	panic("not implemented")
}

func (v KeywordsFieldValue) testCheck(t *testing.T, g *TestMessageGenerator, ev FieldValue) {
	// TODO
	panic("not implemented")
}

// Optional fields
type OptionalFieldValue string

func (v OptionalFieldValue) String() string {
	return fmt.Sprintf("%q", string(v))
}

func (v *OptionalFieldValue) Read(r *DataReader) error {
	value, err := r.ReadUnstructured()
	if err != nil {
		return err
	}

	*v = OptionalFieldValue(value)
	return nil
}

func (v OptionalFieldValue) Write(w *DataWriter) error {
	w.WriteUnstructured(string(v))
	return nil
}

func (v *OptionalFieldValue) testGenerate(g *TestMessageGenerator) {
	// TODO
	panic("not implemented")
}

func (v OptionalFieldValue) testCheck(t *testing.T, g *TestMessageGenerator, ev FieldValue) {
	// TODO
	panic("not implemented")
}
