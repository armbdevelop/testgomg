package main

import (
	"log"

	"github.com/armbdevelop/testgomg/config"
	"github.com/armbdevelop/testgomg/internal/initx"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	app, err := initx.NewApp(cfg)
	if err != nil {
		log.Fatalf("init: %v", err)
	}

	if err = app.Run(); err != nil {
		log.Fatalf("run: %v", err)
	}
}
