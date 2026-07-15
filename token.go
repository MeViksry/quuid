package quuid

import (
	"crypto/rand"
	"crypto/sha512"
	"crypto/subtle"
	"fmt"
	"io"
)

// TokenSize is the binary size of a Token in bytes.
const (
	TokenSize   = 32
	tokenPrefix = "qtk_"
)

// Token is a 256-bit bearer secret. Store TokenDigest, not the plaintext Token.
type Token [TokenSize]byte

// TokenDigest is a one-way SHA-512/256 digest suitable for database storage.
type TokenDigest [32]byte

// NewToken returns a new bearer token generated from crypto/rand.
func NewToken() (Token, error) { return NewTokenFromReader(rand.Reader) }

// NewTokenFromReader returns a new bearer token using r as its entropy source.
func NewTokenFromReader(r io.Reader) (Token, error) {
	var token Token
	if r == nil {
		return token, fmt.Errorf("quuid: nil entropy reader")
	}
	if _, err := io.ReadFull(r, token[:]); err != nil {
		return Token{}, fmt.Errorf("quuid: read entropy: %w", err)
	}
	if allZero(token[:]) {
		return Token{}, ErrZeroEntropy
	}
	return token, nil
}

// MustNewToken is like NewToken but panics when secure randomness fails.
func MustNewToken() Token {
	token, err := NewToken()
	if err != nil {
		panic(err)
	}
	return token
}

// ParseToken parses a bearer token from text.
func ParseToken(s string) (Token, error) {
	raw, err := decodeText(s, tokenPrefix, TokenSize)
	if err != nil {
		return Token{}, err
	}
	var token Token
	copy(token[:], raw)
	return token, nil
}

// Secret returns the complete bearer-token plaintext and should be used only at issuance or verification boundaries.
func (token Token) Secret() string { return encodeText(tokenPrefix, token[:]) }

// Bytes returns a defensive copy of the binary representation.
func (token Token) Bytes() []byte { return append([]byte(nil), token[:]...) }

// String implements fmt.Stringer with a redacted representation so ordinary
// logging does not expose the bearer secret. Use Secret only at issuance time.
func (token Token) String() string { return token.Redacted() }

// Redacted returns a log-safe token representation that does not reveal the complete secret.
func (token Token) Redacted() string {
	text := token.Secret()
	if len(text) <= 10 {
		return "qtk_[redacted]"
	}
	return "qtk_…" + text[len(text)-6:]
}

// Digest returns the one-way digest that should be stored for later verification.
func (token Token) Digest() TokenDigest {
	h := sha512.New512_256()
	writeFrame(h, []byte("github.com/MeViksry/quuid/token-digest/v1"), token[:])
	var digest TokenDigest
	copy(digest[:], h.Sum(nil))
	return digest
}

// String returns the canonical text representation.
func (digest TokenDigest) String() string { return fmt.Sprintf("%x", digest[:]) }

// VerifyToken parses presented and compares its digest with expected in constant time.
func VerifyToken(presented string, expected TokenDigest) bool {
	token, err := ParseToken(presented)
	if err != nil {
		return false
	}
	actual := token.Digest()
	return subtle.ConstantTimeCompare(actual[:], expected[:]) == 1
}
