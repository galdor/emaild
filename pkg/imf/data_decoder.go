package imf

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/galdor/emaild/pkg/utils"
)

type DataDecoder struct {
	MixedEOL bool

	buf []byte
}

func NewDataDecoder(data []byte) *DataDecoder {
	return &DataDecoder{
		buf: data,
	}
}

func (d *DataDecoder) Empty() bool {
	return len(d.buf) == 0
}

func (d *DataDecoder) Try(fn func() error) error {
	buf := d.buf

	if err := fn(); err != nil {
		d.buf = buf
		return err
	}

	return nil
}

func (d *DataDecoder) Skip(n int) {
	if len(d.buf) < n {
		utils.Panicf("cannot skip %d bytes in a %d byte buffer", n, len(d.buf))
	}

	d.buf = d.buf[n:]
}

func (d *DataDecoder) StartsWithByte(c byte) bool {
	return len(d.buf) > 0 && d.buf[0] == c
}

func (d *DataDecoder) SkipByte(c byte) bool {
	if !d.StartsWithByte(c) {
		return false
	}

	d.Skip(1)
	return true
}

func (d *DataDecoder) ReadFWS() ([]byte, error) {
	var ws []byte

	for len(d.buf) > 0 {
		if IsWSP(d.buf[0]) {
			ws = append(ws, d.buf[0])

			d.Skip(1)
		} else if d.buf[0] == '\r' {
			if len(d.buf) < 3 {
				return nil, fmt.Errorf("truncated fws sequence")
			}

			if d.buf[1] != '\n' {
				return nil, fmt.Errorf("missing lf character after cr character")
			}

			if !IsWSP(d.buf[2]) {
				return nil, fmt.Errorf("invalid fws sequence: missing " +
					"whitespace after crlf sequence")
			}

			ws = append(ws, d.buf[2])

			d.Skip(3)
		} else if d.MixedEOL && d.buf[0] == '\n' {
			// In mixed EOL mode, we accept \n as an EOL sequence, meaning that
			// FWS becomes \n followed by either a WSP characted.

			if len(d.buf) < 2 {
				return nil, fmt.Errorf("truncated fws sequence")
			}

			if !IsWSP(d.buf[1]) {
				return nil, fmt.Errorf("invalid fws sequence: missing " +
					"whitespace after lf character")
			}

			ws = append(ws, d.buf[1])

			d.Skip(2)
		} else {
			break
		}
	}

	return ws, nil
}

func (d *DataDecoder) ReadCFWS() ([]byte, error) {
	var ws []byte

	for len(d.buf) > 0 {
		ws2, err := d.ReadFWS()
		if err != nil {
			return nil, err
		}

		ws = append(ws, ws2...)

		if d.StartsWithByte('(') {
			if err := d.SkipComment(0); err != nil {
				return nil, err
			}
		} else if len(ws2) == 0 {
			break
		}
	}

	return ws, nil
}

func (d *DataDecoder) SkipComment(depth int) error {
	if depth > 20 {
		return fmt.Errorf("too many nested comments")
	}

	d.SkipByte('(')

	for {
		if _, err := d.ReadFWS(); err != nil {
			return err
		}

		if len(d.buf) == 0 {
			return fmt.Errorf("truncated comment")
		}

		c := d.buf[0]

		if c == '(' {
			if err := d.SkipComment(depth + 1); err != nil {
				return err
			}
		} else if c == ')' {
			d.Skip(1)
			break
		} else if IsCommentChar(c) || IsWSP(c) {
			d.Skip(1)
		} else if c == '\\' {
			if len(d.buf) == 1 {
				return fmt.Errorf("truncated quoted pair")
			}
			d.Skip(2)
		} else {
			return fmt.Errorf("invalid comment character %s",
				utils.QuoteByte(d.buf[0]))
		}
	}

	return nil
}

func (d *DataDecoder) ReadAll() []byte {
	data := d.buf
	d.buf = nil
	return data
}

