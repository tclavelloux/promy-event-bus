package ptr

import "time"

// String returns a pointer to the given string.
// Returns nil if the string is empty.
func String(s string) *string {
	if s == "" {
		return nil
	}

	return &s
}

// StringValue safely dereferences a string pointer.
// Returns empty string if pointer is nil.
func StringValue(s *string) string {
	if s == nil {
		return ""
	}

	return *s
}

// StringOrNil returns a pointer to the string, or nil if empty.
// Alias for String() for clarity in some contexts.
func StringOrNil(s string) *string {
	return String(s)
}

// Int returns a pointer to the given int.
func Int(i int) *int {
	return &i
}

// IntValue safely dereferences an int pointer.
// Returns 0 if pointer is nil.
func IntValue(i *int) int {
	if i == nil {
		return 0
	}

	return *i
}

// Int64 returns a pointer to the given int64.
func Int64(i int64) *int64 {
	return &i
}

// Int64Value safely dereferences an int64 pointer.
// Returns 0 if pointer is nil.
func Int64Value(i *int64) int64 {
	if i == nil {
		return 0
	}

	return *i
}

// Float64 returns a pointer to the given float64.
func Float64(f float64) *float64 {
	return &f
}

// Float64Value safely dereferences a float64 pointer.
// Returns 0.0 if pointer is nil.
func Float64Value(f *float64) float64 {
	if f == nil {
		return 0.0
	}

	return *f
}

// Bool returns a pointer to the given bool.
func Bool(b bool) *bool {
	return &b
}

// BoolValue safely dereferences a bool pointer.
// Returns false if pointer is nil.
func BoolValue(b *bool) bool {
	if b == nil {
		return false
	}

	return *b
}

// Time returns a pointer to the given time.Time.
func Time(t time.Time) *time.Time {
	return &t
}

// TimeValue safely dereferences a time.Time pointer.
// Returns zero time if pointer is nil.
func TimeValue(t *time.Time) time.Time {
	if t == nil {
		return time.Time{}
	}

	return *t
}

// IsNilOrEmpty returns true if the string pointer is nil or points to empty string.
func IsNilOrEmpty(s *string) bool {
	return s == nil || *s == ""
}

// StringSlice returns a pointer to the given []string.
func StringSlice(s []string) *[]string {
	return &s
}

// StringSliceValue safely dereferences a []string pointer.
func StringSliceValue(s *[]string) []string {
	if s == nil {
		return []string{}
	}

	return *s
}
