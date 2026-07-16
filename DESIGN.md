# Design

## Goals

`quuid` is designed around the following constraints:

1. implement RFC 9562 UUID layouts without inheriting undocumented parser or SQL behavior;
2. keep strict validation as the default application boundary;
3. make compatibility behavior explicit and locally visible;
4. support text and binary database columns in the same process without global flags;
5. provide monotonic time-ordered identifiers with injectable clocks;
6. use the operating-system CSPRNG by default;
7. avoid legacy cryptographic hashes in deterministic APIs and the dependency tree;
8. separate public identifiers from bearer credentials;
9. expose Go encoding interfaces without reflection;
10. remain safe under concurrent use.

## Dependency policy

The module has no third-party runtime dependency. UUID generation, parsing, encoding, SQL integration, and deterministic derivation use the Go standard library.

The public `UUID` type is `[16]byte`. This preserves explicit conversion compatibility with other Go UUID types that use the same representation while allowing `quuid` to control its own behavior.

## UUID representation

`UUID` contains 16 bytes in RFC network order. Canonical text is lowercase hexadecimal with separators at byte-group boundaries:

```text
xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
```

`ParseUUID` and `ParseUUIDBytes` accept only this 36-character shape. `ParseUUIDLoose` is a separately named migration API for a documented, bounded set of legacy text formats.

No parser performs Microsoft mixed-endian field reinterpretation.

## UUID versions

### UUIDv4

UUIDv4 reads 16 bytes with `io.ReadFull`, rejects an all-zero entropy block, then applies version and RFC variant bits.

### UUIDv6

UUIDv6 uses the 60-bit count of 100-nanosecond intervals since 1582-10-15. Timestamp bits are reordered according to RFC 9562. The remaining clock-sequence and node bits are sampled randomly once per generator. Hardware MAC addresses are never queried. The package-level constructor and `Generator.NewUUIDv6` keep the timestamp strictly increasing inside one generator by advancing one 100-nanosecond tick when the clock is equal or moves backward.

### UUIDv7

UUIDv7 stores a 48-bit Unix millisecond timestamp followed by 74 random bits after version and variant fields.

The package-level constructor and `Generator.NewUUIDv7` provide process-local monotonicity:

- when time advances, a new random field is sampled;
- in the same millisecond, the 74-bit random field is incremented;
- during clock rollback, the previous millisecond is retained and the field is incremented;
- exhaustion returns `ErrCounterExhausted`.

This does not create a distributed total order.

### UUIDv8

Random UUIDv8 fills all bytes from a CSPRNG and sets version and variant fields.

Deterministic UUIDv8 uses SHA-512/256 or HMAC-SHA-512/256 with length-prefixed framing and a versioned domain string. UUIDv3 and UUIDv5 are intentionally absent because they require MD5 and SHA-1.

## Larger identifiers

### ID

`ID` is exactly 32 bytes. Random construction provides 256 random bits. Deterministic construction uses SHA-512/256. Keyed construction uses HMAC-SHA-512/256.

### StrongID

`StrongID` is exactly 48 bytes. Random construction provides 384 random bits. Deterministic construction uses SHA-384. Keyed construction uses HMAC-SHA-512 truncated to 48 bytes.

### SortableID

`SortableID` is 40 bytes:

- bytes 0 through 7: unsigned big-endian Unix milliseconds;
- bytes 8 through 39: 256 random bits.

The canonical Base32 alphabet preserves byte ordering for equal-length encoded values.

### Token

`Token` is a separate 32-byte bearer-secret type. `String` is redacted. Full plaintext is available only through `Secret` for issuance. `TokenDigest` is SHA-512/256 over framed, domain-separated token bytes. Verification uses `crypto/subtle.ConstantTimeCompare`.

## Deterministic framing

Each hash input is encoded as a sequence of frames:

```text
uint64_be(length) || bytes
```

This prevents concatenation ambiguity. Each function includes a versioned domain string so outputs from different APIs or revisions remain separated.

## Text encoding

`ID`, `StrongID`, `SortableID`, and `Token` use unpadded Crockford Base32 with a type prefix:

- `qid_` for `ID`;
- `qsid_` for `StrongID`;
- `qsort_` for `SortableID`;
- `qtk_` for `Token`.

Canonical output is lowercase. Parsers accept uppercase and Crockford human-entry aliases, but reject whitespace and non-canonical trailing bits.

## SQL representation

`UUID.Value` returns canonical text. `BinaryUUID.Value` returns a defensive 16-byte slice. Both scanners accept canonical text and exact 16-byte binary values.

The explicit wrapper design avoids a global mode. Applications can use multiple database drivers and representations simultaneously.

`NullUUID` and `NullBinaryUUID` treat SQL `NULL`, empty string, and empty byte slice as invalid. Malformed non-empty values return an error and leave the value invalid.

## Encoding interfaces

Identifier types implement relevant standard interfaces:

- `encoding.TextMarshaler` and `encoding.TextUnmarshaler`;
- `encoding.BinaryMarshaler` and `encoding.BinaryUnmarshaler`;
- `AppendText` and `AppendBinary` methods;
- JSON marshal and unmarshal;
- `database/sql.Scanner` and `driver.Valuer` where database storage is appropriate.

Non-nullable types reject JSON `null` and SQL `NULL`.

## Randomness

Default generation uses `crypto/rand.Reader`. There is no mutable global random-pool toggle. A `Generator` owns its reader and clock, making behavior explicit and testable.

All constructors use `io.ReadFull` and return errors rather than silently accepting partial entropy reads.

## Error behavior

Parsing and scanning are fail-closed. A receiver is updated only after a complete successful parse. Nullable receivers are reset to invalid on malformed input.

`Must...` functions are convenience methods for initialization paths and panic on error. Regular application paths should use error-returning constructors.
