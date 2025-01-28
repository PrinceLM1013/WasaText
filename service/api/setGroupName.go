package api

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (rt *_router) setGroupName(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Parse request body
	var request struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate the request
	if request.Name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}

	// Retrieve group ID from route parameters
	groupID := ps.ByName("id")

	// Update the group name
	if err := rt.db.UpdateGroupName(groupID, request.Name); err != nil {
		http.Error(w, "Failed to update group name", http.StatusInternalServerError)
		return
	}

	// Respond with success
	w.Header().Set