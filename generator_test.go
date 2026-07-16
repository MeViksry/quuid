package quuid

import (
	"bytes"
	"errors"
	"sort"
	"sync"
	"testing"
	"time"
)

func TestGeneratorConcurrentUniqueness(t *testing.T) {
	g := NewGenerator()
	const total = 1000

	ids := make(chan ID, total)
	errs := make(chan error, total)
	var wg sync.WaitGroup
	wg.Add(total)
	for i := 0; i < total; i++ {
		go func() {
			defer wg.Done()
			id, err := g.NewID()
			if err != nil {
				errs <- err
				return
			}
			ids <- id
		}()
	}
	wg.Wait()
	close(ids)
	close(errs)

	for err := range errs {
		t.Fatal(err)
	}
	seen := make(map[ID]struct{}, total)
	for id := range ids {
		if _, exists := seen[id]; exists {
			t.Fatalf("duplicate ID: %s", id)
		}
		seen[id] = struct{}{}
	}
	if len(seen) != total {
		t.Fatalf("got %d IDs, want %d", len(seen), total)
	}
}

func TestGeneratorConcurrentUUIDv7Stress(t *testing.T) {
	const (
		workers   = 64
		perWorker = 1024
		total     = workers * perWorker
	)
	fixed := time.UnixMilli(1_800_000_000_000)
	g, err := NewGeneratorWith(
		bytes.NewReader(bytes.Repeat([]byte{0x7a}, 10)),
		func() time.Time { return fixed },
	)
	if err != nil {
		t.Fatal(err)
	}

	start := make(chan struct{})
	ids := make(chan UUID, total)
	errs := make(chan error, total)
	var wg sync.WaitGroup
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			<-start
			for j := 0; j < perWorker; j++ {
				id, err := g.NewUUIDv7()
				if err != nil {
					errs <- err
					return
				}
				ids <- id
			}
		}()
	}
	close(start)
	wg.Wait()
	close(ids)
	close(errs)

	for err := range errs {
		t.Fatal(err)
	}

	seen := make(map[UUID]struct{}, total)
	ordered := make([]UUID, 0, total)
	for id := range ids {
		if id.Version() != Version7 || id.Variant() != VariantRFC9562 {
			t.Fatalf("invalid UUIDv7 metadata: %s", id)
		}
		got, err := id.Time()
		if err != nil {
			t.Fatal(err)
		}
		if !got.Equal(fixed.UTC()) {
			t.Fatalf("timestamp changed under fixed clock: got %s, want %s", got, fixed.UTC())
		}
		if _, ok := seen[id]; ok {
			t.Fatalf("duplicate UUIDv7 under concurrency: %s", id)
		}
		seen[id] = struct{}{}
		ordered = append(ordered, id)
	}
	if len(seen) != total {
		t.Fatalf("got %d UUIDs, want %d", len(seen), total)
	}

	sort.Slice(ordered, func(i, j int) bool {
		return ordered[i].Compare(ordered[j]) < 0
	})
	for i := 1; i < len(ordered); i++ {
		if ordered[i-1].Compare(ordered[i]) >= 0 {
			t.Fatalf("sorted UUIDv7 sequence regressed at %d: %s >= %s", i, ordered[i-1], ordered[i])
		}
	}
}

func TestGeneratorMonotonicUUIDSequences(t *testing.T) {
	fixed := time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)

	v6Generator, err := NewGeneratorWith(
		bytes.NewReader(bytes.Repeat([]byte{0x55}, 8)),
		func() time.Time { return fixed },
	)
	if err != nil {
		t.Fatal(err)
	}
	previousV6, err := v6Generator.NewUUIDv6()
	if err != nil {
		t.Fatal(err)
	}
	previousV6Time, err := previousV6.Time()
	if err != nil {
		t.Fatal(err)
	}
	for i := 1; i < 10_000; i++ {
		current, err := v6Generator.NewUUIDv6()
		if err != nil {
			t.Fatal(err)
		}
		currentTime, err := current.Time()
		if err != nil {
			t.Fatal(err)
		}
		if current.Compare(previousV6) <= 0 || !currentTime.After(previousV6Time) {
			t.Fatalf("UUIDv6 sequence regressed at %d", i)
		}
		previousV6 = current
		previousV6Time = currentTime
	}

	v7Generator, err := NewGeneratorWith(
		bytes.NewReader(bytes.Repeat([]byte{0x66}, 10)),
		func() time.Time { return fixed },
	)
	if err != nil {
		t.Fatal(err)
	}
	previousV7, err := v7Generator.NewUUIDv7()
	if err != nil {
		t.Fatal(err)
	}
	for i := 1; i < 10_000; i++ {
		current, err := v7Generator.NewUUIDv7()
		if err != nil {
			t.Fatal(err)
		}
		if current.Compare(previousV7) <= 0 {
			t.Fatalf("UUIDv7 sequence regressed at %d", i)
		}
		previousV7 = current
	}
}

func TestGeneratorUUIDCounterExhaustion(t *testing.T) {
	fixed := time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)

	v6 := &Generator{
		reader:          bytes.NewReader(bytes.Repeat([]byte{1}, 8)),
		clock:           func() time.Time { return fixed },
		v6Initialized:   true,
		lastV6Timestamp: 1<<60 - 1,
	}
	if _, err := v6.NewUUIDv6(); !errors.Is(err, ErrCounterExhausted) {
		t.Fatalf("expected UUIDv6 exhaustion, got %v", err)
	}

	lastV7 := UUID{0, 0, 0, 0, 0, 0, 0x7f, 0xff, 0xbf, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	v7 := &Generator{
		reader:        bytes.NewReader(bytes.Repeat([]byte{1}, 10)),
		clock:         func() time.Time { return fixed },
		v7Initialized: true,
		lastV7MS:      uint64(fixed.UnixMilli()),
		lastV7:        lastV7,
	}
	putUUIDv7Milliseconds(&v7.lastV7, v7.lastV7MS)
	if _, err := v7.NewUUIDv7(); !errors.Is(err, ErrCounterExhausted) {
		t.Fatalf("expected UUIDv7 exhaustion, got %v", err)
	}
}
