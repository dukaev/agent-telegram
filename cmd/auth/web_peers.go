package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"agent-telegram/internal/authflow"
	"agent-telegram/internal/policy"
	tgapp "agent-telegram/telegram"
	tgtypes "agent-telegram/telegram/types"
)

func loadAuthPeers(ctx context.Context, state *authflow.State, sessionData []byte) ([]authPeer, error) {
	if len(sessionData) == 0 {
		return nil, fmt.Errorf("auth session is empty")
	}
	client := tgapp.NewClient(state.AppID, state.AppHash).WithSessionStorage(tgapp.NewMemoryStorage(sessionData))

	clientCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- client.Start(clientCtx)
	}()

	select {
	case <-client.Ready():
	case err := <-errCh:
		if err == nil {
			return nil, fmt.Errorf("telegram client stopped before it became ready")
		}
		return nil, err
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	result, err := client.Chat().GetChats(ctx, &tgtypes.GetChatsParams{Limit: 100})
	cancel()
	if err != nil {
		return nil, err
	}
	select {
	case err := <-errCh:
		if err != nil && !errors.Is(err, context.Canceled) {
			return nil, err
		}
	case <-time.After(2 * time.Second):
	}

	return authPeersFromChats(result.Chats), nil
}

func authPeersFromChats(chats []map[string]any) []authPeer {
	peers := make([]authPeer, 0, len(chats))
	for _, chat := range chats {
		peer := authPeerFromChat(chat)
		if peer.Peer == "" {
			continue
		}
		peers = append(peers, peer)
	}
	return peers
}

func authPeerFromChat(chat map[string]any) authPeer {
	rawType := stringValue(chat, "type")
	username := stringValue(chat, "username")
	title := stringValue(chat, "title")
	firstName := stringValue(chat, "first_name")
	lastName := stringValue(chat, "last_name")
	if title == "" {
		title = strings.TrimSpace(firstName + " " + lastName)
	}
	if title == "" && username != "" {
		title = "@" + username
	}

	var id int64
	var kind string
	var peer string
	switch rawType {
	case "user":
		id = int64Value(chat, "user_id")
		kind = "user"
		if boolValue(chat, "bot") {
			kind = "bot"
		}
		if id != 0 {
			peer = fmt.Sprintf("user:%d", id)
		}
	case "chat":
		id = int64Value(chat, "chat_id")
		kind = "group"
		if id != 0 {
			peer = fmt.Sprintf("chat:%d", id)
		}
	case "channel":
		id = int64Value(chat, "channel_id")
		kind = "channel"
		if boolValue(chat, "megagroup") {
			kind = "group"
		}
		if id != 0 {
			peer = fmt.Sprintf("channel:%d", id)
		}
	}

	if peer == "" {
		peer = stringValue(chat, "peer")
	}
	peer = policy.NormalizePeer(peer)
	if title == "" {
		title = peer
	}
	return authPeer{
		Peer:     peer,
		Title:    title,
		Username: username,
		Type:     kind,
		ID:       id,
	}
}

func stringValue(m map[string]any, key string) string {
	if value, ok := m[key].(string); ok {
		return value
	}
	return ""
}

func boolValue(m map[string]any, key string) bool {
	if value, ok := m[key].(bool); ok {
		return value
	}
	return false
}

func int64Value(m map[string]any, key string) int64 {
	switch value := m[key].(type) {
	case int64:
		return value
	case int:
		return int64(value)
	case float64:
		return int64(value)
	case json.Number:
		out, _ := value.Int64()
		return out
	default:
		return 0
	}
}
