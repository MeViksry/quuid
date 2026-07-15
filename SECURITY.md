# Security Policy

## Supported versions

Until version 1.0, security fixes are released for the newest published minor version. After version 1.0, the latest minor release of the current major version is supported.

## Reporting a vulnerability

Do not disclose an unpatched vulnerability in a public issue, discussion, or pull request.

Use GitHub private vulnerability reporting for this repository when available:

1. open the repository Security tab;
2. choose **Report a vulnerability**;
3. include the affected version or commit, a minimal reproduction, expected behavior, observed behavior, and practical impact.

A complete report should also state:

- Go version and operating system;
- whether untrusted input is required;
- whether confidentiality, integrity, availability, authentication, or authorization is affected;
- any known workaround;
- suggested remediation, when available.

The maintainer will attempt to reproduce the issue, prepare a regression test and fix, and coordinate disclosure. This open-source project does not promise a fixed response-time service level.

## Cryptographic scope

`quuid` does not implement post-quantum public-key encryption, key exchange, or digital signatures.

The library uses:

- `crypto/rand` for default random generation;
- SHA-512/256 and SHA-384 for deterministic public identifiers;
- HMAC-SHA-512 family primitives for keyed derivation;
- constant-time digest comparison for bearer-token verification.

The module intentionally does not import MD5 or SHA-1.

References to a quantum-search margin describe generic work-factor estimates for uniformly random outputs. They are not a certification or guarantee against every future quantum algorithm, compromised entropy source, endpoint compromise, side channel, implementation defect, or operational failure.

## Security boundaries

Identifiers are not authorization credentials. An attacker who learns an object ID must not automatically gain access to that object.

Applications must independently enforce:

- authentication;
- authorization;
- tenant isolation;
- database uniqueness constraints;
- rate limits where guessing is relevant;
- secure transport;
- secret management;
- audit logging that excludes bearer secrets.

## Token handling

- return token plaintext only once at issuance;
- store `TokenDigest`, not plaintext `Token`;
- never write `Token.Secret()` to application logs;
- rotate compromised tokens;
- use application-level expiry and revocation;
- protect databases and backups containing token digests.

## Keyed derivation

Keyed deterministic identifiers require at least 32 bytes of secret material from an approved random source. Store the secret in a KMS, HSM, or managed secret store. Do not commit it to source control or embed it in a client application.

## Database behavior

Choose the SQL representation deliberately:

- use `UUID` for native UUID or text columns;
- use `BinaryUUID` for exact 16-byte columns;
- do not mix MySQL field-swapped UUID binary order with standard RFC network order without an explicit migration layer.

## Dependency and toolchain hygiene

- use a supported Go release;
- review CI and vulnerability-scanner findings;
- keep GitHub Actions dependencies updated;
- verify release tags and checksums;
- run `go test -race ./...` for changes affecting concurrent generation.
