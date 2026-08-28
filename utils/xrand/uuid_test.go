package xrand

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUUIDv7(t *testing.T) {
	id1 := UUIDv7()
	id2 := UUIDv7()

	assert.Len(t, id1, 36)
	assert.Len(t, id2, 36)
	assert.NotEqual(t, id1, id2)

	simple := UUIDv7Simple()
	assert.Len(t, simple, 32)
	assert.NotContains(t, simple, "-")
}
