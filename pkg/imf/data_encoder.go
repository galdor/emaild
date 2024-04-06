package imf

import (
	"bytes"
	"fmt"
	"regexp"
	"time"
	"unicode/utf8"

	"github.com/galdor/emaild/pkg/utils"
)

var (
	phraseWordSeparatorRE = regexp.MustCompile("\\s+")
)

type DataEncoder struct {
	MaxLineLength int // 0 if no maximum line length

	buf        bytes.Buffer
	lineLength int
}

func MustEncodeInlineData(fn func(*DataEncoder) error) string {
	e := NewDataEncoder(0)

	if err := fn(e); err != nil {
		utils.Panicf("cannot encode data: %v", err)
	}

	return e.buf.String()
}

func (e *DataEncoder) Bytes() []byte {
	return e.buf.Bytes()
}

func NewDataEncoder(maxLineLength int) *DataEncoder {
	return &DataEncoder{
		MaxLineLength: maxLineLength,
	}
}

func (e *DataEncoder) WriteRune(c rune) {
	sz := utf8.RuneLen(c)

	if e.MaxLineLength > 0 && e.lineLength+sz > e.MaxLineLength {
		e.buf.WriteString("\r\n ")
		e.lineLength = 1
	}

	e.buf.WriteRune(c)
	e.lineLength += sz
}

func (e *DataEncoder) WriteString(s string) {
	if e.MaxLineLength > 0 && e.lineLength+len(s) > e.MaxLineLength {
		e.buf.WriteString("\r\n ")
		e.lineLength = 1
	}

	e.buf.WriteString(s)
	e.lineLength += len(s)
}

func (e *DataEncoder) WriteHeader(header []*Field) error {
	for _, field := range header {
		e.WriteField(field)
	}

	return nil
}

func (e *DataEncoder) WriteField(f *Field) error {
	e.WriteString(f.Name)
	e.WriteString(": ")

	if err := f.Value.Encode(e); err != nil {
		return fmt.Errorf("invalid value: %w", err)
	}

	e.WriteEOL()

	return nil
}

func (e *DataEncoder) WriteEOL() {
	e.buf.WriteString("\r\n")
	e.lineLength = 0
}

func (e *DataEncoder) WriteUnstructured(s string) {
	for _, c := range s {
		e.WriteRune(c)
	}
}

func (e *DataEncoder) WriteQuotedString(s string) error {
	// All printable ASCII characters are allowed; some characters must be
	// escaped with a single backslash. UTF-8 sequences are allowed (see RFC
	// 6532 Internationalized Email Headers).
	//
	// Quoted strings can be folded. We can split between characters as long as
	// we do not split quoted pairs (e.g. "\\") or UTF-8 sequences.

	e.WriteRune('"')

	for _, c := range s {
		switch {
		case c < 33 && c != '\t' && c != ' ':
			return fmt.Errorf("unencodable control character 0x%x", c)
		case c == '"' || c == '\\':
			e.WriteString("\\" + string(c))
		default:
			e.WriteRune(c)
		}
	}

	e.WriteRune('"')

	return nil
}

func (e *DataEncoder) WriteAtomOrQuotedString(s string) error {
	if IsAtom(s) {
		e.WriteString(s)
		return nil
	}

	return e.WriteQuotedString(s)
}

func (e *DataEncoder) WriteDotAtomOrQuotedString(s string) error {
	if IsDotAtom(s) {
		e.WriteString(s)
		return nil
	}

	return e.WriteQuotedString(s)
}

func (e *DataEncoder) WriteDomain(domain Domain) {
	s := string(domain)

	if IsDotAtom(s) {
		e.WriteString(s)
	} else {
		// Domain literals are stored with their bracket delimiters, so we do
		// not have to add them here.
		e.WriteString(s)
	}
}

func (e *DataEncoder) WriteWord(word string) error {
	return e.WriteAtomOrQuotedString(word)
}

func (e *DataEncoder) WritePhrase(phrase string) error {
	// Phrases are defined as a list of at least one word. The correct
	// representation would be []string, but it is inconvenient. So we extract
	// words from the string ourselves then encode them.

	words := phraseWordSeparatorRE.Split(phrase, -1)
	for i, word := range words {
		if i > 0 {
			e.WriteRune(' ')
		}

		if err := e.WriteWord(word); err != nil {
			return fmt.Errorf("invalid word %q: %w", word, err)
		}
	}

	return nil
}

