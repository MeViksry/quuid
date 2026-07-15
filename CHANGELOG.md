# Changelog

All notable changes are documented in this file.

The project follows semantic versioning after version 1.0. Before version 1.0, compatibility-impacting refinements are documented explicitly.

## [Unreleased]

## [0.1.0] - 2026-07-15

### Added

- Standalone RFC 9562 UUID type with no third-party runtime dependency.
- UUIDv4, UUIDv6, process-local monotonic UUIDv7, and UUIDv8 generation.
- Injectable entropy readers and clocks.
- Strict canonical UUID parser and explicit migration parser.
- RFC UUIDv6 example regression test.
- Explicit `BinaryUUID`, `NullUUID`, and `NullBinaryUUID` SQL representations.
- Empty-value-safe nullable scanning.
- `IsZero`, `Equal`, `Compare`, `ToNullUUID`, and `ToNullBinaryUUID` helpers.
- `AppendText` and `AppendBinary` methods.
- 256-bit `ID`, 320-bit `SortableID`, and 384-bit `StrongID`.
- SHA-512-family deterministic and keyed derivation.
- Redacted 256-bit bearer tokens with constant-time digest verification.
- CLI generation for all supported identifier families.
- Unit tests, race tests, fuzz targets, benchmarks, vulnerability scanning, and multi-platform CI.
- CI policy rejecting MD5 and SHA-1 in the dependency tree.