func (d *DataDecoder) ReadWhile(fn func(byte) bool) []byte {
	return d.ReadWhileN(fn, len(d.buf))
}

func (d *DataDecoder) ReadWhileN(fn func(byte) bool, maxLength int) []byte {
	limit := min(len(d.buf), maxLength)

	end := limit

	for i := 0; i < limit; i++ {
		if !fn(d.buf[i]) {
			end = i
			break
		}
	}

	data := d.buf[:end]
	d.buf = d.buf[end:]

	return data
}

func (d *DataDecoder) ReadFromChar(c byte) []byte {
	idx := bytes.LastIndexByte(d.buf, c)
	if idx == -1 {
		return nil
	}

	data := d.buf[idx+1:]
	d.buf = d.buf[:idx]

	return data
}

func (d *DataDecoder) ReadUnstructured() ([]byte, error) {
	var value []byte

	for {
		ws, err := d.ReadFWS()
		if err != nil {
			return nil, err
		}

		value = append(value, ws...)

		if len(d.buf) == 0 {
			break
		}

		value = append(value, d.buf[0])
		d.buf = d.buf[1:]
	}

	return value, nil
}

func (d *DataDecoder) ReadAtom() ([]byte, error) {
	if _, err := d.ReadCFWS(); err != nil {
		return nil, err
	}

	if len(d.buf) == 0 {
		return nil, fmt.Errorf("invalid empty value")
	}

	if !IsAtomChar(d.buf[0]) {
		return nil, fmt.Errorf("invalid character %s",
			utils.QuoteByte(d.buf[0]))
	}

	atom := d.ReadWhile(IsAtomChar)

	return atom, nil
}

func (d *DataDecoder) ReadDotAtom() ([]byte, error) {
	var atoms [][]byte

	for {
		if len(d.buf) == 0 {
			return nil, fmt.Errorf("truncated value")
		}

		if !IsAtomChar(d.buf[0]) {
			return nil, fmt.Errorf("invalid character %s",
				utils.QuoteByte(d.buf[0]))
		}

		part := d.ReadWhile(IsAtomChar)
		atoms = append(atoms, part)

		if !d.SkipByte('.') {
			break
		}
	}

	return bytes.Join(atoms, []byte{'.'}), nil
}

func (d *DataDecoder) ReadQuotedString() ([]byte, error) {
	if _, err := d.ReadCFWS(); err != nil {
		return nil, err
	}

	if !d.SkipByte('"') {
		return nil, fmt.Errorf("missing initial '\"' character")
	}

	var value []byte

	for {
		ws, err := d.ReadFWS()
		if err != nil {
			return nil, err
		}
		value = append(value, ws...)

		if len(d.buf) == 0 {
			return nil, fmt.Errorf("missing final '\"' character")
		} else if d.buf[0] == '"' {
			d.Skip(1)
			break
		} else if d.buf[0] == '\\' {
			if len(d.buf) < 2 {
				return nil, fmt.Errorf("truncated quoted pair")
			}

			// RFC 5322 3.2.1. "Where any quoted-pair appears, it is to be
			// interpreted as the character alone".
			value = append(value, d.buf[1])
			d.Skip(2)
		} else {
			value = append(value, d.buf[0])
			d.Skip(1)
		}
	}

	return value, nil
}

func (d *DataDecoder) ReadWord() ([]byte, error) {
	if _, err := d.ReadFWS(); err != nil {
		return nil, err
	}

	var value []byte
	var err error

	if d.StartsWithByte('"') {
		value, err = d.ReadQuotedString()
	} else {
		value, err = d.ReadAtom()
	}

	if err != nil {
		return nil, err
	}

	return value, nil
}

