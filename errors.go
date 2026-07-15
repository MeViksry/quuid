package quuid

import "errors"

var (
	// ErrInvalidLength indicates that an identifier has an unexpected text or binary length.
	ErrInvalidLength = errors.New("quuid: invalid identifier length")
	// ErrInvalidEncoding indicates malformed or non-canonical identifier encoding.
	ErrInvalidEncoding = errors.New("quuid: invalid identifier encoding")
	// ErrInvalidPrefix indicates that a type-specific text prefix is incorrect.
	ErrInvalidPrefix = errors.New("quuid: invalid identifier prefix")
	// ErrNilValue indicates that SQL NULL or JSON null was assigned to a non-nullable type.
	ErrNilValue = errors.New("quuid: nil cannot be assigned to a non-null identifier")
	// ErrZeroEntropy indicates that a test or custom entropy source returned only zero bytes.
	ErrZeroEntropy = errors.New("quuid: entropy source returned an all-zero value")
	// ErrWeakSecret indicates that a keyed-derivation secret is shorter than 32 bytes.
	ErrWeakSecret = errors.New("quuid: secret must contain at least 32 bytes")
	// ErrTimeOutOfRange indicates that a timestamp cannot be represented by the requested identifier layout.
	ErrTimeOutOfRange = errors.New("quuid: time is outside the supported identifier range")
	// ErrCounterExhausted indicates that a monotonic random or timestamp field cannot be incremented further.
	ErrCounterExhausted = errors.New("quuid: monotonic identifier counter exhausted")
	// ErrInvalidSource indicates an unsupported database scanner source type.
	ErrInvalidSource = errors.New("quuid: unsupported database source type")
)
