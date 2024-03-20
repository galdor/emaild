package imf

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/galdor/emaild/pkg/utils"
)

type DataReader struct {
	buf []byte
}

func NewDataReader(data []byte) *DataReader {
	r := DataReader{
		buf: data,
	}

	return &r
}

func (r *DataReader) Empty() bool {
	return len(r.buf) == 0
}

func (r *DataReader) Try(fn func() error) error {
	buf := r.buf

	if err := fn(); err != nil {
		r.buf = buf
		return err
	}

	return nil
}

func (r *DataReader) Skip(n int) {
	if len(r.buf) < n {
		utils.Panicf("cannot skip %d bytes in a %d byte buffer", n, len(r.buf))
	}

	r.buf = r.buf[n:]
}

func (r *DataReader) StartsWithByte(c byte) bool {
	return len(r.buf) > 0 && r.buf[0] == c
}

func (r *DataReader) SkipByte(c byte) bool {
	if !r.StartsWithByte(c) {
		return false
	}

	r.Skip(1)
	return true
}

func (r *DataReader) MaybeSkipFWS() error {
	for len(r.buf) > 0 {
		if IsWSP(r.buf[0]) {
			r.Skip(1)
		} else if r.buf[0] == '\r' {
			if len(r.buf) < 3 {
				return fmt.Errorf("truncated fws sequence")
			}

			if r.buf[1] != '\n' {
				return fmt.Errorf("missing lf character after cr character")
			}

			if !IsWSP(r.buf[2]) {
				return fmt.Errorf("invalid fws sequence: missing whitespace after " +
					"crlf sequence")
			}

			r.Skip(3)
		} else {
			break
		}
	}

	return nil
}

func (r *DataReader) SkipCFWS() error {
	for len(r.buf) > 0 {
		if !IsWSP(r.buf[0]) && r.buf[0] != '\r' && r.buf[0] != '(' {
			break
		}

		if err := r.MaybeSkipFWS(); err != nil {
			return err
		}

		if r.StartsWithByte('(') {
			if err := r.SkipComment(0); err != nil {
				return err
			}
		}
	}

	return nil
}

func (r *DataReader) SkipComment(depth int) error {
	if depth > 20 {
		return fmt.Errorf("too many nested comments")
	}

	r.SkipByte('(')

	for {
		if err := r.MaybeSkipFWS(); err != nil {
			return err
		}

		if len(r.buf) == 0 {
			return fmt.Errorf("truncated comment")
		}

		c := r.buf[0]

		if c == '(' {
			if err := r.SkipComment(depth + 1); err != nil {
				return err
			}
		} else if c == ')' {
			r.Skip(1)
			break
		} else if IsCommentChar(c) || IsWSP(c) {
			r.Skip(1)
		} else if c == '\\' {
			if len(r.buf) == 1 {
				return fmt.Errorf("truncated quoted pair")
			}
			r.Skip(2)
		} else {
			return fmt.Errorf("invalid character %s", utils.QuoteByte(r.buf[0]))
		}
	}

	return nil
}

func (r *DataReader) ReadAll() []byte {
	data := r.buf
	r.buf = nil
	return data
}

func (r *DataReader) ReadWhile(fn func(byte) bool) []byte {
	return r.ReadWhileN(fn, len(r.buf))
}

func (r *DataReader) ReadWhileN(fn func(byte) bool, maxLength int) []byte {
	limit := min(len(r.buf), maxLength)

	end := limit

	for i := 0; i < limit; i++ {
		if !fn(r.buf[i]) {
			end = i
			break
		}
	}

	data := r.buf[:end]
	r.buf = r.buf[end:]

	return data
}

func (r *DataReader) ReadFromChar(c byte) []byte {
	idx := bytes.LastIndexByte(r.buf, c)
	if idx == -1 {
		return nil
	}

	data := r.buf[idx+1:]
	r.buf = r.buf[:idx]

	return data
}

func (r *DataReader) ReadUnstructured() ([]byte, error) {
	var value []byte

	for {
		if err := r.MaybeSkipFWS(); err != nil {
			return nil, err
		}

		if len(r.buf) == 0 {
			break
		}

		value = append(value, r.buf[0])
		r.buf = r.buf[1:]
	}

	return value, nil
}

