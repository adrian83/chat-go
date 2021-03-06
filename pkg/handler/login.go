package handler

import (
	"fmt"
	"net/http"

	"github.com/adrian83/chat/pkg/user"
	session "github.com/adrian83/go-redis-session"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type userService interface {
	FindUser(string) (*user.User, error)
}

// LoginHandler struct responsible for handling actions
// made on login html page.
type LoginHandler struct {
	userService  userService
	sessionStore *session.Store
	templates    *TemplateRepository
}

// NewLoginHandler returns new LoginHandler struct.
func NewLoginHandler(templates *TemplateRepository, userService userService, sessionStore *session.Store) *LoginHandler {
	return &LoginHandler{
		userService:  userService,
		sessionStore: sessionStore,
		templates:    templates,
	}
}

// ShowLoginPage renders login html page.
func (h *LoginHandler) ShowLoginPage(w http.ResponseWriter, req *http.Request) {
	RenderTemplate(w, h.templates.Login)
}

// LoginUser processes user login form.
func (h *LoginHandler) LoginUser(w http.ResponseWriter, req *http.Request) {
	model := NewModel()

	if err := req.ParseForm(); err != nil {
		model.AddError(fmt.Sprintf("Cannot parse form: %v", err))
		RenderTemplateWithModel(w, h.templates.ServerError, model)
		return
	}

	username, password := h.validateLoginForm(req, model)

	if model.HasErrors() {
		RenderTemplateWithModel(w, h.templates.Login, model)
		return
	}

	usr, err := h.userService.FindUser(username)
	if err != nil {
		model.AddError(fmt.Sprintf("Cannot get data about user: %v", err))
		RenderTemplateWithModel(w, h.templates.Login, model)
		return
	}

	if usr.Empty() {
		model.AddError("User with this username doesn't exist")
		RenderTemplateWithModel(w, h.templates.Login, model)
		return
	}

	if err = bcrypt.CompareHashAndPassword([]byte(usr.Password), []byte(password)); err != nil {
		model.AddError(fmt.Sprintf("Passwords don't match: %v", err))
		RenderTemplateWithModel(w, h.templates.Login, model)
		return
	}

	if err = h.storeInSession(*usr, w); err != nil {
		model.AddError(fmt.Sprintf("Cannot create session: %v", err))
		RenderTemplateWithModel(w, h.templates.Login, model)
		return
	}

	http.Redirect(w, req, "/conversation", http.StatusFound)

}

func (h *LoginHandler) validateLoginForm(req *http.Request, model Model) (string, string) {
	username := req.FormValue("username")
	password := req.FormValue("password")

	if username == "" {
		model.AddError("Username cannot be empty")
	}

	if password == "" {
		model.AddError("Password cannot be empty")
	}

	return username, password
}

func (h *LoginHandler) storeInSession(usr user.User, w http.ResponseWriter) error {
	sessionID := uuid.New().String()

	sess, err := h.sessionStore.Create(sessionID)
	if err != nil {
		return err
	}

	if err := sess.Add("user", usr); err != nil {
		return err
	}

	if err := h.sessionStore.Save(sess); err != nil {
		return err
	}

	StoreSessionCookie(sessionID, w)

	return nil
}
