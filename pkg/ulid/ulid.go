package ulid

import (
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"io"
	"sync"
	"time"
)

// Generator creates ULID-like, lexicographically sortable identifiers, uses Crockford's Base32 (no padding)
// Format:
// - 48-bit big-endian Unix time in milliseconds
// - 80 bits of entropy (monotonic within the same millisecond)

type ID string

type Generator struct {
	mutex           sync.Mutex
	enc             *base32.Encoding
	lastMilliSecond uint64
	lastEntrophy    [10]byte
}

func NewGenerator() *Generator {
	// remove I, L, O, U
	const crockford = "0123456789ABCDEFGHJKLMNPQRSTVWXYZ"
	enc := base32.NewEncoding(crockford).WithPadding(base32.NoPadding)
	return &Generator{
		enc: enc,
	}
}

func (g *Generator) New() (ID, error) {
	ms := uint64(time.Now().UnixMilli())

	g.mutex.Lock()
	defer g.mutex.Unlock()

	var entrophy [10]byte
	// collision with the last time we generate id
	if ms == g.lastMilliSecond {
		entrophy = g.lastEntrophy
		if increment(&entrophy) {
			return "", fmt.Errorf("generator id counter overflow")
		}
	} else {
		if _, err := io.ReadFull(rand.Reader, entrophy[:]); err != nil {
			return "", err
		}
		g.lastMilliSecond = ms
	}
	g.lastEntrophy = entrophy

	var b [16]byte
	b[0] = byte(ms >> 40)
	b[1] = byte(ms >> 32)
	b[2] = byte(ms >> 24)
	b[3] = byte(ms >> 16)
	b[4] = byte(ms >> 8)
	b[5] = byte(ms >> 0)
	copy(b[6:], entrophy[:])

	return ID(g.enc.EncodeToString(b[:])), nil
}

func (g *Generator) MustNew() ID {
	id, err := g.New()
	if err != nil {
		panic(err)
	}
	return id
}

func increment(x *[10]byte) (overflow bool) {
	for i := len(x) - 1; i >= 0; i-- {
		x[i]++
		if x[i] != 0 {
			return false
		}
	}
	return true
}
