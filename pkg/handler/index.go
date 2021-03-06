package handler

import (
	"fmt"
	"net/http"

	"github.com/adrian83/chat/pkg/user"
	session "github.com/adrian83/go-redis-session"
)

// IndexHandler struct responsible for handling actions
// made on index html page.
type IndexHandler struct {
	sessionStore *session.Store
	templates    *TemplateRepository
}

// NewIndexHandler returns new IndexHandler struct.
func NewIndexHandler(templates *TemplateRepository, sessionStore *session.Store) *IndexHandler {
	return &IndexHandler{
		sessionStore: sessionStore,
		templates:    templates,
	}
}

// ShowIndexPage renders Index page.
func (h *IndexHandler) ShowIndexPage(w http.ResponseWriter, req *http.Request) {
	model := NewModel()

	if afterLogout(req) {
		model.AddInfo("You have been logged out.")
		RenderTemplateWithModel(w, h.templates.Index, model)
		return
	}

	sessionID, err := ReadSessionIDFromCookie(req)
	if err != nil {
		RenderTemplateWithModel(w, h.templates.Login, model)
		return
	}

	userSession, err := h.sessionStore.Find(sessionID)
	if err != nil {
		model.AddError(fmt.Sprintf("Cannot find user session: %v", err))
		RenderTemplateWithModel(w, h.templates.Login, model)
		return
	}

	var usr user.User
	if err = userSession.Get("user", &usr); err != nil {
		model.AddError(fmt.Sprintf("Cannot get data about user: %v", err))
		RenderTemplateWithModel(w, h.templates.Login, model)
		return
	}

	model.AddUser(&usr)

	RenderTemplateWithModel(w, h.templates.Index, model)
}

const (
	reason = "reason"
	logout = "logout"
)

func afterLogout(req *http.Request) bool {
	reason := req.URL.Query().Get(reason)
	return reason == logout
}
