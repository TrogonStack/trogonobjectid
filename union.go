package trogonobjectid

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"log/slog"
	"reflect"
	"sort"
	"strings"
	"sync"
)

// Member is an ID that a Union can hold. Every ID[T] implements it.
type Member interface {
	fmt.Stringer
	Prefix() Prefix
	IsZero() bool
}

// MemberType describes one ID type that a Union accepts. Make it with MemberOf.
type MemberType struct {
	objectType ObjectType
	prefixes   []Prefix
	err        error
	parse      func(s string) (Member, error)
	holds      func(m Member) bool
}

// MemberOf returns the MemberType for ID[T].
func MemberOf[T Definer]() MemberType {
	s, err := specFor[T]()
	return MemberType{
		objectType: s.ObjectType,
		prefixes:   s.prefixes,
		err:        err,
		parse:      func(str string) (Member, error) { return Parse[T](str) },
		holds:      func(m Member) bool { _, ok := m.(ID[T]); return ok },
	}
}

// UnionDefiner is implemented by the marker type that names a Union:
//
//	type principalUnion struct{}
//
//	func (principalUnion) Types() []trogonobjectid.MemberType {
//		return []trogonobjectid.MemberType{
//			trogonobjectid.MemberOf[userType](),
//			trogonobjectid.MemberOf[systemType](),
//		}
//	}
//
//	type PrincipalID = trogonobjectid.Union[principalUnion]
type UnionDefiner interface {
	Types() []MemberType
}

// Union holds one ID of any of its member types, chosen by prefix when parsing.
// Use UnionAs, or a type switch on Member, to get the typed ID back.
//
// Every encoding uses the member's full form, because the prefix is the only thing
// that tells the member types apart. The zero value is an absent ID.
type Union[U UnionDefiner] struct {
	member Member
}

type unionSpec struct {
	objectTypes []ObjectType
	members     []MemberType
	// entries pairs every member prefix, aliases included, with its member, longest first.
	entries []unionEntry
}

type unionEntry struct {
	prefix Prefix
	member MemberType
}

var unionSpecs sync.Map

func unionSpecFor[U UnionDefiner]() (*unionSpec, error) {
	return loadOrBuild(&unionSpecs, reflect.TypeFor[U](), newUnionSpec[U])
}

func newUnionSpec[U UnionDefiner]() (unionSpec, error) {
	var marker U
	members := marker.Types()
	if len(members) == 0 {
		return unionSpec{}, fmt.Errorf("%w: %T has no members", ErrInvalidDefinition, marker)
	}

	objectTypes := make([]ObjectType, 0, len(members))
	seen := make(map[Prefix]bool, len(members))
	var entries []unionEntry
	for _, m := range members {
		if m.err != nil {
			return unionSpec{}, fmt.Errorf("%T: %w", marker, m.err)
		}
		for _, p := range m.prefixes {
			if seen[p] {
				return unionSpec{}, fmt.Errorf("%w: %T has duplicate prefix %q", ErrInvalidDefinition, marker, p)
			}
			seen[p] = true
			entries = append(entries, unionEntry{prefix: p, member: m})
		}
		objectTypes = append(objectTypes, m.objectType)
	}

	// Longest prefix first, so "acme_admin_" wins over "acme_" for "acme_admin_xyz".
	sort.SliceStable(entries, func(i, j int) bool { return len(entries[i].prefix) > len(entries[j].prefix) })
	return unionSpec{objectTypes: objectTypes, members: members, entries: entries}, nil
}

// CheckUnion reports whether U's member list is usable: at least one member, valid
// member definitions and no duplicate prefixes.
func CheckUnion[U UnionDefiner]() error {
	_, err := unionSpecFor[U]()
	return err
}

// NewUnion wraps m, which must be a non-zero ID of one of the union's member types.
func NewUnion[U UnionDefiner](m Member) (Union[U], error) {
	s, err := unionSpecFor[U]()
	if err != nil {
		return Union[U]{}, err
	}
	if m == nil || m.IsZero() {
		return Union[U]{}, &ValidationError{ObjectTypes: s.objectTypes, Err: ErrEmptyValue}
	}
	for _, mt := range s.members {
		if mt.holds(m) {
			return Union[U]{member: m}, nil
		}
	}
	return Union[U]{}, &ValidationError{ObjectTypes: s.objectTypes, Value: m.String(), Err: ErrNotMember}
}

