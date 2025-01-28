package api

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (rt *_router) setMyPhoto(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Parse the multipart form
	if err := r.ParseMultipartForm(10 << 20); err != nil { // 10 MB limit
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	// Retrieve the photo file
	file, _, err := r.FormFile("photo")
	if err != nil {
		http.Error(w, "Photo is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Retrieve user ID from context (set by middleware)
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "User ID not found in context", http.StatusUnauthorized)
		return
	}

	// Save the photo (e.g., to a file storage service or database)
	if err := rt.db.SaveUserPhoto(userID, file); err != nil {
		http.Error(w, "Failed to save photo", http.StatusInternalServerError)
		return
	}

	// Respond with success
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]bool{
		"success": true,
	})
}
