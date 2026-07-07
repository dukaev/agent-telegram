package observability

import (
	"strings"
	"testing"
)

func TestRedactAnyMasksSecretsAndPhones(t *testing.T) {
	input := map[string]any{
		"phone":         "+88806283792",
		"token":         "secret-token",
		"password":      "secret-password",
		"phoneCodeHash": "hash",
		"message":       "hello",
	}

	out := RedactAny(input).(map[string]any)
	if out["phone"] != "***3792" {
		t.Fatalf("phone = %v, want masked phone", out["phone"])
	}
	for _, key := range []string{"token", "password", "phoneCodeHash"} {
		if out[key] != redacted {
			t.Fatalf("%s = %v, want redacted", key, out[key])
		}
	}
	if out["message"] != "hello" {
		t.Fatalf("short message should remain useful, got %v", out["message"])
	}
}

func TestRedactAnyTruncatesLongText(t *testing.T) {
	input := map[string]any{"message": "abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyz"}
	out := RedactAny(input).(map[string]any)
	message, _ := out["message"].(string)
	if !strings.Contains(message, "[truncated") {
		t.Fatalf("message should be truncated: %q", message)
	}
}

func TestRedactAuditEventsForDisplaySafe(t *testing.T) {
	events := []AuditEvent{{
		Params: map[string]any{
			"message":   "11.11.2000",
			"peer":      "@bot",
			"city":      "Тбилиси",
			"latitude":  41.7151,
			"longitude": 44.8271,
		},
	}}

	out := RedactAuditEventsForDisplay(events, RedactionSafe)
	params := out[0].Params.(map[string]any)
	if params["message"] != "[TEXT REDACTED]" {
		t.Fatalf("message = %v, want redacted", params["message"])
	}
	if params["city"] != "[LOCATION REDACTED]" {
		t.Fatalf("city = %v, want redacted", params["city"])
	}
	if params["latitude"] != "[PERSONAL DATA REDACTED]" || params["longitude"] != "[PERSONAL DATA REDACTED]" {
		t.Fatalf("coords not redacted: %+v", params)
	}
	if params["peer"] != "@bot" {
		t.Fatalf("peer should remain useful: %+v", params)
	}
}

func TestRedactLogLineForDisplaySafe(t *testing.T) {
	line := `{"level":"INFO","params":"{\"message\":\"hello\",\"peer\":\"@bot\"}"}`
	out := RedactLogLineForDisplay(line, RedactionSafe)
	if strings.Contains(out, "hello") {
		t.Fatalf("line should hide text: %s", out)
	}
	if !strings.Contains(out, "@bot") {
		t.Fatalf("line should keep peer: %s", out)
	}
}
