// Package policy stores and enforces local operation permissions.
package policy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"agent-telegram/internal/ipc"
	"agent-telegram/internal/operations"
	"agent-telegram/internal/paths"
)

const (
	version = 1

	PeerTypeUser    = "user"
	PeerTypeGroup   = "group"
	PeerTypeChannel = "channel"
	PeerTypeBot     = "bot"
	PeerTypeUnknown = "unknown"
)

// Safeties controls which operation safety classes are allowed.
type Safeties struct {
	Read        bool `json:"read"`
	Write       bool `json:"write"`
	Destructive bool `json:"destructive"`
	Paid        bool `json:"paid"`
}

// PeerTypes controls broad dialog classes.
type PeerTypes struct {
	Users    bool `json:"users"`
	Groups   bool `json:"groups"`
	Channels bool `json:"channels"`
	Bots     bool `json:"bots"`
}

// Policy is persisted in ~/.agent-telegram/policy.json.
type Policy struct {
	Version    int       `json:"version"`
	Safeties   Safeties  `json:"safeties"`
	PeerTypes  PeerTypes `json:"peerTypes"`
	AllowPeers []string  `json:"allowPeers,omitempty"`
	DenyPeers  []string  `json:"denyPeers,omitempty"`
}

// PeerResolver can normalize a peer using the authenticated Telegram client.
type PeerResolver interface {
	ResolvePeerID(ctx context.Context, peer string) (string, error)
}

// Enforcer checks raw RPC requests against a policy.
type Enforcer struct {
	policy   Policy
	resolver PeerResolver
}

// Default returns the conservative default local policy.
func Default() Policy {
	return Policy{
		Version: version,
		Safeties: Safeties{
			Read:  true,
			Write: true,
		},
		PeerTypes: PeerTypes{
			Users:    true,
			Groups:   true,
			Channels: true,
			Bots:     true,
		},
	}
}

// DefaultPath returns the default policy path.
func DefaultPath() (string, error) {
	dir, err := paths.EnsureConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "policy.json"), nil
}

// LoadDefault loads the default policy file, returning defaults if it is absent.
func LoadDefault() (Policy, error) {
	path, err := DefaultPath()
	if err != nil {
		return Policy{}, err
	}
	return Load(path)
}

// Load reads a policy from path.
func Load(path string) (Policy, error) {
	p := Default()
	//nolint:gosec // path is a caller-selected local policy file path.
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return p, nil
		}
		return Policy{}, fmt.Errorf("read policy: %w", err)
	}
	if err := json.Unmarshal(data, &p); err != nil {
		return Policy{}, fmt.Errorf("parse policy: %w", err)
	}
	p.Normalize()
	return p, nil
}

// SaveDefault writes policy to the default path.
func SaveDefault(p Policy) error {
	path, err := DefaultPath()
	if err != nil {
		return err
	}
	return Save(path, p)
}

// Save writes policy to path with owner-only permissions.
func Save(path string, p Policy) error {
	p.Normalize()
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return fmt.Errorf("create policy dir: %w", err)
	}
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal policy: %w", err)
	}
	return os.WriteFile(path, data, 0600)
}

// Normalize canonicalizes peer lists and fills the policy version.
func (p *Policy) Normalize() {
	if p.Version == 0 {
		p.Version = version
	}
	p.AllowPeers = normalizePeerList(p.AllowPeers)
	p.DenyPeers = normalizePeerList(p.DenyPeers)
}

// NewEnforcer creates a policy enforcer.
func NewEnforcer(p Policy, resolver PeerResolver) *Enforcer {
	p.Normalize()
	return &Enforcer{policy: p, resolver: resolver}
}

// Policy returns the normalized policy.
func (e *Enforcer) Policy() Policy {
	return e.policy
}

// Check validates method and params against the local policy.
func (e *Enforcer) Check(ctx context.Context, method string, params json.RawMessage) error {
	if e == nil {
		return nil
	}
	op, ok := operations.Get(method)
	if !ok {
		return ipc.NewPolicyDeniedError(method, "operation is not registered")
	}
	if !e.safetyAllowed(op.Safety) {
		return ipc.NewPolicyDeniedError(method, op.Safety+" operations are disabled")
	}
	if op.RequiresConfirmation && !ipc.IsConfirmed(ctx) {
		return ipc.NewPolicyDeniedError(method, "explicit confirmation is required")
	}

	peers := ExtractPeers(params)
	if len(peers) == 0 {
		return nil
	}

	for _, peer := range peers {
		if err := e.checkPeer(ctx, method, peer); err != nil {
			return err
		}
	}
	return nil
}

func (e *Enforcer) safetyAllowed(safety string) bool {
	switch safety {
	case "", operations.SafetyRead:
		return e.policy.Safeties.Read
	case operations.SafetyWrite:
		return e.policy.Safeties.Write
	case operations.SafetyDestructive:
		return e.policy.Safeties.Destructive
	case operations.SafetyPaid:
		return e.policy.Safeties.Paid
	default:
		return false
	}
}

func (e *Enforcer) checkPeer(ctx context.Context, method, peer string) error {
	normalized := NormalizePeer(peer)
	if normalized == "" {
		return nil
	}
	if containsPeer(e.policy.DenyPeers, normalized) {
		return ipc.NewPolicyDeniedError(method, "peer "+normalized+" is denied")
	}
	explicitlyAllowed := containsPeer(e.policy.AllowPeers, normalized)
	if len(e.policy.AllowPeers) > 0 && !explicitlyAllowed {
		resolved := e.resolvePeer(ctx, normalized)
		explicitlyAllowed = resolved != "" && containsPeer(e.policy.AllowPeers, resolved)
		if !explicitlyAllowed {
			return ipc.NewPolicyDeniedError(method, "peer "+normalized+" is not in allow list")
		}
	}
	if explicitlyAllowed {
		return nil
	}

	kind := inferPeerType(normalized)
	if kind == PeerTypeUnknown {
		if resolved := e.resolvePeer(ctx, normalized); resolved != "" {
			kind = inferPeerType(resolved)
		}
	}
	if !e.peerTypeAllowed(kind) {
		return ipc.NewPolicyDeniedError(method, "peer type "+kind+" is disabled for "+normalized)
	}
	return nil
}

