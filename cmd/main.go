package main

import (
	. "minichain"
	"flag"
	"github.com/BurntSushi/toml"
)

var (
	configFile string
)

func init() {
	flag.StringVar(&configFile, "config", "config.toml", "config file name")

	flag.Parse()
}

func readConfig() (*Config, error) {
	config := &Config{}

	_, err := toml.DecodeFile(configFile, config)

	if err != nil {
		return nil, err
	}

	return config, nil
}

func main() {
	config, err := readConfig()

	if err != nil {
		panic(err)
	}

	println(config)
}
