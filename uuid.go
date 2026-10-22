package trogonobjectid

import (
	"fmt"

	"github.com/google/uuid"
)

// FromUUID returns the ID for u, after the type's Normalize and Validate.
func FromUUID[T Definer](u uuid.UUID) (ID[T], error) {
	return New[T](u.String())
}

// UUID returns the bare value as a uuid.UUID, so callers that work with UUIDs do not
// parse Bare themselves. The zero ID fails with ErrEmptyValue.
func (id ID[T]) UUID() (uuid.UUID, error) {
	if id.IsZero() {
		return uuid.Nil, &ValidationError{ObjectType: id.ObjectType(), Err: ErrEmptyValue}
	}
	u, err := uuid.Parse(id.id)
	if err != nil {
		return uuid.Nil, &ValidationError{ObjectType: id.ObjectType(), Value: id.id, Err: fmt.Errorf("%w: %w", ErrInvalidUUID, err)}
	}
	return u, nil
}
