package trogonobjectid

import "github.com/google/uuid"

// ValidateUUID accepts the UUID forms that uuid.Parse accepts. It keeps the value as
// given; use NormalizeUUID instead to store one canonical form.
func ValidateUUID(id string) error {
	_, err := parseUUID(id)
	return err
}

func parseUUID(id string) (uuid.UUID, error) {
	u, err := uuid.Parse(id)
	if err != nil {
		return uuid.Nil, ErrInvalidUUID
	}
	return u, nil
}

// ValidateInteger accepts an optional sign followed by decimal digits, with no size limit,
// to match the Elixir package.
func ValidateInteger(id string) error {
	digits := id
	if len(digits) > 0 && (digits[0] == '+' || digits[0] == '-') {
		digits = digits[1:]
	}
	if digits == "" {
		return ErrInvalidInteger
	}
	for i := 0; i < len(digits); i++ {
		if digits[i] < '0' || digits[i] > '9' {
			return ErrInvalidInteger
		}
	}
	return nil
}

// GenerateUUIDv7 returns a new time-ordered UUID string.
func GenerateUUIDv7() (string, error) {
	u, err := uuid.NewV7()
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

// GenerateUUIDv4 returns a new random UUID string.
func GenerateUUIDv4() (string, error) {
	u, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

// NormalizeUUID returns the canonical lowercase hyphenated form of any UUID form that
// uuid.Parse accepts, so upper-case and URN inputs name the same ID. It also rejects
// values that are not UUIDs, so a type that sets it needs no ValidateUUID.
func NormalizeUUID(id string) (string, error) {
	u, err := parseUUID(id)
	if err != nil {
		return "", err
	}
	if isCanonicalUUID(id) {
		return id, nil
	}
	return u.String(), nil
}

// isCanonicalUUID reports whether a value that uuid.Parse accepted is already in the
// 36-character lowercase form, so the common read path skips building a new string.
func isCanonicalUUID(id string) bool {
	if len(id) != 36 {
		return false
	}
	for i := 0; i < len(id); i++ {
		if id[i] >= 'A' && id[i] <= 'F' {
			return false
		}
	}
	return true
}
