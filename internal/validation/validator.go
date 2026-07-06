package validation

import (
	"fmt"
	"strings"
)

// Validator checks a raw string value, returning a descriptive error if it
// is invalid, or nil if it is acceptable.
type Validator func(value string) error

// Required rejects a value that is empty once trimmed.
func Required() Validator {
	return func(value string) error {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("this field is required")
		}
		return nil
	}
}

// Optional wraps v so that an empty (once trimmed) value is always accepted,
// while a non-empty value still has to satisfy v.
func Optional(v Validator) Validator {
	return func(value string) error {
		if strings.TrimSpace(value) == "" {
			return nil
		}
		return v(value)
	}
}

// All runs every validator in order, failing on the first error.
func All(validators ...Validator) Validator {
	return func(value string) error {
		for _, v := range validators {
			if err := v(value); err != nil {
				return err
			}
		}
		return nil
	}
}
