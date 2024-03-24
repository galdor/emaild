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

// TODO Resent-From

func TestReadResentSenderField(t *testing.T) {
	g := NewTestMessageGenerator(t)
	g.GenerateAndTestFieldN("Resent-Sender")
}

// TODO Resent-To

// TODO Resent-Cc

// TODO Resent-Bcc

func TestReadResentMessageIdField(t *testing.T) {
	g := NewTestMessageGenerator(t)
	g.GenerateAndTestFieldN("Resent-Message-ID")
}

func TestReadDateField(t *testing.T) {
	g := NewTestMessageGenerator(t)
	g.GenerateAndTestFieldN("Date")
}

// TODO From

func TestReadSenderField(t *testing.T) {
	g := NewTestMessageGenerator(t)
	g.GenerateAndTestFieldN("Sender")
}

// TODO Reply-To

// TODO To

// TODO Cc

// TODO Bcc

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

// TODO Subject

// TODO Comments

// TODO Keywords

// TODO Optional fields
