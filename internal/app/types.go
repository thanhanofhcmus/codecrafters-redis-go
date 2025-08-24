package app

import (
	"time"

	"github.com/codecrafters-io/redis-starter-go/pkg/ulid"
)

type ValueType int

const (
	ValueTypeSimple ValueType = iota
	ValueTypeList
)

type Value struct {
	Key       string
	ValueType ValueType
	String    string
	List      []string
}

type BLPOPConsumer struct {
	id  ulid.ID
	key string
	ch  chan struct{}
}

type App struct {
	// TODO: make this thread safe

	dict   map[string]Value
	expiry map[string]time.Time

	keySpaceConsumer map[ulid.ID]chan Value
	// TODO: add uild
	blpopConsumers map[string][]BLPOPConsumer

	idGenerator *ulid.Generator
}

func NewApp() *App {
	return &App{
		dict:   map[string]Value{},
		expiry: map[string]time.Time{},

		keySpaceConsumer: map[ulid.ID]chan Value{},
		blpopConsumers:   map[string][]BLPOPConsumer{},

		idGenerator: ulid.NewGenerator(),
	}
}
