package router

import (
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/pbaettig/raspi-dash/assets"
	"github.com/pbaettig/raspi-dash/auth"
	"github.com/pbaettig/raspi-dash/config"
)

var (
	allowedUsers *auth.BasicAuthUsers = new(auth.BasicAuthUsers)
)

func New() *mux.Router {
	allowedUsers.PopulateFromEnv()

	r := mux.NewRouter()
	r.HandleFunc("/plot/{name}", plotHandler)
	r.PathPrefix("/assets").Handler(http.StripPrefix("/assets", http.FileServer(http.FS(assets.FS))))
	r.HandleFunc("/", indexHandler)
	r.HandleFunc("/api/docs/{id}", docByIdHandler)

	private := r.PathPrefix("/private").Subrouter()
	private.Path("/docs/id/{id}").HandlerFunc(docByIdHandler)
	private.Use(allowedUsers.Middleware)

	docs := r.PathPrefix("/documents").Subrouter()
	docs.NewRoute().Handler(http.StripPrefix("/documents", http.FileServer(
		http.FS(os.DirFS(config.DocumentsPath)))))
	docs.Use(allowedUsers.Middleware)

	return r
}
