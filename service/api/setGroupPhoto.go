package api

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (rt *_router) setGroupPhoto(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
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

	// Retrieve group ID from route parameters
	groupID := ps.ByName("id")

	// Save the photo (e.g., to a file storage service or database)
	if err := rt.db.SaveGroupPhoto(groupID, file); err != nil {
		http.Error(w, "Failed to save group photo", http.StatusInternalServerError)
		return
	}

	// Respond with success
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]bool{
		"success": true,
	})
}
