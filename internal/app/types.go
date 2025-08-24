package app

import (
	"fmt"
	"time"

	"github.com/codecrafters-io/redis-starter-go/pkg/ulid"
)

type ValueType int

const (
	ValueTypeString ValueType = iota
	ValueTypeList
	ValueTypeSet
	ValueTypeZSet
	ValueTypeHash
	ValueTypeStream
	ValueTypeVectorSet
)

func ValueTypeToName(valueType ValueType) string {
	switch valueType {
	case ValueTypeString:
		return "string"
	case ValueTypeList:
		return "list"
	case ValueTypeSet:
		return "set"
	case ValueTypeZSet:
		return "zset"
	case ValueTypeHash:
		return "hash"
	case ValueTypeStream:
		return "stream"
	case ValueTypeVectorSet:
		return "vector_set"
	default:
		return fmt.Sprintf("unknown-%d", int(valueType))
	}
}

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
	blpopConsumers   map[string][]BLPOPConsumer

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
