package trogonobjectid_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/TrogonStack/trogonobjectid"
)

func TestFromUUID(t *testing.T) {
	u := uuid.MustParse(invoiceUUID)

	id, err := trogonobjectid.FromUUID[invoiceType](u)
	require.NoError(t, err)
	assert.Equal(t, trogonobjectid.MustNew[invoiceType](invoiceUUID), id)

	t.Run("validates against the type", func(t *testing.T) {
		_, err := trogonobjectid.FromUUID[seqType](u)
		require.ErrorIs(t, err, trogonobjectid.ErrInvalidInteger)
	})
}

func TestIDUUID(t *testing.T) {
	got, err := trogonobjectid.MustNew[invoiceType](invoiceUUID).UUID()
	require.NoError(t, err)
	assert.Equal(t, uuid.MustParse(invoiceUUID), got)

	t.Run("the zero ID has no UUID", func(t *testing.T) {
		_, err := InvoiceID{}.UUID()
		require.ErrorIs(t, err, trogonobjectid.ErrEmptyValue)
	})

	t.Run("a bare value that is not a UUID has no UUID", func(t *testing.T) {
		_, err := trogonobjectid.MustNew[userType]("abc-123").UUID()
		require.ErrorIs(t, err, trogonobjectid.ErrInvalidUUID)
	})
}

func TestScanNativeUUID(t *testing.T) {
	want := trogonobjectid.MustNew[invoiceType](invoiceUUID)
	u := uuid.MustParse(invoiceUUID)

	for name, src := range map[string]any{"uuid.UUID": u, "[16]byte": [16]byte(u)} {
		t.Run("Scan accepts "+name, func(t *testing.T) {
			var got InvoiceID
			require.NoError(t, got.Scan(src))
			assert.Equal(t, want, got)
		})
		t.Run("DecodeSpanner accepts "+name, func(t *testing.T) {
			var got InvoiceID
			require.NoError(t, got.DecodeSpanner(src))
			assert.Equal(t, want, got)
		})
	}

	t.Run("a full storage format rejects a native UUID", func(t *testing.T) {
		var got UserID
		require.ErrorIs(t, got.Scan(u), trogonobjectid.ErrInvalidFormat)
	})
}
