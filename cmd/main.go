package main

import (
	"flag"
	"github.com/BurntSushi/toml"
	. "minichain"
	"net/http"
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

	InitLogger(config)
	GetLogger().Info(config)
	server := BlockChainServer{}

	http.HandleFunc("/", server.TransactionHandler)
	GetLogger().Infof("Listen and serve %s", config.Main.ListenStr)

	if err := http.ListenAndServe(config.Main.ListenStr, nil); err != nil {
		GetLogger().Fatal(err)
	}
}
