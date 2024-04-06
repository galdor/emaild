package imf

import (
	"bytes"
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

func (v *ReturnPathFieldValue) Decode(d *DataDecoder) error {
	spec, err := d.ReadAngleAddress(true)
	if err != nil {
		return err
	}

	v.Address = spec
	return nil
}

func (v ReturnPathFieldValue) Encode(e *DataEncoder) error {
	e.WriteRune('<')

	if v.Address != nil {
		e.WriteSpecificAddress(*v.Address)
	}

	e.WriteRune('>')

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
	Tokens ReceivedTokens
	Date   time.Time
}

func (v *ReceivedFieldValue) String() string {
	return fmt.Sprintf("%v %s", v.Tokens, v.Date.Format(time.RFC3339))
}

func (v *ReceivedFieldValue) Decode(d *DataDecoder) error {
	tokens, err := d.ReadReceivedTokens()
	if err != nil {
		return err
	}

	if _, err := d.ReadCFWS(); err != nil {
		return err
	}

	if !d.SkipByte(';') {
		return fmt.Errorf("missing ';' character after tokens")
	}

	date, err := d.ReadDateTime()
	if err != nil {
		return fmt.Errorf("invalid datetime: %w", err)
	}

	v.Tokens = tokens
	v.Date = *date

	return nil
}

func (v ReceivedFieldValue) Encode(e *DataEncoder) error {
	if err := e.WriteReceivedTokens(v.Tokens); err != nil {
		return err
	}
	e.WriteString("; ")
	e.WriteDateTime(v.Date)
	return nil
}

func (v *ReceivedFieldValue) testGenerate(g *TestMessageGenerator) {
	// Received tokens are irrelevant because we do not parse them

	g.writeByte(';')

	if g.maybe(0.1) {
		g.generateCFWS()
	}

	v.Date = g.generateDate()
}

func (v ReceivedFieldValue) testCheck(t *testing.T, g *TestMessageGenerator, ev FieldValue) {
	g.checkDate(t, ev.(*ReceivedFieldValue).Date, v.Date)
}

// Resent-Date
type ResentDateFieldValue time.Time

func (v *ResentDateFieldValue) String() string {
	return fmt.Sprintf("%s", time.Time(*v).Format(time.RFC3339))
}

func (v *ResentDateFieldValue) Decode(d *DataDecoder) error {
	date, err := d.ReadDateTime()
	if err != nil {
		return fmt.Errorf("invalid datetime: %w", err)
	}

	*v = ResentDateFieldValue(*date)
	return nil
}

func (v ResentDateFieldValue) Encode(e *DataEncoder) error {
	e.WriteDateTime(time.Time(v))
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

func (v *ResentFromFieldValue) Decode(d *DataDecoder) error {
	addrs, err := d.ReadAddressList(false)
	if err != nil {
		return err
	}

	*v = ResentFromFieldValue(addrs)
	return nil
}

func (v ResentFromFieldValue) Encode(e *DataEncoder) error {
	if len(v) == 0 {
		return fmt.Errorf("invalid empty address list")
	}

	return e.WriteAddressList(Addresses(v))
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

func (v *ResentSenderFieldValue) Decode(d *DataDecoder) error {
	addr, err := d.ReadAddress()
	if err != nil {
		return err
	}

	v.Address = addr
	return nil
}

func (v ResentSenderFieldValue) Encode(e *DataEncoder) error {
	return e.WriteAddress(v.Address)
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

func (v *ResentToFieldValue) Decode(d *DataDecoder) error {
	addrs, err := d.ReadAddressList(false)
	if err != nil {
		return err
	}

	*v = ResentToFieldValue(addrs)
	return nil
}

func (v ResentToFieldValue) Encode(e *DataEncoder) error {
	if len(v) == 0 {
		return fmt.Errorf("invalid address list")
	}

	return e.WriteAddressList(Addresses(v))
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

func (v *ResentCcFieldValue) Decode(d *DataDecoder) error {
	addrs, err := d.ReadAddressList(false)
	if err != nil {
		return err
	}

	*v = ResentCcFieldValue(addrs)
	return nil
}

func (v ResentCcFieldValue) Encode(e *DataEncoder) error {
	if len(v) == 0 {
		return fmt.Errorf("invalid address list")
	}

	return e.WriteAddressList(Addresses(v))
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

func (v *ResentBccFieldValue) Decode(d *DataDecoder) error {
	addrs, err := d.ReadAddressList(true)
	if err != nil {
		return err
	}

	*v = ResentBccFieldValue(Addresses(addrs))
	return nil
}

func (v ResentBccFieldValue) Encode(e *DataEncoder) error {
	// The Resent-Bcc field can be empty

	return e.WriteAddressList(Addresses(v))
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

func (v *ResentMessageIdFieldValue) Decode(d *DataDecoder) error {
	id, err := d.ReadMessageId()
	if err != nil {
		return err
	}

	*v = ResentMessageIdFieldValue(*id)
	return nil
}

func (v ResentMessageIdFieldValue) Encode(e *DataEncoder) error {
	return e.WriteMessageId(MessageId(v))
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

func (v *DateFieldValue) Decode(d *DataDecoder) error {
	date, err := d.ReadDateTime()
	if err != nil {
		return fmt.Errorf("invalid datetime: %w", err)
	}

	*v = DateFieldValue(*date)
	return nil
}

func (v DateFieldValue) Encode(e *DataEncoder) error {
	e.WriteDateTime(time.Time(v))
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

func (v *FromFieldValue) Decode(d *DataDecoder) error {
	addrs, err := d.ReadAddressList(false)
	if err != nil {
		return err
	}

	*v = FromFieldValue(addrs)
	return nil
}

func (v FromFieldValue) Encode(e *DataEncoder) error {
	if len(v) == 0 {
		return fmt.Errorf("invalid empty address list")
	}

	return e.WriteAddressList(Addresses(v))
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

func (v *SenderFieldValue) Decode(d *DataDecoder) error {
	addr, err := d.ReadAddress()
	if err != nil {
		return err
	}

	v.Address = addr
	return nil
}

func (v SenderFieldValue) Encode(e *DataEncoder) error {
	return e.WriteAddress(v.Address)
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

func (v *ReplyToFieldValue) Decode(d *DataDecoder) error {
	addrs, err := d.ReadAddressList(false)
	if err != nil {
		return err
	}

	*v = ReplyToFieldValue(addrs)
	return nil
}

func (v ReplyToFieldValue) Encode(e *DataEncoder) error {
	if len(v) == 0 {
		return fmt.Errorf("invalid empty address list")
	}

	return e.WriteAddressList(Addresses(v))
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

func (v *ToFieldValue) Decode(d *DataDecoder) error {
	addrs, err := d.ReadAddressList(false)
	if err != nil {
		return err
	}

	*v = ToFieldValue(addrs)
	return nil
}

func (v ToFieldValue) Encode(e *DataEncoder) error {
	if len(v) == 0 {
		return fmt.Errorf("invalid empty address list")
	}

	return e.WriteAddressList(Addresses(v))
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

func (v *CcFieldValue) Decode(d *DataDecoder) error {
	addrs, err := d.ReadAddressList(false)
	if err != nil {
		return err
	}

	*v = CcFieldValue(addrs)
	return nil
}

func (v CcFieldValue) Encode(e *DataEncoder) error {
	if len(v) == 0 {
		return fmt.Errorf("invalid empty address list")
	}

	return e.WriteAddressList(Addresses(v))
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

func (v *BccFieldValue) Decode(d *DataDecoder) error {
	addrs, err := d.ReadAddressList(true)
	if err != nil {
		return err
	}

	*v = BccFieldValue(addrs)
	return nil
}

func (v BccFieldValue) Encode(e *DataEncoder) error {
	// The Bcc field can be empty
	return e.WriteAddressList(Addresses(v))
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

func (v *MessageIdFieldValue) Decode(d *DataDecoder) error {
	id, err := d.ReadMessageId()
	if err != nil {
		return err
	}

	*v = MessageIdFieldValue(*id)
	return nil
}

func (v MessageIdFieldValue) Encode(e *DataEncoder) error {
	return e.WriteMessageId(MessageId(v))
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

func (v *InReplyToFieldValue) Decode(d *DataDecoder) error {
	ids, err := d.ReadMessageIdList(true)
	if err != nil {
		return err
	}

	*v = InReplyToFieldValue(ids)
	return nil
}

func (v InReplyToFieldValue) Encode(e *DataEncoder) error {
	if len(v) == 0 {
		return fmt.Errorf("invalid empty message id list")
	}

	return e.WriteMessageIdList(MessageIds(v))
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

func (v *ReferencesFieldValue) Decode(d *DataDecoder) error {
	ids, err := d.ReadMessageIdList(true)
	if err != nil {
		return err
	}

	*v = ReferencesFieldValue(ids)
	return nil
}

func (v ReferencesFieldValue) Encode(e *DataEncoder) error {
	if len(v) == 0 {
		return fmt.Errorf("invalid empty message id list")
	}

	return e.WriteMessageIdList(MessageIds(v))
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

func (v *SubjectFieldValue) Decode(d *DataDecoder) error {
	value, err := d.ReadUnstructured()
	if err != nil {
		return err
	}

	*v = SubjectFieldValue(value)
	return nil
}

func (v SubjectFieldValue) Encode(e *DataEncoder) error {
	e.WriteUnstructured(string(v))
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

func (v *CommentsFieldValue) Decode(d *DataDecoder) error {
	value, err := d.ReadUnstructured()
	if err != nil {
		return err
	}

	*v = CommentsFieldValue(value)
	return nil
}

func (v CommentsFieldValue) Encode(e *DataEncoder) error {
	e.WriteUnstructured(string(v))
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

func (v *KeywordsFieldValue) Decode(d *DataDecoder) error {
	phrases, err := d.ReadPhraseList()
	if err != nil {
		return err
	}

	*v = phrases
	return nil
}

func (v KeywordsFieldValue) Encode(e *DataEncoder) error {
	if len(v) == 0 {
		return fmt.Errorf("invalid empty phrase list")
	}

	return e.WritePhraseList(v)
}

func (v *KeywordsFieldValue) testGenerate(g *TestMessageGenerator) {
	*v = KeywordsFieldValue(g.generatePhrases())
}

func (v KeywordsFieldValue) testCheck(t *testing.T, g *TestMessageGenerator, ev FieldValue) {
	ev2 := ev.(*KeywordsFieldValue)
	g.checkPhrases(t, []string(*ev2), []string(v))
}

// Optional fields
type OptionalFieldValue string

func (v OptionalFieldValue) String() string {
	return fmt.Sprintf("%q", string(v))
}

func (v *OptionalFieldValue) Decode(d *DataDecoder) error {
	value, err := d.ReadUnstructured()
	if err != nil {
		return err
	}

	*v = OptionalFieldValue(value)
	return nil
}

func (v OptionalFieldValue) Encode(e *DataEncoder) error {
	e.WriteUnstructured(string(v))
	return nil
}

func (v *OptionalFieldValue) testGenerate(g *TestMessageGenerator) {
	*v = OptionalFieldValue(g.generateUnstructured())
}

func (v OptionalFieldValue) testCheck(t *testing.T, g *TestMessageGenerator, ev FieldValue) {
	ev2 := ev.(*OptionalFieldValue)
	g.checkUnstructured(t, string(*ev2), string(v))
}
