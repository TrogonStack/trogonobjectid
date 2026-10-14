package trogonobjectid

import "github.com/google/uuid"

// ValidateUUID accepts the UUID forms that uuid.Parse accepts. It keeps the value as given.
func ValidateUUID(id string) error {
	if _, err := uuid.Parse(id); err != nil {
		return ErrInvalidUUID
	}
	return nil
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
