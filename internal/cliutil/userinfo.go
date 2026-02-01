package cliutil

import (
	"encoding/json"
	"os"
)

// GetUserInfo fetches user info and outputs as JSON.
// If username is empty, returns current user info (get_me).
// Otherwise returns info about the specified user.
func GetUserInfo(runner *Runner, username string) {
	var result any
	if username == "" {
		result = runner.Call("get_me", nil)
	} else {
		result = runner.CallWithParams("get_user_info", map[string]any{
			"username": username,
		})
	}
	//nolint:errchkjson // Output to stdout, error handling not required
	_ = json.NewEncoder(os.Stdout).Encode(result)
}
