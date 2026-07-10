package types

import (
	"errors"
	"strings"
	"testing"
)

type EmbeddedValidator struct {
	err error
}

func (e *EmbeddedValidator) Validate() error {
	return e.err
}

func TestValidateStructRequiredFields(t *testing.T) {
	type sample struct {
		StringValue string         `validate:"required"`
		SliceValue  []string       `validate:"required"`
		MapValue    map[string]int `validate:"required"`
		IntValue    int            `validate:"required"`
		UintValue   uint           `validate:"required"`
		FloatValue  float64        `validate:"required"`
		BoolValue   bool           `validate:"required"`
		PtrValue    *int           `validate:"required"`
		_           string         `validate:"required"`
	}

	if err := ValidateStruct(42); err != nil {
		t.Fatalf("non-struct validation returned error: %v", err)
	}
	if err := ValidateStruct(sample{}); err == nil || !strings.Contains(err.Error(), "string_value") {
		t.Fatalf("empty struct error = %v, want string_value required", err)
	}

	value := 1
	valid := sample{
		StringValue: "x",
		SliceValue:  []string{"x"},
		MapValue:    map[string]int{"x": 1},
		IntValue:    1,
		UintValue:   1,
		FloatValue:  1,
		BoolValue:   true,
		PtrValue:    &value,
	}
	if err := ValidateStruct(&valid); err != nil {
		t.Fatalf("valid struct error = %v", err)
	}
}

func TestValidateStructEmbeddedValidator(t *testing.T) {
	want := errors.New("embedded")
	type sample struct {
		EmbeddedValidator
	}

	if err := ValidateStruct(&sample{EmbeddedValidator: EmbeddedValidator{err: want}}); !errors.Is(err, want) {
		t.Fatalf("embedded error = %v, want %v", err, want)
	}
	if err := ValidateStruct(sample{}); err != nil {
		t.Fatalf("value embedded validation should be ok: %v", err)
	}
}

func TestValidationHelpers(t *testing.T) {
	if err := RequiredString("", "name"); err == nil {
		t.Fatal("RequiredString should reject empty strings")
	}
	if err := RequiredString("ok", "name"); err != nil {
		t.Fatalf("RequiredString valid = %v", err)
	}
	if err := RequiredSlice([]int{}, "items"); err == nil {
		t.Fatal("RequiredSlice should reject empty slices")
	}
	if err := RequiredSlice([]int{1}, "items"); err != nil {
		t.Fatalf("RequiredSlice valid = %v", err)
	}
	for _, lat := range []float64{-91, 91} {
		if err := ValidateLatitude(lat); err == nil {
			t.Fatalf("latitude %v should be invalid", lat)
		}
	}
	for _, lng := range []float64{-181, 181} {
		if err := ValidateLongitude(lng); err == nil {
			t.Fatalf("longitude %v should be invalid", lng)
		}
	}
	if err := ValidateLatitude(45); err != nil {
		t.Fatalf("latitude valid = %v", err)
	}
	if err := ValidateLongitude(90); err != nil {
		t.Fatalf("longitude valid = %v", err)
	}
}

func TestCommonParamValidation(t *testing.T) {
	validPeer := PeerInfo{Peer: "@peer"}
	if err := (PeerInfo{}).Validate(); err == nil {
		t.Fatal("empty PeerInfo should be invalid")
	}
	if err := validPeer.Validate(); err != nil {
		t.Fatalf("valid PeerInfo = %v", err)
	}
	if err := (MsgID{}).Validate(); err == nil {
		t.Fatal("empty MsgID should be invalid")
	}
	if err := (MsgID{MessageID: 1}).Validate(); err != nil {
		t.Fatalf("valid MsgID = %v", err)
	}
	if err := (RequiredText{}).ValidateText(); err == nil {
		t.Fatal("empty RequiredText should be invalid")
	}
	if err := (RequiredText{Text: "hello"}).ValidateText(); err != nil {
		t.Fatalf("valid RequiredText = %v", err)
	}
}
