package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/moolen/logistis/pkg/blame/formatter"
	"github.com/moolen/logistis/pkg/cmd/blame"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

type Config struct {
	debug           bool
	podNamespace    string
	podSelector     string
	targetNamespace string
	targetKind      string
	targetName      string
	logLevel        string
	kubeconfig      string
	format          string
	maxHistory      int
}

var cfg Config

var rootCmd = &cobra.Command{
	Use: "kubectl-blame",

	Run: func(cmd *cobra.Command, args []string) {
		ok, _ := cmd.Flags().GetBool("help")
		if ok {
			cmd.Help()
			return
		}

		logger := logrus.New()
		lvl, err := logrus.ParseLevel(cfg.logLevel)
		if err != nil {
			logger.Fatalf("cannot set LOG_LEVEL to %q", cfg.logLevel)
		}
		logger.SetLevel(lvl)

		if len(args) < 1 {
			logger.Errorf("You must specify the type of resource to get. Example: kubectl blame pod nginx")
			os.Exit(1)
			return
		}

		cfg.targetKind = strings.ToLower(args[0])
		if len(args) >= 2 {
			cfg.targetName = strings.ToLower(args[1])
		}

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
		err = formatter.FormatStream(formatter.Format(cfg.format), eventMap, cfg.debug)
		if err != nil {
			logger.Fatal(err)
		}
	},
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	home := homedir.HomeDir()
	rootCmd.Flags().StringVarP(&cfg.podNamespace, "logistis-namespace", "", "default", "logistis pod namespace")
	rootCmd.Flags().StringVarP(&cfg.targetNamespace, "namespace", "n", "default", "target namespace to pull events from")
	rootCmd.Flags().StringVarP(&cfg.podSelector, "logistis-pod-selector", "", "app=logistis", "pod selector")
	rootCmd.Flags().StringVarP(&cfg.logLevel, "loglevel", "", "info", "")
	rootCmd.Flags().StringVarP(&cfg.kubeconfig, "kubeconfig", "k", filepath.Join(home, ".kube", "config"), "path to kubeconfig file")
	rootCmd.Flags().StringVarP(&cfg.format, "format", "f", "patch", "show `patch` or `diff`")
	rootCmd.Flags().IntVarP(&cfg.maxHistory, "max-history", "h", 3, "number of events to pull from history")
	rootCmd.Flags().Bool("help", false, "show help message")
	rootCmd.Flags().BoolVarP(&cfg.debug, "debug", "", false, "enable debug mode for verbose output")
}

func initConfig() {
	viper.AutomaticEnv()
}
