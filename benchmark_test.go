package quuid

import (
	"sync"
	"testing"
)

func BenchmarkNewID(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := NewID(); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkIDString(b *testing.B) {
	id := MustNewID()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = id.String()
	}
}

func BenchmarkParseID(b *testing.B) {
	text := MustNewID().String()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := ParseID(text); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParseUUID(b *testing.B) {
	text := MustNewUUIDv7().String()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := ParseUUID(text); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParseUUIDBytes(b *testing.B) {
	text := []byte(MustNewUUIDv7().String())
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := ParseUUIDBytes(text); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkUUIDAppendText(b *testing.B) {
	id := MustNewUUIDv7()
	buf := make([]byte, 0, 36)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf = buf[:0]
		var err error
		buf, err = id.AppendText(buf)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkNewUUIDv6(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := NewUUIDv6(); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkNewUUIDv7(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := NewUUIDv7(); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkNewUUIDv7Parallel(b *testing.B) {
	b.ReportAllocs()
	runParallelUUIDBenchmark(b, NewUUIDv7)
}

func runParallelUUIDBenchmark(b *testing.B, fn func() (UUID, error)) {
	var (
		once sync.Once
		err  error
	)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if _, callErr := fn(); callErr != nil {
				once.Do(func() { err = callErr })
				return
			}
		}
	})
	if err != nil {
		b.Fatal(err)
	}
}
