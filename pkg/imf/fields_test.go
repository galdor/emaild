package imf

import (
	"math/rand"
	"testing"
)

func TestReadReturnPathField(t *testing.T) {
	g := NewTestMessageGenerator()
	g.GenerateAndTestFieldN(t, "Return-Path")
}

func TestReadReceivedField(t *testing.T) {
	g := NewTestMessageGenerator()
	g.GenerateAndTestFieldN(t, "Received")
}

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

func TestReadKeywordsField(t *testing.T) {
	g := NewTestMessageGenerator()
	g.GenerateAndTestFieldN(t, "Keywords")
}

func TestReadOptionalField(t *testing.T) {
	g := NewTestMessageGenerator()

	for i := 0; i < NbFieldTests; i++ {
		// We always use 'X' as first character to make sure to never generate
		// an existing non-optional field.

		name := make([]byte, rand.Intn(16)+1)
		name[0] = 'X'

		for i := 1; i < len(name); i++ {
			name[i] = fieldChars[rand.Intn(len(fieldChars))]
		}

		g.GenerateAndTestField(t, string(name))
	}
}
