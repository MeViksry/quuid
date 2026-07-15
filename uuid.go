package quuid

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha512"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"
)

// UUIDSize is the binary size of a UUID in bytes.
const UUIDSize = 16

// UUID is a 128-bit universally unique identifier encoded according to RFC 9562.
//
// UUID is intentionally implemented by quuid instead of being an alias to an
// upstream type. This lets quuid provide strict parsing, deterministic SQL
// behavior, allocation-aware append methods, clock injection, and explicit
// binary database wrappers without global process settings.
type UUID [UUIDSize]byte

// Version is the four-bit UUID version field.
type Version byte

const (
	// VersionUnknown represents an unassigned or unavailable UUID version.
	VersionUnknown Version = 0
	// Version1 identifies UUIDv1.
	Version1 Version = 1
	// Version2 identifies UUIDv2.
	Version2 Version = 2
	// Version3 identifies name-based UUIDv3.
	Version3 Version = 3
	// Version4 identifies random UUIDv4.
	Version4 Version = 4
	// Version5 identifies name-based UUIDv5.
	Version5 Version = 5
	// Version6 identifies reordered-time UUIDv6.
	Version6 Version = 6
	// Version7 identifies Unix-time UUIDv7.
	Version7 Version = 7
	// Version8 identifies application-defined UUIDv8.
	Version8 Version = 8
)

// String returns the conventional UUID version name.
func (v Version) String() string {
	if v >= Version1 && v <= Version8 {
		return fmt.Sprintf("UUIDv%d", byte(v))
	}
	return "unknown"
}

// Variant identifies the UUID layout family.
type Variant byte

const (
	// VariantNCS identifies the historical NCS UUID variant.
	VariantNCS Variant = iota
	// VariantRFC9562 identifies the RFC 9562 variant, historically called RFC 4122.
	VariantRFC9562
	// VariantMicrosoft identifies the historical Microsoft UUID variant.
	VariantMicrosoft
	// VariantFuture identifies the reserved future UUID variant.
	VariantFuture
)

// String returns the conventional UUID variant name.
func (v Variant) String() string {
	switch v {
	case VariantNCS:
		return "NCS"
	case VariantRFC9562:
		return "RFC9562"
	case VariantMicrosoft:
		return "Microsoft"
	case VariantFuture:
		return "Future"
	default:
		return "unknown"
	}
}

