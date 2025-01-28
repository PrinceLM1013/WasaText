package api

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (rt *_router) sendMessage(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Parse request body
	var request struct {
		ConversationID string `json:"conversationId"`
		Content        string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate the request
	if request.ConversationID == "" || request.Content == "" {
		http.Error(w, "Conversation ID and content are required", http.StatusBadRequest)
		return
	}

	// Retrieve user ID from context (set by middleware)
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "User ID not found in context", http.StatusUnauthorized)
		return
	}

	// Save the message to the database
	messageID, err := rt.db.SaveMessage(request.ConversationID, userID, request.Content)
	if err != nil {
		http.Error(w, "Failed to send message", http.StatusInternalServerError)
		return
	}

	// Respond with success
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"messageID": messageID,
	})
}
