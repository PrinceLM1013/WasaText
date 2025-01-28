package api

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (rt *_router) getMyConversations(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Retrieve user ID from context (set by middleware)
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "User ID not found in context", http.StatusUnauthorized)
		return
	}

	// Fetch conversations from the database
	conversations, err := rt.db.GetConversations(userID)
	if err != nil {
		http.Error(w, "Failed to retrieve conversations", http.StatusInternalServerError)
		return
	}

	// Respond with the list of conversations
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(conversations)
}