var (
	// NilUUID is the all-zero UUID. It is not SQL NULL.
	NilUUID UUID
	// MaxUUID is the all-one UUID defined by RFC 9562.
	MaxUUID = UUID{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
)

// NewUUIDv4 returns a random UUIDv4 generated from crypto/rand.
func NewUUIDv4() (UUID, error) { return NewUUIDv4FromReader(rand.Reader) }

// NewUUIDv4FromReader returns a random UUIDv4 generated from r.
func NewUUIDv4FromReader(r io.Reader) (UUID, error) {
	id, err := randomUUIDBytes(r)
	if err != nil {
		return NilUUID, err
	}
	id[6] = (id[6] & 0x0f) | 0x40
	id[8] = (id[8] & 0x3f) | 0x80
	return id, nil
}

// MustNewUUIDv4 is like NewUUIDv4 but panics when secure randomness fails.
func MustNewUUIDv4() UUID {
	id, err := NewUUIDv4()
	if err != nil {
		panic(err)
	}
	return id
}

// NewUUIDv6 returns a process-local monotonic RFC 9562 UUIDv6. The embedded
// timestamp remains strictly increasing when multiple values are generated in
// the same 100-nanosecond interval or the wall clock moves backwards. Random
// clock-sequence and node bits are sampled once for the process generator.
func NewUUIDv6() (UUID, error) { return defaultGenerator.NewUUIDv6() }

// NewUUIDv6At returns an RFC 9562 UUIDv6 for t without shared monotonic state.
// The random portion is read from r.
func NewUUIDv6At(t time.Time, r io.Reader) (UUID, error) {
	timestamp, err := uuidV6Timestamp(t)
	if err != nil {
		return NilUUID, err
	}
	return newUUIDv6FromTimestamp(timestamp, r)
}

// MustNewUUIDv6 is like NewUUIDv6 but panics when generation fails.
func MustNewUUIDv6() UUID {
	id, err := NewUUIDv6()
	if err != nil {
		panic(err)
	}
	return id
}

func uuidV6Timestamp(t time.Time) (uint64, error) {
	const uuidEpochToUnixSeconds int64 = 12_219_292_800
	const ticksPerSecond uint64 = 10_000_000
	const maxTimestamp uint64 = 1<<60 - 1

	sec := t.Unix()
	maxShiftedSeconds := maxTimestamp / ticksPerSecond
	if sec < -uuidEpochToUnixSeconds || sec > int64(maxShiftedSeconds)-uuidEpochToUnixSeconds {
		return 0, ErrTimeOutOfRange
	}
	shiftedSeconds := uint64(sec + uuidEpochToUnixSeconds)
	timestamp := shiftedSeconds*ticksPerSecond + uint64(t.Nanosecond()/100)
	if timestamp > maxTimestamp {
		return 0, ErrTimeOutOfRange
	}
	return timestamp, nil
}

func newUUIDv6FromTimestamp(timestamp uint64, r io.Reader) (UUID, error) {
	if r == nil {
		return NilUUID, fmt.Errorf("quuid: nil entropy reader")
	}
	var id UUID
	if _, err := io.ReadFull(r, id[8:]); err != nil {
		return NilUUID, fmt.Errorf("quuid: read entropy: %w", err)
	}
	if allZero(id[8:]) {
		return NilUUID, ErrZeroEntropy
	}
	putUUIDv6Timestamp(&id, timestamp)
	id[8] = (id[8] & 0x3f) | 0x80
	return id, nil
}

func putUUIDv6Timestamp(id *UUID, timestamp uint64) {
	id[0] = byte(timestamp >> 52)
	id[1] = byte(timestamp >> 44)
	id[2] = byte(timestamp >> 36)
	id[3] = byte(timestamp >> 28)
	id[4] = byte(timestamp >> 20)
	id[5] = byte(timestamp >> 12)
	id[6] = 0x60 | byte((timestamp>>8)&0x0f)
	id[7] = byte(timestamp)
}

// NewUUIDv7 returns a process-local monotonic UUIDv7. Values remain ordered
// when multiple UUIDs are generated in the same millisecond or the wall clock
// moves backwards. The guarantee does not extend across processes or hosts.
func NewUUIDv7() (UUID, error) { return defaultGenerator.NewUUIDv7() }

// NewUUIDv7FromReader creates a UUIDv7 using time.Now and r without preserving
// monotonic state between calls. Use Generator for injected clocks and monotonicity.
func NewUUIDv7FromReader(r io.Reader) (UUID, error) { return NewUUIDv7At(time.Now(), r) }

// NewUUIDv7At creates an RFC 9562 UUIDv7 for t using random bits from r.
func NewUUIDv7At(t time.Time, r io.Reader) (UUID, error) {
	ms := t.UnixMilli()
	if ms < 0 || ms > 1<<48-1 {
		return NilUUID, ErrTimeOutOfRange
	}
	if r == nil {
		return NilUUID, fmt.Errorf("quuid: nil entropy reader")
	}

	var id UUID
	if _, err := io.ReadFull(r, id[6:]); err != nil {
		return NilUUID, fmt.Errorf("quuid: read entropy: %w", err)
	}
	if allZero(id[6:]) {
		return NilUUID, ErrZeroEntropy
	}
	putUUIDv7Milliseconds(&id, uint64(ms))
	id[6] = (id[6] & 0x0f) | 0x70
	id[8] = (id[8] & 0x3f) | 0x80
	return id, nil
}

// MustNewUUIDv7 is like NewUUIDv7 but panics when generation fails.
func MustNewUUIDv7() UUID {
	id, err := NewUUIDv7()
	if err != nil {
		panic(err)
	}
	return id
}

// NewUUIDv8 returns a random RFC 9562 UUIDv8.
func NewUUIDv8() (UUID, error) { return NewUUIDv8FromReader(rand.Reader) }

// NewUUIDv8FromReader returns a random application-defined UUIDv8 using r.
func NewUUIDv8FromReader(r io.Reader) (UUID, error) {
	id, err := randomUUIDBytes(r)
	if err != nil {
		return NilUUID, err
	}
	setUUIDv8Bits(&id)
	return id, nil
}

// MustNewUUIDv8 is like NewUUIDv8 but panics when generation fails.
func MustNewUUIDv8() UUID {
	id, err := NewUUIDv8()
	if err != nil {
		panic(err)
	}
	return id
}

// DeriveUUIDv8 creates a deterministic UUIDv8 using SHA-512/256 and framed
// domain separation. It does not use MD5 or SHA-1.
func DeriveUUIDv8(namespace string, data []byte) UUID {
	h := sha512.New512_256()
	writeFrame(h, []byte("github.com/MeViksry/quuid/derive-uuid-v8/v1"), []byte(namespace), data)
	digest := h.Sum(nil)
	var id UUID
	copy(id[:], digest[:UUIDSize])
	setUUIDv8Bits(&id)
	return id
}

// DeriveUUIDv8Keyed creates a deterministic UUIDv8 using HMAC-SHA-512/256.
// secret must contain at least 32 bytes from a cryptographically secure source.
func DeriveUUIDv8Keyed(secret []byte, namespace string, data []byte) (UUID, error) {
	if len(secret) < 32 {
		return NilUUID, ErrWeakSecret
	}
	h := hmac.New(sha512.New512_256, secret)
	writeFrame(h, []byte("github.com/MeViksry/quuid/derive-uuid-v8-keyed/v1"), []byte(namespace), data)
	digest := h.Sum(nil)
	var id UUID
	copy(id[:], digest[:UUIDSize])
	setUUIDv8Bits(&id)
	return id, nil
}

// ParseUUID accepts only the canonical 36-character UUID representation.
// It rejects surrounding whitespace, URNs, braces, raw hexadecimal strings,
// and Microsoft mixed-endian binary encodings.
func ParseUUID(s string) (UUID, error) {
	if len(s) != 36 {
		return NilUUID, ErrInvalidLength
	}
	if s[8] != '-' || s[13] != '-' || s[18] != '-' || s[23] != '-' {
		return NilUUID, ErrInvalidEncoding
	}

	var id UUID
	sections := [][2]int{{0, 8}, {9, 13}, {14, 18}, {19, 23}, {24, 36}}
	offset := 0
	for _, section := range sections {
		part := s[section[0]:section[1]]
		decoded := id[offset : offset+len(part)/2]
		if _, err := hex.Decode(decoded, []byte(part)); err != nil {
			return NilUUID, fmt.Errorf("%w: %v", ErrInvalidEncoding, err)
		}
		offset += len(decoded)
	}
	return id, nil
}

// ParseUUIDLoose is an explicit migration helper. It accepts canonical UUIDs,
// 32 hexadecimal digits, urn:uuid: values, and brace-wrapped canonical UUIDs.
// Surrounding ASCII whitespace is trimmed. New application input should use
// ParseUUID instead.
func ParseUUIDLoose(s string) (UUID, error) {
	s = strings.TrimSpace(s)
	if len(s) >= 9 && strings.EqualFold(s[:9], "urn:uuid:") {
		s = s[9:]
	}
	if len(s) == 38 && s[0] == '{' && s[37] == '}' {
		s = s[1:37]
	}
	if len(s) == 32 {
		var id UUID
		if _, err := hex.Decode(id[:], []byte(s)); err != nil {
			return NilUUID, fmt.Errorf("%w: %v", ErrInvalidEncoding, err)
		}
		return id, nil
	}
	return ParseUUID(s)
}

// MustParseUUID is like ParseUUID but panics on malformed input.
func MustParseUUID(s string) UUID {
	id, err := ParseUUID(s)
	if err != nil {
		panic(err)
	}
	return id
}

// ValidateUUID validates the canonical UUID text representation.
func ValidateUUID(s string) error {
	_, err := ParseUUID(s)
	return err
}

// ValidateRFC9562UUID validates canonical syntax, RFC variant bits, and a
// currently assigned version number from 1 through 8.
func ValidateRFC9562UUID(s string) error {
	id, err := ParseUUID(s)
	if err != nil {
		return err
	}
	if id.Variant() != VariantRFC9562 {
		return fmt.Errorf("quuid: invalid RFC 9562 variant")
	}
	version := id.Version()
	if version < Version1 || version > Version8 {
		return fmt.Errorf("quuid: unsupported UUID version %d", version)
	}
	return nil
}

// String returns the canonical text representation.
func (id UUID) String() string {
	text, _ := id.AppendText(nil)
	return string(text)
}

// Bytes returns a defensive copy of the binary representation.
func (id UUID) Bytes() []byte { return append([]byte(nil), id[:]...) }

// IsZero reports whether every byte is zero.
func (id UUID) IsZero() bool { return id == NilUUID }

// Version returns the UUID version field.
func (id UUID) Version() Version {
	return Version(id[6] >> 4)
}

// Variant returns the UUID variant field.
func (id UUID) Variant() Variant {
	switch {
	case id[8]&0x80 == 0:
		return VariantNCS
	case id[8]&0xc0 == 0x80:
		return VariantRFC9562
	case id[8]&0xe0 == 0xc0:
		return VariantMicrosoft
	default:
		return VariantFuture
	}
}

// Equal reports whether two values contain identical bytes.
func (id UUID) Equal(other UUID) bool {
	return subtle.ConstantTimeCompare(id[:], other[:]) == 1
}

// Compare compares two values lexicographically and returns -1, 0, or 1.
func (id UUID) Compare(other UUID) int {
	for i := range id {
		if id[i] < other[i] {
			return -1
		}
		if id[i] > other[i] {
			return 1
		}
	}
	return 0
}

// Time returns the embedded timestamp for UUIDv6 and UUIDv7.
func (id UUID) Time() (time.Time, error) {
	switch id.Version() {
	case Version6:
		timestamp := uint64(id[0])<<52 |
			uint64(id[1])<<44 |
			uint64(id[2])<<36 |
			uint64(id[3])<<28 |
			uint64(id[4])<<20 |
			uint64(id[5])<<12 |
			uint64(id[6]&0x0f)<<8 |
			uint64(id[7])
		const uuidEpochToUnixTicks int64 = 122_192_928_000_000_000
		unixTicks := int64(timestamp) - uuidEpochToUnixTicks
		seconds := unixTicks / 10_000_000
		remainder := unixTicks % 10_000_000
		if remainder < 0 {
			seconds--
			remainder += 10_000_000
		}
		return time.Unix(seconds, remainder*100).UTC(), nil
	case Version7:
		ms := uint64(id[0])<<40 |
			uint64(id[1])<<32 |
			uint64(id[2])<<24 |
			uint64(id[3])<<16 |
			uint64(id[4])<<8 |
			uint64(id[5])
		return time.UnixMilli(int64(ms)).UTC(), nil
	default:
		return time.Time{}, fmt.Errorf("quuid: UUID version %d has no supported timestamp", id.Version())
	}
}

// MarshalText implements encoding.TextMarshaler.
func (id UUID) MarshalText() ([]byte, error) { return id.AppendText(nil) }

// AppendText appends the canonical UUID representation to dst.
func (id UUID) AppendText(dst []byte) ([]byte, error) {
	const digits = "0123456789abcdef"
	start := len(dst)
	dst = append(dst, make([]byte, 36)...)
	out := dst[start:]
	positions := [...]int{0, 2, 4, 6, 9, 11, 14, 16, 19, 21, 24, 26, 28, 30, 32, 34}
	for i, p := range positions {
		out[p] = digits[id[i]>>4]
		out[p+1] = digits[id[i]&0x0f]
	}
	out[8], out[13], out[18], out[23] = '-', '-', '-', '-'
	return dst, nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (id *UUID) UnmarshalText(text []byte) error {
	if id == nil {
		return fmt.Errorf("quuid: UnmarshalText on nil *UUID")
	}
	parsed, err := ParseUUID(string(text))
	if err != nil {
		return err
	}
	*id = parsed
	return nil
}

// MarshalBinary implements encoding.BinaryMarshaler.
func (id UUID) MarshalBinary() ([]byte, error) { return id.AppendBinary(nil) }

// AppendBinary appends the 16-byte network-order UUID representation to dst.
func (id UUID) AppendBinary(dst []byte) ([]byte, error) { return append(dst, id[:]...), nil }

// UnmarshalBinary implements encoding.BinaryUnmarshaler.
func (id *UUID) UnmarshalBinary(data []byte) error {
	if id == nil {
		return fmt.Errorf("quuid: UnmarshalBinary on nil *UUID")
	}
	if len(data) != UUIDSize {
		return ErrInvalidLength
	}
	copy(id[:], data)
	return nil
}

// MarshalJSON implements json.Marshaler.
func (id UUID) MarshalJSON() ([]byte, error) { return json.Marshal(id.String()) }

// UnmarshalJSON implements json.Unmarshaler.
func (id *UUID) UnmarshalJSON(data []byte) error {
	if id == nil {
		return fmt.Errorf("quuid: UnmarshalJSON on nil *UUID")
	}
	if string(data) == "null" {
		return ErrNilValue
	}
	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return fmt.Errorf("quuid: decode UUID JSON: %w", err)
	}
	return id.UnmarshalText([]byte(value))
}

func randomUUIDBytes(r io.Reader) (UUID, error) {
	if r == nil {
		return NilUUID, fmt.Errorf("quuid: nil entropy reader")
	}
	var id UUID
	if _, err := io.ReadFull(r, id[:]); err != nil {
		return NilUUID, fmt.Errorf("quuid: read entropy: %w", err)
	}
	if allZero(id[:]) {
		return NilUUID, ErrZeroEntropy
	}
	return id, nil
}

func setUUIDv8Bits(id *UUID) {
	id[6] = (id[6] & 0x0f) | 0x80
	id[8] = (id[8] & 0x3f) | 0x80
}

func putUUIDv7Milliseconds(id *UUID, ms uint64) {
	id[0] = byte(ms >> 40)
	id[1] = byte(ms >> 32)
	id[2] = byte(ms >> 24)
	id[3] = byte(ms >> 16)
	id[4] = byte(ms >> 8)
	id[5] = byte(ms)
}
