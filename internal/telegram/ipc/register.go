// Package ipc provides Telegram IPC handlers registration.
package ipc

import (
	"encoding/json"

	"agent-telegram/internal/ipc"
)

// RegisterHandlers registers all Telegram IPC handlers.
func RegisterHandlers(srv ipc.MethodRegistrar, client Client) {
	srv.Register("get_me", func(params json.RawMessage) (interface{}, *ipc.ErrorObject) {
		result, err := GetMeHandler(client)(params)
		if err != nil {
			return nil, &ipc.ErrorObject{
				Code:    -32000,
				Message: err.Error(),
			}
		}
		return result, nil
	})

	srv.Register("get_chats", func(params json.RawMessage) (interface{}, *ipc.ErrorObject) {
		result, err := GetChatsHandler(client)(params)
		if err != nil {
			return nil, &ipc.ErrorObject{
				Code:    -32000,
				Message: err.Error(),
			}
		}
		return result, nil
	})

	srv.Register("get_updates", func(params json.RawMessage) (interface{}, *ipc.ErrorObject) {
		result, err := GetUpdatesHandler(client)(params)
		if err != nil {
			return nil, &ipc.ErrorObject{
				Code:    -32000,
				Message: err.Error(),
			}
		}
		return result, nil
	})
}
