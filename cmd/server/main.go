// cmd/server/main.go
package main

import (
	"flag"
	"fmt"
	"log"

	"frp-proxy/internal/config"
)

func main() {
	cfgPath := flag.String("config", "configs/app.toml", "config file path")
	flag.Parse()

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	fmt.Printf("Server starting on %s:%d\n", cfg.Server.Host, cfg.Server.Port)
}
