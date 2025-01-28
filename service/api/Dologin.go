package api

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (rt *_router) Dologin(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Parse request body
	var user struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate user ID and name
	if user.ID == "" || user.Name == "" {
		http.Error(w, "ID and name are required", http.StatusBadRequest)
		return
	}

	// Check if the user exists or create a new one
	userID, err := rt.db.GetOrCreateUser(user.ID, user.Name)
	if err != nil {
		http.Error(w, "Failed to create or retrieve user", http.StatusInternalServerError)
		return
	}

	// Respond with the user ID
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"identifier": userID,
	})
}
