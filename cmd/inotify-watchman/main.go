package main

import (
	"flag"
	"fmt"
	"os"

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
	configPath := flag.String("C", "", "Path to config file")
	showConfig := flag.Bool("show-config", false, "Show loaded configuration")
	flag.Parse()

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

}
