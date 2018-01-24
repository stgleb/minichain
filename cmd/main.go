package main

import (
	"context"
	"flag"
	"github.com/BurntSushi/toml"
	. "minichain"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"
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
	blockChainServer := NewBlockChainServer(config)

	http.HandleFunc("/", blockChainServer.TransactionHandler)

	server := &http.Server{
		ReadTimeout:  time.Duration(config.Http.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(config.Http.WriteTimeout) * time.Second,
	}

	l, err := net.Listen("tcp", config.Http.ListenStr)
	RegisterShutDownHandler(server, blockChainServer)

	GetLogger().Infof("Listen and serve %s", config.Http.ListenStr)
	if err := server.Serve(l); err != nil {
		GetLogger().Fatal(err)
	}
}

func RegisterShutDownHandler(server *http.Server, blockChainServer *BlockChainServer) {
	stopChan := make(chan os.Signal)
	signal.Notify(stopChan, os.Interrupt)

	go func() {
		<-stopChan // wait for SIGINT
		GetLogger().Info("Shutting down server...")

		ctx, cancel := context.WithTimeout(context.Background(), server.ReadTimeout)
		defer cancel()

		server.Shutdown(ctx)
		close(blockChainServer.MemPool.ShutDown)

		GetLogger().Info("Server has been stopped")
	}()
}
