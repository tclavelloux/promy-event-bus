package eventbus

import (
	"errors"
	"fmt"
	"sync"

	"github.com/go-playground/validator/v10"
)

var (
	// Global validator instance (singleton)
	validatorInstance *validator.Validate
	validatorOnce     sync.Once
)

// GetValidator returns the global validator instance.
// It's initialized once and reused for performance.
func GetValidator() *validator.Validate {
	validatorOnce.Do(func() {
		validatorInstance = validator.New(validator.WithRequiredStructEnabled())
	})

	return validatorInstance
}

// ValidateStruct validates a struct using go-playground/validator tags.
func ValidateStruct(s any) error {
	if err := GetValidator().Struct(s); err != nil {
		var validationErrors validator.ValidationErrors
		if errors.As(err, &validationErrors) {
			// Format validation errors nicely
			return fmt.Errorf("%w: %s", ErrInvalidEvent, formatValidationErrors(validationErrors))
		}

		return fmt.Errorf("%w: %w", ErrInvalidEvent, err)
	}

	return nil
}

// formatValidationErrors formats validator.ValidationErrors into a readable string.
func formatValidationErrors(errs validator.ValidationErrors) string {
	if len(errs) == 0 {
		return "validation failed"
	}

	// Format first error
	err := errs[0]
	field := err.Field()
	tag := err.Tag()

	switch tag {
	case "required":
		return fmt.Sprintf("field '%s' is required", field)
	case "email":
		return fmt.Sprintf("field '%s' must be a valid email", field)
	case "min":
		return fmt.Sprintf("field '%s' must be at least %s", field, err.Param())
	case "max":
		return fmt.Sprintf("field '%s' must be at most %s", field, err.Param())
	case "uuid":
		return fmt.Sprintf("field '%s' must be a valid UUID", field)
	default:
		return fmt.Sprintf("field '%s' failed validation '%s'", field, tag)
	}
}