func (r *DataReader) ReadAtom() ([]byte, error) {
	if err := r.SkipCFWS(); err != nil {
		return nil, err
	}

	if len(r.buf) == 0 {
		return nil, fmt.Errorf("invalid empty value")
	}

	if !IsAtomChar(r.buf[0]) {
		return nil, fmt.Errorf("invalid character %s",
			utils.QuoteByte(r.buf[0]))
	}

	atom := r.ReadWhile(IsAtomChar)

	return atom, nil
}

func (r *DataReader) ReadDotAtom() ([]byte, error) {
	var atoms [][]byte

	for {
		if len(r.buf) == 0 {
			return nil, fmt.Errorf("truncated value")
		}

		if !IsAtomChar(r.buf[0]) {
			return nil, fmt.Errorf("invalid character %s",
				utils.QuoteByte(r.buf[0]))
		}

		part := r.ReadWhile(IsAtomChar)
		atoms = append(atoms, part)

		if !r.SkipByte('.') {
			break
		}
	}

	return bytes.Join(atoms, []byte{'.'}), nil
}

func (r *DataReader) ReadQuotedString() ([]byte, error) {
	if err := r.SkipCFWS(); err != nil {
		return nil, err
	}

	if !r.SkipByte('"') {
		return nil, fmt.Errorf("missing initial '\"' character")
	}

	var value []byte

	for {
		if err := r.MaybeSkipFWS(); err != nil {
			return nil, err
		}

		if len(r.buf) == 0 {
			return nil, fmt.Errorf("missing final '\"' character")
		} else if r.buf[0] == '"' {
			r.Skip(1)
			break
		} else if r.buf[0] == '\\' {
			if len(r.buf) < 2 {
				return nil, fmt.Errorf("truncated quoted pair")
			}

			// RFC 5322 3.2.1. "Where any quoted-pair appears, it is to be
			// interpreted as the character alone".
			value = append(value, r.buf[1])
			r.Skip(2)
		} else {
			value = append(value, r.buf[0])
			r.Skip(1)
		}
	}

	return value, nil
}

func (r *DataReader) ReadWord() ([]byte, error) {
	if err := r.SkipCFWS(); err != nil {
		return nil, err
	}

	var value []byte
	var err error

	if r.StartsWithByte('"') {
		value, err = r.ReadQuotedString()
	} else {
		value, err = r.ReadAtom()
	}

	if err != nil {
		return nil, err
	}

	return value, nil
}

func (r *DataReader) ReadPhrase() ([]byte, error) {
	var value []byte

	appendWord := func(word []byte) {
		if len(word) == 0 {
			return
		}

		if len(value) > 0 {
			value = append(value, ' ')
		}

		value = append(value, word...)
	}

	for {
		if err := r.SkipCFWS(); err != nil {
			return nil, err
		}

		if len(r.buf) == 0 {
			break
		}

		// With obsolete syntax, a phrase element can be a single '.' character
		// which is not a valid word, so we have to handle it separately.
		if r.buf[0] == '.' {
			appendWord([]byte{'.'})
			r.Skip(1)
		}

		var finished bool

		err := r.Try(func() error {
			word, err := r.ReadWord()
			if err != nil {
				// Phrases must contain at least one word
				if len(value) == 0 {
					return fmt.Errorf("invalid word: %w", err)
				}

				finished = true
			}

			// Note that a word can be empty if it is a quoted string
			appendWord(word)

			return nil
		})
		if err != nil {
			return nil, err
		}

		if finished {
			break
		}
	}

	return value, nil
}

func (r *DataReader) ReadPhraseList() ([]string, error) {
	var phrases []string

	for {
		if err := r.SkipCFWS(); err != nil {
			return nil, err
		}

		if len(r.buf) == 0 {
			break
		}

		if r.SkipByte(',') {
			// Obsolete syntax allows empty list elements
			continue
		}

		phrase, err := r.ReadPhrase()
		if err != nil {
			return nil, fmt.Errorf("invalid phase: %w", err)
		}

		phrases = append(phrases, string(phrase))

		if err := r.SkipCFWS(); err != nil {
			return nil, err
		}

		if !r.SkipByte(',') {
			break
		}
	}

	if len(phrases) == 0 {
		return nil, fmt.Errorf("empty list")
	}

	return phrases, nil
}

