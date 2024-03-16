package imf

import "testing"

func TestReadReturnPathField(t *testing.T) {
	g := NewTestMessageGenerator(t)
	g.GenerateAndTestFieldN("Return-Path")
}
