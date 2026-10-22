package trogonobjectid_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/TrogonStack/trogonobjectid"
)

type paymentMethodType struct{}

func (paymentMethodType) Definition() trogonobjectid.Definition {
	return trogonobjectid.Definition{ObjectType: "pm", Aliases: []trogonobjectid.ObjectType{"card", "src"}}
}

type selfAliasType struct{}

func (selfAliasType) Definition() trogonobjectid.Definition {
	return trogonobjectid.Definition{ObjectType: "pm", Aliases: []trogonobjectid.ObjectType{"pm"}}
}

type badAliasType struct{}

func (badAliasType) Definition() trogonobjectid.Definition {
	return trogonobjectid.Definition{ObjectType: "pm", Aliases: []trogonobjectid.ObjectType{"Card"}}
}

type paymentUnion struct{}

func (paymentUnion) Types() []trogonobjectid.MemberType {
	return []trogonobjectid.MemberType{
		trogonobjectid.TypeOf[paymentMethodType](),
		trogonobjectid.TypeOf[userType](),
	}
}

type clashUnion struct{}

func (clashUnion) Types() []trogonobjectid.MemberType {
	return []trogonobjectid.MemberType{
		trogonobjectid.TypeOf[paymentMethodType](),
		trogonobjectid.TypeOf[cardType](),
	}
}

type cardType struct{}

func (cardType) Definition() trogonobjectid.Definition {
	return trogonobjectid.Definition{ObjectType: "card"}
}

func TestAliases(t *testing.T) {
	want := trogonobjectid.MustNew[paymentMethodType]("abc")

	for _, input := range []string{"pm_abc", "card_abc", "src_abc"} {
		t.Run("Parse accepts "+input, func(t *testing.T) {
			got, err := trogonobjectid.Parse[paymentMethodType](input)
			require.NoError(t, err)
			assert.Equal(t, want, got)
			assert.Equal(t, "pm_abc", got.String())
		})
	}

	t.Run("full storage reads an aliased value", func(t *testing.T) {
		var got trogonobjectid.ID[paymentMethodType]
		require.NoError(t, got.Scan("card_abc"))
		assert.Equal(t, want, got)
	})

	t.Run("an unknown prefix still fails", func(t *testing.T) {
		_, err := trogonobjectid.Parse[paymentMethodType]("user_abc")
		require.ErrorIs(t, err, trogonobjectid.ErrInvalidFormat)
	})

	t.Run("a union matches an alias", func(t *testing.T) {
		u, err := trogonobjectid.ParseUnion[paymentUnion]("card_abc")
		require.NoError(t, err)
		got, ok := trogonobjectid.UnionAs[paymentMethodType](u)
		require.True(t, ok)
		assert.Equal(t, want, got)
		assert.Equal(t, "pm_abc", u.String())
	})

	t.Run("Check rejects an alias equal to the object type", func(t *testing.T) {
		require.ErrorIs(t, trogonobjectid.Check[selfAliasType](), trogonobjectid.ErrInvalidDefinition)
	})

	t.Run("Check rejects an invalid alias", func(t *testing.T) {
		require.ErrorIs(t, trogonobjectid.Check[badAliasType](), trogonobjectid.ErrInvalidDefinition)
	})

	t.Run("CheckUnion rejects an alias that is another member's prefix", func(t *testing.T) {
		require.ErrorIs(t, trogonobjectid.CheckUnion[clashUnion](), trogonobjectid.ErrInvalidDefinition)
	})
}