func (r *DataReader) ReadLocalPart() ([]byte, error) {
	// Since we have to support obsolete syntax, we accept any sequence of one
	// or more words (each word being either an atom or a quoted string)
	// separated by a '.' character.

	var value []byte

	for len(r.buf) > 0 {
		if err := r.SkipCFWS(); err != nil {
			return nil, err
		}

		word, err := r.ReadWord()
		if err != nil {
			return nil, err
		}

		value = append(value, word...)

		if err := r.SkipCFWS(); err != nil {
			return nil, err
		}

		if !r.SkipByte('.') {
			break
		}

		value = append(value, '.')
	}

	return value, nil
}

func (r *DataReader) ReadDomainLiteral() ([]byte, error) {
	if err := r.SkipCFWS(); err != nil {
		return nil, err
	}

	if !r.SkipByte('[') {
		return nil, fmt.Errorf("missing initial '[' character")
	}

	domain := []byte{'['}

	if r.StartsWithByte('"') {
		value, err := r.ReadQuotedString()
		if err != nil {
			return nil, err
		}

		domain = append(domain, value...)
	} else {
		for {
			if err := r.MaybeSkipFWS(); err != nil {
				return nil, err
			}

			if len(r.buf) == 0 || r.buf[0] == ']' {
				break
			}

			// In theory we must support obs-NO-WS-CTL which contains ASCII
			// control characters. No valid domain or IP address contains
			// control characters, period. Accepting them is pretty much
			// guaranteed to cause issues down the line.

			c := r.buf[0]

			valid := c >= 33 && c <= 90 || c >= 94 && c <= 126
			if !valid {
				return nil, fmt.Errorf("invalid character %s in domain "+
					"literal", utils.QuoteByte(c))
			}

			domain = append(domain, r.buf[0])

			r.buf = r.buf[1:]
		}
	}

	if !r.SkipByte(']') {
		return nil, fmt.Errorf("missing final ']' character")
	}

	domain = append(domain, ']')

	return domain, nil
}

func (r *DataReader) ReadDomain() ([]byte, error) {
	// Start by looking for a domain-literal. Note that it can contain a
	// quoted string.

	if err := r.SkipCFWS(); err != nil {
		return nil, err
	}

	if r.StartsWithByte('[') {
		domain, err := r.ReadDomainLiteral()
		if err != nil {
			return nil, fmt.Errorf("invalid domain literal: %w", err)
		}

		// Surprisingly, a domain literal can be empty (see RFC 5322 3.4.1.
		// Addr-Spec Specification). It is not clear what it is supposed to
		// represent.

		return domain, nil
	}

	// If it is not a domain-literal, it is a dot-atom

	domain, err := r.ReadDotAtom()
	if err != nil {
		return nil, fmt.Errorf("invalid dot atom: %w", err)
	}

	return domain, nil
}

func (r *DataReader) ReadSpecificAddress() (*SpecificAddress, error) {
	localPart, err := r.ReadLocalPart()
	if err != nil {
		return nil, fmt.Errorf("invalid local part: %w", err)
	}

	if err := r.SkipCFWS(); err != nil {
		return nil, err
	}

	if !r.SkipByte('@') {
		return nil, fmt.Errorf("missing '@' character after local part")
	}

	if err := r.SkipCFWS(); err != nil {
		return nil, err
	}

	domain, err := r.ReadDomain()
	if err != nil {
		return nil, fmt.Errorf("invalid domain: %w", err)
	}

	spec := SpecificAddress{
		LocalPart: string(localPart),
		Domain:    string(domain),
	}

	return &spec, nil
}

func (r *DataReader) ReadAngleAddress(allowEmpty bool) (*SpecificAddress, error) {
	if err := r.SkipCFWS(); err != nil {
		return nil, err
	}

	if !r.SkipByte('<') {
		return nil, fmt.Errorf("missing '<' character before specific " +
			"address")
	}

	if err := r.SkipCFWS(); err != nil {
		return nil, err
	}

	if allowEmpty && r.StartsWithByte('>') {
		r.Skip(1)
		return nil, nil
	}

	// TODO obs-route

	if err := r.SkipCFWS(); err != nil {
		return nil, err
	}

	spec, err := r.ReadSpecificAddress()
	if err != nil {
		return nil, fmt.Errorf("invalid specific address: %w", err)
	}

	if err := r.SkipCFWS(); err != nil {
		return nil, err
	}

	if !r.SkipByte('>') {
		return nil, fmt.Errorf("missing '>' character after specific " +
			"address")
	}

	return spec, nil
}

