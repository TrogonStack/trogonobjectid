package trogonobjectid_test

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/TrogonStack/trogonobjectid"
)

type systemType struct{}

func (systemType) Definition() trogonobjectid.Definition {
	return trogonobjectid.Definition{ObjectType: "system"}
}

type SystemID = trogonobjectid.ID[systemType]

type principalUnion struct{}

func (principalUnion) Types() []trogonobjectid.MemberType {
	return []trogonobjectid.MemberType{
		trogonobjectid.MemberOf[userType](),
		trogonobjectid.MemberOf[systemType](),
		trogonobjectid.MemberOf[invoiceType](),
	}
}

type PrincipalID = trogonobjectid.Union[principalUnion]

type acmeType struct{}

func (acmeType) Definition() trogonobjectid.Definition {
	return trogonobjectid.Definition{ObjectType: "acme"}
}

type acmeAdminType struct{}

func (acmeAdminType) Definition() trogonobjectid.Definition {
	return trogonobjectid.Definition{ObjectType: "acme_admin"}
}

type overlapUnion struct{}

func (overlapUnion) Types() []trogonobjectid.MemberType {
	return []trogonobjectid.MemberType{
		trogonobjectid.MemberOf[acmeType](),
		trogonobjectid.MemberOf[acmeAdminType](),
	}
}

type emptyUnion struct{}

func (emptyUnion) Types() []trogonobjectid.MemberType { return nil }

type duplicateUnion struct{}

func (duplicateUnion) Types() []trogonobjectid.MemberType {
	return []trogonobjectid.MemberType{
		trogonobjectid.MemberOf[userType](),
		trogonobjectid.MemberOf[userType](),
	}
}

type badMemberUnion struct{}

func (badMemberUnion) Types() []trogonobjectid.MemberType {
	return []trogonobjectid.MemberType{trogonobjectid.MemberOf[badType]()}
}

func TestNewUnion(t *testing.T) {
	t.Run("wraps a member", func(t *testing.T) {
		user := trogonobjectid.MustNew[userType]("abc")
		p, err := trogonobjectid.NewUnion[principalUnion](user)
		require.NoError(t, err)
		assert.Equal(t, user, p.Member())
		assert.Equal(t, "user_abc", p.String())
		assert.False(t, p.IsZero())
	})

	tests := []struct {
		name    string
		member  trogonobjectid.Member
		wantErr error
	}{
		{"non-member type", trogonobjectid.MustNew[orderType]("1"), trogonobjectid.ErrNotMember},
		{"zero member", trogonobjectid.ID[userType]{}, trogonobjectid.ErrEmptyValue},
		{"nil member", nil, trogonobjectid.ErrEmptyValue},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := trogonobjectid.NewUnion[principalUnion](tt.member)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}

	t.Run("MustNewUnion panics on error", func(t *testing.T) {
		assert.Panics(t, func() { trogonobjectid.MustNewUnion[principalUnion](trogonobjectid.MustNew[orderType]("1")) })
	})
}

func TestUnionDefinition(t *testing.T) {
	tests := []struct {
		name  string
		parse func() error
	}{
		{"no members", func() error { _, err := trogonobjectid.ParseUnion[emptyUnion]("user_abc"); return err }},
		{"duplicate prefixes", func() error { _, err := trogonobjectid.ParseUnion[duplicateUnion]("user_abc"); return err }},
		{"invalid member", func() error { _, err := trogonobjectid.ParseUnion[badMemberUnion]("x_abc"); return err }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.ErrorIs(t, tt.parse(), trogonobjectid.ErrInvalidDefinition)
		})
	}
}

func TestParseUnion(t *testing.T) {
	t.Run("picks the member by prefix", func(t *testing.T) {
		p, err := trogonobjectid.ParseUnion[principalUnion]("system_cron")
		require.NoError(t, err)

		switch m := p.Member().(type) {
		case SystemID:
			assert.Equal(t, "cron", m.Bare())
		default:
			t.Fatalf("got member %T, want SystemID", m)
		}
	})

	t.Run("longest prefix wins", func(t *testing.T) {
		p, err := trogonobjectid.ParseUnion[overlapUnion]("acme_admin_xyz")
		require.NoError(t, err)
		assert.Equal(t, trogonobjectid.MustNew[acmeAdminType]("xyz"), p.Member())

		p, err = trogonobjectid.ParseUnion[overlapUnion]("acme_xyz")
		require.NoError(t, err)
		assert.Equal(t, trogonobjectid.MustNew[acmeType]("xyz"), p.Member())
	})

	t.Run("validates with the member's rules", func(t *testing.T) {
		_, err := trogonobjectid.ParseUnion[principalUnion]("inv_nope")
		require.ErrorIs(t, err, trogonobjectid.ErrInvalidUUID)
	})

	for _, input := range []string{"", "order_1", "abc", "user_"} {
		t.Run(fmt.Sprintf("rejects %q", input), func(t *testing.T) {
			_, err := trogonobjectid.ParseUnion[principalUnion](input)
			require.ErrorIs(t, err, trogonobjectid.ErrInvalidFormat)
		})
	}

	t.Run("MustParseUnion panics on error", func(t *testing.T) {
		assert.Panics(t, func() { trogonobjectid.MustParseUnion[principalUnion]("abc") })
	})
}

