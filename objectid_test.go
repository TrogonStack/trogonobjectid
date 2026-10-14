package trogonobjectid_test

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/TrogonStack/trogonobjectid"
)

type userType struct{}

func (userType) Definition() trogonobjectid.Definition {
	return trogonobjectid.Definition{ObjectType: "user"}
}

type UserID = trogonobjectid.ID[userType]

type orderType struct{}

func (orderType) Definition() trogonobjectid.Definition {
	return trogonobjectid.Definition{ObjectType: "order", Separator: "#"}
}

type invoiceType struct{}

func (invoiceType) Definition() trogonobjectid.Definition {
	return trogonobjectid.Definition{
		ObjectType:    "inv",
		StorageFormat: trogonobjectid.FormatDropPrefix,
		JSONFormat:    trogonobjectid.FormatDropPrefix,
		Validate:      trogonobjectid.ValidateUUID,
		Generate:      trogonobjectid.GenerateUUIDv7,
	}
}

type InvoiceID = trogonobjectid.ID[invoiceType]

type seqType struct{}

func (seqType) Definition() trogonobjectid.Definition {
	return trogonobjectid.Definition{
		ObjectType: "seq",
		Validate:   trogonobjectid.ValidateInteger,
		Generate:   trogonobjectid.GenerateUUIDv4,
	}
}

type badType struct{}

func (badType) Definition() trogonobjectid.Definition {
	return trogonobjectid.Definition{}
}

const invoiceUUID = "0199e3a1-7c2b-7d4e-9f10-2a3b4c5d6e7f"

func TestNew(t *testing.T) {
	t.Run("keeps the bare value and adds the prefix in String", func(t *testing.T) {
		id, err := trogonobjectid.New[userType]("abc-123")
		require.NoError(t, err)
		assert.Equal(t, "abc-123", id.Bare())
		assert.Equal(t, "user_abc-123", id.String())
		assert.Equal(t, "user", id.ObjectType())
		assert.Equal(t, "user_", id.Prefix())
		assert.False(t, id.IsZero())
	})

	t.Run("uses the configured separator", func(t *testing.T) {
		id := trogonobjectid.MustNew[orderType]("42")
		assert.Equal(t, "order#42", id.String())
	})

	tests := []struct {
		name    string
		create  func() error
		wantErr error
	}{
		{"empty value", func() error { _, err := trogonobjectid.New[userType](""); return err }, trogonobjectid.ErrEmptyValue},
		{"invalid uuid", func() error { _, err := trogonobjectid.New[invoiceType]("nope"); return err }, trogonobjectid.ErrInvalidUUID},
		{"invalid integer", func() error { _, err := trogonobjectid.New[seqType]("12a"); return err }, trogonobjectid.ErrInvalidInteger},
		{"empty object type", func() error { _, err := trogonobjectid.New[badType]("abc"); return err }, trogonobjectid.ErrInvalidDefinition},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.create()
			require.ErrorIs(t, err, tt.wantErr)
		})
	}

	t.Run("reports the object type and value", func(t *testing.T) {
		_, err := trogonobjectid.New[invoiceType]("nope")
		var verr *trogonobjectid.ValidationError
		require.ErrorAs(t, err, &verr)
		assert.Equal(t, "inv", verr.ObjectType)
		assert.Equal(t, "nope", verr.Value)
		assert.EqualError(t, err, `trogonobjectid: invalid inv value "nope": invalid uuid`)
	})

	t.Run("wraps custom validator errors", func(t *testing.T) {
		_, err := trogonobjectid.New[customType]("a")
		require.ErrorIs(t, err, errTooShort)
	})

	t.Run("MustNew panics on invalid input", func(t *testing.T) {
		assert.Panics(t, func() { trogonobjectid.MustNew[userType]("") })
	})
}

var errTooShort = errors.New("too short")

type customType struct{}

func (customType) Definition() trogonobjectid.Definition {
	return trogonobjectid.Definition{
		ObjectType: "custom",
		Validate: func(id string) error {
			if len(id) < 3 {
				return errTooShort
			}
			return nil
		},
	}
}

