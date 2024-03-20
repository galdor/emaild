package imf

import (
	"bytes"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"testing"

	"github.com/galdor/emaild/pkg/utils"
)

const NbFieldTests = 100

var (
	atomChars = CharRange('a', 'z') + CharRange('A', 'Z') +
		CharRange('0', '9') + "!#$%&'*+-/=?^_`{|}~"

	quotedStringChars = CharRange(33, 39) + CharRange(42, 91) +
		CharRange(93, 126) +
		CharRange(1, 8) + CharRange(11, 12) + CharRange(14, 31) +
		CharRange(127, 127)

	commentChars = quotedStringChars
)

type TestMessageGenerator struct {
	t *testing.T

	buf bytes.Buffer
}

func NewTestMessageGenerator(t *testing.T) *TestMessageGenerator {
	return &TestMessageGenerator{
		t: t,
	}
}

func (g *TestMessageGenerator) GenerateAndTestMessage() {
	g.buf.Reset()

	eMsg := g.generateMessage()

	r := NewMessageReader()
	msg, err := r.ReadAll(g.buf.Bytes())
	if err != nil {
		g.t.Fatalf("%v", err)
	}

	g.checkMessage(eMsg, msg)
}

func (g *TestMessageGenerator) GenerateAndTestFieldN(name string) {
	for i := 0; i < NbFieldTests; i++ {
		g.GenerateAndTestField(name)
	}
}

func (g *TestMessageGenerator) GenerateAndTestField(name string) {
	g.buf.Reset()

	eField := g.generateField(name)

	g.t.Logf("field: %q", g.buf.String())

	r := NewMessageReader()
	msg, err := r.ReadAll(g.buf.Bytes())
	if err != nil {
		g.t.Fatalf("%v", err)
	}

	if len(msg.Header) == 0 {
		g.t.Errorf("parsing succeeded but no field was found")
		return
	}

	g.checkField(eField, msg.Header[0])
}

func (g *TestMessageGenerator) generateMessage() *Message {
	var msg Message

	// TODO

	return &msg
}

func (g *TestMessageGenerator) generateField(name string) *Field {
	var field Field

	field.Name = name

	g.writeString(field.Name)
	if g.maybe(0.1) {
		g.generateFWS()
	}
	g.writeString(":")
	if g.maybe(0.9) {
		g.generateWS()
	} else if g.maybe(0.5) {
		g.generateFWS()
	}

	switch strings.ToLower(field.Name) {
	case "return-path":
		field.Value = &ReturnPathFieldValue{}
	case "received":
		field.Value = &ReceivedFieldValue{}
	case "resent-date":
		field.Value = &ResentDateFieldValue{}
	case "resent-from":
		field.Value = &ResentFromFieldValue{}
	case "resent-sender":
		field.Value = &ResentSenderFieldValue{}
	case "resent-to":
		field.Value = &ResentToFieldValue{}
	case "resent-cc":
		field.Value = &ResentCcFieldValue{}
	case "resent-bcc":
		field.Value = &ResentBccFieldValue{}
	case "resent-message-id":
		field.Value = &ResentMessageIdFieldValue{}
	case "date":
		field.Value = &DateFieldValue{}
	case "from":
		field.Value = &FromFieldValue{}
	case "sender":
		field.Value = &SenderFieldValue{}
	case "reply-to":
		field.Value = &ReplyToFieldValue{}
	case "to":
		field.Value = &ToFieldValue{}
	case "cc":
		field.Value = &CcFieldValue{}
	case "bcc":
		field.Value = &BccFieldValue{}
	case "message-id":
		field.Value = &MessageIdFieldValue{}
	case "in-reply-to":
		field.Value = &InReplyToFieldValue{}
	case "references":
		field.Value = &ReferencesFieldValue{}
	case "subject":
		field.Value = utils.Ref(SubjectFieldValue(""))
	case "comments":
		field.Value = utils.Ref(CommentsFieldValue(""))
	case "keywords":
		field.Value = &KeywordsFieldValue{}
	default:
		field.Value = utils.Ref(OptionalFieldValue(""))
	}

	field.Value.testGenerate(g)

	if g.maybe(0.1) {
		g.generateFWS()
	}
	g.writeString("\r\n")

	return &field
}

