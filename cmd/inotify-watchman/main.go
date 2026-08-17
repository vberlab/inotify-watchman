package main

import (
	"flag"
	"os"
)

func main() {
	configPath := flag.String("C", "", "Path to config file")

	flag.Parse()

	if *configPath == "" {
		flag.Usage()
		os.Exit(1)
	}
}
