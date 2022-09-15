package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/moolen/logistis/pkg/cmd/blame"
	"github.com/moolen/logistis/pkg/formatter"
	"github.com/sirupsen/logrus"
	authenticationv1 "k8s.io/api/authentication/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

type Config struct {
	podNamespace    string
	podSelector     string
	targetNamespace string
	targetKind      string
	targetName      string
	logLevel        string
	kubeconfig      string
	format          string
	maxHistory      int
	diffPadding     int
}

const (
	formatDiff  = "diff"
	formatPatch = "patch"
)

func main() {
	var cfg Config
	home := homedir.HomeDir()
	flag.StringVar(&cfg.podNamespace, "namespace", "default", "logistis pod namespace")
	flag.StringVar(&cfg.targetNamespace, "target-namespace", "", "target namespace to pull events from")
	flag.StringVar(&cfg.targetKind, "target-kind", "", "target kind to pull events from. target-namespace must be defined.")
	flag.StringVar(&cfg.targetName, "target-name", "", "target name to pull events from. target-namespace amd target-kind must be defined.")
	flag.StringVar(&cfg.podSelector, "pod-selector", "app=logistis", "pod selector")
	flag.StringVar(&cfg.logLevel, "loglevel", "info", "")
	flag.StringVar(&cfg.kubeconfig, "kubeconfig", filepath.Join(home, ".kube", "config"), "path to kubeconfig file")
	flag.StringVar(&cfg.format, "format", "patch", "show `patch` or `diff`")
	flag.IntVar(&cfg.maxHistory, "max-history", 3, "number of events to pull from history")
	flag.IntVar(&cfg.diffPadding, "diff-padding", 5, "number of rows to show for context")
	flag.Parse()

	logger := logrus.New()
	lvl, err := logrus.ParseLevel(cfg.logLevel)
	if err != nil {
		logger.Fatalf("cannot set LOG_LEVEL to %q", cfg.logLevel)
	}
	logger.SetLevel(lvl)

	config, err := clientcmd.BuildConfigFromFlags("", cfg.kubeconfig)
	if err != nil {
		logger.Fatal(err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		logger.Fatal(err)
	}

	eventMap, err := blame.Fetch(
		logger,
		clientset,
		config,
		cfg.podSelector,
		cfg.podNamespace,
		cfg.targetNamespace,
		cfg.targetKind,
		cfg.targetName,
		cfg.maxHistory)
	if err != nil {
		logger.Fatal(err)
	}

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"key", "operation", "user", "time", cfg.format})

	for key, events := range eventMap {
		for _, ev := range events {
			formattedText := ""
			switch cfg.format {
			case formatDiff:
				df, err := formatter.DiffEvent(ev)
				if err != nil {
					logger.Fatalf("unable to diff events: %s", err.Error())
				}
				formattedText = formatter.FormatDiff(df, cfg.diffPadding)
			case formatPatch:
				formattedText, err = formatter.FormatPatch(ev)
				if err != nil {
					logger.Fatalf("unable to format patch: %s", err.Error())
				}
			}
			t.AppendRow([]interface{}{
				key,
				ev.Operation,
				formatUser(ev.UserInfo),
				ev.Timestamp.Format(time.RFC3339),
				formattedText})
			t.AppendSeparator()
		}
	}
	t.Render()
}

func formatUser(userInfo authenticationv1.UserInfo) string {
	out := userInfo.Username + "\n"
	if len(userInfo.Groups) > 0 {
		out += "groups:\n"
		for _, g := range userInfo.Groups {
			out += "- " + g + "\n"
		}
	}
	if len(userInfo.Extra) > 0 {
		out += "extra:\n"
		for k, vals := range userInfo.Extra {
			out += fmt.Sprintf("%s=%s\n", k, strings.Join(vals, ","))
		}

	}
	return out
}