// MustNewUnion is like NewUnion but panics on error. Use it for constants and tests.
func MustNewUnion[U UnionDefiner](m Member) Union[U] {
	v, err := NewUnion[U](m)
	if err != nil {
		panic(err)
	}
	return v
}

// ParseUnion parses the full form of any member type. When prefixes overlap,
// the longest matching prefix wins.
func ParseUnion[U UnionDefiner](str string) (Union[U], error) {
	s, err := unionSpecFor[U]()
	if err != nil {
		return Union[U]{}, err
	}
	for _, e := range s.entries {
		if strings.HasPrefix(str, string(e.prefix)) {
			m, err := e.member.parse(str)
			if err != nil {
				return Union[U]{}, err
			}
			return Union[U]{member: m}, nil
		}
	}
	prefixes := make([]Prefix, len(s.entries))
	for i, e := range s.entries {
		prefixes[i] = e.prefix
	}
	return Union[U]{}, &ValidationError{ObjectTypes: s.objectTypes, Value: str, Err: &WrongPrefixError{Want: prefixes}}
}

// MustParseUnion is like ParseUnion but panics on error. Use it for constants and tests.
func MustParseUnion[U UnionDefiner](str string) Union[U] {
	v, err := ParseUnion[U](str)
	if err != nil {
		panic(err)
	}
	return v
}

// Member returns the wrapped ID, or nil for the zero union.
func (u Union[U]) Member() Member { return u.member }

// UnionAs returns u's member as an ID[T], and false when u holds another type or
// is zero. It narrows a union without a type switch.
func UnionAs[T Definer, U UnionDefiner](u Union[U]) (ID[T], bool) {
	id, ok := u.member.(ID[T])
	return id, ok
}

// IsZero reports whether u is the zero (absent) union.
func (u Union[U]) IsZero() bool { return u.member == nil }

// String returns the member's full form, or "" for the zero union.
func (u Union[U]) String() string {
	if u.member == nil {
		return ""
	}
	return u.member.String()
}

// LogValue implements slog.LogValuer with the full form.
func (u Union[U]) LogValue() slog.Value { return slog.StringValue(u.String()) }

// MarshalText implements encoding.TextMarshaler with the full form.
func (u Union[U]) MarshalText() ([]byte, error) { return []byte(u.String()), nil }

// UnmarshalText implements encoding.TextUnmarshaler. Empty text gives the zero union.
func (u *Union[U]) UnmarshalText(text []byte) error {
	return u.set(string(text))
}

// MarshalJSON implements json.Marshaler with the full form, ignoring the member's
// JSONFormat so that UnmarshalJSON can pick the member type again. The zero union is null.
func (u Union[U]) MarshalJSON() ([]byte, error) {
	if u.IsZero() {
		return []byte("null"), nil
	}
	return json.Marshal(u.String())
}

// UnmarshalJSON implements json.Unmarshaler for the full form. Null gives the zero union.
func (u *Union[U]) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*u = Union[U]{}
		return nil
	}
	var raw string
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	return u.set(raw)
}

// Value implements driver.Valuer with the full form, ignoring the member's StorageFormat.
// The zero union is NULL.
func (u Union[U]) Value() (driver.Value, error) {
	if u.IsZero() {
		return nil, nil
	}
	return u.String(), nil
}

// Scan implements sql.Scanner for the full form. NULL and the empty string give the zero union.
func (u *Union[U]) Scan(src any) error {
	switch v := src.(type) {
	case nil:
		*u = Union[U]{}
		return nil
	case string:
		return u.set(v)
	case []byte:
		return u.set(string(v))
	default:
		return fmt.Errorf("%w: cannot scan %T", ErrInvalidFormat, src)
	}
}

func (u *Union[U]) set(str string) error {
	if str == "" {
		*u = Union[U]{}
		return nil
	}
	v, err := ParseUnion[U](str)
	if err != nil {
		return err
	}
	*u = v
	return nil
}

// EncodeSpanner implements spanner.Encoder with the full form. The zero union is a NULL
// STRING, the same as the zero ID.
func (u Union[U]) EncodeSpanner() (any, error) {
	if u.IsZero() {
		return (*string)(nil), nil
	}
	return u.String(), nil
}

// DecodeSpanner implements spanner.Decoder for STRING values in the full form.
func (u *Union[U]) DecodeSpanner(val any) error { return u.Scan(val) }
