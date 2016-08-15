package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/matrix-org/go-neb/clients"
	"github.com/matrix-org/go-neb/database"
	_ "github.com/matrix-org/go-neb/realms/github"
	_ "github.com/matrix-org/go-neb/realms/jira"
	"github.com/matrix-org/go-neb/server"
	_ "github.com/matrix-org/go-neb/services/echo"
	_ "github.com/matrix-org/go-neb/services/github"
	_ "github.com/matrix-org/go-neb/services/jira"
	"github.com/matrix-org/go-neb/types"
	_ "github.com/mattn/go-sqlite3"
	"net/http"
	_ "net/http/pprof"
	"os"
)

func main() {
	bindAddress := os.Getenv("BIND_ADDRESS")
	databaseType := os.Getenv("DATABASE_TYPE")
	databaseURL := os.Getenv("DATABASE_URL")
	baseURL := os.Getenv("BASE_URL")

	err := types.BaseURL(baseURL)
	if err != nil {
		log.Panic(err)
	}

	db, err := database.Open(databaseType, databaseURL)
	if err != nil {
		log.Panic(err)
	}
	database.SetServiceDB(db)

	clients := clients.New(db)
	if err := clients.Start(); err != nil {
		log.Panic(err)
	}

	http.Handle("/test", server.MakeJSONAPI(&heartbeatHandler{}))
	http.Handle("/admin/getService", server.MakeJSONAPI(&getServiceHandler{db: db}))
	http.Handle("/admin/configureClient", server.MakeJSONAPI(&configureClientHandler{db: db, clients: clients}))
	http.Handle("/admin/configureService", server.MakeJSONAPI(&configureServiceHandler{db: db, clients: clients}))
	http.Handle("/admin/configureAuthRealm", server.MakeJSONAPI(&configureAuthRealmHandler{db: db}))
	http.Handle("/admin/requestAuthSession", server.MakeJSONAPI(&requestAuthSessionHandler{db: db}))
	wh := &webhookHandler{db: db, clients: clients}
	http.HandleFunc("/services/hooks/", wh.handle)
	rh := &realmRedirectHandler{db: db}
	http.HandleFunc("/realms/redirects/", rh.handle)

	http.ListenAndServe(bindAddress, nil)
}
