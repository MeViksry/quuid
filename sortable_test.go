package quuid

import (
	"bytes"
	"testing"
	"time"
)

func TestSortableIDRoundTripAndTime(t *testing.T) {
	when := time.Date(2026, 7, 15, 12, 34, 56, 789_000_000, time.FixedZone("WIB", 7*3600))
	id, err := NewSortableIDAt(when, bytes.NewReader(bytes.Repeat([]byte{0x22}, IDSize)))
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := ParseSortableID(id.String())
	if err != nil {
		t.Fatal(err)
	}
	if !id.Equal(parsed) {
		t.Fatal("round trip mismatch")
	}
	if got, want := parsed.Timestamp(), time.UnixMilli(when.UnixMilli()).UTC(); !got.Equal(want) {
		t.Fatalf("timestamp got %s want %s", got, want)
	}
}

func TestSortableLexicalOrder(t *testing.T) {
	reader := bytes.NewReader(bytes.Repeat([]byte{0x33}, IDSize*2))
	a, err := NewSortableIDAt(time.UnixMilli(100), reader)
	if err != nil {
		t.Fatal(err)
	}
	b, err := NewSortableIDAt(time.UnixMilli(101), reader)
	if err != nil {
		t.Fatal(err)
	}
	if !(a.Compare(b) < 0 && a.String() < b.String()) {
		t.Fatalf("expected sortable order: %s >= %s", a, b)
	}
}

func TestMonotonicGenerator(t *testing.T) {
	fixed := time.UnixMilli(123456)
	g, err := NewGeneratorWith(
		bytes.NewReader(bytes.Repeat([]byte{0x44}, IDSize)),
		func() time.Time { return fixed },
	)
	if err != nil {
		t.Fatal(err)
	}
	a, err := g.NewMonotonicSortableID()
	if err != nil {
		t.Fatal(err)
	}
	b, err := g.NewMonotonicSortableID()
	if err != nil {
		t.Fatal(err)
	}
	if a.Compare(b) >= 0 {
		t.Fatalf("not monotonic: %s >= %s", a, b)
	}
	if !a.Timestamp().Equal(b.Timestamp()) {
		t.Fatal("expected same timestamp")
	}
}
