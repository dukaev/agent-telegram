// Package types provides validation utilities for Telegram client types.
package types

import (
	"fmt"
	"reflect"
	"strings"
)

// NoValidation is an embeddable type that provides a no-op Validate() method.
// Embed this in Params types that rely solely on struct-tag validation
// (the Handler auto-runs ValidateStruct before calling Validate).
type NoValidation struct{}

// Validate is a no-op — struct-tag validation is handled by Handler.
func (NoValidation) Validate() error { return nil }

// ValidateStruct validates a struct using struct tags.
// Supported tags:
//   - validate:"required" - field must not be empty/zero
//
// Note: Embedded structs with Validate() method are called automatically.
func ValidateStruct(v any) error {
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return nil
	}

	typ := val.Type()

	for i := range val.NumField() {
		field := val.Field(i)
		fieldType := typ.Field(i)

		if err := validateField(field, fieldType); err != nil {
			return err
		}
	}

	return nil
}

// validateField validates a single struct field.
func validateField(field reflect.Value, fieldType reflect.StructField) error {
	// Skip unexported fields
	if !fieldType.IsExported() {
		return nil
	}

	// Handle embedded structs with Validate() method
	if fieldType.Anonymous && field.Kind() == reflect.Struct {
		return validateEmbedded(field)
	}

	// Get validate tag
	tag := fieldType.Tag.Get("validate")
	if tag != "" {
		if err := processValidateTag(field, fieldType.Name, tag); err != nil {
			return err
		}
	}
	return validateNested(field)
}

func validateNested(field reflect.Value) error {
	if field.Kind() == reflect.Ptr {
		if field.IsNil() {
			return nil
		}
		field = field.Elem()
	}
	switch field.Kind() {
	case reflect.Struct:
		return ValidateStruct(field.Interface())
	case reflect.Slice, reflect.Array:
		for i := 0; i < field.Len(); i++ {
			if err := validateNested(field.Index(i)); err != nil {
				return fmt.Errorf("item %d: %w", i, err)
			}
		}
	}
	return nil
}

// validateEmbedded validates an embedded struct if it has a Validate method.
func validateEmbedded(field reflect.Value) error {
	// Check if field is addressable before calling Addr()
	if !field.CanAddr() {
		// If not addressable, try to validate via the value directly
		// This happens when the parent struct was passed by value
		if validator, ok := field.Interface().(interface{ Validate() error }); ok {
			return validator.Validate()
		}
		return nil
	}
	validator, ok := field.Addr().Interface().(interface{ Validate() error })
	if !ok {
		return nil
	}
	return validator.Validate()
}

// processValidateTag processes validate tag options.
func processValidateTag(field reflect.Value, fieldName, tag string) error {
	options := strings.Split(tag, ",")
	for _, opt := range options {
		if opt == "required" {
			if err := validateRequired(field, fieldName); err != nil {
				return err
			}
		}
	}
	return nil
}

// validateRequired checks if a field is empty and returns an error if so.
func validateRequired(field reflect.Value, name string) error {
	if isEmpty(field) {
		return fmt.Errorf("%s is required", toSnakeCase(name))
	}
	return nil
}

// isEmpty checks if a value is considered empty.
//
//nolint:exhaustive // We only care about common types, others return false.
func isEmpty(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.String:
		return v.String() == ""
	case reflect.Slice, reflect.Map, reflect.Array:
		return v.Len() == 0
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Ptr, reflect.Interface:
		return v.IsNil()
	default:
		return false
	}
}

// toSnakeCase converts CamelCase to snake_case.
func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteByte('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}

// RequiredString returns an error if s is empty.
func RequiredString(s, name string) error {
	if s == "" {
		return fmt.Errorf("%s is required", name)
	}
	return nil
}

// RequiredSlice returns an error if slice is empty.
func RequiredSlice[T any](slice []T, name string) error {
	if len(slice) == 0 {
		return fmt.Errorf("%s is required", name)
	}
	return nil
}

func eitherRequiredSchema(first, second string) map[string]any {
	return map[string]any{
		"anyOf": []map[string][]string{
			{"required": []string{first}},
			{"required": []string{second}},
		},
	}
}

// ValidateLatitude returns an error if latitude is out of range.
func ValidateLatitude(lat float64) error {
	if lat < -90 || lat > 90 {
		return fmt.Errorf("latitude must be between -90 and 90")
	}
	return nil
}

// ValidateLongitude returns an error if longitude is out of range.
func ValidateLongitude(lng float64) error {
	if lng < -180 || lng > 180 {
		return fmt.Errorf("longitude must be between -180 and 180")
	}
	return nil
}
