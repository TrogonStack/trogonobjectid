package trogonobjectid

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestObjectTypeValidate(t *testing.T) {
	valid := []ObjectType{"inv", "credit_note", ObjectType(strings.Repeat("a", 63))}
	for _, objectType := range valid {
		assert.NoError(t, objectType.validate(), objectType)
	}

	invalid := []ObjectType{"", "User", "user2", "credit-note", "_user", "user_", ObjectType(strings.Repeat("a", 64))}
	for _, objectType := range invalid {
		assert.Error(t, objectType.validate(), objectType)
	}
}
