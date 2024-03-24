package imf

import "testing"

func TestReadReturnPathField(t *testing.T) {
	g := NewTestMessageGenerator(t)
	g.GenerateAndTestFieldN("Return-Path")
}

// TODO Received

func TestReadResentDateField(t *testing.T) {
	g := NewTestMessageGenerator(t)
	g.GenerateAndTestFieldN("Resent-Date")
}

func TestReadResentFromField(t *testing.T) {
	g := NewTestMessageGenerator(t)
	g.GenerateAndTestFieldN("Resent-From")
}

func TestReadResentSenderField(t *testing.T) {
	g := NewTestMessageGenerator(t)
	g.GenerateAndTestFieldN("Resent-Sender")
}

func TestReadResentToField(t *testing.T) {
	g := NewTestMessageGenerator(t)
	g.GenerateAndTestFieldN("Resent-To")
}

func TestReadResentCcField(t *testing.T) {
	g := NewTestMessageGenerator(t)
	g.GenerateAndTestFieldN("Resent-Cc")
}

func TestReadResentBccField(t *testing.T) {
	g := NewTestMessageGenerator(t)
	g.GenerateAndTestFieldN("Resent-Bcc")
}

func TestReadResentMessageIdField(t *testing.T) {
	g := NewTestMessageGenerator(t)
	g.GenerateAndTestFieldN("Resent-Message-ID")
}

func TestReadDateField(t *testing.T) {
	g := NewTestMessageGenerator(t)
	g.GenerateAndTestFieldN("Date")
}

func TestReadFromField(t *testing.T) {
	g := NewTestMessageGenerator(t)
	g.GenerateAndTestFieldN("From")
}

func TestReadSenderField(t *testing.T) {
	g := NewTestMessageGenerator(t)
	g.GenerateAndTestFieldN("Sender")
}

func TestReadReplyToField(t *testing.T) {
	g := NewTestMessageGenerator(t)
	g.GenerateAndTestFieldN("Reply-To")
}

func TestReadToField(t *testing.T) {
	g := NewTestMessageGenerator(t)
	g.GenerateAndTestFieldN("To")
}

func TestReadCcField(t *testing.T) {
	g := NewTestMessageGenerator(t)
	g.GenerateAndTestFieldN("Cc")
}

func TestReadBccField(t *testing.T) {
	g := NewTestMessageGenerator(t)
	g.GenerateAndTestFieldN("Bcc")
}

func TestReadMessageIdField(t *testing.T) {
	g := NewTestMessageGenerator(t)
	g.GenerateAndTestFieldN("Message-ID")
}

func TestReadInReplyToField(t *testing.T) {
	g := NewTestMessageGenerator(t)
	g.GenerateAndTestFieldN("In-Reply-To")
}

func TestReadReferencesField(t *testing.T) {
	g := NewTestMessageGenerator(t)
	g.GenerateAndTestFieldN("References")
}

func TestReadSubjectField(t *testing.T) {
	g := NewTestMessageGenerator(t)
	g.GenerateAndTestFieldN("Subject")
}

func TestReadCommentsField(t *testing.T) {
	g := NewTestMessageGenerator(t)
	g.GenerateAndTestFieldN("Comments")
}

// TODO Keywords

// TODO Optional fields
