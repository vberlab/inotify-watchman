package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"

	"github.com/vberlabs/inotify-watchman/internal"
	"github.com/vberlabs/inotify-watchman/internal/config"
	"gopkg.in/yaml.v3"
)

func readConfig(configPath string, config *config.Config) error {
	var data []byte
	var err error

	data, err = os.ReadFile(configPath)

	if err != nil {
		return err
	}

	err = yaml.Unmarshal(data, config)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	var nextRoutineID atomic.Uint32
	configPath := flag.String("C", "", "Path to config file")
	showConfig := flag.Bool("show-config", false, "Show loaded configuration")
	showDebug := flag.Bool("debug", false, "Show debug messages")
	flag.Parse()

	if *showDebug {
		logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))
		slog.SetDefault(logger)
		slog.Debug("Logging level set to Debug")
	}
	if *configPath == "" {
		flag.Usage()
		os.Exit(1)
	}

	var cfg config.Config
	var cfgReadErr error = readConfig(*configPath, &cfg)
	if cfgReadErr != nil {
		fmt.Fprintln(os.Stderr, cfgReadErr)
		os.Exit(2)
	}
	if *showConfig {
		out, err := yaml.Marshal(cfg)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(3)
		}
		fmt.Print(string(out))
	}

	for _, trackCfg := range cfg.Tracking {
		routineID := nextRoutineID.Add(1)
		go internal.Watcher(trackCfg, cfg.WatcherExitOnError, routineID)
	}

	signalChannel := make(chan os.Signal, 1)
	signal.Notify(
		signalChannel,
		os.Interrupt,
		syscall.SIGTERM,
	)

	receivedSignal := <-signalChannel
	slog.Info("Shutdown signal received", "signal", receivedSignal)

}