func (d *DataDecoder) ReadPhrase() ([]byte, error) {
	// Are initial and final spaces part of a phrase? In this example:
	//
	// Keywords: foo bar ,  hello world\r\n
	//
	// Is the first phrase "foo bar" or " foo bar ", and is the second phrase
	// "hello world" or " hello world"?
	//
	// RFC 5322 is very bad at specifying how syntaxic elements are combined
	// into semantic elements. Here the logical answer is that initial and final
	// spaces are not part of a phrase. If they were, then the following field:
	//
	// From: Bob Howard  <bob@example.com>\r\n
	//
	// Would yield the display name " Bob Howard ", which is absolutely not what
	// anyone sane has in mind.

	var phrase []byte

	word, err := d.ReadWord()
	if err != nil {
		return nil, fmt.Errorf("invalid word: %w", err)
	}
	phrase = append(phrase, word...)

	for {
		// Careful, we only want to include this whitespace if there is a valid
		// phrase element ('.' character or word) afterward.
		ws, err := d.ReadCFWS()
		if err != nil {
			return nil, err
		}

		if len(d.buf) == 0 {
			break
		}

		// With obsolete syntax, a phrase element can be a single '.' character
		// which is not a valid word, so we have to handle it separately.
		if d.buf[0] == '.' {
			phrase = append(phrase, ws...)
			phrase = append(phrase, '.')
			d.Skip(1)
			continue
		}

		var finished bool

		err = d.Try(func() error {
			word, err := d.ReadWord()
			if err != nil {
				finished = true
				return nil
			}

			phrase = append(phrase, ws...)
			phrase = append(phrase, word...)
			return nil
		})
		if err != nil {
			return nil, err
		}

		if finished {
			break
		}
	}

	return phrase, nil
}

func (d *DataDecoder) ReadPhraseList() ([]string, error) {
	var phrases []string

	for {
		if _, err := d.ReadCFWS(); err != nil {
			return nil, err
		}

		if len(d.buf) == 0 {
			break
		}

		if d.SkipByte(',') {
			// Obsolete syntax allows empty list elements
			continue
		}

		phrase, err := d.ReadPhrase()
		if err != nil {
			return nil, fmt.Errorf("invalid phase: %w", err)
		}

		phrases = append(phrases, string(phrase))

		if _, err := d.ReadCFWS(); err != nil {
			return nil, err
		}

		if !d.SkipByte(',') {
			break
		}
	}

	if len(phrases) == 0 {
		return nil, fmt.Errorf("empty list")
	}

	return phrases, nil
}

func (d *DataDecoder) ReadLocalPart() ([]byte, error) {
	// Since we have to support obsolete syntax, we accept any sequence of one
	// or more words (each word being either an atom or a quoted string)
	// separated by a '.' characted.

	var value []byte

	for len(d.buf) > 0 {
		if _, err := d.ReadCFWS(); err != nil {
			return nil, err
		}

		word, err := d.ReadWord()
		if err != nil {
			return nil, err
		}
		value = append(value, word...)

		if _, err := d.ReadCFWS(); err != nil {
			return nil, err
		}

		if !d.SkipByte('.') {
			break
		}
		value = append(value, '.')
	}

	return value, nil
}

func (d *DataDecoder) ReadDomainLiteral() ([]byte, error) {
	if _, err := d.ReadCFWS(); err != nil {
		return nil, err
	}

	if !d.SkipByte('[') {
		return nil, fmt.Errorf("missing initial '[' character")
	}

	domain := []byte{'['}

	if d.StartsWithByte('"') {
		value, err := d.ReadQuotedString()
		if err != nil {
			return nil, err
		}

		domain = append(domain, value...)
	} else {
		for {
			ws, err := d.ReadFWS()
			if err != nil {
				return nil, err
			}
			domain = append(domain, ws...)

			if len(d.buf) == 0 || d.buf[0] == ']' {
				break
			}

			if !IsDomainLiteralChar(d.buf[0]) {
				return nil, fmt.Errorf("invalid character %s in domain "+
					"literal", utils.QuoteByte(d.buf[0]))
			}

			domain = append(domain, d.buf[0])

			d.buf = d.buf[1:]
		}
	}

	if !d.SkipByte(']') {
		return nil, fmt.Errorf("missing final ']' character")
	}

	domain = append(domain, ']')

	return domain, nil
}