func (r *DataReader) ReadNameAddress() (*Mailbox, error) {
	if err := r.SkipCFWS(); err != nil {
		return nil, err
	}

	var displayName []byte

	if !r.StartsWithByte('<') {
		phrase, err := r.ReadPhrase()
		if err != nil {
			return nil, fmt.Errorf("invalid display name: %w", err)
		}

		displayName = phrase
	}

	if err := r.SkipCFWS(); err != nil {
		return nil, err
	}

	spec, err := r.ReadAngleAddress(false)
	if err != nil {
		return nil, fmt.Errorf("invalid angle address: %w", err)
	}

	mb := Mailbox{
		SpecificAddress: *spec,
		DisplayName:     string(displayName),
	}

	return &mb, nil
}

func (r *DataReader) ReadMailbox() (*Mailbox, error) {
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

	err := r.Try(func() error {
		addr, err := r.ReadSpecificAddress()
		if err != nil {
			return err
		}

		mailbox = &Mailbox{SpecificAddress: *addr}
		return nil
	})
	if err != nil {
		mb, err := r.ReadNameAddress()
		if err != nil {
			return nil, fmt.Errorf("invalid name address: %w", err)
		}

		mailbox = mb
	}

	return mailbox, nil
}

func (r *DataReader) ReadMailboxList() ([]*Mailbox, error) {
	var mailboxes []*Mailbox

	for {
		if err := r.SkipCFWS(); err != nil {
			return nil, err
		}

		if len(r.buf) == 0 {
			break
		}

		if r.SkipByte(',') {
			// Obsolete syntax allows empty list elements
			continue
		}

		mailbox, err := r.ReadMailbox()
		if err != nil {
			return nil, fmt.Errorf("invalid mailbox: %w", err)
		}

		mailboxes = append(mailboxes, mailbox)

		if err := r.SkipCFWS(); err != nil {
			return nil, err
		}

		if !r.SkipByte(',') {
			break
		}
	}

	if len(mailboxes) == 0 {
		return nil, fmt.Errorf("empty list")
	}

	return mailboxes, nil
}

func (r *DataReader) ReadGroup() (*Group, error) {
	if err := r.SkipCFWS(); err != nil {
		return nil, err
	}

	displayName, err := r.ReadPhrase()
	if err != nil {
		return nil, fmt.Errorf("invalid display name: %w", err)
	}

	if err := r.SkipCFWS(); err != nil {
		return nil, err
	}

	if !r.SkipByte(':') {
		return nil, fmt.Errorf("missing ':' character after display name")
	}

	if err := r.SkipCFWS(); err != nil {
		return nil, err
	}

	mailboxes, err := r.ReadMailboxList()
	if err != nil {
		return nil, fmt.Errorf("invalid mailbox list: %w", err)
	}

	if err := r.SkipCFWS(); err != nil {
		return nil, err
	}

	if !r.SkipByte(';') {
		return nil, fmt.Errorf("missing ';' character after mailbox list")
	}

	group := Group{
		DisplayName: string(displayName),
		Mailboxes:   mailboxes,
	}

	return &group, nil
}

func (r *DataReader) ReadAddress() (Address, error) {
	// An address is either a mailbox or a group, and again we cannot
	// differentiate them with a simple lookup.

	var addr Address

	err := r.Try(func() error {
		group, err := r.ReadGroup()
		if err != nil {
			return err
		}

		addr = group
		return nil
	})
	if err != nil {
		mailbox, err := r.ReadMailbox()
		if err != nil {
			return nil, fmt.Errorf("invalid mailbox: %w", err)
		}

		addr = mailbox
	}

	return addr, nil
}

func (r *DataReader) ReadAddressList(allowEmpty bool) ([]Address, error) {
	var addrs []Address

	for {
		if err := r.SkipCFWS(); err != nil {
			return nil, err
		}

		if len(r.buf) == 0 {
			break
		}

		if r.SkipByte(',') {
			// Obsolete syntax allows empty list elements
			continue
		}

		addr, err := r.ReadAddress()
		if err != nil {
			return nil, fmt.Errorf("invalid address: %w", err)
		}

		addrs = append(addrs, addr)

		if err := r.SkipCFWS(); err != nil {
			return nil, err
		}

		if !r.SkipByte(',') {
			break
		}
	}

	if len(addrs) == 0 && !allowEmpty {
		return nil, fmt.Errorf("empty list")
	}

	return addrs, nil
}

