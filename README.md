<div align="center">

# quuid

**Strict RFC 9562 UUIDs, explicit SQL representations, monotonic UUIDv7, and larger security-oriented identifiers for Go.**

[![CI](https://img.shields.io/github/actions/workflow/status/MeViksry/quuid/ci.yml?branch=main&style=flat-square&label=CI)](https://github.com/MeViksry/quuid/actions/workflows/ci.yml)
[![Go Reference](https://img.shields.io/badge/Go%20Reference-pkg.go.dev-00ADD8?style=flat-square&logo=go&logoColor=white)](https://pkg.go.dev/github.com/MeViksry/quuid)
[![Go Version](https://img.shields.io/github/go-mod/go-version/MeViksry/quuid?style=flat-square&logo=go&logoColor=white)](https://github.com/MeViksry/quuid/blob/main/go.mod)
[![Release](https://img.shields.io/github/v/release/MeViksry/quuid?include_prereleases&style=flat-square)](https://github.com/MeViksry/quuid/releases)
[![License](https://img.shields.io/github/license/MeViksry/quuid?style=flat-square)](https://github.com/MeViksry/quuid/blob/main/LICENSE)
[![Open Issues](https://img.shields.io/github/issues/MeViksry/quuid?style=flat-square)](https://github.com/MeViksry/quuid/issues)

[Installation](#installation) · [Quick start](#quick-start) · [Database integration](#database-integration) · [Security](#security-statement) · [Migration](#migration-from-googleuuid)

</div>

## Overview

`quuid` is a standalone Go identifier library designed for applications that need stricter behavior and more explicit control than a minimal UUID package normally provides.

It includes:

- RFC 9562 UUIDv4, UUIDv6, UUIDv7, and UUIDv8;
- strict canonical UUID parsing by default;
- a separately named compatibility parser for migrations;
- process-local monotonic UUIDv7 generation;
- injectable clocks and entropy readers;
- explicit text and binary SQL representations without global flags;
- nullable text and binary UUID types;
- allocation-aware `AppendText` and `AppendBinary` methods;
- 256-bit, 320-bit, and 384-bit identifiers for requirements that do not fit inside UUID's fixed 128-bit layout;
- deterministic identifiers using SHA-512 family primitives and HMAC;
- redacted bearer-token handling with constant-time digest verification;
- no third-party runtime dependency;
- no MD5 or SHA-1 import in the module dependency tree.

`quuid` is not a fork of `github.com/google/uuid`, and it does not require that module at runtime. Its `UUID` type has the same underlying `[16]byte` representation, so explicit conversion remains straightforward.

## Security statement

A UUID is an identifier. It is not encryption, authentication, authorization, a signature, or an access-control decision.

RFC 9562 UUIDs are fixed at 128 bits. UUIDv4 and random UUIDv8 contain 122 unconstrained bits after version and variant fields are applied. No implementation can turn that fixed format into an unlimited-strength or universally quantum-resistant primitive.

For applications that need a larger generic search margin, `quuid` provides:

| Type | Binary size | Intended use |
|---|---:|---|
| `UUID` | 16 bytes | Standards-compatible identifiers |
| `ID` | 32 bytes | 256-bit opaque public identifiers |
| `SortableID` | 40 bytes | 64-bit timestamp plus 256-bit entropy |
| `StrongID` | 48 bytes | 384-bit identifiers with larger collision margin |
| `Token` | 32 bytes | 256-bit bearer credentials; store only the digest |

An idealized Grover search reduces the generic preimage work factor of a uniformly random 256-bit value to roughly 128 bits. This is a conservative margin, not a guarantee against every future algorithm, compromised random source, side channel, implementation defect, or operational failure.

## Design principles

1. **Strict by default.** Canonical parsers reject whitespace, braces, URNs, raw hexadecimal UUIDs, malformed separators, and undocumented mixed-endian encodings.

2. **Compatibility is explicit.** `ParseUUIDLoose` exists for migration boundaries. It is intentionally separate from `ParseUUID`.

3. **Database representation is local to the field.** `UUID` writes text. `BinaryUUID` writes exactly 16 bytes. No package-global switch changes behavior for unrelated database connections.

4. **Time-bearing generation is testable.** `NewUUIDv6At`, `NewUUIDv7At`, and `Generator` accept explicit time and entropy sources.

5. **Modern deterministic primitives only.** Deterministic UUIDv8 and larger identifiers use SHA-512/256, SHA-384, or HMAC-SHA-512 family primitives. UUIDv3 and UUIDv5 are intentionally not implemented.

6. **Secrets are separate from identifiers.** `Token.String()` is redacted. The full token is returned only by the explicit `Secret()` method.

## Installation

```bash
go get github.com/MeViksry/quuid@latest
```

Requirements:

- Go 1.23 or newer;
- no external runtime dependency.

## Quick start

```go
package main

import (
    "fmt"
    "log"

    "github.com/MeViksry/quuid"
)

func main() {
    uuid7, err := quuid.NewUUIDv7()
    if err != nil {
        log.Fatal(err)
    }

    publicID, err := quuid.NewID()
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(uuid7)
    fmt.Println(publicID)
}
```

## Choosing an identifier

| Requirement | Recommended API | Ordering | Security notes |
|---|---|---|---|
| Random standards-compatible UUID | `NewUUIDv4` | Random | 122 random bits |
| Time-ordered standards UUID | `NewUUIDv7` | Process-local monotonic | Timestamp is visible |
| RFC UUID compatible with legacy v1 ordering | `NewUUIDv6` | Time ordered | Random node; no MAC address |
| Application-defined deterministic UUID | `DeriveUUIDv8` | Deterministic | Equality of equal input is visible |
| Opaque public identifier | `NewID` | Random | 256 random bits |
| Large time-sortable identifier | `NewSortableID` | Timestamp ordered | Timestamp is visible |
| Larger collision-security margin | `NewStrongID` | Random | 384 random bits |
| Bearer API credential | `NewToken` | Random | Store `TokenDigest`, not plaintext |

## UUID API

### UUIDv4

```go
id, err := quuid.NewUUIDv4()
```

For deterministic testing or an approved custom DRBG/HSM reader:

```go
id, err := quuid.NewUUIDv4FromReader(reader)
```

### UUIDv6

`UUIDv6` reorders the Gregorian timestamp fields used by UUIDv1 into a lexicographically sortable layout. `quuid` uses random clock-sequence and node bits, sampled once per generator, and never reads a network-interface MAC address. The package-level constructor is process-local monotonic and advances the embedded timestamp by one 100-nanosecond tick when the clock does not advance.

```go
id, err := quuid.NewUUIDv6()
```

Generate for an explicit time:

```go
id, err := quuid.NewUUIDv6At(timestamp, reader)
```

The implementation includes a conformance regression test using the RFC UUIDv6 example:

```text
1ec9414c-232a-6b00-b3c8-9e6bdeced846
```

### UUIDv7

The package-level constructor uses a concurrency-safe process-local monotonic generator:

```go
id, err := quuid.NewUUIDv7()
```

When multiple values are generated in the same millisecond, the 74-bit random field is incremented. If the wall clock moves backwards, generation remains on the last emitted millisecond and continues incrementing. This improves index locality and guarantees ordering inside one process.

It does not provide global ordering across processes, containers, hosts, or regions.

For explicit time without shared monotonic state:

```go
id, err := quuid.NewUUIDv7At(timestamp, reader)
```

For an injected clock with monotonic state:

```go
generator, err := quuid.NewGeneratorWith(reader, clock)
if err != nil {
    return err
}

id, err := generator.NewUUIDv7()
```

Read the timestamp:

```go
timestamp, err := id.Time()
```

### UUIDv8

Random UUIDv8:

```go
id, err := quuid.NewUUIDv8()
```

Deterministic UUIDv8 using SHA-512/256 and domain separation:

```go
id := quuid.DeriveUUIDv8("invoice", []byte("INV-2026-00042"))
```

Keyed deterministic UUIDv8:

```go
id, err := quuid.DeriveUUIDv8Keyed(
    secret,
    "customer",
    []byte("customer@example.com"),
)
```

The keyed secret must contain at least 32 bytes from a secure random source and should be stored in a KMS, HSM, or managed secret store.

## Strict and compatibility parsing

### Strict parser

```go
id, err := quuid.ParseUUID("018fbd2e-7b46-7cc0-98c4-89e6f6dc0c22")
```

`ParseUUID` accepts exactly the canonical 36-character representation. `ParseUUIDBytes` applies the same strict rules to canonical `[]byte` input from SQL drivers, queues, and wire protocols without converting through string first. Both reject:

```text
" 018fbd2e-7b46-7cc0-98c4-89e6f6dc0c22"
"018fbd2e-7b46-7cc0-98c4-89e6f6dc0c22 "
{018fbd2e-7b46-7cc0-98c4-89e6f6dc0c22}
urn:uuid:018fbd2e-7b46-7cc0-98c4-89e6f6dc0c22
018fbd2e7b467cc098c489e6f6dc0c22
```

Uppercase hexadecimal digits are accepted and canonical output is always lowercase.

### Compatibility parser

Use the loose parser only at a controlled migration boundary:

```go
id, err := quuid.ParseUUIDLoose(input)
```

It accepts:

- canonical UUID text;
- surrounding whitespace;
- `urn:uuid:` values;
- brace-wrapped canonical values;
- exactly 32 hexadecimal digits.

It does not implement Microsoft mixed-endian reinterpretation.

## UUID inspection

```go
id := quuid.MustParseUUID("018fbd2e-7b46-7cc0-98c4-89e6f6dc0c22")

fmt.Println(id.Version())
fmt.Println(id.Variant())
fmt.Println(id.IsZero())
fmt.Println(id.Equal(other))
fmt.Println(id.Compare(other))
```

Validation options:

```go
err := quuid.ValidateUUID(text)
err = quuid.ValidateRFC9562UUID(text)
```

`ValidateUUID` checks canonical syntax. `ValidateRFC9562UUID` additionally requires the RFC variant and an assigned version from 1 through 8.

## Database integration

### Portable text representation

`UUID` implements `database/sql.Scanner` and `driver.Valuer`.

```go
type Account struct {
    ID quuid.UUID
}
```

`UUID.Value()` returns the canonical string. This is suitable for:

- PostgreSQL `uuid`;
- MySQL `CHAR(36)`;
- SQLite `TEXT`;
- text-compatible database drivers.

Scanning accepts canonical UUID text or exactly 16 binary bytes.

### Binary representation

Use `BinaryUUID` when the database column expects exactly 16 bytes:

```go
type Account struct {
    ID quuid.BinaryUUID
}
```

Convert between logical and binary representations:

```go
id := quuid.MustNewUUIDv7()

binary := id.Binary()
logical := binary.UUID()
```

`BinaryUUID.Value()` returns a defensive `[]byte` copy with length 16. It is suitable for:

- MySQL or MariaDB `BINARY(16)`;
- SQLite `BLOB`;
- PostgreSQL `bytea`;
- drivers that explicitly require `[]byte`.

The bytes use standard RFC network order. `quuid` does not silently apply MySQL's optional `UUID_TO_BIN(..., 1)` field-swapping convention.

### Text and binary columns in the same application

No global mode is required:

```go
type Record struct {
    PostgreSQLID quuid.UUID
    MySQLID      quuid.BinaryUUID
}
```

Each field controls its own `driver.Value`. Multiple drivers and ORMs can coexist without process-wide behavior changes.

### Nullable UUID

```go
type Record struct {
    ParentID quuid.NullUUID
}
```

`NullUUID.Scan` treats these values as invalid/null:

- SQL `NULL`;
- empty string;
- empty byte slice.

A malformed non-empty value returns an error and leaves `Valid` false.

```go
nullable := quuid.ToNullUUID(id)
```

`ToNullUUID(quuid.NilUUID)` returns an invalid value. `NewNullUUID(id)` marks the value valid even when `id` is `NilUUID`.

### Nullable binary UUID

```go
type Record struct {
    ParentID quuid.NullBinaryUUID
}
```

Valid values are written as exactly 16 bytes. Invalid values are written as SQL `NULL`.

## JSON, text, and binary encoding

`UUID`, `ID`, `StrongID`, and `SortableID` implement the standard marshal/unmarshal interfaces for JSON, text, and binary data.

```go
encoded, err := json.Marshal(id)
```

UUID JSON uses canonical text:

```json
"018fbd2e-7b46-7cc0-98c4-89e6f6dc0c22"
```

Non-nullable types reject JSON `null`. Use their nullable counterpart when null is a valid application state.

### Allocation-aware append methods

The identifier types expose:

```go
AppendText(dst []byte) ([]byte, error)
AppendBinary(dst []byte) ([]byte, error)
```

Example:

```go
buffer := make([]byte, 0, 64)
buffer, err := id.AppendText(buffer)
```

These methods allow encoders to append into reusable buffers instead of allocating a new slice for each value.

## Larger identifiers

### ID

`ID` is a 256-bit opaque identifier with canonical Crockford Base32 text prefixed by `qid_`.

```go
id, err := quuid.NewID()
parsed, err := quuid.ParseID(id.String())
```

Deterministic public ID:

```go
id := quuid.DeriveID("customer", []byte("customer-42"))
```

Keyed deterministic ID:

```go
id, err := quuid.DeriveIDKeyed(secret, "customer", input)
```

### SortableID

`SortableID` contains:

```text
8-byte unsigned Unix millisecond timestamp || 32-byte random entropy
```

```go
id, err := quuid.NewSortableID()
fmt.Println(id.Timestamp())
```

Use a generator for strict process-local monotonic order:

```go
g := quuid.NewGenerator()
id, err := g.NewMonotonicSortableID()
```

### StrongID

`StrongID` is a 384-bit identifier prefixed by `qsid_`.

```go
id, err := quuid.NewStrongID()
```

It is intended for systems that justify a larger quantum collision-search margin and can accept the additional storage and index cost.

## Bearer tokens

Generate a token once and return its full value only to the caller:

```go
token, err := quuid.NewToken()
if err != nil {
    return err
}

plaintext := token.Secret()
digest := token.Digest()
```

Store `digest`, not `plaintext`.

Verify a presented token:

```go
valid := quuid.VerifyToken(presented, digest)
```

`Token.String()` and normal formatting are redacted to reduce accidental log disclosure:

```text
qtk_…a1b2c3
```

Do not use UUID, ID, SortableID, or StrongID possession as authorization.

## Migration from google/uuid

The underlying representation of both UUID types is `[16]byte`, so explicit conversion is direct:

```go
import (
    googleuuid "github.com/google/uuid"
    "github.com/MeViksry/quuid"
)

googleID := googleuuid.New()
quuidID := quuid.UUID(googleID)
back := googleuuid.UUID(quuidID)
```

Important behavior changes:

| Area | Migration impact |
|---|---|
| Parsing | `ParseUUID` is canonical and strict; use `ParseUUIDLoose` only for controlled legacy input |
| SQL | `UUID` writes text; use `BinaryUUID` for `BINARY(16)` or BLOB columns |
| Null values | Empty string and empty bytes produce invalid `NullUUID` rather than a valid zero value |
| UUIDv7 | Package constructor is process-local monotonic |
| UUIDv6 | Process-local monotonic timestamp, random node bits, and no MAC address collection |
| Deterministic UUID | Use UUIDv8 with SHA-512/256 or HMAC; no UUIDv3/UUIDv5 helpers |
| Randomness | No global random pool toggle |

## Hardening against known UUID-library failure modes

The following table documents how `quuid` handles the classes of problems raised in public UUID-library issue trackers.

| Concern | Related upstream issue | `quuid` behavior |
|---|---:|---|
| SQL binary representation | [`google/uuid#46`](https://github.com/google/uuid/issues/46) | Explicit `BinaryUUID` and `NullBinaryUUID`; no global flag |
| Legacy MD5/SHA-1 hashes | [`google/uuid#218`](https://github.com/google/uuid/issues/218) | No UUIDv3/v5 API, no MD5/SHA-1 imports, SHA-512-family UUIDv8 derivation |
| Future Go standard UUID package | [`google/uuid#221`](https://github.com/google/uuid/issues/221) | No dependency lock-in; UUID backend is isolated and can adopt a verified standard implementation without changing public types |
| Allocation-aware encoding | [`google/uuid#193`](https://github.com/google/uuid/issues/193) | `AppendText` and `AppendBinary` implemented |
| PostgreSQL-style UUIDv7 locality | [`google/uuid#191`](https://github.com/google/uuid/issues/191) | Process-local monotonic UUIDv7 with rollback handling |
| Injectable time source | [`google/uuid#165`](https://github.com/google/uuid/issues/165) | `NewUUIDv6At`, `NewUUIDv7At`, and `NewGeneratorWith` |
| Incorrect UUIDv6 field layout | [`google/uuid#159`](https://github.com/google/uuid/issues/159) | RFC layout plus published UUIDv6 example regression test |
| UUIDv6 decoded time moves backward | [`google/uuid#199`](https://github.com/google/uuid/issues/199) | Process-local monotonic UUIDv6 timestamp with 100-nanosecond stepping |
| Leading/trailing spaces accepted | [`google/uuid#147`](https://github.com/google/uuid/issues/147) | Strict parser rejects whitespace; compatibility parser is separately named |
| Nil UUID documentation ambiguity | [`google/uuid#112`](https://github.com/google/uuid/issues/112) | `NilUUID` is explicitly documented as the all-zero UUID and not SQL `NULL` |
| Missing zero/null helpers | [`google/uuid#113`](https://github.com/google/uuid/issues/113) | `IsZero`, `ToNullUUID`, `ToNullBinaryUUID` |
| Empty string marked valid in nullable scan | [`google/uuid#109`](https://github.com/google/uuid/issues/109) | Empty string and empty bytes set `Valid=false` |
| Missing UUIDv8 reference API | [`google/uuid#103`](https://github.com/google/uuid/issues/103) | Random, deterministic, and keyed UUIDv8 APIs with version/variant tests |
| Missing equality helper | [`google/uuid#100`](https://github.com/google/uuid/issues/100) | `Equal` and `Compare` methods |
| Process-global random pool behavior | [`google/uuid#86`](https://github.com/google/uuid/issues/86) | No mutable global random-pool switch; generators own their reader |
| Unexpected panic from convenience constructor | [`google/uuid#97`](https://github.com/google/uuid/issues/97) | Default constructors return errors; panic behavior is limited to explicitly named `Must...` APIs |
| Overly permissive or mixed-endian parsing | [`google/uuid#60`](https://github.com/google/uuid/issues/60) | Canonical strict parser and explicitly bounded loose parser |
| Ambiguous output casing | [`google/uuid#94`](https://github.com/google/uuid/issues/94) | Canonical lowercase output; uppercase hexadecimal input remains valid |
| Variant documentation ambiguity | [`google/uuid#91`](https://github.com/google/uuid/issues/91) | Public `Variant` type and explicit RFC validation |

A mitigation table is not a substitute for review. Every behavior above is covered by tests or a CI policy check.

## Concurrency and ordering guarantees

- Package-level random constructors are safe for concurrent use.
- `Generator` methods are protected by a mutex.
- `NewUUIDv6` and `Generator.NewUUIDv6` are strictly increasing inside one generator instance.
- `NewUUIDv7` and `Generator.NewUUIDv7` are strictly increasing inside one generator instance.
- `NewMonotonicSortableID` is strictly increasing inside one generator instance.
- No distributed total-order claim is made.
- Database uniqueness constraints remain mandatory.

## Error handling

Random constructors return errors:

```go
id, err := quuid.NewID()
```

`Must...` helpers panic and should be limited to initialization paths where entropy failure is unrecoverable:

```go
id := quuid.MustNewUUIDv7()
```

Parsers return stable sentinel errors where appropriate:

```go
errors.Is(err, quuid.ErrInvalidLength)
errors.Is(err, quuid.ErrInvalidEncoding)
errors.Is(err, quuid.ErrInvalidPrefix)
errors.Is(err, quuid.ErrNilValue)
errors.Is(err, quuid.ErrWeakSecret)
errors.Is(err, quuid.ErrTimeOutOfRange)
```

## Command-line tool

Build the CLI:

```bash
go install github.com/MeViksry/quuid/cmd/quuid@latest
```

Generate values:

```bash
quuid -type uuid7
quuid -type uuid6 -n 10
quuid -type id -n 5
quuid -type strong
quuid -type sortable
quuid -type token
quuid -type uuid8 -namespace invoice -data INV-2026-00042
quuid -type uuid7 -json
quuid -version
```

## Automated releases and packages

GitHub automation publishes release assets and packages without adding runtime dependencies to the module. The workflows run on the repository's `self-hosted` runner label, matching the CI setup.

Every push to `main` updates a single prerelease named `nightly` with fresh CLI archives and checksums. This keeps the Releases page active without creating a noisy release for every commit.

Push a semantic version tag to create an official GitHub Release:

```bash
git tag v0.1.0
git push origin v0.1.0
```

The release workflow verifies the candidate, builds `quuid` CLI archives for Linux, macOS, and Windows on amd64 and arm64, generates `checksums.txt`, and uploads everything to the GitHub Release.

The package workflow publishes a GitHub Container Registry image for the CLI:

- `ghcr.io/meviksry/quuid:edge` on pushes to `main`;
- `ghcr.io/meviksry/quuid:sha-<commit>` on every packaged push;
- `ghcr.io/meviksry/quuid:vX.Y.Z` on release tags;
- `ghcr.io/meviksry/quuid:latest` on stable release tags.

The package runner must have Docker available and permission to push packages with `GITHUB_TOKEN`.

## Testing

Run the complete local verification suite:

```bash
make check
```

Individual commands:

```bash
go test ./...
go test -race ./...
go vet ./...
go test -run '^$' -bench . -benchmem ./...
```

The unit suite includes a fixed-clock concurrent UUIDv7 stress test. The benchmark suite includes sequential and parallel UUIDv7 generation plus parser allocation checks.

Parser fuzzing:

```bash
go test -fuzz='^FuzzParseUUID$' -fuzztime=30s .
go test -fuzz='^FuzzParseUUIDBytes$' -fuzztime=30s .
go test -fuzz='^FuzzParseUUIDLoose$' -fuzztime=30s .
go test -fuzz='^FuzzParseID$' -fuzztime=30s .
go test -fuzz='^FuzzParseStrongID$' -fuzztime=30s .
go test -fuzz='^FuzzParseSortableID$' -fuzztime=30s .
```

CI verifies multiple supported Go versions and operating systems, runs the race detector, `go vet`, vulnerability scanning, and a dependency-tree check that rejects `crypto/md5` and `crypto/sha1`.

## Compatibility policy

Before v1.0, small API refinements may occur in minor releases and are documented in `CHANGELOG.md`.

After v1.0:

- exported API compatibility follows semantic versioning;
- parsing behavior remains strict by default;
- security-sensitive behavior changes require regression tests and release notes;
- deprecated behavior is not silently re-enabled through global switches.

## Project documentation

- [`DESIGN.md`](DESIGN.md) describes layouts, generation, parsing, and database representation.
- [`SECURITY.md`](SECURITY.md) defines the vulnerability-reporting process and cryptographic scope.
- [`CONTRIBUTING.md`](CONTRIBUTING.md) describes development and pull-request requirements.
- [`CHANGELOG.md`](CHANGELOG.md) records release changes.

## Contributing

Focused contributions are welcome. New functionality should include:

- tests for success and failure paths;
- parser fuzz seeds when input handling changes;
- documentation for security assumptions;
- no custom cryptographic primitive;
- no hidden global configuration that changes unrelated callers.

See [`CONTRIBUTING.md`](CONTRIBUTING.md).

## License

`quuid` is available under the MIT License. See [`LICENSE`](LICENSE).
