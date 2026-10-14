package trogonobjectid

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
)

// DefaultSeparator is the separator used when Definition.Separator is empty.
const DefaultSeparator = "_"

// Format selects whether an encoding keeps the prefix.
type Format int

const (
	// FormatFull emits the prefixed form, for example "user_abc-123".
	FormatFull Format = iota
	// FormatDropPrefix emits only the bare value, for example "abc-123".
	FormatDropPrefix
)

// Definition configures an object ID type.
type Definition struct {
	// ObjectType is the prefix without the separator, for example "user". It is required.
	ObjectType string
	// Separator goes between the object type and the bare value. Empty means DefaultSeparator.
	Separator string
	// StorageFormat controls Storage, FromStorage, Value and Scan.
	StorageFormat Format
	// JSONFormat controls MarshalJSON and UnmarshalJSON.
	JSONFormat Format
	// Validate checks the bare value. Nil accepts any non-empty value.
	Validate func(id string) error
	// Generate returns a new bare value for Generate. Nil makes Generate fail.
	Generate func() (string, error)
}

// Definer is implemented by the marker type that names an object ID type.
// Implement it on an empty struct with a value receiver:
//
//	type userType struct{}
//
//	func (userType) Definition() trogonobjectid.Definition {
//		return trogonobjectid.Definition{ObjectType: "user"}
//	}
//
//	type UserID = trogonobjectid.ID[userType]
type Definer interface {
	Definition() Definition
}

// ID is a type-safe object ID. IDs with different marker types are different Go types,
// so a UserID cannot be passed where an OrderID is expected.
//
// The zero value is an absent ID: it encodes as SQL NULL and JSON null.
// String, MarshalText and LogValue always use the prefixed form, so logs and URLs
// stay unambiguous no matter how the type is stored or sent as JSON.
type ID[T Definer] struct {
	id string
}

type spec struct {
	Definition
	prefix string
}

func specFor[T Definer]() (spec, error) {
	var marker T
	def := marker.Definition()
	if def.ObjectType == "" {
		return spec{}, fmt.Errorf("%w: empty object type for %T", ErrInvalidDefinition, marker)
	}
	if def.Separator == "" {
		def.Separator = DefaultSeparator
	}
	return spec{Definition: def, prefix: def.ObjectType + def.Separator}, nil
}

func (s spec) build(id string) error {
	if id == "" {
		return &ValidationError{ObjectType: s.ObjectType, Value: id, Err: ErrEmptyValue}
	}
	if s.Validate != nil {
		if err := s.Validate(id); err != nil {
			return &ValidationError{ObjectType: s.ObjectType, Value: id, Err: err}
		}
	}
	return nil
}

func (s spec) parse(full string) (string, error) {
	bare, ok := strings.CutPrefix(full, s.prefix)
	if !ok || bare == "" {
		return "", &ValidationError{ObjectType: s.ObjectType, Value: full, Err: ErrInvalidFormat}
	}
	if err := s.build(bare); err != nil {
		return "", err
	}
	return bare, nil
}

func (s spec) decode(format Format, value string) (string, error) {
	if format == FormatDropPrefix {
		if err := s.build(value); err != nil {
			return "", err
		}
		return value, nil
	}
	return s.parse(value)
}

func (s spec) encode(format Format, id string) string {
	if format == FormatDropPrefix {
		return id
	}
	return s.prefix + id
}

// New returns an ID for the bare value id, after validation.
func New[T Definer](id string) (ID[T], error) {
	s, err := specFor[T]()
	if err != nil {
		return ID[T]{}, err
	}
	if err := s.build(id); err != nil {
		return ID[T]{}, err
	}
	return ID[T]{id: id}, nil
}

// MustNew is like New but panics on error. Use it for constants and tests.
func MustNew[T Definer](id string) ID[T] {
	v, err := New[T](id)
	if err != nil {
		panic(err)
	}
	return v
}

// Parse returns the ID for the prefixed form s, for example "user_abc-123".
func Parse[T Definer](s string) (ID[T], error) {
	sp, err := specFor[T]()
	if err != nil {
		return ID[T]{}, err
	}
	bare, err := sp.parse(s)
	if err != nil {
		return ID[T]{}, err
	}
	return ID[T]{id: bare}, nil
}

// MustParse is like Parse but panics on error. Use it for constants and tests.
func MustParse[T Definer](s string) ID[T] {
	v, err := Parse[T](s)
	if err != nil {
		panic(err)
	}
	return v
}

