package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/moolen/logistis/pkg/recorder"
	"github.com/moolen/logistis/pkg/store/fs"
	"github.com/sirupsen/logrus"
)

type Config struct {
	listenAddr  string
	certFile    string
	keyFile     string
	logLevel    string
	dbFile      string
	maxVersions int
}

func main() {
	var cfg Config
	flag.StringVar(&cfg.listenAddr, "listen", ":10250", "address/port to listen on")
	flag.StringVar(&cfg.certFile, "cert-file", "", "path to TLS certificate file")
	flag.StringVar(&cfg.keyFile, "key-file", "", "path to TLS private key file")
	flag.StringVar(&cfg.logLevel, "log-level", "debug", "")
	flag.StringVar(&cfg.dbFile, "db", "/tmp/logistis", "path to database file")
	flag.IntVar(&cfg.maxVersions, "max-versions", 100, "number of max versions to keep per entry")
	flag.Parse()

	logger := logrus.New()
	// prep logger
	llev, err := logrus.ParseLevel(cfg.logLevel)
	if err != nil {
		logrus.Fatalf("cannot set LOG_LEVEL to %q", cfg.logLevel)
	}
	logger.SetLevel(llev)
	logger.SetFormatter(&logrus.JSONFormatter{})

	fsStore, err := fs.New(cfg.dbFile, logger, cfg.maxVersions)
	if err != nil {
		logger.Fatal(err)
	}

	defer fsStore.Close()

	rec, err := recorder.New(logger, fsStore)
	if err != nil {
		logger.Fatal(err)
	}

	// handle our core application
	http.HandleFunc("/", rec.CaptureEvents)
	http.HandleFunc("/health", ServeHealth)
	http.HandleFunc("/events", rec.ListEvents)
	http.HandleFunc("/diff", rec.DiffEvents)

	logger.Printf("Listening on port %s", cfg.listenAddr)
	logger.Fatal(http.ListenAndServeTLS(cfg.listenAddr, cfg.certFile, cfg.keyFile, nil))
}

// ServeHealth returns 200 when things are good
func ServeHealth(w http.ResponseWriter, r *http.Request) {
	logrus.WithField("uri", r.RequestURI).Debug("healthy")
	fmt.Fprint(w, "OK")
}