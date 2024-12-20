package api

import (
	"encoding/json"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/PrinceLM1013/WasaText/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()
var write_err = "error writing response"

// User login
func (rt *_router) Dologin(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	var user models.Username
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&user); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, err := w.Write([]byte(err.Error()))
		if err != nil {
			ctx.Logger.WithError(err).Error(write_err)
		}
		return
	}
	validationErr := validate.Struct(user)
	if validationErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, err := w.Write([]byte(validationErr.Error()))
		if err != nil {
			ctx.Logger.WithError(err).Error(write_err)
		}
		return
	}
	rt.db.CreateUser(user.Username, w, ctx)
}

// Update user name
func (rt *_router) setMyUserName(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	var request struct {
		Name string `json:"name" validate:"required,min=3,max=16"`
	}
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, err := w.Write([]byte(err.Error()))
		if err != nil {
			ctx.Logger.WithError(err).Error(write_err)
		}
		return
	}
	validationErr := validate.Struct(request)
	if validationErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, err := w.Write([]byte(validationErr.Error()))
		if err != nil {
			ctx.Logger.WithError(err).Error(write_err)
		}
		return
	}
	token := r.Header.Get("Authorization")
	is_valid := rt.db.Authorize("", token, w, ctx)
	if is_valid {
		rt.db.UpdateUserName(request.Name, w, ctx)
	}
}

// Update profile photo
func (rt *_router) setMyPhoto(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	err := r.ParseMultipartForm(10 << 20) // 10 MB limit
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, err := w.Write([]byte(`{"error": "Unable to parse form data"}`))
		if err != nil {
			ctx.Logger.WithError(err).Error(write_err)
		}
		return
	}
	file, handler, err := r.FormFile("photo")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, err := w.Write([]byte(`{"error": "Missing or invalid photo file in the request"}`))
		if err != nil {
			ctx.Logger.WithError(err).Error(write_err)
		}
		return
	}
	defer file.Close()
	allowedExtensions := []string{"png", "jpg", "jpeg"}
	ext := filepath.Ext(handler.Filename)
	ext = strings.ToLower(ext[1:])
	isValidExtension := false
	for _, allowedExt := range allowedExtensions {
		if ext == allowedExt {
			isValidExtension = true
			break
		}
	}
	if !isValidExtension {
		w.WriteHeader(http.StatusUnsupportedMediaType)
		_, err := w.Write([]byte(`{"error": "Only PNG, JPG, and JPEG images are allowed"}`))
		if err != nil {
			ctx.Logger.WithError(err).Error(write_err)
		}
		return
	}
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, err := w.Write([]byte(`{"error": "Failed to read the file", "details": "` + err.Error() + `"}`))
		if err != nil {
			ctx.Logger.WithError(err).Error(write_err)
		}
		return
	}
	token := r.Header.Get("Authorization")
	is_valid := rt.db.Authorize("", token, w, ctx)
	if is_valid {
		rt.db.SetUserProfilePhoto(fileBytes, w, ctx)
	}
}

// Retrieve all conversations
func (rt *_router) getMyConversations(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	username := ps.ByName("username")
	token := r.Header.Get("Authorization")
	is_valid := rt.db.Authorize(username, token, w, ctx)
	if is_valid {
		rt.db.GetConversations(username, w, ctx)
	}
}

// getConversation handles GET requests to retrieve a conversation by ID
func (rt *_router) getConversation(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	// Get the conversation ID from the URL parameters
	conversationID := ps.ByName("id")

	// Validate the conversation ID
	if conversationID == "" {
		w.WriteHeader(http.StatusBadRequest)
		_, err := w.Write([]byte("Conversation ID is required"))
		if err != nil {
			ctx.Logger.WithError(err).Error("Error writing response")
		}
		return
	}

	// Authorize the request
	token := r.Header.Get("Authorization")
	if !rt.db.Authorize(conversationID, token, w, ctx) {
		return
	}

	// Retrieve the conversation from the database
	conversation, err := rt.db.GetConversation(conversationID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, err := w.Write([]byte(err.Error()))
		if err != nil {
			ctx.Logger.WithError(err).Error("Error writing response")
		}
		return
	}

	// Write the conversation to the response
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(conversation)
	if err != nil {
		ctx.Logger.WithError(err).Error("Error encoding response")
	}
}