// Generate returns a new ID from the type's Generate function.
func Generate[T Definer]() (ID[T], error) {
	s, err := specFor[T]()
	if err != nil {
		return ID[T]{}, err
	}
	if s.Generate == nil {
		return ID[T]{}, fmt.Errorf("%w: %s", ErrNoGenerator, s.ObjectType)
	}
	id, err := s.Generate()
	if err != nil {
		return ID[T]{}, fmt.Errorf("trogonobjectid: generate %s: %w", s.ObjectType, err)
	}
	if err := s.build(id); err != nil {
		return ID[T]{}, err
	}
	return ID[T]{id: id}, nil
}

// FromStorage returns the ID for a value in the type's StorageFormat.
// It is the inverse of Storage.
func FromStorage[T Definer](value string) (ID[T], error) {
	s, err := specFor[T]()
	if err != nil {
		return ID[T]{}, err
	}
	bare, err := s.decode(s.StorageFormat, value)
	if err != nil {
		return ID[T]{}, err
	}
	return ID[T]{id: bare}, nil
}

// mustSpec is for methods on an ID that already exists, which proves the definition is valid.
func mustSpec[T Definer]() spec {
	s, err := specFor[T]()
	if err != nil {
		panic(err)
	}
	return s
}

// IsZero reports whether id is the zero (absent) ID.
func (id ID[T]) IsZero() bool { return id.id == "" }

// Bare returns the value without the prefix, for example "abc-123".
func (id ID[T]) Bare() string { return id.id }

// ObjectType returns the type's object type, for example "user".
func (id ID[T]) ObjectType() string { return mustSpec[T]().ObjectType }

// Prefix returns the object type and separator, for example "user_".
func (id ID[T]) Prefix() string { return mustSpec[T]().prefix }

// String returns the prefixed form, for example "user_abc-123", or "" for the zero ID.
func (id ID[T]) String() string {
	if id.IsZero() {
		return ""
	}
	return mustSpec[T]().prefix + id.id
}

// Storage returns the value in the type's StorageFormat, or "" for the zero ID.
func (id ID[T]) Storage() string {
	if id.IsZero() {
		return ""
	}
	s := mustSpec[T]()
	return s.encode(s.StorageFormat, id.id)
}

// LogValue implements slog.LogValuer. Without it, slog.JSONHandler would call
// MarshalJSON and log the bare value for types that drop the prefix in JSON.
func (id ID[T]) LogValue() slog.Value { return slog.StringValue(id.String()) }

// MarshalText implements encoding.TextMarshaler with the prefixed form.
// JSON map keys and flag values use it, and they must parse back without other context.
func (id ID[T]) MarshalText() ([]byte, error) { return []byte(id.String()), nil }

// UnmarshalText implements encoding.TextUnmarshaler. Empty text gives the zero ID.
func (id *ID[T]) UnmarshalText(text []byte) error {
	if len(text) == 0 {
		*id = ID[T]{}
		return nil
	}
	v, err := Parse[T](string(text))
	if err != nil {
		return err
	}
	*id = v
	return nil
}

// MarshalJSON implements json.Marshaler with the type's JSONFormat. The zero ID is null.
func (id ID[T]) MarshalJSON() ([]byte, error) {
	if id.IsZero() {
		return []byte("null"), nil
	}
	s := mustSpec[T]()
	return json.Marshal(s.encode(s.JSONFormat, id.id))
}

// UnmarshalJSON implements json.Unmarshaler with the type's JSONFormat. Null gives the zero ID.
func (id *ID[T]) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*id = ID[T]{}
		return nil
	}
	var raw string
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	s, err := specFor[T]()
	if err != nil {
		return err
	}
	bare, err := s.decode(s.JSONFormat, raw)
	if err != nil {
		return err
	}
	*id = ID[T]{id: bare}
	return nil
}

// Value implements driver.Valuer with the type's StorageFormat. The zero ID is NULL.
func (id ID[T]) Value() (driver.Value, error) {
	if id.IsZero() {
		return nil, nil
	}
	return id.Storage(), nil
}

// Scan implements sql.Scanner for values in the type's StorageFormat.
// NULL and the empty string give the zero ID.
func (id *ID[T]) Scan(src any) error {
	var value string
	switch v := src.(type) {
	case nil:
		*id = ID[T]{}
		return nil
	case string:
		value = v
	case []byte:
		value = string(v)
	default:
		return fmt.Errorf("%w: cannot scan %T", ErrInvalidFormat, src)
	}
	if value == "" {
		*id = ID[T]{}
		return nil
	}
	v, err := FromStorage[T](value)
	if err != nil {
		return err
	}
	*id = v
	return nil
}