func (d *DataDecoder) ReadDomain() (*Domain, error) {
	// Start by looking for a domain-literal. Note that it can contain a
	// quoted string.

	if _, err := d.ReadCFWS(); err != nil {
		return nil, err
	}

	if d.StartsWithByte('[') {
		domain, err := d.ReadDomainLiteral()
		if err != nil {
			return nil, fmt.Errorf("invalid domain literal: %w", err)
		}

		// Surprisingly, a domain literal can be empty (see RFC 5322 3.4.1.
		// Addr-Spec Specification). It is not clear what it is supposed to
		// represent.

		return utils.Ref(Domain(domain)), nil
	}

	// If it is not a domain-literal, it is a dot-atom

	domain, err := d.ReadDotAtom()
	if err != nil {
		return nil, fmt.Errorf("invalid dot atom: %w", err)
	}

	return utils.Ref(Domain(domain)), nil
}

func (d *DataDecoder) ReadSpecificAddress() (*SpecificAddress, error) {
	localPart, err := d.ReadLocalPart()
	if err != nil {
		return nil, fmt.Errorf("invalid local part: %w", err)
	}

	if _, err := d.ReadCFWS(); err != nil {
		return nil, err
	}

	if !d.SkipByte('@') {
		return nil, fmt.Errorf("missing '@' character after local part")
	}

	if _, err := d.ReadCFWS(); err != nil {
		return nil, err
	}

	domain, err := d.ReadDomain()
	if err != nil {
		return nil, fmt.Errorf("invalid domain: %w", err)
	}

	spec := SpecificAddress{
		LocalPart: string(localPart),
		Domain:    *domain,
	}

	return &spec, nil
}

func (d *DataDecoder) ReadAngleAddress(allowEmpty bool) (*SpecificAddress, error) {
	if _, err := d.ReadCFWS(); err != nil {
		return nil, err
	}

	if !d.SkipByte('<') {
		return nil, fmt.Errorf("missing '<' character before specific " +
			"address")
	}

	if _, err := d.ReadCFWS(); err != nil {
		return nil, err
	}

	if allowEmpty && d.StartsWithByte('>') {
		d.Skip(1)
		return nil, nil
	}

	// TODO obs-route

	if _, err := d.ReadCFWS(); err != nil {
		return nil, err
	}

	spec, err := d.ReadSpecificAddress()
	if err != nil {
		return nil, fmt.Errorf("invalid specific address: %w", err)
	}

	if _, err := d.ReadCFWS(); err != nil {
		return nil, err
	}

	if !d.SkipByte('>') {
		return nil, fmt.Errorf("missing '>' character after specific " +
			"address")
	}

	return spec, nil
}

func (d *DataDecoder) ReadNamedAddress() (*Mailbox, error) {
	if _, err := d.ReadCFWS(); err != nil {
		return nil, err
	}

	var displayName []byte

	if !d.StartsWithByte('<') {
		phrase, err := d.ReadPhrase()
		if err != nil {
			return nil, fmt.Errorf("invalid display name: %w", err)
		}

		displayName = phrase
	}

	if _, err := d.ReadCFWS(); err != nil {
		return nil, err
	}

	spec, err := d.ReadAngleAddress(false)
	if err != nil {
		return nil, fmt.Errorf("invalid angle address: %w", err)
	}

	mb := Mailbox{
		SpecificAddress: *spec,
		DisplayName:     utils.Ref(string(displayName)),
	}

	return &mb, nil
}

