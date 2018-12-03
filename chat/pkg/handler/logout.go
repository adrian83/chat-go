package handler

import (
	"fmt"
	"net/http"

	"github.com/adrian83/go-redis-session"
)

// LogoutHandler struct responsible for handling logout action.
type LogoutHandler struct {
	sessionStore session.Store
}

// NewLogoutHandler returns new LogoutHandler struct.
func NewLogoutHandler(sessionStore session.Store) *LogoutHandler {
	return &LogoutHandler{
		sessionStore: sessionStore,
	}
}

// Logout processes logout action.
func (h *LogoutHandler) Logout(w http.ResponseWriter, req *http.Request) {

	model := NewModel()

	sessionID, err := ReadSessionIDFromCookie(req)
	if err != nil {
		model.AddError(fmt.Sprintf("Cannot find session id: %v", err))
		RenderTemplateWithModel(w, loginTmpl, model)
		return
	}

	if err := h.sessionStore.Delete(sessionID); err != nil {
		model.AddError(fmt.Sprintf("Cannot remove user session: %v", err))
		RenderTemplateWithModel(w, loginTmpl, model)
		return
	}

	http.Redirect(w, req, "/?reason=logout", http.StatusFound)
}