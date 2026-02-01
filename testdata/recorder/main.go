// Package main provides a utility to record Telegram API responses as fixtures.
//
// Usage:
//
//	go run ./testdata/recorder -method messages.getHistory -peer @username
//
// This requires an authenticated Telegram session.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
)

// FixtureMeta contains metadata about the recorded fixture.
type FixtureMeta struct {
	Method        string    `json:"method"`
	RecordedAt    time.Time `json:"recorded_at"`
	TelegramLayer int       `json:"telegram_layer"`
	Notes         string    `json:"notes,omitempty"`
	Sanitized     bool      `json:"sanitized"`
}

// Fixture represents a recorded API request/response pair.
type Fixture struct {
	Meta     FixtureMeta     `json:"meta"`
	Request  json.RawMessage `json:"request"`
	Response json.RawMessage `json:"response"`
}

var (
	method    = flag.String("method", "", "API method to record (e.g., messages.getHistory)")
	peer      = flag.String("peer", "", "Peer to use (username or ID)")
	outputDir = flag.String("output", "./testdata/fixtures", "Output directory")
	notes     = flag.String("notes", "", "Notes about this fixture")
	limit     = flag.Int("limit", 50, "Limit for list operations")
)

func main() {
	flag.Parse()

	if *method == "" {
		log.Fatal("method is required")
	}

	ctx := context.Background()

	// TODO: Initialize Telegram client with existing session
	// This is a template - actual implementation depends on your auth setup
	client, err := initClient(ctx)
	if err != nil {
		log.Fatalf("failed to init client: %v", err)
	}

	fixture, err := recordMethod(ctx, client, *method)
	if err != nil {
		log.Fatalf("failed to record: %v", err)
	}

	if err := saveFixture(fixture); err != nil {
		log.Fatalf("failed to save: %v", err)
	}

	fmt.Printf("Fixture saved to %s\n", getOutputPath(*method))
}

func initClient(ctx context.Context) (*tg.Client, error) {
	// TODO: Load session from your session storage
	// This depends on how agent-telegram stores sessions

	// Placeholder - replace with actual initialization
	return nil, fmt.Errorf("not implemented: configure session loading")
}

func recordMethod(ctx context.Context, api *tg.Client, method string) (*Fixture, error) {
	var request, response any
	var err error

	switch method {
	case "messages.getHistory":
		request, response, err = recordGetHistory(ctx, api)
	case "messages.getDialogs":
		request, response, err = recordGetDialogs(ctx, api)
	case "contacts.resolveUsername":
		request, response, err = recordResolveUsername(ctx, api)
	case "users.getFullUser":
		request, response, err = recordGetFullUser(ctx, api)
	case "channels.getParticipants":
		request, response, err = recordGetParticipants(ctx, api)
	default:
		return nil, fmt.Errorf("unknown method: %s", method)
	}

	if err != nil {
		return nil, err
	}

	reqJSON, err := json.MarshalIndent(request, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	respJSON, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal response: %w", err)
	}

	return &Fixture{
		Meta: FixtureMeta{
			Method:        method,
			RecordedAt:    time.Now().UTC(),
			TelegramLayer: tg.Layer,
			Notes:         *notes,
			Sanitized:     false,
		},
		Request:  reqJSON,
		Response: respJSON,
	}, nil
}