func (e *Enforcer) resolvePeer(ctx context.Context, peer string) string {
	if e.resolver == nil {
		return ""
	}
	resolved, err := e.resolver.ResolvePeerID(ctx, peer)
	if err != nil {
		return ""
	}
	return NormalizePeer(resolved)
}

func (e *Enforcer) peerTypeAllowed(kind string) bool {
	switch kind {
	case PeerTypeUser:
		return e.policy.PeerTypes.Users
	case PeerTypeGroup:
		return e.policy.PeerTypes.Groups
	case PeerTypeChannel:
		return e.policy.PeerTypes.Channels
	case PeerTypeBot:
		return e.policy.PeerTypes.Bots
	default:
		return e.policy.PeerTypes.Users &&
			e.policy.PeerTypes.Groups &&
			e.policy.PeerTypes.Channels &&
			e.policy.PeerTypes.Bots
	}
}

// ExtractPeers extracts peer-like fields from a JSON params object.
func ExtractPeers(raw json.RawMessage) []string {
	if len(raw) == 0 {
		return nil
	}
	var m map[string]any
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	if err := decoder.Decode(&m); err != nil {
		return nil
	}
	values := []string{}
	for _, key := range []string{
		"peer", "username", "fromPeer", "toPeer", "channel", "user",
	} {
		values = append(values, valueStrings(m[key])...)
	}
	for _, key := range []string{
		"members", "includedChats", "excludedChats", "include", "exclude",
	} {
		values = append(values, valueStrings(m[key])...)
	}
	return normalizePeerList(values)
}

func valueStrings(value any) []string {
	switch v := value.(type) {
	case string:
		if v == "" {
			return nil
		}
		return []string{v}
	case float64:
		return []string{strconv.FormatInt(int64(v), 10)}
	case json.Number:
		return []string{v.String()}
	case []any:
		out := make([]string, 0, len(v))
		for _, item := range v {
			out = append(out, valueStrings(item)...)
		}
		return out
	default:
		return nil
	}
}

// SplitPeerList parses comma, newline, or whitespace separated peer lists.
func SplitPeerList(value string) []string {
	fields := strings.FieldsFunc(value, func(r rune) bool {
		return r == ',' || r == '\n' || r == '\r' || r == '\t' || r == ' '
	})
	return normalizePeerList(fields)
}

// JoinPeerList formats peers for textarea display.
func JoinPeerList(peers []string) string {
	return strings.Join(normalizePeerList(peers), "\n")
}

// NormalizePeer canonicalizes common peer formats.
func NormalizePeer(peer string) string {
	value := strings.TrimSpace(peer)
	if value == "" {
		return ""
	}
	value = strings.TrimSuffix(value, "/")
	if parsed, err := url.Parse(value); err == nil && parsed.Host != "" {
		host := strings.TrimPrefix(strings.ToLower(parsed.Host), "www.")
		if host == "t.me" || host == "telegram.me" {
			part := strings.Trim(parsed.Path, "/")
			if part != "" && !strings.HasPrefix(part, "+") {
				value = "@" + strings.Split(part, "/")[0]
			}
		}
	}

	lower := strings.ToLower(value)
	switch {
	case lower == "me" || lower == "self" || lower == "current_user":
		return lower
	case strings.HasPrefix(lower, "@"):
		return lower
	case strings.HasPrefix(lower, "user:") ||
		strings.HasPrefix(lower, "chat:") ||
		strings.HasPrefix(lower, "group:") ||
		strings.HasPrefix(lower, "channel:") ||
		strings.HasPrefix(lower, "bot:"):
		return lower
	case isNumeric(lower):
		return lower
	case strings.ContainsAny(lower, "/:"):
		return lower
	default:
		return "@" + lower
	}
}

func normalizePeerList(peers []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(peers))
	for _, peer := range peers {
		normalized := NormalizePeer(peer)
		if normalized == "" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		out = append(out, normalized)
	}
	return out
}

func containsPeer(peers []string, peer string) bool {
	for _, item := range peers {
		if item == peer {
			return true
		}
	}
	return false
}

func inferPeerType(peer string) string {
	switch {
	case peer == "me" || peer == "self" || peer == "current_user":
		return PeerTypeUser
	case strings.HasPrefix(peer, "bot:"):
		return PeerTypeBot
	case strings.HasPrefix(peer, "user:"):
		return PeerTypeUser
	case strings.HasPrefix(peer, "chat:") || strings.HasPrefix(peer, "group:"):
		return PeerTypeGroup
	case strings.HasPrefix(peer, "channel:") || strings.HasPrefix(peer, "-100"):
		return PeerTypeChannel
	case strings.HasPrefix(peer, "-"):
		return PeerTypeGroup
	case strings.HasPrefix(peer, "@") && strings.HasSuffix(peer, "bot"):
		return PeerTypeBot
	default:
		return PeerTypeUnknown
	}
}

func isNumeric(value string) bool {
	if value == "" {
		return false
	}
	if value[0] == '-' || value[0] == '+' {
		value = value[1:]
	}
	if value == "" {
		return false
	}
	for _, r := range value {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
