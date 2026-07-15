package quuid

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"io"
	"sync"
	"time"
)

// Generator owns an entropy source, clock, and monotonic state. Its methods are
// safe for concurrent use. A Generator never uses global randomness settings.
type Generator struct {
	mu     sync.Mutex
	reader io.Reader
	clock  func() time.Time

	sortableInitialized bool
	lastSortableMS      uint64
	lastSortableEntropy ID

	v6Initialized   bool
	lastV6Timestamp uint64
	lastV6          UUID

	v7Initialized bool
	lastV7MS      uint64
	lastV7        UUID
}

var defaultGenerator = NewGenerator()

// NewGenerator returns a generator backed by crypto/rand.Reader and time.Now.
func NewGenerator() *Generator {
	return &Generator{reader: rand.Reader, clock: time.Now}
}

// NewGeneratorWith returns a generator with explicit entropy and clock sources.
// It is intended for deterministic tests, approved DRBG/HSM integrations, and
// applications that need control over time-bearing identifiers.
func NewGeneratorWith(reader io.Reader, clock func() time.Time) (*Generator, error) {
	if reader == nil {
		return nil, fmt.Errorf("quuid: nil entropy reader")
	}
	if clock == nil {
		return nil, fmt.Errorf("quuid: nil clock")
	}
	return &Generator{reader: reader, clock: clock}, nil
}

// NewID returns a new identifier using the generator entropy source.
func (g *Generator) NewID() (ID, error) {
	if g == nil {
		return ID{}, fmt.Errorf("quuid: nil Generator")
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	return NewIDFromReader(g.reader)
}

// NewStrongID returns a new strong identifier using the generator entropy source.
func (g *Generator) NewStrongID() (StrongID, error) {
	if g == nil {
		return StrongID{}, fmt.Errorf("quuid: nil Generator")
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	return NewStrongIDFromReader(g.reader)
}

// NewSortableID returns a new sortable identifier using the generator clock and entropy source.
func (g *Generator) NewSortableID() (SortableID, error) {
	if g == nil {
		return SortableID{}, fmt.Errorf("quuid: nil Generator")
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	return NewSortableIDAt(g.clock(), g.reader)
}

// NewMonotonicSortableID returns a strictly increasing process-local SortableID.
func (g *Generator) NewMonotonicSortableID() (SortableID, error) {
	if g == nil {
		return SortableID{}, fmt.Errorf("quuid: nil Generator")
	}
	g.mu.Lock()
	defer g.mu.Unlock()

	nowMS := g.clock().UnixMilli()
	if nowMS < 0 {
		return SortableID{}, ErrTimeOutOfRange
	}
	ms := uint64(nowMS)

	if !g.sortableInitialized || ms > g.lastSortableMS {
		entropy, err := NewIDFromReader(g.reader)
		if err != nil {
			return SortableID{}, err
		}
		g.lastSortableMS = ms
		g.lastSortableEntropy = entropy
		g.sortableInitialized = true
	} else {
		ms = g.lastSortableMS
		if !incrementBigEndian(g.lastSortableEntropy[:]) {
			return SortableID{}, ErrCounterExhausted
		}
	}

	var id SortableID
	binary.BigEndian.PutUint64(id[:8], ms)
	copy(id[8:], g.lastSortableEntropy[:])
	return id, nil
}

// NewUUIDv4 returns a UUIDv4 using the generator's entropy source.
func (g *Generator) NewUUIDv4() (UUID, error) {
	if g == nil {
		return NilUUID, fmt.Errorf("quuid: nil Generator")
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	return NewUUIDv4FromReader(g.reader)
}

// NewUUIDv6 returns a strictly increasing process-local UUIDv6. Its random
// clock-sequence and node bits are sampled once per Generator. When the clock
// does not advance, the 60-bit timestamp is advanced by one 100ns tick.
func (g *Generator) NewUUIDv6() (UUID, error) {
	if g == nil {
		return NilUUID, fmt.Errorf("quuid: nil Generator")
	}
	g.mu.Lock()
	defer g.mu.Unlock()

	timestamp, err := uuidV6Timestamp(g.clock())
	if err != nil {
		return NilUUID, err
	}
	if !g.v6Initialized {
		id, err := newUUIDv6FromTimestamp(timestamp, g.reader)
		if err != nil {
			return NilUUID, err
		}
		g.v6Initialized = true
		g.lastV6Timestamp = timestamp
		g.lastV6 = id
		return id, nil
	}
	if timestamp <= g.lastV6Timestamp {
		if g.lastV6Timestamp == 1<<60-1 {
			return NilUUID, ErrCounterExhausted
		}
		g.lastV6Timestamp++
	} else {
		g.lastV6Timestamp = timestamp
	}
	putUUIDv6Timestamp(&g.lastV6, g.lastV6Timestamp)
	return g.lastV6, nil
}

// NewUUIDv7 returns a strictly increasing process-local UUIDv7. On clock
// rollback, it holds the previous millisecond and increments the 74-bit random
// field. This mirrors the monotonic behavior expected for database locality.
func (g *Generator) NewUUIDv7() (UUID, error) {
	if g == nil {
		return NilUUID, fmt.Errorf("quuid: nil Generator")
	}
	g.mu.Lock()
	defer g.mu.Unlock()

	nowMS := g.clock().UnixMilli()
	if nowMS < 0 || nowMS > 1<<48-1 {
		return NilUUID, ErrTimeOutOfRange
	}
	ms := uint64(nowMS)

	if !g.v7Initialized || ms > g.lastV7MS {
		id, err := NewUUIDv7At(time.UnixMilli(nowMS), g.reader)
		if err != nil {
			return NilUUID, err
		}
		g.lastV7MS = ms
		g.lastV7 = id
		g.v7Initialized = true
		return id, nil
	}

	if !incrementUUIDv7Random(&g.lastV7) {
		return NilUUID, ErrCounterExhausted
	}
	putUUIDv7Milliseconds(&g.lastV7, g.lastV7MS)
	g.lastV7[6] = (g.lastV7[6] & 0x0f) | 0x70
	g.lastV7[8] = (g.lastV7[8] & 0x3f) | 0x80
	return g.lastV7, nil
}

// NewUUIDv8 returns a random UUIDv8 using the generator's entropy source.
func (g *Generator) NewUUIDv8() (UUID, error) {
	if g == nil {
		return NilUUID, fmt.Errorf("quuid: nil Generator")
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	return NewUUIDv8FromReader(g.reader)
}

func incrementBigEndian(value []byte) bool {
	for i := len(value) - 1; i >= 0; i-- {
		value[i]++
		if value[i] != 0 {
			return true
		}
	}
	return false
}

// incrementUUIDv7Random increments UUIDv7's 74 random bits while preserving
// its version and variant fields.
func incrementUUIDv7Random(id *UUID) bool {
	for i := 15; i >= 9; i-- {
		id[i]++
		if id[i] != 0 {
			return true
		}
	}

	lowSix := (id[8] & 0x3f) + 1
	if lowSix <= 0x3f {
		id[8] = (id[8] & 0xc0) | lowSix
		return true
	}
	id[8] &= 0xc0

	id[7]++
	if id[7] != 0 {
		return true
	}

	lowFour := (id[6] & 0x0f) + 1
	if lowFour <= 0x0f {
		id[6] = (id[6] & 0xf0) | lowFour
		return true
	}
	id[6] &= 0xf0
	return false
}