func (d *DataDecoder) ReadMailbox() (*Mailbox, error) {
	// A mailbox is either a named address (in which case it starts with either
	// a display name or an angle address) or a specific address (in which case
	// it starts with a local part). The problem is that we cannot differentiate
	// them with a simple character lookup. For example both "Bob
	// <bob@example.com>" (named address) and "Bob@example.com" (specific
	// address) start with atom characters.
	//
	// So we try to read a specific address first, and on failure fallback to
	// reading a named address.

	var mailbox *Mailbox

	err := d.Try(func() error {
		addr, err := d.ReadSpecificAddress()
		if err != nil {
			return err
		}

		mailbox = &Mailbox{SpecificAddress: *addr}
		return nil
	})
	if err != nil {
		mb, err := d.ReadNamedAddress()
		if err != nil {
			return nil, fmt.Errorf("invalid named address: %w", err)
		}

		mailbox = mb
	}

	return mailbox, nil
}

func (d *DataDecoder) ReadMailboxList(allowEmpty bool) ([]*Mailbox, error) {
	var mailboxes []*Mailbox

	for {
		if _, err := d.ReadCFWS(); err != nil {
			return nil, err
		}

		if len(d.buf) == 0 || d.StartsWithByte(';') {
			break
		}

		if d.SkipByte(',') {
			// Obsolete syntax allows empty list elements
			continue
		}

		mailbox, err := d.ReadMailbox()
		if err != nil {
			return nil, fmt.Errorf("invalid mailbox: %w", err)
		}

		mailboxes = append(mailboxes, mailbox)

		if _, err := d.ReadCFWS(); err != nil {
			return nil, err
		}

		if !d.SkipByte(',') {
			break
		}
	}

	if len(mailboxes) == 0 && !allowEmpty {
		return nil, fmt.Errorf("empty list")
	}

	return mailboxes, nil
}

func (d *DataDecoder) ReadGroup() (*Group, error) {
	if _, err := d.ReadCFWS(); err != nil {
		return nil, err
	}

	displayName, err := d.ReadPhrase()
	if err != nil {
		return nil, fmt.Errorf("invalid display name: %w", err)
	}

	if _, err := d.ReadCFWS(); err != nil {
		return nil, err
	}

	if !d.SkipByte(':') {
		return nil, fmt.Errorf("missing ':' character after display name")
	}

	if _, err := d.ReadCFWS(); err != nil {
		return nil, err
	}

	// A group list can be a mailbox list (in which case it must contain at
	// least one mailbox), but it can also be a single CFWS element. We handle
	// both by reading a mailbox list which can be empty.
	mailboxes, err := d.ReadMailboxList(true)
	if err != nil {
		return nil, fmt.Errorf("invalid mailbox list: %w", err)
	}

	if _, err := d.ReadCFWS(); err != nil {
		return nil, err
	}

	if !d.SkipByte(';') {
		return nil, fmt.Errorf("missing ';' character after mailbox list")
	}

	group := Group{
		DisplayName: string(displayName),
		Mailboxes:   mailboxes,
	}

	return &group, nil
}

func (d *DataDecoder) ReadAddress() (Address, error) {
	// An address is either a mailbox or a group, and again we cannot
	// differentiate them with a simple lookup.

	var addr Address

	err := d.Try(func() error {
		group, err := d.ReadGroup()
		if err != nil {
			return err
		}

		addr = group
		return nil
	})
	if err != nil {
		mailbox, err := d.ReadMailbox()
		if err != nil {
			return nil, fmt.Errorf("invalid mailbox: %w", err)
		}

		addr = mailbox
	}

	return addr, nil
}

func (d *DataDecoder) ReadAddressList(allowEmpty bool) ([]Address, error) {
	var addrs []Address

	for {
		if _, err := d.ReadCFWS(); err != nil {
			return nil, err
		}

		if len(d.buf) == 0 {
			break
		}

		if d.SkipByte(',') {
			// Obsolete syntax allows empty list elements
			continue
		}

		addr, err := d.ReadAddress()
		if err != nil {
			return nil, fmt.Errorf("invalid address: %w", err)
		}

		addrs = append(addrs, addr)

		if _, err := d.ReadCFWS(); err != nil {
			return nil, err
		}

		if !d.SkipByte(',') {
			break
		}
	}

	if len(addrs) == 0 && !allowEmpty {
		return nil, fmt.Errorf("empty list")
	}

	return addrs, nil
}

