// Package types provides validation utilities for Telegram client types.
package types

import (
	"fmt"
	"reflect"
	"strings"
)

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
	if tag == "" {
		return nil
	}

	return processValidateTag(field, fieldType.Name, tag)
}

// validateEmbedded validates an embedded struct if it has a Validate method.
func validateEmbedded(field reflect.Value) error {
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
