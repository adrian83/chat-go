package main

import (
	"net/http"

	"db"
	"logger"

	handler "handler"
	gsession "session"

	//gws "websocket"
	//gwsclient "websocket/client"
	//gwsconnection "websocket/connection"
	//gwsserver "websocket/server"

	redisSession "github.com/adrian83/go-redis-session"
	"github.com/gorilla/mux"
)

func main() {

	rethinkConfig := &db.Config{
		Host:             "localhost",
		Port:             28016,
		DBName:           "chat_go",
		UsersTableName:   "users",
		UsersTablePKName: "name",
	}

	database, err := db.New(rethinkConfig)
	if err != nil {
		logger.Errorf("Main", "main", "Error while creating RethinkDB session! Error: %v", err)
		panic(err)
	}
	defer func() {
		if err1 := database.Close(); err1 != nil {
			logger.Errorf("Main", "main", "Error while closing RethinkDB session! Error: %v", err1)
		}
	}()

	logger.Info("Main", "main", "RethinkDB session created")

	if err = database.Setup(); err != nil {
		logger.Errorf("Main", "main", "Error during RethinkDB database setup! Error: %v", err)
		panic(err)
	}
	logger.Info("Main", "main", "RethinkDB database initialized")

	sessionStoreConfig := redisSession.Config{
		DB:       0,
		Password: "",
		Host:     "localhost",
		Port:     6380,
		IDLength: 50,
	}

	sessionStore, err := redisSession.NewStore(sessionStoreConfig)
	if err != nil {
		logger.Errorf("Main", "main", "Error while creating SessionStore. Error: %v", err)
		return
	}
	defer func() {
		if err1 := sessionStore.Close(); err1 != nil {
			logger.Errorf("Main", "main", "Error while closing SessionStore session! Error: %v", err1)
		}
	}()
	logger.Info("Main", "main", "SessionStore created.")

	simpleSession := gsession.New(sessionStore)

	userRepository := db.NewUserRepository(database)

	loginHandler := handler.NewLoginHandler(userRepository, simpleSession)
	logoutHandler := handler.NewLogoutHandler(simpleSession)
	registerHandler := handler.NewRegisterHandler(userRepository)
	indexHandler := handler.NewIndexHandler(simpleSession)
	/*
		wsServer := gwsserver.New()
		wsServer.Start()
	*/
	mux := mux.NewRouter()
	mux.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static/"))))

	mux.HandleFunc("/", indexHandler.ShowIndexPage)

	mux.HandleFunc("/login", loginHandler.ShowLoginPage).Methods("GET")
	mux.HandleFunc("/login", loginHandler.LoginUser).Methods("POST")

	mux.HandleFunc("/logout", logoutHandler.Logout).Methods("GET")

	mux.HandleFunc("/register", registerHandler.ShowRegisterPage).Methods("GET")
	mux.HandleFunc("/register", registerHandler.RegisterUser).Methods("POST")
	/*
		mux.Handle("/talk", ws.Handler(connect(sessionStore, wsServer, simpleSession)))
	*/
	server := &http.Server{Addr: "0.0.0.0:7070", Handler: mux}
	if err2 := server.ListenAndServe(); err2 != nil {
		logger.Errorf("Main", "main", "Error while starting server! Error: %v", err2)
	}

}

/*
func connect(sessionStore redisSession.Store, wsServer gws.Server, simpleSession *gsession.Session) func(*ws.Conn) {
	return func(wsc *ws.Conn) {

		sessionID := gsession.FindSessionID(wsc.Request())
		if sessionID == "" {
			logger.Errorf("Main", "Connect", "Error while getting sessionID from WebSocket.")
			return
		}

		user, err2 := simpleSession.FindUserData(sessionID)
		if err2 != nil {
			logger.Errorf("Main", "Connect", "Error while getting user data from session. Error: %v", err2)
			return
		}

		conn := gwsconnection.WsConnection{Conn: wsc}
		client := gwsclient.New(sessionID, user.Name, conn, wsServer)

		logger.Infof("Main", "Connect", "New connection received from %v - %v", client)

		wsServer.RegisterClient(client)

		client.Start()

	}
}
*/
