package web

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	log "github.com/sirupsen/logrus"
	"go.uber.org/atomic"
)

const defaultPort = 28086

var port = flag.Int("web.port", defaultPort, "port to listen on")

var IsMaster atomic.Bool

func StartServer() {
	log.Info(fmt.Sprintf("Starting on port %d...", *port))

	err := http.ListenAndServe(fmt.Sprintf(":%d", *port), GetHandler())
	if err != nil {
		log.WithError(err).Fatal()
	}
}

func GetHandler() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/ready", handlerReady)
	mux.HandleFunc("/healthz", handlerHealthz)
	mux.HandleFunc("/debug", handlerDebug)

	return mux
}

func handlerReady(w http.ResponseWriter, r *http.Request) {
	if IsMaster.Load() {
		_, _ = w.Write([]byte("is master"))
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)

		_, _ = w.Write([]byte("wait for master"))
	}
}

func handlerHealthz(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte("live"))
}

func handlerDebug(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte(os.Getenv("HOSTNAME")))
}
