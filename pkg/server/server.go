package server

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/moolen/logistis/pkg/recorder"
	"github.com/sirupsen/logrus"
)

type Server struct {
	srv        *http.Server
	listenAddr string
	certFile   string
	keyFile    string
}

func New(rec *recorder.Recorder, listenAddr, certFile, keyFile string) *Server {
	r := mux.NewRouter()
	r.HandleFunc("/healthz", ServeHealth)
	r.HandleFunc("/readyz", ServeReady)
	r.HandleFunc("/events", rec.ListEvents)
	r.HandleFunc("/capture", rec.RecordEvents)

	srv := &http.Server{Addr: listenAddr, Handler: r}
	return &Server{
		srv:        srv,
		listenAddr: listenAddr,
		certFile:   certFile,
		keyFile:    keyFile,
	}
}

func (s *Server) Close() error {
	return s.srv.Close()
}

func (s *Server) Listen() error {
	return s.srv.ListenAndServeTLS(s.certFile, s.keyFile)
}

// ServeHealth returns 200 when things are good
func ServeHealth(w http.ResponseWriter, r *http.Request) {
	logrus.WithField("uri", r.RequestURI).Debug("healthy")
	fmt.Fprint(w, "OK")
}

// ServeReady returns 200 when things are good
func ServeReady(w http.ResponseWriter, r *http.Request) {
	logrus.WithField("uri", r.RequestURI).Debug("ready")
	fmt.Fprint(w, "OK")
}