// Leave a group
func (rt *_router) leaveGroup(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	groupID := ps.ByName("id")
	token := r.Header.Get("Authorization")
	is_valid := rt.db.Authorize(groupID, token, w, ctx)
	if is_valid {
		rt.db.LeaveGroup(groupID, w, ctx)
	}
}

// Send a new message
func (rt *_router) sendMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	var message models.Message
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&message); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, err := w.Write([]byte(err.Error()))
		if err != nil {
			ctx.Logger.WithError(err).Error(write_err)
		}
		return
	}
	validationErr := validate.Struct(message)
	if validationErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, err := w.Write([]byte(validationErr.Error()))
		if err != nil {
			ctx.Logger.WithError(err).Error(write_err)
		}
		return
	}
	token := r.Header.Get("Authorization")
	is_valid := rt.db.Authorize(message.ConversationID, token, w, ctx)
	if is_valid {
		rt.db.SendMessage(message, w, ctx)
	}
}

// Forward a message
func (rt *_router) forwardMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	messageID := ps.ByName("id")
	var forward struct {
		ToConversationID string `json:"toConversationID" validate:"required"`
	}
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&forward); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, err := w.Write([]byte(err.Error()))
		if err != nil {
			ctx.Logger.WithError(err).Error(write_err)
		}
		return
	}
	validationErr := validate.Struct(forward)
	if validationErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, err := w.Write([]byte(validationErr.Error()))
		if err != nil {
			ctx.Logger.WithError(err).Error(write_err)
		}
		return
	}
	token := r.Header.Get("Authorization")
	is_valid := rt.db.Authorize(messageID, token, w, ctx)
	if is_valid {
		rt.db.ForwardMessage(messageID, forward.ToConversationID, w, ctx)
	}
}

// React to a message
func (rt *_router) commentMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	messageID := ps.ByName("id")
	var reaction struct {
		Emoji string `json:"emoji" validate:"required"`
	}
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&reaction); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, err := w.Write([]byte(err.Error()))
		if err != nil {
			ctx.Logger.WithError(err).Error(write_err)
		}
		return
	}
	validationErr := validate.Struct(reaction)
	if validationErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, err := w.Write([]byte(validationErr.Error()))
		if err != nil {
			ctx.Logger.WithError(err).Error(write_err)
		}
		return
	}
	token := r.Header.Get("Authorization")
	is_valid := rt.db.Authorize(messageID, token, w, ctx)
	if is_valid {
		rt.db.CommentMessage(messageID, reaction.Emoji, w, ctx)
	}
}

// Remove a reaction
func (rt *_router) uncommentMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	messageID := ps.ByName("id")
	token := r.Header.Get("Authorization")
	is_valid := rt.db.Authorize(messageID, token, w, ctx)
	if is_valid {
		rt.db.UncommentMessage(messageID, w, ctx)
	}
}

// Update group photo
func (rt *_router) setGroupPhoto(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	groupID := ps.ByName("id")
	err := r.ParseMultipartForm(10 << 20) // 10 MB limit
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, err := w.Write([]byte(`{"error": "Unable to parse form data"}`))
		if err != nil {
			ctx.Logger.WithError(err).Error(write_err)
		}
		return
	}
	file, handler, err := r.FormFile("photo")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, err := w.Write([]byte(`{"error": "Missing or invalid photo file in the request"}`))
		if err != nil {
			ctx.Logger.WithError(err).Error(write_err)
		}
		return
	}
	defer file.Close()
	allowedExtensions := []string{"png", "jpg", "jpeg"}
	ext := filepath.Ext(handler.Filename)
	ext = strings.ToLower(ext[1:])
	isValidExtension := false
	for _, allowedExt := range allowedExtensions {
		if ext == allowedExt {
			isValidExtension = true
			break
		}
	}
	if !isValidExtension {
		w.WriteHeader(http.StatusUnsupportedMediaType)
		_, err := w.Write([]byte(`{"error": "Only PNG, JPG, and JPEG images are allowed"}`))
		if err != nil {
			ctx.Logger.WithError(err).Error(write_err)
		}
		return
	}
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, err := w.Write([]byte(`{"error": "Failed to read the file", "details": "` + err.Error() + `"}`))
		if err != nil {
			ctx.Logger.WithError(err).Error(write_err)
		}
		return
	}
	token := r.Header.Get("Authorization")
	is_valid := rt.db.Authorize(groupID, token, w, ctx)
	if is_valid {
		rt.db.SetGroupPhoto(groupID, fileBytes, w, ctx)
	}
}