func TestValidateInteger(t *testing.T) {
	for _, v := range []string{"0", "42", "-7", "+7", "123456789012345678901234567890"} {
		assert.NoError(t, trogonobjectid.ValidateInteger(v), v)
	}
	for _, v := range []string{"", "-", "+", "1.5", "12a", " 1", "0x10"} {
		assert.ErrorIs(t, trogonobjectid.ValidateInteger(v), trogonobjectid.ErrInvalidInteger, v)
	}
}

func TestParse(t *testing.T) {
	t.Run("round trips String", func(t *testing.T) {
		want := trogonobjectid.MustNew[invoiceType](invoiceUUID)
		got, err := trogonobjectid.Parse[invoiceType](want.String())
		require.NoError(t, err)
		assert.Equal(t, want, got)
	})

	tests := []struct {
		name    string
		input   string
		wantErr error
	}{
		{"no prefix", "abc-123", trogonobjectid.ErrInvalidFormat},
		{"wrong prefix", "order_abc-123", trogonobjectid.ErrInvalidFormat},
		{"prefix only", "user_", trogonobjectid.ErrInvalidFormat},
		{"empty", "", trogonobjectid.ErrInvalidFormat},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := trogonobjectid.Parse[userType](tt.input)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}

	t.Run("validates the bare value", func(t *testing.T) {
		_, err := trogonobjectid.Parse[invoiceType]("inv_nope")
		require.ErrorIs(t, err, trogonobjectid.ErrInvalidUUID)
	})

	t.Run("MustParse panics on invalid input", func(t *testing.T) {
		assert.Panics(t, func() { trogonobjectid.MustParse[userType]("abc") })
	})
}

func TestGenerate(t *testing.T) {
	t.Run("uses the configured generator", func(t *testing.T) {
		id, err := trogonobjectid.Generate[invoiceType]()
		require.NoError(t, err)
		u, err := uuid.Parse(id.Bare())
		require.NoError(t, err)
		assert.Equal(t, uuid.Version(7), u.Version())
	})

	t.Run("fails without a generator", func(t *testing.T) {
		_, err := trogonobjectid.Generate[userType]()
		require.ErrorIs(t, err, trogonobjectid.ErrNoGenerator)
	})

	t.Run("validates the generated value", func(t *testing.T) {
		_, err := trogonobjectid.Generate[seqType]()
		require.ErrorIs(t, err, trogonobjectid.ErrInvalidInteger)
	})
}

func TestStorage(t *testing.T) {
	t.Run("full format keeps the prefix", func(t *testing.T) {
		id := trogonobjectid.MustNew[userType]("abc-123")
		assert.Equal(t, "user_abc-123", id.Storage())
		got, err := trogonobjectid.FromStorage[userType]("user_abc-123")
		require.NoError(t, err)
		assert.Equal(t, id, got)
	})

	t.Run("drop prefix format stores the bare value", func(t *testing.T) {
		id := trogonobjectid.MustNew[invoiceType](invoiceUUID)
		assert.Equal(t, invoiceUUID, id.Storage())
		got, err := trogonobjectid.FromStorage[invoiceType](invoiceUUID)
		require.NoError(t, err)
		assert.Equal(t, id, got)
	})

	t.Run("drop prefix format rejects a prefixed value", func(t *testing.T) {
		_, err := trogonobjectid.FromStorage[invoiceType]("inv_" + invoiceUUID)
		require.ErrorIs(t, err, trogonobjectid.ErrInvalidUUID)
	})
}

func TestSQL(t *testing.T) {
	t.Run("Value emits the storage form", func(t *testing.T) {
		v, err := trogonobjectid.MustNew[invoiceType](invoiceUUID).Value()
		require.NoError(t, err)
		assert.Equal(t, driver.Value(invoiceUUID), v)

		v, err = trogonobjectid.MustNew[userType]("abc").Value()
		require.NoError(t, err)
		assert.Equal(t, driver.Value("user_abc"), v)
	})

	t.Run("Value emits NULL for the zero ID", func(t *testing.T) {
		v, err := UserID{}.Value()
		require.NoError(t, err)
		assert.Nil(t, v)
	})

	scans := []struct {
		name string
		src  any
		want InvoiceID
	}{
		{"string", invoiceUUID, trogonobjectid.MustNew[invoiceType](invoiceUUID)},
		{"bytes", []byte(invoiceUUID), trogonobjectid.MustNew[invoiceType](invoiceUUID)},
		{"nil", nil, InvoiceID{}},
		{"empty string", "", InvoiceID{}},
	}
	for _, tt := range scans {
		t.Run("Scan "+tt.name, func(t *testing.T) {
			got := trogonobjectid.MustNew[invoiceType](invoiceUUID)
			require.NoError(t, got.Scan(tt.src))
			assert.Equal(t, tt.want, got)
		})
	}

	t.Run("Scan rejects other types", func(t *testing.T) {
		var got InvoiceID
		require.ErrorIs(t, got.Scan(42), trogonobjectid.ErrInvalidFormat)
	})

	t.Run("Scan rejects invalid values", func(t *testing.T) {
		var got InvoiceID
		require.ErrorIs(t, got.Scan("nope"), trogonobjectid.ErrInvalidUUID)
	})
}

