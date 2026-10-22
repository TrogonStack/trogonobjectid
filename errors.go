package trogonobjectid

import (
	"errors"
	"fmt"
	"strings"
)

var (
	// ErrEmptyValue means the bare value is empty.
	ErrEmptyValue = errors.New("empty value")
	// ErrInvalidFormat means the input does not have the expected prefix or Go type.
	ErrInvalidFormat = errors.New("invalid format")
	// ErrInvalidUUID means ValidateUUID rejected the bare value.
	ErrInvalidUUID = errors.New("invalid uuid")
	// ErrInvalidInteger means ValidateInteger rejected the bare value.
	ErrInvalidInteger = errors.New("invalid integer")
	// ErrInvalidDefinition means the marker type's Definition is not usable.
	ErrInvalidDefinition = errors.New("trogonobjectid: invalid definition")
	// ErrNotMember means the ID passed to NewUnion is not one of the union's member types.
	ErrNotMember = errors.New("not a union member")
	// ErrNoGenerator means Generate was called on a type without a Generate function.
	ErrNoGenerator = errors.New("trogonobjectid: no generator")
)

// ValidationError reports a value that is not a valid ID of the listed object types: one
// for an ID, every member's for a union. Err is ErrEmptyValue, ErrInvalidFormat, a
// *WrongPrefixError, or the error from Definition.Validate.
type ValidationError struct {
	ObjectTypes []ObjectType
	Value       string
	Err         error
}

func (e *ValidationError) Error() string {
	names := make([]string, len(e.ObjectTypes))
	for i, t := range e.ObjectTypes {
		names[i] = string(t)
	}
	return fmt.Sprintf("trogonobjectid: invalid %s value %q: %v", strings.Join(names, "|"), e.Value, e.Err)
}

func (e *ValidationError) Unwrap() error { return e.Err }

// WrongPrefixError reports a value without one of the prefixes the type accepts, for
// example an order ID where an invoice ID is expected. ValidationError.Value holds the
// input. It unwraps to ErrInvalidFormat.
type WrongPrefixError struct {
	Want []Prefix
}

func (e *WrongPrefixError) Error() string {
	return fmt.Sprintf("%v: want prefix %q", ErrInvalidFormat, e.Want)
}

func (e *WrongPrefixError) Unwrap() error { return ErrInvalidFormat }