func (d *DataDecoder) ReadMessageId() (*MessageId, error) {
	if _, err := d.ReadCFWS(); err != nil {
		return nil, err
	}

	// Start delimiter
	if !d.SkipByte('<') {
		return nil, fmt.Errorf("missing initial '<' character")
	}

	// Left part
	if _, err := d.ReadCFWS(); err != nil {
		return nil, err
	}

	left, err := d.ReadLocalPart()
	if err != nil {
		return nil, fmt.Errorf("invalid left part: %w", err)
	}

	// Separator
	if _, err := d.ReadCFWS(); err != nil {
		return nil, err
	}

	if !d.SkipByte('@') {
		return nil, fmt.Errorf("missing '@' character after left part")
	}

	// Right part
	if _, err := d.ReadCFWS(); err != nil {
		return nil, err
	}

	right, err := d.ReadDomain()
	if err != nil {
		return nil, fmt.Errorf("invalid right part: %w", err)
	}

	// End delimiter
	if _, err := d.ReadCFWS(); err != nil {
		return nil, err
	}

	if !d.SkipByte('>') {
		return nil, fmt.Errorf("missing final '>' character")
	}

	id := MessageId{
		Left:  string(left),
		Right: *right,
	}

	return &id, nil
}

func (d *DataDecoder) ReadMessageIdList(allowEmpty bool) (MessageIds, error) {
	// Unlike other lists, message id lists do not use a comma separator

	var ids MessageIds

	for {
		if _, err := d.ReadCFWS(); err != nil {
			return nil, err
		}

		if len(d.buf) == 0 {
			break
		}

		id, err := d.ReadMessageId()
		if err != nil {
			return nil, fmt.Errorf("invalid message id: %w", err)
		}

		ids = append(ids, *id)
	}

	if len(ids) == 0 && !allowEmpty {
		return nil, fmt.Errorf("empty list")
	}

	return ids, nil
}

func (d *DataDecoder) ReadDateTime() (*time.Time, error) {
	// Optional day name
	if _, err := d.ReadCFWS(); err != nil {
		return nil, err
	}

	if len(d.buf) == 0 {
		return nil, fmt.Errorf("empty value")
	}

	if IsAlphaChar(d.buf[0]) {
		if _, err := d.MaybeReadDayName(); err != nil {
			return nil, fmt.Errorf("invalid day name: %w", err)
		}
	}

	// Day
	day, err := d.ReadDay()
	if err != nil {
		return nil, fmt.Errorf("invalid day: %w", err)
	}

	// Month
	month, err := d.ReadMonth()
	if err != nil {
		return nil, fmt.Errorf("invalid month: %w", err)
	}

	// Year
	year, err := d.ReadYear()
	if err != nil {
		return nil, fmt.Errorf("invalid year: %w", err)
	}

	// Hour
	hour, err := d.ReadHour()
	if err != nil {
		return nil, fmt.Errorf("invalid hour: %w", err)
	}

	// Separator
	if _, err := d.ReadCFWS(); err != nil {
		return nil, err
	}

	if !d.SkipByte(':') {
		return nil, fmt.Errorf("missing ':' character after hour")
	}

	// Minute
	minute, err := d.ReadMinute()
	if err != nil {
		return nil, fmt.Errorf("invalid minute: %w", err)
	}

	// Separator and optional second
	if _, err := d.ReadCFWS(); err != nil {
		return nil, err
	}

	var second int
	if d.SkipByte(':') {
		second, err = d.ReadSecond()
		if err != nil {
			return nil, fmt.Errorf("invalid second: %w", err)
		}
	}

	// Timezone
	loc, err := d.ReadTimezone()
	if err != nil {
		return nil, fmt.Errorf("invalid timezone: %w", err)
	}

	date := time.Date(year, month, day, hour, minute, second, 0, loc)
	return &date, nil
}

