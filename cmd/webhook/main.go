package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/namsral/flag"

	"github.com/moolen/logistis/pkg/recorder"
	"github.com/moolen/logistis/pkg/store/fs"
	"github.com/sirupsen/logrus"
)

type Config struct {
	listenAddr     string
	certFile       string
	keyFile        string
	logLevel       string
	matchUser      string
	matchGroup     string
	matchUserExtra Pair
	dbFile         string
	maxVersions    int
}

func main() {
	var cfg Config
	flag.StringVar(&cfg.listenAddr, "listen", ":10250", "address/port to listen on")
	flag.StringVar(&cfg.certFile, "cert-file", "", "path to TLS certificate file")
	flag.StringVar(&cfg.keyFile, "key-file", "", "path to TLS private key file")
	flag.StringVar(&cfg.logLevel, "log-level", "debug", "")
	flag.StringVar(&cfg.matchUser, "match-user", "", "match user name")
	flag.StringVar(&cfg.matchGroup, "match-group", "", "match user group")
	flag.Var(&cfg.matchUserExtra, "match-user-extra", "match user extra key/value pairs")
	flag.StringVar(&cfg.dbFile, "db", "/data/logistis", "path to database file")
	flag.IntVar(&cfg.maxVersions, "max-versions", 10, "number of max versions to keep per entry")
	flag.Parse()

	logger := logrus.New()
	// prep logger
	lvl, err := logrus.ParseLevel(cfg.logLevel)
	if err != nil {
		logrus.Fatalf("cannot set LOG_LEVEL to %q", cfg.logLevel)
	}
	logger.SetLevel(lvl)
	logger.SetFormatter(&logrus.JSONFormatter{})

	fsStore, err := fs.New(cfg.dbFile, logger, cfg.maxVersions)
	if err != nil {
		logger.Fatal(err)
	}

	defer fsStore.Close()

	logger.Debugf("found matching config: user=%q group=%q extra=%#v", cfg.matchUser, cfg.matchGroup, cfg.matchUserExtra.Value)
	matcher := recorder.MustNewMatchConfig(cfg.matchUser, cfg.matchGroup, cfg.matchUserExtra.Value)
	rec, err := recorder.New(logger, fsStore, matcher)
	if err != nil {
		logger.Fatal(err)
	}

	// handle our core application
	http.HandleFunc("/", rec.CaptureEvents)
	http.HandleFunc("/health", ServeHealth)
	http.HandleFunc("/events", rec.ListEvents)
	logger.Printf("Listening on port %s", cfg.listenAddr)

	server := &http.Server{Addr: cfg.listenAddr, Handler: nil}

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	go func() {
		<-sigc
		logger.Info("received shutdown signal")
		err = server.Close()
		if err != nil {
			logger.Error(err)
		}
	}()

	server.ListenAndServeTLS(cfg.certFile, cfg.keyFile)
}

// ServeHealth returns 200 when things are good
func ServeHealth(w http.ResponseWriter, r *http.Request) {
	logrus.WithField("uri", r.RequestURI).Debug("healthy")
	fmt.Fprint(w, "OK")
}

type Pair struct {
	Value map[string]string
}

func (p Pair) String() string {
	return fmt.Sprintf("%#v", p.Value)
}
func (p Pair) Set(in string) error {
	if p.Value == nil {
		p.Value = make(map[string]string)
	}
	pairs := strings.Split(strings.Trim(in, " "), "=")
	p.Value[strings.Trim(pairs[0], " ")] = strings.Trim(pairs[1], " ")
	return nil
}
