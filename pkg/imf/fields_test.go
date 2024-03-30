package imf

import "testing"

func TestReadReturnPathField(t *testing.T) {
	g := NewTestMessageGenerator()
	g.GenerateAndTestFieldN(t, "Return-Path")
}

// TODO Received

func TestReadResentDateField(t *testing.T) {
	g := NewTestMessageGenerator()
	g.GenerateAndTestFieldN(t, "Resent-Date")
}

func TestReadResentFromField(t *testing.T) {
	g := NewTestMessageGenerator()
	g.GenerateAndTestFieldN(t, "Resent-From")
}

func TestReadResentSenderField(t *testing.T) {
	g := NewTestMessageGenerator()
	g.GenerateAndTestFieldN(t, "Resent-Sender")
}

func TestReadResentToField(t *testing.T) {
	g := NewTestMessageGenerator()
	g.GenerateAndTestFieldN(t, "Resent-To")
}

func TestReadResentCcField(t *testing.T) {
	g := NewTestMessageGenerator()
	g.GenerateAndTestFieldN(t, "Resent-Cc")
}

func TestReadResentBccField(t *testing.T) {
	g := NewTestMessageGenerator()
	g.GenerateAndTestFieldN(t, "Resent-Bcc")
}

func TestReadResentMessageIdField(t *testing.T) {
	g := NewTestMessageGenerator()
	g.GenerateAndTestFieldN(t, "Resent-Message-ID")
}

func TestReadDateField(t *testing.T) {
	g := NewTestMessageGenerator()
	g.GenerateAndTestFieldN(t, "Date")
}

func TestReadFromField(t *testing.T) {
	g := NewTestMessageGenerator()
	g.GenerateAndTestFieldN(t, "From")
}

func TestReadSenderField(t *testing.T) {
	g := NewTestMessageGenerator()
	g.GenerateAndTestFieldN(t, "Sender")
}

func TestReadReplyToField(t *testing.T) {
	g := NewTestMessageGenerator()
	g.GenerateAndTestFieldN(t, "Reply-To")
}

func TestReadToField(t *testing.T) {
	g := NewTestMessageGenerator()
	g.GenerateAndTestFieldN(t, "To")
}

func TestReadCcField(t *testing.T) {
	g := NewTestMessageGenerator()
	g.GenerateAndTestFieldN(t, "Cc")
}

func TestReadBccField(t *testing.T) {
	g := NewTestMessageGenerator()
	g.GenerateAndTestFieldN(t, "Bcc")
}

func TestReadMessageIdField(t *testing.T) {
	g := NewTestMessageGenerator()
	g.GenerateAndTestFieldN(t, "Message-ID")
}

func TestReadInReplyToField(t *testing.T) {
	g := NewTestMessageGenerator()
	g.GenerateAndTestFieldN(t, "In-Reply-To")
}

func TestReadReferencesField(t *testing.T) {
	g := NewTestMessageGenerator()
	g.GenerateAndTestFieldN(t, "References")
}

func TestReadSubjectField(t *testing.T) {
	g := NewTestMessageGenerator()
	g.GenerateAndTestFieldN(t, "Subject")
}

func TestReadCommentsField(t *testing.T) {
	g := NewTestMessageGenerator()
	g.GenerateAndTestFieldN(t, "Comments")
}

// TODO Keywords

// TODO Optional fields