func (g *TestMessageGenerator) checkMessage(eMsg, msg *Message) {
	if len(msg.Header) == len(eMsg.Header) {
		for i, efield := range eMsg.Header {
			field := msg.Header[i]
			g.checkField(efield, field)
		}
	} else {
		g.t.Errorf("header contains %d fields but should contain %d fields",
			len(msg.Header), len(eMsg.Header))
	}

	eBody := string(eMsg.Body)
	body := string(msg.Body)
	if body != eBody {
		g.t.Errorf("body is %q but should be %q", body, eBody)
	}
}

func (g *TestMessageGenerator) checkField(eField, field *Field) {
	if field.Name != eField.Name {
		g.t.Errorf("field is named %q but should be named %q",
			eField.Name, field.Name)
		return
	}

	field.Value.testCheck(g, eField.Value)
}

func (g *TestMessageGenerator) maybe(p float32) bool {
	if p < 0.0 || p > 1.0 {
		utils.Panicf("probability must be in [0, 1.0)")
	}

	return rand.Float32() <= p
}

func (g *TestMessageGenerator) randData(n int) []byte {
	data := make([]byte, n)
	if _, err := rand.Read(data); err != nil {
		utils.Panicf("cannot generate random data: %v", err)
	}

	return data
}

func (g *TestMessageGenerator) writeByte(c byte) string {
	g.buf.WriteByte(c)
	return string(c)
}

func (g *TestMessageGenerator) writeString(s string) string {
	g.buf.WriteString(s)
	return s
}

func (g *TestMessageGenerator) generateWS() {
	if g.maybe(0.25) {
		g.writeString("\t")
	} else {
		g.writeString(" ")
	}
}

func (g *TestMessageGenerator) generateFWS() {
	for i := 0; i < rand.Intn(3)+1; i++ {
		g.generateWS()
	}

	for i := 0; i < rand.Intn(2); i++ {
		g.writeString("\r\n")

		for i := 0; i < rand.Intn(3)+1; i++ {
			g.generateWS()
		}
	}
}

func (g *TestMessageGenerator) generateComment() {
	g.writeByte('(')

	for i := 0; i < rand.Intn(3); i++ {
		if g.maybe(0.1) {
			g.generateFWS()
		}

		if g.maybe(0.8) {
			for i := 0; i < rand.Intn(8); i++ {
				c := commentChars[rand.Intn(len(commentChars))]

				if c == '\\' || c == '(' || c == ')' || g.maybe(0.05) {
					g.writeByte('\\')
				}

				g.writeByte(c)
			}
		} else {
			g.generateComment()
		}
	}

	if g.maybe(0.05) {
		g.generateFWS()
	}

	g.writeByte(')')
}

func (g *TestMessageGenerator) generateCFWS() {
	if g.maybe(0.1) {
		for i := 0; i < rand.Intn(2)+1; i++ {
			if g.maybe(0.5) {
				g.generateFWS()
			}

			g.generateComment()
		}

		if g.maybe(0.5) {
			g.generateFWS()
		}
	} else {
		g.generateFWS()
	}
}

func (g *TestMessageGenerator) generateAtom() string {
	atom := make([]byte, rand.Intn(8)+1)

	for i := 0; i < len(atom); i++ {
		atom[i] = atomChars[rand.Intn(len(atomChars))]
	}

	return g.writeString(string(atom))
}

func (g *TestMessageGenerator) generateDotAtom() string {
	var dotAtom bytes.Buffer

	atom := g.generateAtom()
	dotAtom.WriteString(atom)

	for i := 0; i < rand.Intn(3); i++ {
		g.writeString(".")
		dotAtom.WriteByte('.')

		atom := g.generateAtom()
		dotAtom.WriteString(atom)
	}

	return dotAtom.String()
}

