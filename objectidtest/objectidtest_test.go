package objectidtest_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/TrogonStack/trogonobjectid"
	"github.com/TrogonStack/trogonobjectid/objectidtest"
)

type fanType struct{}

func (fanType) Definition() trogonobjectid.Definition {
	return trogonobjectid.Definition{ObjectType: "fan", Validate: trogonobjectid.ValidateUUID}
}

type shopType struct{}

func (shopType) Definition() trogonobjectid.Definition {
	return trogonobjectid.Definition{ObjectType: "shop", Validate: trogonobjectid.ValidateUUID}
}

type seqType struct{}

func (seqType) Definition() trogonobjectid.Definition {
	return trogonobjectid.Definition{ObjectType: "seq", Validate: trogonobjectid.ValidateInteger}
}

func TestFromSeed(t *testing.T) {
	t.Run("the same seed gives the same ID", func(t *testing.T) {
		assert.Equal(t, objectidtest.FromSeed[fanType]("buyer"), objectidtest.FromSeed[fanType]("buyer"))
	})

	t.Run("different seeds give different IDs", func(t *testing.T) {
		assert.NotEqual(t, objectidtest.FromSeed[fanType]("buyer"), objectidtest.FromSeed[fanType]("seller"))
	})

	t.Run("the bare value is a UUIDv5 in Namespace", func(t *testing.T) {
		got := objectidtest.FromSeed[fanType]("buyer")
		assert.Equal(t, uuid.NewSHA1(objectidtest.Namespace, []byte("buyer")).String(), got.Bare())
		u, err := got.UUID()
		require.NoError(t, err)
		assert.Equal(t, uuid.Version(5), u.Version())
	})

	t.Run("types share bare values for the same seed", func(t *testing.T) {
		assert.Equal(t, objectidtest.FromSeed[fanType]("x").Bare(), objectidtest.FromSeed[shopType]("x").Bare())
	})

	t.Run("panics for a type that rejects UUIDs", func(t *testing.T) {
		assert.Panics(t, func() { objectidtest.FromSeed[seqType]("x") })
	})
}

func TestFromSeedIn(t *testing.T) {
	namespace := uuid.MustParse("75211e1f-6d69-5d80-b91f-9e9c5c8aa697")
	got := objectidtest.FromSeedIn[fanType](namespace, "buyer")
	assert.Equal(t, uuid.NewSHA1(namespace, []byte("buyer")).String(), got.Bare())
}