func (r *DataReader) ReadMessageId() (*MessageId, error) {
	if err := r.SkipCFWS(); err != nil {
		return nil, err
	}

	// Start delimiter
	if !r.SkipByte('<') {
		return nil, fmt.Errorf("missing initial '<' character")
	}

	// Left part
	if err := r.SkipCFWS(); err != nil {
		return nil, err
	}

	var left []byte
	var err error

	if IsAtomChar(r.buf[0]) {
		left, err = r.ReadDotAtom()
	} else {
		left, err = r.ReadQuotedString()
	}

	if err != nil {
		return nil, fmt.Errorf("invalid left part: %w", err)
	}

	// Separator
	if err := r.SkipCFWS(); err != nil {
		return nil, err
	}

	if !r.SkipByte('@') {
		return nil, fmt.Errorf("missing '@' character after left part")
	}

	// Right part
	if err := r.SkipCFWS(); err != nil {
		return nil, err
	}

	right, err := r.ReadDomain()
	if err != nil {
		return nil, fmt.Errorf("invalid right part: %w", err)
	}

	// End delimiter
	if err := r.SkipCFWS(); err != nil {
		return nil, err
	}

	if !r.SkipByte('>') {
		return nil, fmt.Errorf("missing final '>' character")
	}

	id := MessageId{
		Left:  string(left),
		Right: string(right),
	}

	return &id, nil
}

func (r *DataReader) ReadMessageIdList() (MessageIds, error) {
	// Unlike other lists, message id lists do not use a comma separator

	var ids MessageIds

	for {
		if err := r.SkipCFWS(); err != nil {
			return nil, err
		}

		if len(r.buf) == 0 {
			break
		}

		id, err := r.ReadMessageId()
		if err != nil {
			return nil, fmt.Errorf("invalid message id: %w", err)
		}

		ids = append(ids, *id)
	}

	if len(ids) == 0 {
		return nil, fmt.Errorf("empty list")
	}

	return ids, nil
}

func (r *DataReader) ReadDateTime() (*time.Time, error) {
	// Optional day name
	if err := r.SkipCFWS(); err != nil {
		return nil, err
	}

	if len(r.buf) == 0 {
		return nil, fmt.Errorf("empty value")
	}

	if IsAlphaChar(r.buf[0]) {
		if _, err := r.MaybeReadDayName(); err != nil {
			return nil, err
		}
	}

	// Day
	day, err := r.ReadDay()
	if err != nil {
		return nil, fmt.Errorf("invalid day: %w", err)
	}

	// Month
	month, err := r.ReadMonth()
	if err != nil {
		return nil, fmt.Errorf("invalid month: %w", err)
	}

	// Year
	year, err := r.ReadYear()
	if err != nil {
		return nil, fmt.Errorf("invalid year: %w", err)
	}

	// Hour
	hour, err := r.ReadHour()
	if err != nil {
		return nil, fmt.Errorf("invalid hour: %w", err)
	}

	// Separator
	if err := r.SkipCFWS(); err != nil {
		return nil, err
	}

	if !r.SkipByte(':') {
		return nil, fmt.Errorf("missing ':' character after hour")
	}

	// Minute
	minute, err := r.ReadMinute()
	if err != nil {
		return nil, fmt.Errorf("invalid minute: %w", err)
	}

	// Separator and optional second
	if err := r.SkipCFWS(); err != nil {
		return nil, err
	}

	var second int
	if r.SkipByte(':') {
		second, err = r.ReadSecond()
		if err != nil {
			return nil, fmt.Errorf("invalid second: %w", err)
		}
	}

	// Timezone
	loc, err := r.ReadTimezone()
	if err != nil {
		return nil, fmt.Errorf("invalid timezone: %w", err)
	}

	date := time.Date(year, month, day, hour, minute, second, 0, loc)
	return &date, nil
}

