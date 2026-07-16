// Package quuid provides strict RFC 9562 UUIDs and larger identifiers for Go.
//
// Use UUIDv7 for standards-compatible, time-ordered database identifiers. Use
// BinaryUUID when a SQL driver requires exactly 16 bytes. Use ID for 256-bit
// public identifiers, StrongID for a larger collision-security margin, and
// Token for bearer secrets whose plaintext must not be logged or stored.
//
// ParseUUID and ParseUUIDBytes are intentionally strict. ParseUUIDLoose is
// available only for controlled migration boundaries.
package quuid
