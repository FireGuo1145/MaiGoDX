package handler

import (
	"encoding/json"
	"net/http"
)

// HandleCurrentUser returns the account bound to the active browser session.
func HandleCurrentUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	account, ok := requireAccount(w, r)
	if !ok {
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"email":    account.Email,
		"username": account.Username,
		"isAdmin":  account.IsAdmin,
	})
}