func (d *DataDecoder) MaybeReadDayName() (string, error) {
	// This function is called after checking that the buffer starts with an
	// alpha character, so we do not have to skip anything or check that the
	// result is not empty.

	name := d.ReadWhile(IsAlphaChar)

	if _, err := d.ReadCFWS(); err != nil {
		return "", err
	}

	if !d.SkipByte(',') {
		return "", fmt.Errorf("missing ',' character after day name")
	}

	return string(name), nil
}

func (d *DataDecoder) ReadInteger(maxNbDigits int, minValue, maxValue int64) (int, error) {
	if _, err := d.ReadCFWS(); err != nil {
		return 0, err
	}

	if len(d.buf) == 0 {
		return 0, fmt.Errorf("empty value")
	}

	if !IsDigitChar(d.buf[0]) {
		return 0, fmt.Errorf("invalid character %s", utils.QuoteByte(d.buf[0]))
	}

	s := string(d.ReadWhileN(IsDigitChar, maxNbDigits))

	i64, err := strconv.ParseInt(s, 10, 64)
	if err != nil || i64 < minValue || i64 > maxValue {
		return 0, fmt.Errorf("invalid value %q", s)
	}

	return int(i64), nil
}

func (d *DataDecoder) ReadDay() (int, error) {
	return d.ReadInteger(2, 1, 31)
}

func (d *DataDecoder) ReadMonth() (time.Month, error) {
	if _, err := d.ReadCFWS(); err != nil {
		return 0, err
	}

	if len(d.buf) == 0 {
		return 0, fmt.Errorf("empty value")
	}

	if !IsAlphaChar(d.buf[0]) {
		return 0, fmt.Errorf("invalid character %s", utils.QuoteByte(d.buf[0]))
	}

	s := string(d.ReadWhileN(IsAlphaChar, 3))

	var month time.Month

	switch strings.ToLower(s) {
	case "jan":
		month = time.January
	case "feb":
		month = time.February
	case "mar":
		month = time.March
	case "apr":
		month = time.April
	case "may":
		month = time.May
	case "jun":
		month = time.June
	case "jul":
		month = time.July
	case "aug":
		month = time.August
	case "sep":
		month = time.September
	case "oct":
		month = time.October
	case "nov":
		month = time.November
	case "dec":
		month = time.December
	default:
		return 0, fmt.Errorf("invalid value %q", s)
	}

	return month, nil
}

func (d *DataDecoder) ReadYear() (int, error) {
	if _, err := d.ReadCFWS(); err != nil {
		return 0, err
	}

	if len(d.buf) == 0 {
		return 0, fmt.Errorf("empty value")
	}

	if !IsDigitChar(d.buf[0]) {
		return 0, fmt.Errorf("invalid character %s", utils.QuoteByte(d.buf[0]))
	}

	s := string(d.ReadWhileN(IsDigitChar, 4))

	nbDigits := len(s)
	if nbDigits < 2 {
		return 0, fmt.Errorf("invalid year %q", s)
	}

	// RFC 5322 4.3. Obsolete Date and Time
	//
	// Where a two or three digit year occurs in a date, the year is to be
	// interpreted as follows: If a two digit year is encountered whose value is
	// between 00 and 49, the year is interpreted by adding 2000, ending up with
	// a value between 2000 and 2049. If a two digit year is encountered with a
	// value between 50 and 99, or any three digit year is encountered, the year
	// is interpreted by adding 1900.

	i64, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid value %q: %w", s, err)
	}

	i := int(i64)

	var year int

	switch {
	case nbDigits == 2 && i <= 49:
		year = 2000 + i

	case nbDigits == 2 && i >= 50 || nbDigits == 3:
		year = 1900 + i

	case nbDigits == 4:
		year = i
	}

	return year, nil
}

