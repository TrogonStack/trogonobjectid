package trogonobjectid_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/TrogonStack/trogonobjectid"
)

type upperCaseType struct{}

func (upperCaseType) Definition() trogonobjectid.Definition {
	return trogonobjectid.Definition{ObjectType: "User"}
}

func TestCheck(t *testing.T) {
	require.NoError(t, trogonobjectid.Check[invoiceType]())
	require.ErrorIs(t, trogonobjectid.Check[badType](), trogonobjectid.ErrInvalidDefinition)
	require.ErrorIs(t, trogonobjectid.Check[upperCaseType](), trogonobjectid.ErrInvalidDefinition)

	t.Run("rejects a generator that its own validation rejects", func(t *testing.T) {
		err := trogonobjectid.Check[seqType]()
		require.ErrorIs(t, err, trogonobjectid.ErrInvalidDefinition)
		require.ErrorIs(t, err, trogonobjectid.ErrInvalidInteger)
	})

	t.Run("an invalid object type fails every constructor", func(t *testing.T) {
		_, err := trogonobjectid.New[upperCaseType]("a")
		require.ErrorIs(t, err, trogonobjectid.ErrInvalidDefinition)
	})
}

func TestCheckUnion(t *testing.T) {
	require.NoError(t, trogonobjectid.CheckUnion[principalUnion]())
	require.ErrorIs(t, trogonobjectid.CheckUnion[duplicateUnion](), trogonobjectid.ErrInvalidDefinition)
	require.ErrorIs(t, trogonobjectid.CheckUnion[emptyUnion](), trogonobjectid.ErrInvalidDefinition)
}

func TestWrongPrefix(t *testing.T) {
	for name, input := range map[string]string{"another type's prefix": "order_" + invoiceUUID, "no prefix": invoiceUUID} {
		t.Run(name, func(t *testing.T) {
			_, err := trogonobjectid.Parse[invoiceType](input)
			var wrong *trogonobjectid.WrongPrefixError
			require.True(t, errors.As(err, &wrong))
			assert.Equal(t, []trogonobjectid.Prefix{"inv_"}, wrong.Want)
			require.ErrorIs(t, err, trogonobjectid.ErrInvalidFormat)
		})
	}

	t.Run("a union lists every member prefix", func(t *testing.T) {
		_, err := trogonobjectid.ParseUnion[principalUnion]("order_abc")
		var wrong *trogonobjectid.WrongPrefixError
		require.True(t, errors.As(err, &wrong))
		assert.ElementsMatch(t, []trogonobjectid.Prefix{"user_", "system_", "inv_"}, wrong.Want)
	})
}
