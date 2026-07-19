package main

import (
	"flag"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"server/config"
	"server/server"
)

func setUpFlags() {
	config.ForceInit(&config.ServerConfig{}) // #genai
	flag.StringVar(&config.Config.Host, "host", "0.0.0.0", "host to listen on")
	flag.IntVar(&config.Config.Port, "port", 8080, "port to listen on")
	flag.Parse()
}

func main() {
	setUpFlags()

	var sigs = make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	var wg sync.WaitGroup
	wg.Add(2)

	go server.RunAsyncTCPServer(&wg)
	go server.WaitForSignals(&wg, sigs)

	wg.Wait()
}
