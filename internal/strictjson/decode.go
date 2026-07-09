// Package strictjson provides the canonical JSON decoder for public contracts.
package strictjson

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

// Decode rejects unknown fields and trailing JSON values.
func Decode(data []byte, dst any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			return fmt.Errorf("multiple JSON values are not allowed")
		}
		return err
	}
	return nil
}
