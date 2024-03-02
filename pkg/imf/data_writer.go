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

type DataWriter struct {
	MaxLineLength int // 0 if no maximum line length

	buf        bytes.Buffer
	lineLength int
}

func MustWriteInlineData(fn func(*DataWriter) error) string {
	w := NewDataWriter(0)

	if err := fn(w); err != nil {
		utils.Panicf("cannot write data: %w", err)
	}

	return w.buf.String()
}

func (w *DataWriter) Bytes() []byte {
	return w.buf.Bytes()
}

func NewDataWriter(maxLineLength int) *DataWriter {
	return &DataWriter{
		MaxLineLength: maxLineLength,
	}
}

func (w *DataWriter) WriteRune(c rune) {
	sz := utf8.RuneLen(c)

	if w.MaxLineLength > 0 && w.lineLength+sz > w.MaxLineLength {
		w.buf.WriteString("\r\n ")
		w.lineLength = 1
	}

	w.buf.WriteRune(c)
	w.lineLength += sz
}

func (w *DataWriter) WriteString(s string) {
	if w.MaxLineLength > 0 && w.lineLength+len(s) > w.MaxLineLength {
		w.buf.WriteString("\r\n ")
		w.lineLength = 1
	}

	w.buf.WriteString(s)
	w.lineLength += len(s)
}

func (w *DataWriter) WriteHeader(header []*Field) error {
	for _, field := range header {
		w.WriteField(field)
	}

	return nil
}

func (w *DataWriter) WriteField(f *Field) error {
	// TODO Raw
	w.WriteString(f.Name)
	w.WriteString(": ")

	if err := f.Value.Write(w); err != nil {
		return fmt.Errorf("invalid value: %w", err)
	}

	w.WriteEOL()

	return nil
}

func (w *DataWriter) WriteEOL() {
	w.buf.WriteString("\r\n")
	w.lineLength = 0
}

func (w *DataWriter) WriteUnstructured(s string) {
	for _, c := range s {
		w.WriteRune(c)
	}
}

func (w *DataWriter) WriteQuotedString(s string) error {
	// All printable ASCII characters are allowed; space and backslash
	// characters must be escaped with a single backslash. UTF-8 sequences are
	// allowed (see RFC 6532 Internationalized Email Headers).
	//
	// Quoted strings can be folded. We can split between characters as long as
	// we do not split quoted pairs (e.g. "\\") or UTF-8 sequences.

	w.WriteRune('"')

	for _, c := range s {
		switch {
		case c < 33 && c != ' ' && c != '\t':
			return fmt.Errorf("unencodable control character 0x%x", c)
		case c == '"' || c == '\\':
			w.WriteString("\\" + string(c))
		default:
			w.WriteRune(c)
		}
	}

	w.WriteRune('"')

	return nil
}

func (w *DataWriter) WriteAtomOrQuotedString(s string) error {
	if IsAtom(s) {
		w.WriteString(s)
		return nil
	}

	return w.WriteQuotedString(s)
}

func (w *DataWriter) WriteDotAtomOrQuotedString(s string) error {
	if IsDotAtom(s) {
		w.WriteString(s)
		return nil
	}

	return w.WriteQuotedString(s)
}

func (w *DataWriter) WriteDomain(domain string) {
	if IsDotAtom(domain) {
		w.WriteString(domain)
	} else {
		w.WriteRune('[')
		w.WriteString(domain)
		w.WriteRune(']')
	}
}

func (w *DataWriter) WritePhrase(phrase string) error {
	// Phrases are defined as a list of at least one word. The correct
	// representation would be []string, but it is inconvenient. So we extract
	// words from the string ourselves then encode them.

	words := phraseWordSeparatorRE.Split(phrase, -1)
	for i, word := range words {
		if i > 0 {
			w.WriteRune(' ')
		}

		if err := w.WriteAtomOrQuotedString(word); err != nil {
			return fmt.Errorf("invalid word %q: %w", word, err)
		}
	}

	return nil
}

