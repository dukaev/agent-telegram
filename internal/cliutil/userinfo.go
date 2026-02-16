package cliutil

import (
	"strconv"
)

// GetUserInfo fetches user info and outputs as JSON.
// If identifier is empty, returns current user info (get_me).
// If identifier is numeric, looks up by user ID.
// Otherwise looks up by username.
func GetUserInfo(runner *Runner, identifier string) {
	var result any
	if identifier == "" {
		result = runner.Call("get_me", nil)
	} else if userID, err := strconv.ParseInt(identifier, 10, 64); err == nil && userID > 0 {
		result = runner.CallWithParams("get_user_info", map[string]any{
			"userId": userID,
		})
	} else {
		result = runner.CallWithParams("get_user_info", map[string]any{
			"username": identifier,
		})
	}
	runner.PrintResult(result, nil)
}
