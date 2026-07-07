package operations

import (
	"encoding/json"
	"testing"
)

func TestManifestIncludesSchemasAndSafety(t *testing.T) {
	op, ok := Get("send_message")
	if !ok {
		t.Fatal("send_message not registered")
	}
	manifest := ManifestFor(op)

	if manifest.Safety != SafetyWrite {
		t.Fatalf("send_message safety = %q, want %q", manifest.Safety, SafetyWrite)
	}
	if manifest.InputSchema["type"] != "object" {
		t.Fatalf("input schema type = %v, want object", manifest.InputSchema["type"])
	}
	if manifest.OutputSchema["type"] != "object" {
		t.Fatalf("output schema type = %v, want object", manifest.OutputSchema["type"])
	}
}

func TestPeerInfoSchemaDocumentsPeerOrUsername(t *testing.T) {
	op, ok := Get("send_message")
	if !ok {
		t.Fatal("send_message not registered")
	}
	manifest := ManifestFor(op)
	if _, ok := manifest.InputSchema["anyOf"]; !ok {
		t.Fatal("expected anyOf for peer/username validation")
	}
}

func TestOutputSchemaDoesNotIncludeInputOnlyPeerAnyOf(t *testing.T) {
	op, ok := Get("add_contact")
	if !ok {
		t.Fatal("add_contact not registered")
	}
	manifest := ManifestFor(op)
	properties, ok := manifest.OutputSchema["properties"].(map[string]any)
	if !ok {
		t.Fatal("output properties missing")
	}
	contact, ok := properties["contact"].(JSONSchema)
	if !ok {
		t.Fatalf("contact schema type = %T, want JSONSchema", properties["contact"])
	}
	if _, ok := contact["anyOf"]; ok {
		t.Fatal("output schema should not include peer/username anyOf")
	}
}

func TestValidateParams(t *testing.T) {
	valid := json.RawMessage(`{"peer":"@user","message":"hello"}`)
	if err := ValidateParams("send_message", valid); err != nil {
		t.Fatalf("valid params rejected: %v", err)
	}

	invalid := json.RawMessage(`{"peer":"@user"}`)
	if err := ValidateParams("send_message", invalid); err == nil {
		t.Fatal("missing message should be rejected")
	}

	missingPeer := json.RawMessage(`{"message":"hello"}`)
	if err := ValidateParams("send_message", missingPeer); err == nil {
		t.Fatal("missing peer/username should be rejected")
	}
}

func TestOpenAPIIncludesRPCOperation(t *testing.T) {
	doc := OpenAPI("test", "dev")
	paths, ok := doc["paths"].(map[string]any)
	if !ok {
		t.Fatal("openapi paths missing")
	}
	if _, ok := paths["/rpc/send_message"]; !ok {
		t.Fatal("send_message path missing from OpenAPI")
	}
	if _, ok := paths["/manifest"]; !ok {
		t.Fatal("manifest path missing from OpenAPI")
	}
}