func (r *DataReader) MaybeReadDayName() (string, error) {
	// This function is called after checking that the buffer starts with an
	// alpha character, so we do not have to skip anything or check that the
	// result is not empty.

	name := r.ReadWhile(IsAlphaChar)

	if err := r.SkipCFWS(); err != nil {
		return "", err
	}

	if !r.SkipByte(',') {
		return "", fmt.Errorf("missing ',' character after day name")
	}

	return string(name), nil
}

func (r *DataReader) ReadInteger(maxNbDigits int, minValue, maxValue int64) (int, error) {
	if err := r.SkipCFWS(); err != nil {
		return 0, err
	}

	if len(r.buf) == 0 {
		return 0, fmt.Errorf("empty value")
	}

	if !IsDigitChar(r.buf[0]) {
		return 0, fmt.Errorf("invalid character %s", utils.QuoteByte(r.buf[0]))
	}

	s := string(r.ReadWhileN(IsDigitChar, maxNbDigits))

	i64, err := strconv.ParseInt(s, 10, 64)
	if err != nil || i64 < minValue || i64 > maxValue {
		return 0, fmt.Errorf("invalid value %q: %w", s, err)
	}

	return int(i64), nil
}

func (r *DataReader) ReadDay() (int, error) {
	return r.ReadInteger(2, 1, 31)
}

func (r *DataReader) ReadMonth() (time.Month, error) {
	if err := r.SkipCFWS(); err != nil {
		return 0, err
	}

	if len(r.buf) == 0 {
		return 0, fmt.Errorf("empty value")
	}

	if !IsAlphaChar(r.buf[0]) {
		return 0, fmt.Errorf("invalid character %s", utils.QuoteByte(r.buf[0]))
	}

	s := string(r.ReadWhileN(IsAlphaChar, 3))

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
	case "nov":
		month = time.November
	case "dec":
		month = time.December
	default:
		return 0, fmt.Errorf("invalid value %q", s)
	}

	return month, nil
}

func (r *DataReader) ReadYear() (int, error) {
	if err := r.SkipCFWS(); err != nil {
		return 0, err
	}

	if len(r.buf) == 0 {
		return 0, fmt.Errorf("empty value")
	}

	if !IsDigitChar(r.buf[0]) {
		return 0, fmt.Errorf("invalid character %s", utils.QuoteByte(r.buf[0]))
	}

	s := string(r.ReadWhileN(IsDigitChar, 4))

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

func (r *DataReader) ReadHour() (int, error) {
	return r.ReadInteger(2, 0, 23)
}

func (r *DataReader) ReadMinute() (int, error) {
	return r.ReadInteger(2, 0, 59)
}

func (r *DataReader) ReadSecond() (int, error) {
	// Yes, 60, leap seconds are a thing
	return r.ReadInteger(2, 0, 60)
}

func (r *DataReader) ReadTimezone() (*time.Location, error) {
	if err := r.SkipCFWS(); err != nil {
		return nil, err
	}

	if len(r.buf) == 0 {
		return nil, fmt.Errorf("empty value")
	}

	var loc *time.Location

	if r.buf[0] == '+' || r.buf[0] == '-' {
		// Timezone offset

		sign := 1
		if r.buf[0] == '-' {
			sign = -1
		}

		r.Skip(1)

		maxOffset := int64(12)
		if sign == 1 {
			maxOffset = 14 // Line Islands
		}

		i, err := r.ReadInteger(2, 0, maxOffset)
		if err != nil {
			return nil, fmt.Errorf("invalid timezone hour offset: %w", err)
		}
		hourOffset := sign * i

		i, err = r.ReadInteger(2, 0, 59)
		if err != nil {
			return nil, fmt.Errorf("invalid timezone minute offset: %w", err)
		}
		minuteOffset := i

		loc = time.FixedZone("", hourOffset*3600+minuteOffset*60)
	} else if IsAlphaChar(r.buf[0]) {
		// Timezone name
		//
		// See RFC 5322 4.3. Obsolete Date and Time. Zone names can be up to 5
		// character long. For military time zones (single letter zone names),
		// "they SHOULD all be considered equivalent to "-0000" unless there is
		// out-of-band information confirming their meaning". Unknown timezone
		// names should also be considered equivalent to "-0000".

		s := string(r.ReadWhileN(IsDigitChar, 5))
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
			utils.QuoteByte(r.buf[0]))
	}

	return loc, nil
}
