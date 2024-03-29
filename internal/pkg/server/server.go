package server

import (
	"net/http"
	"strconv"

	"github.com/Nunoki/onetimesharer/internal/pkg/config"
)

type tplData struct {
	ErrorMsg  string
	ShareURL  string
	SecretKey string
}

type server struct {
	config config.Config
	store  Storer
}

type jsonOutput struct {
	Secret string `json:"secret"`
}

type Storer interface {
	ReadSecret(key string) (string, error)
	SaveSecret(secret string) (string, error)
	ValidateSecret(key string) (bool, error)
	Close() error
}

// New returns a new instance of the server.
func New(c config.Config, s Storer) server {
	server := server{
		config: c,
		store:  s,
	}
	return server
}

// Shutdown will attempt a graceful shutdown by calling the Close() method on its store. Any error
// returned by the store's Close() method is being returned.
func (serv server) Shutdown() error {
	return serv.store.Close()
}

// Serve registers the handlers for all the endpoints, and then initializes the listening on the
// port configured in its configuration instance. Any returned error will be from http's Listen-
// methods.
func (serv server) Serve() error {
	// TODO: test all endpoints
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			// this is because the "/" pattern of HandleFunc matches everything
			http.NotFound(w, r)
			return
		}

		if r.Method == "GET" {
			serv.handleIndex(w, r)
		} else if r.Method == "POST" {
			payloadLimit(*serv.config.PayloadLimit, func(w http.ResponseWriter, r *http.Request) {
				serv.handlePost(w, r, serv.store)
			})(w, r)
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}
	})

	// #show_url
	http.HandleFunc("/show", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			serv.handleShow(w, r, serv.store)
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}
	})

	http.HandleFunc("/secret", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			serv.handleFetchSecret(w, r, serv.store)
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}
	})

	portStr := strconv.Itoa(int(*serv.config.Port))
	if *serv.config.HTTPS {
		return http.ListenAndServeTLS(":"+portStr, *serv.config.Certfile, *serv.config.Keyfile, nil)
	} else {
		return http.ListenAndServe(":"+portStr, nil)
	}
}