func recordGetHistory(ctx context.Context, api *tg.Client) (any, any, error) {
	if *peer == "" {
		return nil, nil, fmt.Errorf("peer is required for messages.getHistory")
	}

	inputPeer, err := resolvePeer(ctx, api, *peer)
	if err != nil {
		return nil, nil, err
	}

	req := &tg.MessagesGetHistoryRequest{
		Peer:  inputPeer,
		Limit: *limit,
	}

	resp, err := api.MessagesGetHistory(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	return req, resp, nil
}

func recordGetDialogs(ctx context.Context, api *tg.Client) (any, any, error) {
	req := &tg.MessagesGetDialogsRequest{
		OffsetPeer: &tg.InputPeerEmpty{},
		Limit:      *limit,
	}

	resp, err := api.MessagesGetDialogs(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	return req, resp, nil
}

func recordResolveUsername(ctx context.Context, api *tg.Client) (any, any, error) {
	if *peer == "" {
		return nil, nil, fmt.Errorf("peer (username) is required")
	}

	username := strings.TrimPrefix(*peer, "@")
	req := &tg.ContactsResolveUsernameRequest{
		Username: username,
	}

	resp, err := api.ContactsResolveUsername(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	return req, resp, nil
}

func recordGetFullUser(ctx context.Context, api *tg.Client) (any, any, error) {
	if *peer == "" {
		return nil, nil, fmt.Errorf("peer is required")
	}

	inputPeer, err := resolvePeer(ctx, api, *peer)
	if err != nil {
		return nil, nil, err
	}

	// Convert to InputUser
	var inputUser tg.InputUserClass
	switch p := inputPeer.(type) {
	case *tg.InputPeerUser:
		inputUser = &tg.InputUser{
			UserID:     p.UserID,
			AccessHash: p.AccessHash,
		}
	default:
		return nil, nil, fmt.Errorf("peer must be a user")
	}

	resp, err := api.UsersGetFullUser(ctx, inputUser)
	if err != nil {
		return nil, nil, err
	}

	return inputUser, resp, nil
}

func recordGetParticipants(ctx context.Context, api *tg.Client) (any, any, error) {
	if *peer == "" {
		return nil, nil, fmt.Errorf("peer (channel) is required")
	}

	inputPeer, err := resolvePeer(ctx, api, *peer)
	if err != nil {
		return nil, nil, err
	}

	// Convert to InputChannel
	var inputChannel *tg.InputChannel
	switch p := inputPeer.(type) {
	case *tg.InputPeerChannel:
		inputChannel = &tg.InputChannel{
			ChannelID:  p.ChannelID,
			AccessHash: p.AccessHash,
		}
	default:
		return nil, nil, fmt.Errorf("peer must be a channel")
	}

	req := &tg.ChannelsGetParticipantsRequest{
		Channel: inputChannel,
		Filter:  &tg.ChannelParticipantsRecent{},
		Limit:   *limit,
	}

	resp, err := api.ChannelsGetParticipants(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	return req, resp, nil
}

func resolvePeer(ctx context.Context, api *tg.Client, peerStr string) (tg.InputPeerClass, error) {
	username := strings.TrimPrefix(peerStr, "@")

	resolved, err := api.ContactsResolveUsername(ctx, &tg.ContactsResolveUsernameRequest{
		Username: username,
	})
	if err != nil {
		return nil, fmt.Errorf("resolve username: %w", err)
	}

	rp, ok := resolved.(*tg.ContactsResolvedPeer)
	if !ok {
		return nil, fmt.Errorf("unexpected response type")
	}

	return peerFromResolved(rp)
}

func peerFromResolved(rp *tg.ContactsResolvedPeer) (tg.InputPeerClass, error) {
	switch p := rp.Peer.(type) {
	case *tg.PeerUser:
		for _, u := range rp.Users {
			if user, ok := u.(*tg.User); ok && user.ID == p.UserID {
				return &tg.InputPeerUser{
					UserID:     user.ID,
					AccessHash: user.AccessHash,
				}, nil
			}
		}
	case *tg.PeerChannel:
		for _, c := range rp.Chats {
			if ch, ok := c.(*tg.Channel); ok && ch.ID == p.ChannelID {
				return &tg.InputPeerChannel{
					ChannelID:  ch.ID,
					AccessHash: ch.AccessHash,
				}, nil
			}
		}
	case *tg.PeerChat:
		return &tg.InputPeerChat{ChatID: p.ChatID}, nil
	}

	return nil, fmt.Errorf("could not extract peer")
}

func saveFixture(f *Fixture) error {
	path := getOutputPath(*method)

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0o644)
}

func getOutputPath(method string) string {
	// messages.getHistory -> messages/get_history.json
	parts := strings.Split(method, ".")
	if len(parts) != 2 {
		return filepath.Join(*outputDir, method+".json")
	}

	category := parts[0]
	name := toSnakeCase(parts[1])

	if *notes != "" {
		name = name + "_" + toSnakeCase(*notes)
	}

	return filepath.Join(*outputDir, category, name+".json")
}

func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}
