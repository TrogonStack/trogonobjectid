package trogonobjectid_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/TrogonStack/trogonobjectid"
)

type namedType[N interface {
	name() trogonobjectid.ObjectType
}] struct{}

func (namedType[N]) Definition() trogonobjectid.Definition {
	var n N
	return trogonobjectid.Definition{ObjectType: n.name()}
}

type (
	upperName    struct{}
	digitName    struct{}
	dashName     struct{}
	leadingName  struct{}
	trailingName struct{}
	longName     struct{}
	maxName      struct{}
	snakeName    struct{}
)

func (upperName) name() trogonobjectid.ObjectType    { return "User" }
func (digitName) name() trogonobjectid.ObjectType    { return "user2" }
func (dashName) name() trogonobjectid.ObjectType     { return "credit-note" }
func (leadingName) name() trogonobjectid.ObjectType  { return "_user" }
func (trailingName) name() trogonobjectid.ObjectType { return "user_" }
func (longName) name() trogonobjectid.ObjectType {
	return trogonobjectid.ObjectType(strings.Repeat("a", 64))
}
func (maxName) name() trogonobjectid.ObjectType {
	return trogonobjectid.ObjectType(strings.Repeat("a", 63))
}
func (snakeName) name() trogonobjectid.ObjectType { return "credit_note" }

func TestCheck(t *testing.T) {
	t.Run("accepts valid definitions", func(t *testing.T) {
		require.NoError(t, trogonobjectid.Check[invoiceType]())
		require.NoError(t, trogonobjectid.Check[namedType[snakeName]]())
		require.NoError(t, trogonobjectid.Check[namedType[maxName]]())
	})

	invalid := map[string]error{
		"empty object type":   trogonobjectid.Check[badType](),
		"upper case":          trogonobjectid.Check[namedType[upperName]](),
		"digit":               trogonobjectid.Check[namedType[digitName]](),
		"dash":                trogonobjectid.Check[namedType[dashName]](),
		"leading underscore":  trogonobjectid.Check[namedType[leadingName]](),
		"trailing underscore": trogonobjectid.Check[namedType[trailingName]](),
		"64 characters":       trogonobjectid.Check[namedType[longName]](),
	}
	for name, err := range invalid {
		t.Run("rejects "+name, func(t *testing.T) {
			require.ErrorIs(t, err, trogonobjectid.ErrInvalidDefinition)
		})
	}

	t.Run("rejects a generator that its own validation rejects", func(t *testing.T) {
		err := trogonobjectid.Check[seqType]()
		require.ErrorIs(t, err, trogonobjectid.ErrInvalidDefinition)
		require.ErrorIs(t, err, trogonobjectid.ErrInvalidInteger)
	})

	t.Run("an invalid object type fails every constructor", func(t *testing.T) {
		_, err := trogonobjectid.New[namedType[upperName]]("a")
		require.ErrorIs(t, err, trogonobjectid.ErrInvalidDefinition)
	})
}

func TestCheckUnion(t *testing.T) {
	require.NoError(t, trogonobjectid.CheckUnion[principalUnion]())
	require.ErrorIs(t, trogonobjectid.CheckUnion[duplicateUnion](), trogonobjectid.ErrInvalidDefinition)
	require.ErrorIs(t, trogonobjectid.CheckUnion[emptyUnion](), trogonobjectid.ErrInvalidDefinition)
}

func TestWrongPrefix(t *testing.T) {
	_, err := trogonobjectid.Parse[invoiceType]("order_" + invoiceUUID)

	var wrong *trogonobjectid.WrongPrefixError
	require.True(t, errors.As(err, &wrong))
	assert.Equal(t, "order_", wrong.Got)
	assert.Equal(t, "inv_", wrong.Want)
	require.ErrorIs(t, err, trogonobjectid.ErrInvalidFormat)

	t.Run("no separator is a plain format error", func(t *testing.T) {
		_, err := trogonobjectid.Parse[invoiceType](invoiceUUID)
		require.ErrorIs(t, err, trogonobjectid.ErrInvalidFormat)
		assert.False(t, errors.As(err, &wrong))
	})
}

func TestNamedTypes(t *testing.T) {
	assert.Equal(t, trogonobjectid.ObjectType("order"), trogonobjectid.ObjectTypeOf[orderType]())
	assert.Equal(t, trogonobjectid.Separator("_"), trogonobjectid.DefaultSeparator)
}
