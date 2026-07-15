package quuid

import "testing"

func BenchmarkNewID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if _, err := NewID(); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkIDString(b *testing.B) {
	id := MustNewID()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = id.String()
	}
}

func BenchmarkParseID(b *testing.B) {
	text := MustNewID().String()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := ParseID(text); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParseUUID(b *testing.B) {
	text := MustNewUUIDv7().String()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := ParseUUID(text); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkNewUUIDv6(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if _, err := NewUUIDv6(); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkNewUUIDv7(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if _, err := NewUUIDv7(); err != nil {
			b.Fatal(err)
		}
	}
}
