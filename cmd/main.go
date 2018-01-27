package main

import (
	"context"
	"flag"
	"github.com/BurntSushi/toml"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	. "github.com/stgleb/minichain"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"
)

var (
	configFile string
)

// TODO(stgleb): Add ability to override config from Env variables and command line args
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
	blockChainServer, err := NewBlockChainServer(config)

	if err != nil {
		GetLogger().Fatal(err)
	}

	http.HandleFunc("/tx", blockChainServer.TransactionHandler)
	http.HandleFunc("/search", blockChainServer.SearchByKey)
	http.Handle("/metrics", promhttp.Handler())

	server := &http.Server{
		ReadTimeout:  time.Duration(config.Http.Timeout) * time.Second,
		WriteTimeout: time.Duration(config.Http.Timeout) * time.Second,
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

		shutDownCtx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		doneChan := make(chan struct{})
		blockChainServer.BlockChain.ShutDown <- doneChan

		select {
		case <-doneChan:
			GetLogger().Info("Blockchain has been stopped")
		case <-shutDownCtx.Done():
			GetLogger().Warnf("Flushing context finished with %v", shutDownCtx.Err())
		}

		GetLogger().Info("Server has been stopped")
	}()
}
