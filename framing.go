package quuid

import (
	"encoding/binary"
	"hash"
)

func writeFrame(h hash.Hash, parts ...[]byte) {
	var length [8]byte
	for _, part := range parts {
		binary.BigEndian.PutUint64(length[:], uint64(len(part)))
		_, _ = h.Write(length[:])
		_, _ = h.Write(part)
	}
}
