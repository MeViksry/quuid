package quuid

import (
	"encoding/base32"
	"encoding/hex"
	"fmt"
	"strings"
)

const crockfordAlphabet = "0123456789abcdefghjkmnpqrstvwxyz"

var crockford = base32.NewEncoding(crockfordAlphabet).WithPadding(base32.NoPadding)

func encodeText(prefix string, raw []byte) string {
	return string(appendEncodedText(nil, prefix, raw))
}

func appendEncodedText(dst []byte, prefix string, raw []byte) []byte {
	dst = append(dst, prefix...)
	return crockford.AppendEncode(dst, raw)
}

func decodeText(input, prefix string, size int) ([]byte, error) {
	s := input
	if s == "" {
		return nil, ErrInvalidLength
	}

	if len(s) >= len(prefix) && strings.EqualFold(s[:len(prefix)], prefix) {
		s = s[len(prefix):]
	} else if strings.ContainsRune(s, '_') {
		return nil, ErrInvalidPrefix
	}

	if len(s) == size*2 {
		decoded, err := hex.DecodeString(s)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrInvalidEncoding, err)
		}
		return decoded, nil
	}

	normalized := normalizeCrockford(s)
	if len(normalized) != crockford.EncodedLen(size) {
		return nil, ErrInvalidLength
	}

	decoded, err := crockford.DecodeString(normalized)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidEncoding, err)
	}
	if len(decoded) != size {
		return nil, ErrInvalidLength
	}
	if crockford.EncodeToString(decoded) != normalized {
		return nil, ErrInvalidEncoding
	}
	return decoded, nil
}

func normalizeCrockford(s string) string {
	s = strings.ToLower(s)
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		switch r {
		case 'o':
			b.WriteByte('0')
		case 'i', 'l':
			b.WriteByte('1')
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

func allZero(raw []byte) bool {
	var v byte
	for _, b := range raw {
		v |= b
	}
	return v == 0
}
