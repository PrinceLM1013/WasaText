package api

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (rt *_router) forwardMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Parse request body
	var request struct {
		ToConversationID string `json:"toConversationId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate the request
	if request.ToConversationID == "" {
		http.Error(w, "ToConversationID is required", http.StatusBadRequest)
		return
	}

	// Retrieve message ID from route parameters
	messageID := ps.ByName("id")

	// Retrieve user ID from context (set by middleware)
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "User ID not found in context", http.StatusUnauthorized)
		return
	}

	// Forward the message
	if err := rt.db.ForwardMessage(messageID, request.ToConversationID, userID); err != nil {
		http.Error(w, "Failed to forward message", http.StatusInternalServerError)
		return
	}

	// Respond with success
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]bool{
		"success": true,
	})
}