func (g *TestMessageGenerator) generateQuotedString() string {
	var quotedString []byte

	g.writeByte('"')

	for i := 0; i < rand.Intn(8); i++ {
		if g.maybe(0.05) {
			g.generateFWS()
		}

		c := quotedStringChars[rand.Intn(len(quotedStringChars))]

		if c == '"' || c == '\\' || g.maybe(0.05) {
			g.writeByte('\\')
		}

		g.writeByte(c)
		quotedString = append(quotedString, c)
	}

	if g.maybe(0.05) {
		g.generateFWS()
	}

	g.writeByte('"')

	return string(quotedString)
}

func (g *TestMessageGenerator) generateWord() string {
	if g.maybe(0.75) {
		return g.generateAtom()
	}

	return g.generateQuotedString()
}

func (g *TestMessageGenerator) generateLocalPart() string {
	var part string

	switch rand.Intn(3) {
	case 0:
		part = g.generateDotAtom()
	case 1:
		part = g.generateQuotedString()
	case 2:
		var buf bytes.Buffer

		buf.WriteString(g.generateWord())

		for i := 0; i < rand.Intn(3); i++ {
			if g.maybe(0.05) {
				g.generateFWS()
			}

			g.writeByte('.')
			buf.WriteByte('.')

			if g.maybe(0.05) {
				g.generateFWS()
			}

			buf.WriteString(g.generateWord())
		}

		part = buf.String()
	}

	return part
}

func (g *TestMessageGenerator) generateDomainLiteral() string {
	var buf bytes.Buffer

	buf.WriteByte('[')

	if g.maybe(0.5) {
		for i := 0; i < 4; i++ {
			if i > 0 {
				buf.WriteByte('.')
			}

			buf.WriteString(strconv.Itoa(rand.Intn(256)))
		}
	} else {
		buf.WriteString("IPv6:")
		buf.WriteString(net.IP(g.randData(16)).String())
	}

	buf.WriteByte(']')

	s := buf.String()

	for i := range s {
		if i > 0 && g.maybe(0.05) {
			g.generateFWS()
		}

		g.writeByte(s[i])
	}

	if g.maybe(0.05) {
		g.generateFWS()
	}

	return s
}

func (g *TestMessageGenerator) generateDomainName() string {
	return g.generateDotAtom()
}

func (g *TestMessageGenerator) generateDomain() string {
	var domain string

	if g.maybe(0.5) {
		domain = g.generateDomainLiteral()
	} else {
		domain = g.generateDomainName()
	}

	return domain
}

func (g *TestMessageGenerator) generateSpecificAddress() *SpecificAddress {
	localPart := g.generateLocalPart()

	if g.maybe(0.1) {
		g.generateCFWS()
	}

	g.writeString("@")

	if g.maybe(0.1) {
		g.generateCFWS()
	}

	domain := g.generateDomain()

	addr := SpecificAddress{
		LocalPart: localPart,
		Domain:    domain,
	}

	return &addr
}

func (g *TestMessageGenerator) checkSpecificAddress(eAddr, addr *SpecificAddress) bool {
	valid := true

	switch {
	case addr == nil && eAddr != nil:
		g.t.Errorf("address is null but should be equal to %#v", eAddr)
		valid = false

	case addr != nil && eAddr == nil:
		g.t.Errorf("address is %#v but should be null", addr)
		valid = false

	case addr != nil && eAddr != nil:
		if addr.LocalPart != eAddr.LocalPart {
			g.t.Errorf("local part is %q but should be %q",
				addr.LocalPart, eAddr.LocalPart)
			valid = false
		}

		if addr.Domain != eAddr.Domain {
			g.t.Errorf("domain is %q but should be %q",
				addr.Domain, eAddr.Domain)
			valid = false
		}
	}

	return valid
}
