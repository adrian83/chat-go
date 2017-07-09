package session

import (
	"chat/db"

	redisSession "github.com/adrian83/go-redis-session"

	"net/http"
	"time"
)

const (
	sessionIDName      = "session_id"
	defSessionDuration = time.Duration(1000) * time.Minute
)

// FindSessionID returns session ID from cookie or empty string if such
// cookie doesn't exist, error if something bad has happened.
func FindSessionID(req *http.Request) string {
	c, err := req.Cookie(sessionIDName)
	if err != nil {
		return ""
	}
	return c.Value
}

// Session struct represents simplified session mechanism.
type Session struct {
	sessionStore redisSession.Store
}

// New returns pointer to new Session struct.
func New(sessionStore redisSession.Store) *Session {
	return &Session{sessionStore: sessionStore}
}

// FindUserData returns user data with given sessionID from session or empty
// struct if session doesn't contain such data, error if something
// bad has happened.
func (s *Session) FindUserData(sessionID string) (db.User, error) {
	session, err := s.sessionStore.FindSession(sessionID)
	if err != nil {
		return db.User{}, err
	}

	name, nameOk := session.Get("user.name")
	id, idOk := session.Get("user.id")
	if !nameOk || !idOk {
		return db.User{}, nil
	}

	return db.User{
		ID:   id,
		Name: name,
	}, nil
}

// StoreUserData saves user data to session and returns session ID, error
// if something bad has happened.
func (s *Session) StoreUserData(w http.ResponseWriter, user db.User) (string, error) {
	session, err1 := s.sessionStore.NewSession(defSessionDuration)
	if err1 != nil {
		return "", err1
	}
	session.Add("user.name", user.Name)
	session.Add("user.id", user.ID)
	if err2 := s.sessionStore.SaveSession(session); err2 != nil {
		return "", err2
	}

	cookie := &http.Cookie{
		Name:   sessionIDName,
		Value:  session.ID(),
		MaxAge: int(defSessionDuration.Seconds()),
	}

	http.SetCookie(w, cookie)

	return session.ID(), nil
}

// Remove removes the sessionID from cookie and session from database.
func (s *Session) Remove(w http.ResponseWriter, req *http.Request) error {
	c, err := req.Cookie(sessionIDName)
	if err != nil {
		return err
	}

	cookie := &http.Cookie{
		Name:   sessionIDName,
		Value:  "",
		MaxAge: -1,
	}
	http.SetCookie(w, cookie)

	return s.sessionStore.DeleteSession(c.Value)
}