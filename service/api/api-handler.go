package api

import (
	"net/http"
)

// Handler returns an instance of httprouter.Router that handle APIs registered here
func (rt *_router) Handler() http.Handler {
	// Register routes
	rt.router.GET("/", rt.getHelloWorld)
	rt.router.GET("/context", rt.wrap(rt.getContextReply))

	// Special routes
	rt.router.GET("/liveness", rt.liveness)

	// User routes
	rt.router.POST("/session", rt.wrap(rt.Dologin))
	rt.router.PUT("/users/me/name", rt.wrap(rt.setMyUserName))
	rt.router.PUT("/users/me/photo", rt.wrap(rt.setMyPhoto))

	// Conversation routes
	rt.router.GET("/conversations", rt.wrap(rt.getMyConversations))
	rt.router.GET("/conversations/:id", rt.wrap(rt.getConversation))

	// Message routes
	rt.router.POST("/messages", rt.wrap(rt.sendMessage))
	rt.router.POST("/messages/:id/forward", rt.wrap(rt.forwardMessage))
	rt.router.POST("/messages/:id/comment", rt.wrap(rt.commentMessage))
	rt.router.DELETE("/messages/:id/comment", rt.wrap(rt.uncommentMessage))

	// Group routes
	rt.router.POST("/groups/:id/leave", rt.wrap(rt.leaveGroup))
	rt.router.PUT("/groups/:id/photo", rt.wrap(rt.setGroupPhoto))

	return rt.router
}
