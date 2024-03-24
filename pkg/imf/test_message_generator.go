package imf

import (
	"bytes"
	"fmt"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/galdor/emaild/pkg/utils"
)

const NbFieldTests = 1000

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
	}

	if g.maybe(0.25) {
		g.generateCFWS()
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

	if g.maybe(0.25) {
		g.generateCFWS()
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
	if g.maybe(0.1) {
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

func (g *TestMessageGenerator) generatePhrase() string {
	var buf bytes.Buffer

	for i := 1; i <= 1+rand.Intn(2); i++ {
		if i > 1 {
			g.generateFWS()
			buf.WriteByte(' ')
		}

		if g.maybe(0.8) {
			buf.WriteString(g.generateWord())
		} else {
			g.writeByte('.')
			buf.WriteByte('.')
		}
	}

	return buf.String()
}

func (g *TestMessageGenerator) generateDate() time.Time {
	writeSpace := func() {
		// Obsolete syntax allos CFWS pretty much everywhere in dates and times.
		g.generateCFWS()
	}

	minTimestamp := time.Date(1950, 1, 1, 0, 0, 0, 0, time.UTC).Unix()
	maxTimestamp := time.Date(2100, 1, 1, 0, 0, 0, 0, time.UTC).Unix()
	timestamp := minTimestamp + rand.Int63n(maxTimestamp-minTimestamp)

	date := time.Unix(timestamp, 0)

	tzHourOffset := -12 + rand.Intn(26)
	tzMinuteOffset := rand.Intn(60)

	var tzString string
	if g.maybe(0.75) {
		date = date.In(time.FixedZone("", tzHourOffset*3600+tzMinuteOffset*60))

		sign := "+"
		if tzHourOffset < 0 {
			sign = "-"
		}

		tzString = fmt.Sprintf("%s%02d%02d", sign,
			utils.Abs(tzHourOffset), tzMinuteOffset)
	} else {
		var zones = []struct {
			Name   string
			Offset int
		}{
			{"EDT", -4},
			{"EST", -5},
			{"CDT", -5},
			{"CST", -6},
			{"MDT", -6},
			{"MST", -7},
			{"PDT", -7},
			{"PST", -8},
		}

		n := rand.Intn(8)
		date = date.In(time.FixedZone(zones[n].Name, zones[n].Offset*3600))
		tzString = zones[n].Name
	}

	// Day of the week
	if g.maybe(0.5) {
		g.writeString(date.Format("Mon"))
		if g.maybe(0.5) {
			writeSpace()
		}
		g.writeByte(',')
	}

	// Day
	writeSpace()
	if day := date.Day(); day < 10 && g.maybe(0.5) {
		g.writeString(strconv.Itoa(day))
	} else {
		g.writeString(fmt.Sprintf("%02d", day))
	}

	// Month
	writeSpace()
	g.writeString(date.Format("Jan"))

	// Year
	writeSpace()

	year := date.Year()
	if g.maybe(0.5) {
		g.writeString(fmt.Sprintf("%04d", year))
	} else {
		if g.maybe(0.5) {
			if year%100 < 50 {
				g.writeString(fmt.Sprintf("%02d", year-2000))
			} else {
				g.writeString(fmt.Sprintf("%02d", year-1900))

			}
		} else {
			g.writeString(fmt.Sprintf("%03d", year-1900))
		}
	}

	// Hour
	writeSpace()
	g.writeString(fmt.Sprintf("%02d", date.Hour()))

	// Minute
	if g.maybe(0.1) {
		writeSpace()
	}
	g.writeByte(':')

	if g.maybe(0.25) {
		writeSpace()
	}
	g.writeString(fmt.Sprintf("%02d", date.Minute()))

	// Second
	if g.maybe(0.5) {
		if g.maybe(0.1) {
			writeSpace()
		}
		g.writeByte(':')

		g.writeString(fmt.Sprintf("%02d", date.Second()))
	} else {
		date = date.Truncate(time.Minute)
	}

	// Timezone
	writeSpace()
	g.writeString(tzString)

	return date
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
				g.generateCFWS()
			}

			g.writeByte('.')
			buf.WriteByte('.')

			if g.maybe(0.05) {
				g.generateCFWS()
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

	if g.maybe(0.75) {
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
		g.generateFWS()
	}

	g.writeString("@")

	if g.maybe(0.1) {
		g.generateFWS()
	}

	domain := g.generateDomain()

	addr := SpecificAddress{
		LocalPart: localPart,
		Domain:    domain,
	}

	return &addr
}

func (g *TestMessageGenerator) generateMailbox() *Mailbox {
	var mb Mailbox

	if g.maybe(0.5) {
		mb.DisplayName = utils.Ref(g.generatePhrase())
		g.generateFWS()

		g.writeByte('<')
		g.generateFWS()
	}

	addr := g.generateSpecificAddress()

	mb.SpecificAddress = *addr

	if mb.DisplayName != nil {
		g.generateFWS()
		g.writeByte('>')
	}

	return &mb
}

func (g *TestMessageGenerator) generateGroup() *Group {
	var group Group

	group.DisplayName = g.generatePhrase()
	g.generateFWS()

	g.writeByte(':')
	g.generateFWS()

	for i := 0; i <= rand.Intn(3); i++ {
		if i > 0 {
			g.generateFWS()
			g.writeByte(',')
			g.generateFWS()
		}

		if g.maybe(0.1) {
			g.generateCFWS()
		} else {
			mb := g.generateMailbox()
			group.Mailboxes = append(group.Mailboxes, mb)
		}
	}

	g.generateFWS()
	g.writeByte(';')

	return &group
}

func (g *TestMessageGenerator) generateAddress() Address {
	if g.maybe(0.75) {
		return g.generateMailbox()
	} else {
		return g.generateGroup()
	}
}

func (g *TestMessageGenerator) generateMessageId() MessageId {
	var id MessageId

	g.writeByte('<')
	if g.maybe(0.1) {
		g.generateFWS()
	}

	if g.maybe(0.5) {
		id.Left = g.generateDotAtom()
	} else {
		id.Left = g.generateLocalPart()
	}

	if g.maybe(0.1) {
		g.generateFWS()
	}
	g.writeByte('@')
	if g.maybe(0.1) {
		g.generateFWS()
	}

	if g.maybe(0.5) {
		id.Right = g.generateDotAtom()
	} else {
		id.Right = g.generateDomain()
	}

	if g.maybe(0.1) {
		g.generateFWS()
	}
	g.writeByte('>')

	return id
}

func (g *TestMessageGenerator) generateMessageIds() MessageIds {
	var ids MessageIds

	for i := 0; i < rand.Intn(4); i++ {
		if i > 0 {
			g.generateFWS()
		}

		ids = append(ids, g.generateMessageId())
	}

	return ids
}

func (g *TestMessageGenerator) checkDate(eDate, date time.Time) bool {
	dateString := date.Format(time.RFC3339)
	eDateString := eDate.Format(time.RFC3339)

	if dateString != eDateString {
		g.t.Errorf("date is %q but should be %q", dateString, eDateString)
		return false
	}

	return true
}

func (g *TestMessageGenerator) checkSpecificAddress(eAddr, addr *SpecificAddress) bool {
	valid := true

	switch {
	case addr == nil && eAddr != nil:
		g.t.Errorf("address is null but should be %#v", eAddr)
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

func (g *TestMessageGenerator) checkAddress(eAddr, addr Address) bool {
	mailbox, isMailbox := addr.(*Mailbox)
	group, isGroup := addr.(*Group)

	eMailbox, isEMailbox := addr.(*Mailbox)
	eGroup, isEGroup := addr.(*Group)

	if isGroup && isEMailbox {
		g.t.Errorf("address is a group but should be a mailbox")
		return false
	}

	if isMailbox && isEGroup {
		g.t.Errorf("address is a mailbox but should be a group")
		return false
	}

	valid := true

	checkDisplayNames := func(eName, name *string, label string) {
		switch {
		case name == nil && eName != nil:
			g.t.Errorf("%s does not have a display name but should have "+
				"display name %q", label, *eName)
			valid = false
		case name != nil && eName == nil:
			g.t.Errorf("%s has display name %q but should not have one ",
				label, *name)
			valid = false
		case name != nil && eName != nil:
			if *name != *eName {
				g.t.Errorf("%s display name is %q but should be %q",
					label, *name, *eName)
				valid = false
			}
		}
	}

	if isMailbox {
		checkDisplayNames(eMailbox.DisplayName, mailbox.DisplayName, "mailbox")
	} else if isGroup {
		checkDisplayNames(&eGroup.DisplayName, &group.DisplayName, "group")
	}

	return valid
}

func (g *TestMessageGenerator) checkMessageId(eId, id MessageId) bool {
	valid := true

	if id.Left != eId.Left {
		g.t.Errorf("left part is %q but should be %q", id.Left, eId.Left)
		valid = false
	}

	if id.Right != eId.Right {
		g.t.Errorf("right part is %q but should be %q", id.Right, eId.Right)
		valid = false
	}

	return valid
}

func (g *TestMessageGenerator) checkMessageIds(eIds, ids MessageIds) bool {
	if len(ids) != len(eIds) {
		g.t.Errorf("list contains %d ids but should contain %d ids",
			len(ids), len(eIds))
		return false
	}

	valid := true

	for i, id := range ids {
		eId := eIds[i]

		if !g.checkMessageId(eId, id) {
			g.t.Errorf("invalid id at index %d:", i)
			valid = false
		}
	}

	return valid
}