func (d *DataDecoder) ReadHour() (int, error) {
	return d.ReadInteger(2, 0, 23)
}

func (d *DataDecoder) ReadMinute() (int, error) {
	return d.ReadInteger(2, 0, 59)
}

func (d *DataDecoder) ReadSecond() (int, error) {
	// Yes, 60, leap seconds are a thing
	return d.ReadInteger(2, 0, 60)
}

func (d *DataDecoder) ReadTimezone() (*time.Location, error) {
	if _, err := d.ReadCFWS(); err != nil {
		return nil, err
	}

	if len(d.buf) == 0 {
		return nil, fmt.Errorf("empty value")
	}

	var loc *time.Location

	if d.buf[0] == '+' || d.buf[0] == '-' {
		// Timezone offset

		sign := 1
		if d.buf[0] == '-' {
			sign = -1
		}

		d.Skip(1)

		maxOffset := int64(12)
		if sign == 1 {
			maxOffset = 14 // Line Islands
		}

		i, err := d.ReadInteger(2, 0, maxOffset)
		if err != nil {
			return nil, fmt.Errorf("invalid timezone hour offset: %w", err)
		}
		hourOffset := sign * i

		i, err = d.ReadInteger(2, 0, 59)
		if err != nil {
			return nil, fmt.Errorf("invalid timezone minute offset: %w", err)
		}
		minuteOffset := i

		loc = time.FixedZone("", hourOffset*3600+minuteOffset*60)
	} else if IsAlphaChar(d.buf[0]) {
		// Timezone name
		//
		// See RFC 5322 4.3. Obsolete Date and Time. Zone names can be up to 5
		// character long. For military time zones (single letter zone names),
		// "they SHOULD all be considered equivalent to "-0000" unless there is
		// out-of-band information confirming their meaning". Unknown timezone
		// names should also be considered equivalent to "-0000".

		s := string(d.ReadWhileN(IsAlphaChar, 5))
		name := strings.ToUpper(s)

		var hourOffset int

		switch {
		case name == "UT" || name == "GMT":
			hourOffset = 0
		case name == "EST":
			hourOffset = -5
		case name == "EDT":
			hourOffset = -4
		case name == "CST":
			hourOffset = -6
		case name == "CDT":
			hourOffset = -5
		case name == "MST":
			hourOffset = -7
		case name == "MDT":
			hourOffset = -6
		case name == "PST":
			hourOffset = -8
		case name == "PDT":
			hourOffset = -7
		case len(name) == 1:
			hourOffset = 0
		default:
			hourOffset = 0
		}

		loc = time.FixedZone("", hourOffset*3600)
	} else {
		return nil, fmt.Errorf("invalid character %s",
			utils.QuoteByte(d.buf[0]))
	}

	return loc, nil
}

func (d *DataDecoder) ReadReceivedTokens() (ReceivedTokens, error) {
	var tokens ReceivedTokens

	for {
		if _, err := d.ReadCFWS(); err != nil {
			return nil, err
		}

		if len(d.buf) == 0 || d.StartsWithByte(';') {
			break
		}

		err1 := d.Try(func() error {
			addr, err := d.ReadAngleAddress(false)
			if err != nil {
				return err
			}

			tokens = append(tokens, *addr)
			return nil
		})
		if err1 == nil {
			continue
		}

		err2 := d.Try(func() error {
			addr, err := d.ReadSpecificAddress()
			if err != nil {
				return err
			}

			tokens = append(tokens, *addr)
			return nil
		})
		if err2 == nil {
			continue
		}

		err3 := d.Try(func() error {
			domain, err := d.ReadDomain()
			if err != nil {
				return err
			}

			tokens = append(tokens, *domain)
			return nil
		})
		if err3 == nil {
			continue
		}

		word, err := d.ReadWord()
		if err != nil {
			return nil, fmt.Errorf("invalid word: %w", err)
		}
		tokens = append(tokens, string(word))
	}

	return tokens, nil
}
