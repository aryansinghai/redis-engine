package main

import (
	"flag"
	"log"

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
	log.Printf("Starting server on %s:%d", config.Config.Host, config.Config.Port)
	if err := server.RunAsyncTCPServer(); err != nil { // #genai
		log.Fatalf("Server failed: %v", err)
	}
}
