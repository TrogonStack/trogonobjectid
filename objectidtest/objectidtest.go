// Package objectidtest builds deterministic object IDs for tests, so a test can name
// "the buyer" or "the seller" and get the same valid ID on every run:
//
//	buyer := objectidtest.FromSeed[fanType]("buyer")
//
// The bare value is a UUIDv5 of the seed, so it passes ValidateUUID.
package objectidtest

import (
	"github.com/google/uuid"

	"github.com/TrogonStack/trogonobjectid"
)

// Namespace is the UUIDv5 namespace that FromSeed uses.
var Namespace = uuid.NewSHA1(uuid.NameSpaceURL, []byte("https://github.com/TrogonStack/trogonobjectid/objectidtest"))

// FromSeed returns the ID whose bare value is the UUIDv5 of seed in Namespace. It
// panics if the type rejects the value, because a test fixture must be valid.
func FromSeed[T trogonobjectid.Definer](seed string) trogonobjectid.ID[T] {
	return FromSeedIn[T](Namespace, seed)
}

// FromSeedIn is like FromSeed in another namespace, for matching IDs that another
// service's test helpers derive.
func FromSeedIn[T trogonobjectid.Definer](namespace uuid.UUID, seed string) trogonobjectid.ID[T] {
	id, err := trogonobjectid.FromUUID[T](uuid.NewSHA1(namespace, []byte(seed)))
	if err != nil {
		panic(err)
	}
	return id
}