func (e *DataEncoder) WritePhraseList(phrases []string) error {
	for i, phrase := range phrases {
		if i > 0 {
			e.WriteString(", ")
		}

		if err := e.WritePhrase(phrase); err != nil {
			return fmt.Errorf("invalid phrase %q: %w", phrase, err)
		}
	}

	return nil
}

func (e *DataEncoder) WriteDateTime(date time.Time) {
	// RFC 5322 3.3. Date and Time Specification

	e.WriteString(date.Format("Mon"))
	e.WriteString(", ")
	e.WriteString(date.Format("02 Jan 2006"))
	e.WriteRune(' ')
	e.WriteString(date.Format("15:04:05"))
	e.WriteRune(' ')
	e.WriteString(date.Format("-0700"))
}

func (e *DataEncoder) WriteMessageId(id MessageId) error {
	// RFC 5322 3.6.4. Identification Fields

	e.WriteRune('<')
	e.WriteDotAtomOrQuotedString(id.Left)
	e.WriteRune('@')
	e.WriteDomain(id.Right)
	e.WriteRune('>')

	return nil
}

func (e *DataEncoder) WriteMessageIdList(ids MessageIds) error {
	for i, id := range ids {
		if i > 0 {
			e.WriteRune(' ')
		}

		if err := e.WriteMessageId(id); err != nil {
			return fmt.Errorf("invalid message id %v: %w", id, err)
		}
	}

	return nil
}

func (e *DataEncoder) WriteAddress(addr Address) error {
	switch addr2 := addr.(type) {
	case *Mailbox:
		return e.WriteMailbox(addr2)
	case *Group:
		return e.WriteGroup(addr2)
	}

	utils.Panicf("unhandled address %#v (%T)", addr, addr)
	return nil // the Go compiler still cannot do basic flow analysis...
}

func (e *DataEncoder) WriteAddressList(addrs Addresses) error {
	for i, addr := range addrs {
		if i > 0 {
			e.WriteString(", ")
		}

		if err := e.WriteAddress(addr); err != nil {
			return fmt.Errorf("invalid address %v: %w", addr, err)
		}
	}

	return nil
}

func (e *DataEncoder) WriteMailbox(mailbox *Mailbox) error {
	if mailbox.DisplayName == nil {
		return e.WriteSpecificAddress(mailbox.SpecificAddress)
	}

	if err := e.WriteQuotedString(*mailbox.DisplayName); err != nil {
		return fmt.Errorf("invalid display name: %w", err)
	}
	e.WriteRune(' ')

	e.WriteRune('<')
	err := e.WriteSpecificAddress(mailbox.SpecificAddress)
	if err != nil {
		return err
	}
	e.WriteRune('>')

	return nil
}

func (e *DataEncoder) WriteMailboxList(mailboxes Mailboxes) error {
	for i, mailbox := range mailboxes {
		if i > 0 {
			e.WriteString(", ")
		}

		if err := e.WriteMailbox(mailbox); err != nil {
			return fmt.Errorf("invalid mailbox %v: %w", mailbox, err)
		}
	}

	return nil
}

func (e *DataEncoder) WriteSpecificAddress(spec SpecificAddress) error {
	if err := e.WriteDotAtomOrQuotedString(spec.LocalPart); err != nil {
		return fmt.Errorf("invalid local part: %w", err)
	}

	e.WriteRune('@')

	e.WriteDomain(spec.Domain)

	return nil
}

func (e *DataEncoder) WriteGroup(group *Group) error {
	if group.DisplayName == "" {
		return fmt.Errorf("invalid empty display name")
	}

	if err := e.WritePhrase(group.DisplayName); err != nil {
		return fmt.Errorf("invalid display name: %w", err)
	}

	e.WriteString(": ")

	for i, mailbox := range group.Mailboxes {
		if i > 0 {
			e.WriteString(", ")
		}

		e.WriteMailbox(mailbox)
	}

	e.WriteRune(';')

	return nil
}

func (e *DataEncoder) WriteReceivedTokens(tokens ReceivedTokens) error {
	for i, token := range tokens {
		if i > 0 {
			e.WriteRune(' ')
		}

		switch v := token.(type) {
		case SpecificAddress:
			if err := e.WriteSpecificAddress(v); err != nil {
				return fmt.Errorf("invalid specific address: %w", err)
			}
		case Domain:
			e.WriteDomain(v)
		case string:
			e.WriteWord(v)
		default:
			utils.Panicf("unhandle received token %#v (%T)", token, token)
		}
	}

	return nil
}
