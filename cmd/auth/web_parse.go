package auth

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"agent-telegram/internal/config"
	"agent-telegram/internal/policy"
)

func parseAuthField(r *http.Request, name string, trim func(string) string) (string, error) {
	if isJSONRequest(r) {
		var body map[string]string
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			return "", err
		}
		return trim(body[name]), nil
	}
	if err := r.ParseForm(); err != nil {
		return "", err
	}
	return trim(r.FormValue(name)), nil
}

func parseAuthModeRequest(r *http.Request) (mode string, phone string, err error) {
	if isJSONRequest(r) {
		var body struct {
			Mode  string `json:"mode"`
			Phone string `json:"phone"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			return "", "", err
		}
		return strings.TrimSpace(body.Mode), strings.TrimSpace(body.Phone), nil
	}
	if err := r.ParseForm(); err != nil {
		return "", "", err
	}
	return strings.TrimSpace(r.FormValue("mode")), strings.TrimSpace(r.FormValue("phone")), nil
}

func parsePolicyRequest(r *http.Request) (policy.Policy, error) {
	if isJSONRequest(r) {
		var nextPolicy policy.Policy
		if err := json.NewDecoder(r.Body).Decode(&nextPolicy); err != nil {
			return policy.Policy{}, err
		}
		nextPolicy.Normalize()
		return nextPolicy, nil
	}
	if err := r.ParseForm(); err != nil {
		return policy.Policy{}, err
	}
	nextPolicy := policy.Policy{
		Safeties: policy.Safeties{
			Read:        checkboxOn(r.FormValue("allow_read")),
			Write:       checkboxOn(r.FormValue("allow_write")),
			Destructive: checkboxOn(r.FormValue("allow_destructive")),
			Paid:        checkboxOn(r.FormValue("allow_paid")),
		},
		PeerTypes: policy.PeerTypes{
			Users:    checkboxOn(r.FormValue("allow_users")),
			Groups:   checkboxOn(r.FormValue("allow_groups")),
			Channels: checkboxOn(r.FormValue("allow_channels")),
			Bots:     checkboxOn(r.FormValue("allow_bots")),
		},
		AllowPeers: policy.SplitPeerList(r.FormValue("allow_peers")),
		DenyPeers:  policy.SplitPeerList(r.FormValue("deny_peers")),
	}
	nextPolicy.Normalize()
	return nextPolicy, nil
}

func parseOptionalPolicyRequest(r *http.Request) (policy.Policy, bool, error) {
	if !isJSONRequest(r) {
		return policy.Policy{}, false, nil
	}

	var raw map[string]json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
		if errors.Is(err, io.EOF) {
			return policy.Policy{}, false, nil
		}
		return policy.Policy{}, false, err
	}

	if payload, ok := raw["policy"]; ok {
		var nextPolicy policy.Policy
		if err := json.Unmarshal(payload, &nextPolicy); err != nil {
			return policy.Policy{}, false, err
		}
		nextPolicy.Normalize()
		return nextPolicy, true, nil
	}

	if _, ok := raw["safeties"]; !ok {
		return policy.Policy{}, false, nil
	}
	data, err := json.Marshal(raw)
	if err != nil {
		return policy.Policy{}, false, err
	}
	var nextPolicy policy.Policy
	if err := json.Unmarshal(data, &nextPolicy); err != nil {
		return policy.Policy{}, false, err
	}
	nextPolicy.Normalize()
	return nextPolicy, true, nil
}

func parseAPISettingsRequest(r *http.Request) (int, string, error) {
	if isJSONRequest(r) {
		return parseAPISettingsJSON(r)
	}
	return parseAPISettingsForm(r)
}

func parseAPISettingsJSON(r *http.Request) (int, string, error) {
	var body struct {
		AppID      string `json:"appId"`
		AppHash    string `json:"appHash"`
		UseDefault bool   `json:"useDefault"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return 0, "", err
	}
	if body.UseDefault {
		return defaultAPISettings()
	}
	appID, err := config.ParseAppID(strings.TrimSpace(body.AppID))
	if err != nil {
		return 0, "", err
	}
	return appID, strings.TrimSpace(body.AppHash), nil
}

func parseAPISettingsForm(r *http.Request) (int, string, error) {
	if err := r.ParseForm(); err != nil {
		return 0, "", err
	}
	if checkboxOn(r.FormValue("use_default")) {
		return defaultAPISettings()
	}
	appID, err := config.ParseAppID(strings.TrimSpace(r.FormValue("app_id")))
	if err != nil {
		return 0, "", err
	}
	return appID, strings.TrimSpace(r.FormValue("app_hash")), nil
}

func defaultAPISettings() (int, string, error) {
	appID, err := config.ParseAppID(defaultAppID)
	if err != nil {
		return 0, "", err
	}
	return appID, defaultAppHash, nil
}

func loadWebPolicy() policy.Policy {
	p, err := policy.LoadDefault()
	if err != nil {
		return policy.Default()
	}
	return p
}

func checkboxOn(value string) bool {
	return value == "on" || value == "true" || value == "1"
}