type event struct {
	Actor   PrincipalID `json:"actor"`
	Missing PrincipalID `json:"missing,omitzero"`
}

func TestUnionEncodings(t *testing.T) {
	invoice := trogonobjectid.MustNew[invoiceType](invoiceUUID)
	p := trogonobjectid.MustNewUnion[principalUnion](invoice)
	full := "inv_" + invoiceUUID

	t.Run("JSON uses the full form even when the member drops the prefix", func(t *testing.T) {
		data, err := json.Marshal(event{Actor: p})
		require.NoError(t, err)
		assert.JSONEq(t, `{"actor":"`+full+`"}`, string(data))

		var got event
		require.NoError(t, json.Unmarshal(data, &got))
		assert.Equal(t, p, got.Actor)
		assert.True(t, got.Missing.IsZero())
	})

	t.Run("JSON null is the zero union", func(t *testing.T) {
		data, err := json.Marshal(PrincipalID{})
		require.NoError(t, err)
		assert.Equal(t, "null", string(data))

		got := p
		require.NoError(t, json.Unmarshal([]byte("null"), &got))
		assert.True(t, got.IsZero())
	})

	t.Run("SQL uses the full form even when the member drops the prefix", func(t *testing.T) {
		v, err := p.Value()
		require.NoError(t, err)
		assert.Equal(t, driver.Value(full), v)

		var got PrincipalID
		require.NoError(t, got.Scan(full))
		assert.Equal(t, p, got)
		require.NoError(t, got.Scan([]byte(full)))
		assert.Equal(t, p, got)
	})

	t.Run("SQL NULL is the zero union", func(t *testing.T) {
		v, err := PrincipalID{}.Value()
		require.NoError(t, err)
		assert.Nil(t, v)

		got := p
		require.NoError(t, got.Scan(nil))
		assert.True(t, got.IsZero())
		got = p
		require.NoError(t, got.Scan(""))
		assert.True(t, got.IsZero())
	})

	t.Run("Scan rejects other types", func(t *testing.T) {
		var got PrincipalID
		require.ErrorIs(t, got.Scan(42), trogonobjectid.ErrInvalidFormat)
	})

	t.Run("text uses the full form", func(t *testing.T) {
		text, err := p.MarshalText()
		require.NoError(t, err)
		assert.Equal(t, full, string(text))

		var got PrincipalID
		require.NoError(t, got.UnmarshalText(text))
		assert.Equal(t, p, got)
	})

	t.Run("slog uses the full form", func(t *testing.T) {
		var buf bytes.Buffer
		slog.New(slog.NewJSONHandler(&buf, nil)).Info("acted", "actor", p)

		var entry map[string]any
		require.NoError(t, json.Unmarshal(buf.Bytes(), &entry))
		assert.Equal(t, full, entry["actor"])
	})
}

func TestUnionZero(t *testing.T) {
	var p PrincipalID
	assert.True(t, p.IsZero())
	assert.Nil(t, p.Member())
	assert.Empty(t, p.String())
}

func TestUnionAs(t *testing.T) {
	user := trogonobjectid.MustNew[userType]("abc-123")
	u := trogonobjectid.MustNewUnion[principalUnion](user)

	got, ok := trogonobjectid.UnionAs[userType](u)
	require.True(t, ok)
	assert.Equal(t, user, got)

	_, ok = trogonobjectid.UnionAs[systemType](u)
	assert.False(t, ok)

	_, ok = trogonobjectid.UnionAs[userType](trogonobjectid.Union[principalUnion]{})
	assert.False(t, ok)
}

func TestUnionSpanner(t *testing.T) {
	want := trogonobjectid.MustNewUnion[principalUnion](trogonobjectid.MustNew[userType]("abc-123"))

	encoded, err := want.EncodeSpanner()
	require.NoError(t, err)
	assert.Equal(t, "user_abc-123", encoded)

	var got PrincipalID
	require.NoError(t, got.DecodeSpanner(encoded))
	assert.Equal(t, want, got)

	zero, err := PrincipalID{}.EncodeSpanner()
	require.NoError(t, err)
	assert.Equal(t, (*string)(nil), zero)
}