type payload struct {
	User    UserID    `json:"user"`
	Invoice InvoiceID `json:"invoice"`
	Missing UserID    `json:"missing,omitzero"`
}

func TestJSON(t *testing.T) {
	in := payload{
		User:    trogonobjectid.MustNew[userType]("abc"),
		Invoice: trogonobjectid.MustNew[invoiceType](invoiceUUID),
	}

	data, err := json.Marshal(in)
	require.NoError(t, err)
	assert.JSONEq(t, `{"user":"user_abc","invoice":"`+invoiceUUID+`"}`, string(data))

	var out payload
	require.NoError(t, json.Unmarshal(data, &out))
	assert.Equal(t, in, out)

	t.Run("zero ID is null", func(t *testing.T) {
		data, err := json.Marshal(UserID{})
		require.NoError(t, err)
		assert.Equal(t, "null", string(data))

		got := trogonobjectid.MustNew[userType]("abc")
		require.NoError(t, json.Unmarshal([]byte("null"), &got))
		assert.True(t, got.IsZero())
	})

	t.Run("rejects the wrong form", func(t *testing.T) {
		var u UserID
		require.ErrorIs(t, json.Unmarshal([]byte(`"abc"`), &u), trogonobjectid.ErrInvalidFormat)

		var inv InvoiceID
		require.ErrorIs(t, json.Unmarshal([]byte(`"inv_`+invoiceUUID+`"`), &inv), trogonobjectid.ErrInvalidUUID)
	})

	t.Run("rejects non-strings", func(t *testing.T) {
		var u UserID
		require.Error(t, json.Unmarshal([]byte(`42`), &u))
	})
}

func TestText(t *testing.T) {
	t.Run("always uses the full form", func(t *testing.T) {
		id := trogonobjectid.MustNew[invoiceType](invoiceUUID)
		text, err := id.MarshalText()
		require.NoError(t, err)
		assert.Equal(t, "inv_"+invoiceUUID, string(text))

		var got InvoiceID
		require.NoError(t, got.UnmarshalText(text))
		assert.Equal(t, id, got)
	})

	t.Run("map keys use the full form", func(t *testing.T) {
		m := map[InvoiceID]int{trogonobjectid.MustNew[invoiceType](invoiceUUID): 1}
		data, err := json.Marshal(m)
		require.NoError(t, err)
		assert.JSONEq(t, `{"inv_`+invoiceUUID+`":1}`, string(data))
	})
}

func TestLogging(t *testing.T) {
	id := trogonobjectid.MustNew[invoiceType](invoiceUUID)

	t.Run("fmt uses the full form", func(t *testing.T) {
		assert.Equal(t, "inv_"+invoiceUUID, fmt.Sprintf("%v", id))
	})

	t.Run("slog JSON handler uses the full form even with drop prefix JSON", func(t *testing.T) {
		var buf bytes.Buffer
		slog.New(slog.NewJSONHandler(&buf, nil)).Info("charged", "invoice_id", id)

		var entry map[string]any
		require.NoError(t, json.Unmarshal(buf.Bytes(), &entry))
		assert.Equal(t, "inv_"+invoiceUUID, entry["invoice_id"])
	})
}

func TestZero(t *testing.T) {
	var id UserID
	assert.True(t, id.IsZero())
	assert.Empty(t, id.String())
	assert.Empty(t, id.Bare())
	assert.Empty(t, id.Storage())
	assert.Equal(t, UserID{}, id)
}
