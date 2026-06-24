package resource

import (
	"encoding/binary"
	"hash/fnv"
	"strconv"
	"sync/atomic"
)

// IDGenerator allocates monotonically increasing opaque resource IDs.
type IDGenerator struct {
	next atomic.Uint64
}

// Next returns the next resource ID.
func (g *IDGenerator) Next() uint64 {
	return g.next.Add(1)
}

// Hash returns the stable opaque-resource hash for a type name and ID.
func Hash(typeName string, id uint64) uint64 {
	h := fnv.New64a()
	h.Write([]byte(typeName + ":"))

	bytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(bytes, id)
	h.Write(bytes)

	return h.Sum64()
}

// Display returns the standard display form for an opaque resource type.
func Display(typeName string) string {
	return "<" + typeName + ">"
}

// MarshalDisplayJSON returns the JSON string representation of Display(typeName).
func MarshalDisplayJSON(typeName string) ([]byte, error) {
	return MarshalStringJSON(Display(typeName))
}

// MarshalStringJSON returns the JSON string representation of value.
func MarshalStringJSON(value string) ([]byte, error) {
	return []byte(strconv.Quote(value)), nil
}