func (w *DataWriter) WritePhraseList(phrases []string) error {
	for i, phrase := range phrases {
		if i > 0 {
			w.WriteString(", ")
		}

		if err := w.WritePhrase(phrase); err != nil {
			return fmt.Errorf("invalid phrase %q: %w", phrase, err)
		}
	}

	return nil
}

func (w *DataWriter) WriteDateTime(date time.Time) {
	// RFC 5322 3.3. Date and Time Specification

	w.WriteString(date.Format("Mon"))
	w.WriteString(", ")
	w.WriteString(date.Format("02 Jan 2006"))
	w.WriteRune(' ')
	w.WriteString(date.Format("15:04:05"))
	w.WriteRune(' ')
	w.WriteString(date.Format("-0700"))
}

func (w *DataWriter) WriteMessageId(id MessageId) error {
	// RFC 5322 3.6.4. Identification Fields

	w.WriteRune('<')
	w.WriteDotAtomOrQuotedString(id.Left)
	w.WriteRune('@')
	w.WriteDomain(id.Right)
	w.WriteRune('>')

	return nil
}

func (w *DataWriter) WriteMessageIdList(ids MessageIds) error {
	for i, id := range ids {
		if i > 0 {
			w.WriteRune(' ')
		}

		if err := w.WriteMessageId(id); err != nil {
			return fmt.Errorf("invalid message id %v: %w", id, err)
		}
	}

	return nil
}

func (w *DataWriter) WriteAddress(addr Address) error {
	switch addr2 := addr.(type) {
	case *Mailbox:
		return w.WriteMailbox(addr2)
	case *Group:
		return w.WriteGroup(addr2)
	}

	utils.Panicf("unhandled address %#v (%T)", addr, addr)
	return nil // the Go compiler still cannot do basic flow analysis...
}

func (w *DataWriter) WriteAddressList(addrs Addresses) error {
	for i, addr := range addrs {
		if i > 0 {
			w.WriteString(", ")
		}

		if err := w.WriteAddress(addr); err != nil {
			return fmt.Errorf("invalid address %v: %w", addr, err)
		}
	}

	return nil
}

func (w *DataWriter) WriteMailbox(mailbox *Mailbox) error {
	if mailbox.DisplayName == "" {
		return w.WriteSpecificAddress(mailbox.SpecificAddress)
	}

	if err := w.WriteQuotedString(mailbox.DisplayName); err != nil {
		return fmt.Errorf("invalid display name: %w", err)
	}
	w.WriteRune(' ')

	w.WriteRune('<')
	err := w.WriteSpecificAddress(mailbox.SpecificAddress)
	if err != nil {
		return err
	}
	w.WriteRune('>')

	return nil
}

func (w *DataWriter) WriteMailboxList(mailboxes Mailboxes) error {
	for i, mailbox := range mailboxes {
		if i > 0 {
			w.WriteString(", ")
		}

		if err := w.WriteMailbox(mailbox); err != nil {
			return fmt.Errorf("invalid mailbox %v: %w", mailbox, err)
		}
	}

	return nil
}

func (w *DataWriter) WriteSpecificAddress(spec SpecificAddress) error {
	if err := w.WriteDotAtomOrQuotedString(spec.LocalPart); err != nil {
		return fmt.Errorf("invalid local part: %w", err)
	}

	w.WriteRune('@')

	w.WriteString(spec.Domain)

	return nil
}

func (w *DataWriter) WriteGroup(group *Group) error {
	if group.DisplayName == "" {
		return fmt.Errorf("invalid empty display name")
	}

	if err := w.WritePhrase(group.DisplayName); err != nil {
		return fmt.Errorf("invalid display name: %w", err)
	}

	w.WriteString(": ")

	for i, mailbox := range group.Mailboxes {
		if i > 0 {
			w.WriteString(", ")
		}

		w.WriteMailbox(mailbox)
	}

	w.WriteRune(';')

	return nil
}
